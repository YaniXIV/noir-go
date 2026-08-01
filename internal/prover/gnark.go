package prover

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/witness"

	"github.com/YaniXIV/noir-go/internal/prover/acir"
)

// GnarkProver implements the Prover interface using gnark's Groth16 proving
// system over BN254 -- the curve Noir's ACIR field elements are defined
// over. Only AssertZero opcodes are supported (see buildR1CS); circuits
// using black-box functions (range checks, hashes, ECDSA, ...) are rejected
// at Setup/Prove time rather than silently mis-proved.
type GnarkProver struct{}

var _ Prover = (*GnarkProver)(nil)

// NewGnarkProver creates a new GnarkProver instance.
func NewGnarkProver() *GnarkProver {
	return &GnarkProver{}
}

func (g *GnarkProver) Setup(ctx context.Context, acirBlob ACIRBlob) (ProvingKey, VerificationKey, error) {
	built, err := decodeAndBuild(acirBlob)
	if err != nil {
		return nil, nil, err
	}

	pk, vk, err := groth16.Setup(built.cs)
	if err != nil {
		return nil, nil, fmt.Errorf("groth16 setup: %w", err)
	}

	pkBytes, err := writeToBytes(pk)
	if err != nil {
		return nil, nil, fmt.Errorf("serialize proving key: %w", err)
	}
	vkBytes, err := writeToBytes(vk)
	if err != nil {
		return nil, nil, fmt.Errorf("serialize verification key: %w", err)
	}
	return ProvingKey(pkBytes), VerificationKey(vkBytes), nil
}

func (g *GnarkProver) Prove(ctx context.Context, acirBlob ACIRBlob, pkBytes ProvingKey, wm WitnessMap) (ProofWithPublicInputs, error) {
	built, err := decodeAndBuild(acirBlob)
	if err != nil {
		return ProofWithPublicInputs{}, err
	}

	pk := groth16.NewProvingKey(ecc.BN254)
	if _, err := pk.ReadFrom(bytes.NewReader(pkBytes)); err != nil {
		return ProofWithPublicInputs{}, fmt.Errorf("deserialize proving key: %w", err)
	}

	full, err := buildWitness(built, wm)
	if err != nil {
		return ProofWithPublicInputs{}, err
	}

	proof, err := groth16.Prove(built.cs, pk, full)
	if err != nil {
		return ProofWithPublicInputs{}, fmt.Errorf("groth16 prove: %w", err)
	}

	proofBytes, err := writeToBytes(proof)
	if err != nil {
		return ProofWithPublicInputs{}, fmt.Errorf("serialize proof: %w", err)
	}

	publicInputs := make([][]byte, len(built.publicOrder))
	for i, idx := range built.publicOrder {
		v, ok := wm[idx]
		if !ok {
			return ProofWithPublicInputs{}, fmt.Errorf("witness missing value for public witness index %d", idx)
		}
		publicInputs[i] = append([]byte(nil), v...)
	}

	return ProofWithPublicInputs{Proof: Proof(proofBytes), PublicInputs: publicInputs}, nil
}

func (g *GnarkProver) Verify(ctx context.Context, vkBytes VerificationKey, proof ProofWithPublicInputs) (bool, error) {
	vk := groth16.NewVerifyingKey(ecc.BN254)
	if _, err := vk.ReadFrom(bytes.NewReader(vkBytes)); err != nil {
		return false, fmt.Errorf("deserialize verification key: %w", err)
	}

	gnarkProof := groth16.NewProof(ecc.BN254)
	if _, err := gnarkProof.ReadFrom(bytes.NewReader(proof.Proof)); err != nil {
		return false, fmt.Errorf("deserialize proof: %w", err)
	}

	pubWitness, err := witness.New(ecc.BN254.ScalarField())
	if err != nil {
		return false, err
	}
	ch := make(chan any, len(proof.PublicInputs))
	for _, v := range proof.PublicInputs {
		ch <- new(big.Int).SetBytes(v)
	}
	close(ch)
	if err := pubWitness.Fill(len(proof.PublicInputs), 0, ch); err != nil {
		return false, fmt.Errorf("build public witness: %w", err)
	}

	if err := groth16.Verify(gnarkProof, vk, pubWitness); err != nil {
		// Propagate the real reason: a rejected proof and a structural
		// mismatch (e.g. a public-input count from a different circuit)
		// both fail here, and callers that need to tell them apart need
		// the underlying error. Callers that only care about pass/fail
		// can still just check the bool.
		return false, err
	}
	return true, nil
}

// SerializeVK/DeserializeVK are pass-throughs: VerificationKey is already a
// portable byte encoding (gnark's native WriteTo/ReadFrom format), produced
// once in Setup.
func (g *GnarkProver) SerializeVK(vk VerificationKey) ([]byte, error) {
	return vk, nil
}

func (g *GnarkProver) DeserializeVK(data []byte) (VerificationKey, error) {
	return VerificationKey(data), nil
}

func (g *GnarkProver) Backend() string {
	return "gnark-groth16-bn254"
}

func decodeAndBuild(acirBlob ACIRBlob) (*builtR1CS, error) {
	prog, err := acir.DecodeAcir(acirBlob)
	if err != nil {
		return nil, err
	}
	return buildR1CS(prog)
}

func writeToBytes(w io.WriterTo) ([]byte, error) {
	var buf bytes.Buffer
	if _, err := w.WriteTo(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// buildWitness assembles a full gnark witness (public values, then secret
// values, in the exact registration order buildR1CS used) from a caller
// supplied ACIR-witness-index -> value map. gnark ties witness values to
// variables positionally, so this order must match buildR1CS exactly.
func buildWitness(built *builtR1CS, wm WitnessMap) (witness.Witness, error) {
	full, err := witness.New(ecc.BN254.ScalarField())
	if err != nil {
		return nil, err
	}

	n := len(built.publicOrder) + len(built.secretOrder)
	ch := make(chan any, n)
	for _, idx := range built.publicOrder {
		v, ok := wm[idx]
		if !ok {
			return nil, fmt.Errorf("witness map missing value for public witness index %d", idx)
		}
		ch <- new(big.Int).SetBytes(v)
	}
	for _, idx := range built.secretOrder {
		v, ok := wm[idx]
		if !ok {
			return nil, fmt.Errorf("witness map missing value for secret witness index %d", idx)
		}
		ch <- new(big.Int).SetBytes(v)
	}
	close(ch)

	if err := full.Fill(len(built.publicOrder), len(built.secretOrder), ch); err != nil {
		return nil, fmt.Errorf("fill witness: %w", err)
	}
	return full, nil
}
