package noirgo

import (
	"context"

	"github.com/YaniXIV/noir-go/internal/prover"
	"github.com/YaniXIV/noir-go/result"
)

// ProvingKey, VerificationKey and Proof are gnark-native byte encodings
// produced by Setup/Prove; treat them as opaque and persist them as-is.
type ProvingKey = prover.ProvingKey
type VerificationKey = prover.VerificationKey
type Proof = prover.ProofWithPublicInputs

// Setup runs the gnark Groth16 trusted setup for a compiled circuit using a background context.
func Setup(comp *result.Compilation) (ProvingKey, VerificationKey, error) {
	return SetupWithContext(context.Background(), comp)
}

// SetupWithContext runs the gnark Groth16 trusted setup for a compiled circuit with the provided context.
func SetupWithContext(ctx context.Context, comp *result.Compilation) (ProvingKey, VerificationKey, error) {
	return prover.NewGnarkProver().Setup(ctx, prover.ACIRBlob(comp.ACIR.JSON))
}

// Prove generates a Groth16 proof for a solved witness using a background context.
// witness should contain a value for every ACIR witness index used by the
// circuit (as returned by Execute), not just the ABI parameters.
func Prove(comp *result.Compilation, pk ProvingKey, witness map[uint32][32]byte) (Proof, error) {
	return ProveWithContext(context.Background(), comp, pk, witness)
}

// ProveWithContext generates a Groth16 proof for a solved witness with the provided context.
func ProveWithContext(ctx context.Context, comp *result.Compilation, pk ProvingKey, witness map[uint32][32]byte) (Proof, error) {
	wm := make(prover.WitnessMap, len(witness))
	for idx, v := range witness {
		wm[idx] = v[:]
	}
	return prover.NewGnarkProver().Prove(ctx, prover.ACIRBlob(comp.ACIR.JSON), pk, wm)
}

// Verify checks a Groth16 proof against a verification key using a background context.
func Verify(vk VerificationKey, proof Proof) (bool, error) {
	return VerifyWithContext(context.Background(), vk, proof)
}

// VerifyWithContext checks a Groth16 proof against a verification key with the provided context.
func VerifyWithContext(ctx context.Context, vk VerificationKey, proof Proof) (bool, error) {
	return prover.NewGnarkProver().Verify(ctx, vk, proof)
}
