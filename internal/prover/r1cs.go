package prover

import (
	"fmt"
	"sort"

	"github.com/consensys/gnark/constraint"
	cs_bn254 "github.com/consensys/gnark/constraint/bn254"

	"github.com/YaniXIV/noir-go/internal/prover/acir"
)

// oneWireName is an auxiliary constant-1 wire injected outside of ACIR's own
// witness-index space. ACIR has no reserved "constant" witness -- constants
// live directly in opcode coefficients -- but gnark's low-level R1C builder
// needs a real variable to attach qC-style constant terms to, and (unlike
// frontend.Compile) does not reserve one automatically: gnark's solver
// hardcodes variable ID 0 as the always-solved "1" wire, so it must be the
// very first variable registered.
const oneWireName = "one"

// builtR1CS bundles the compiled constraint system with the bookkeeping
// needed to build matching witness vectors in Prove/Verify. gnark ties
// witness values to variables positionally, so the same ACIR-witness-index
// -> gnark-variable-ID map and public/secret registration order used here
// must be reused whenever a witness vector is built for this circuit.
type builtR1CS struct {
	cs          constraint.ConstraintSystem
	varID       map[uint32]int // ACIR witness index -> gnark variable ID
	publicOrder []uint32       // ACIR witness indices, in gnark public-variable registration order
	secretOrder []uint32       // ACIR witness indices, in gnark secret-variable registration order
}

// buildR1CS converts the main function of an ACIR program into a gnark R1CS.
// Only AssertZero opcodes are supported; BrilligCall opcodes are skipped
// (they are unconstrained witness-generation hints, already resolved by
// ExecuteProgram). Any other opcode kind is rejected rather than silently
// ignored, since silently dropping a constraint is a soundness bug.
func buildR1CS(prog *acir.Program) (*builtR1CS, error) {
	if len(prog.Functions) == 0 {
		return nil, fmt.Errorf("acir program has no functions")
	}
	fn := prog.Functions[0]

	r1cs := cs_bn254.NewR1CS(int(fn.CurrentWitnessIndex) + 1)

	oneWireID := r1cs.AddPublicVariable(oneWireName)

	publicSet := make(map[uint32]bool, len(fn.PublicParameters)+len(fn.ReturnValues))
	for _, idx := range fn.PublicParameters {
		publicSet[idx] = true
	}
	for _, idx := range fn.ReturnValues {
		publicSet[idx] = true
	}

	// Only register a gnark variable for witness indices that are either
	// ABI-visible (so they exist even if unconstrained) or actually
	// referenced by an AssertZero opcode. current_witness_index is just a
	// monotonically increasing counter the compiler bumps for every
	// intermediate value it ever considered, including ones later
	// optimized away -- ACVM's solved witness map is not guaranteed to
	// have a value for those, so registering (and later requiring a
	// witness value for) every index in the dense [0, current_witness_index]
	// range would make some otherwise-valid circuits unprovable. This also
	// sidesteps looping up to a uint32 bound, which would wrap forever if
	// current_witness_index were ever math.MaxUint32.
	needed := neededWitnesses(fn)

	varID := make(map[uint32]int, len(needed))
	var publicOrder, secretOrder []uint32

	// Two passes so all ACIR-public variables get contiguous IDs right after
	// the constant wire, then all ACIR-secret variables -- matching the
	// [public block][secret block] layout gnark's witness.Fill expects.
	for _, idx := range needed {
		if publicSet[idx] {
			varID[idx] = r1cs.AddPublicVariable(fmt.Sprintf("pub_%d", idx))
			publicOrder = append(publicOrder, idx)
		}
	}
	for _, idx := range needed {
		if !publicSet[idx] {
			varID[idx] = r1cs.AddSecretVariable(fmt.Sprintf("sec_%d", idx))
			secretOrder = append(secretOrder, idx)
		}
	}

	bID := r1cs.AddBlueprint(&constraint.BlueprintGenericR1C{})

	for _, op := range fn.Opcodes {
		switch op.Kind {
		case "AssertZero":
			if err := addAssertZero(r1cs, bID, oneWireID, op.AssertZero, varID); err != nil {
				return nil, err
			}
		case "BrilligCall":
			// no algebraic constraint
		default:
			return nil, fmt.Errorf("unsupported ACIR opcode %q: gnark prover only supports AssertZero constraints", op.Kind)
		}
	}

	return &builtR1CS{cs: r1cs, varID: varID, publicOrder: publicOrder, secretOrder: secretOrder}, nil
}

