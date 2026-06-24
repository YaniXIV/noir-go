package prover

import "context"

// GnarkProver implements the Prover interface using the gnark proving system.
type GnarkProver struct {
	// TODO: add gnark-specific fields (e.g. curve, scheme, compiled constraint system)
}

// NewGnarkProver creates a new GnarkProver instance.
func NewGnarkProver() *GnarkProver {
	return &GnarkProver{}
}

func (g *GnarkProver) Setup(ctx context.Context, acir ACIRBlob) (ProvingKey, VerificationKey, error) {
	panic("not implemented")
}

func (g *GnarkProver) Prove(ctx context.Context, acir ACIRBlob, pk ProvingKey, witness WitnessMap) (ProofWithPublicInputs, error) {
	panic("not implemented")
}

func (g *GnarkProver) Verify(ctx context.Context, vk VerificationKey, proof ProofWithPublicInputs) (bool, error) {
	panic("not implemented")
}

func (g *GnarkProver) SerializeVK(vk VerificationKey) ([]byte, error) {
	panic("not implemented")
}

func (g *GnarkProver) DeserializeVK(data []byte) (VerificationKey, error) {
	panic("not implemented")
}

func (g *GnarkProver) Backend() string {
	return "gnark"
}