// neededWitnesses returns, in ascending order, every ACIR witness index
// that must exist as a gnark variable: ABI parameters/return values (which
// must exist even if never referenced by a constraint) plus every witness
// actually referenced by a MulTerm or LinearCombination in an AssertZero
// opcode. Anything else in [0, CurrentWitnessIndex] is dead -- allocated by
// the compiler but never used or solved for -- and is safe to skip.
func neededWitnesses(fn acir.Function) []uint32 {
	needed := make(map[uint32]struct{}, fn.CurrentWitnessIndex+1)
	add := func(idx uint32) { needed[idx] = struct{}{} }

	for _, idx := range fn.PrivateParameters {
		add(idx)
	}
	for _, idx := range fn.PublicParameters {
		add(idx)
	}
	for _, idx := range fn.ReturnValues {
		add(idx)
	}
	for _, op := range fn.Opcodes {
		if op.AssertZero == nil {
			continue
		}
		for _, mt := range op.AssertZero.MulTerms {
			add(mt.LHS)
			add(mt.RHS)
		}
		for _, lc := range op.AssertZero.LinearCombinations {
			add(lc.Witness)
		}
	}

	sorted := make([]uint32, 0, len(needed))
	for idx := range needed {
		sorted = append(sorted, idx)
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	return sorted
}

// addAssertZero maps qM.(xa.xb) + sum(qi.xi) + qC == 0 onto a single gnark
// R1C (L.R == O). R1C supports exactly one multiplication per constraint, so
// the two cases are handled separately:
//
//   - one mul term:  L = qM.xa,  R = 1.xb,  O = -(sum(qi.xi) + qC)
//   - no mul term:   L = 1.one,  R = sum(qi.xi) + qC,  O = 0
func addAssertZero(r1cs *cs_bn254.R1CS, bID constraint.BlueprintID, oneWireID int, az *acir.AssertZero, varID map[uint32]int) error {
	if az == nil {
		return fmt.Errorf("AssertZero opcode missing its payload")
	}
	if len(az.MulTerms) > 1 {
		return fmt.Errorf("assert_zero has %d mul_terms, gnark prover supports at most 1 (a single R1C multiplication)", len(az.MulTerms))
	}

	one := r1cs.FromInterface(1)
	qc := r1cs.FromInterface(az.QC.BigInt())

	if len(az.MulTerms) == 1 {
		mt := az.MulTerms[0]
		xa, ok := varID[mt.LHS]
		if !ok {
			return fmt.Errorf("mul_term references unknown witness %d", mt.LHS)
		}
		xb, ok := varID[mt.RHS]
		if !ok {
			return fmt.Errorf("mul_term references unknown witness %d", mt.RHS)
		}

		qm := r1cs.FromInterface(mt.Coeff.BigInt())
		L := constraint.LinearExpression{r1cs.MakeTerm(qm, xa)}
		R := constraint.LinearExpression{r1cs.MakeTerm(one, xb)}

		O, err := buildLinearExpr(r1cs, varID, az.LinearCombinations, true)
		if err != nil {
			return err
		}
		negQC := r1cs.Neg(qc)
		O = append(O, r1cs.MakeTerm(negQC, oneWireID))

		r1cs.AddR1C(constraint.R1C{L: L, R: R, O: O}, bID)
		return nil
	}

	R, err := buildLinearExpr(r1cs, varID, az.LinearCombinations, false)
	if err != nil {
		return err
	}
	R = append(R, r1cs.MakeTerm(qc, oneWireID))
	L := constraint.LinearExpression{r1cs.MakeTerm(one, oneWireID)}
	O := constraint.LinearExpression{}

	r1cs.AddR1C(constraint.R1C{L: L, R: R, O: O}, bID)
	return nil
}

func buildLinearExpr(r1cs *cs_bn254.R1CS, varID map[uint32]int, terms []acir.LinearCombination, negate bool) (constraint.LinearExpression, error) {
	le := make(constraint.LinearExpression, 0, len(terms))
	for _, t := range terms {
		vid, ok := varID[t.Witness]
		if !ok {
			return nil, fmt.Errorf("linear_combination references unknown witness %d", t.Witness)
		}
		coeff := r1cs.FromInterface(t.Coeff.BigInt())
		if negate {
			coeff = r1cs.Neg(coeff)
		}
		le = append(le, r1cs.MakeTerm(coeff, vid))
	}
	return le, nil
}
