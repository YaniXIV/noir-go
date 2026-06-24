package prover

import "context"

// ACIRBlob is a raw serialized ACIR circuit (bytecode from Noir or similar).
type ACIRBlob []byte

// WitnessMap maps witness indices to their field element values (as big-endian bytes).
type WitnessMap map[uint32][]byte

// Proof is the raw proof bytes returned by the backend.
type Proof []byte

// VerificationKey is the serialized verification key for a compiled circuit.
type VerificationKey []byte

// ProvingKey is the serialized proving key (if the backend requires a separate setup phase).
type ProvingKey []byte

// ProofWithPublicInputs bundles the proof alongside any extracted public inputs.
type ProofWithPublicInputs struct {
	Proof        Proof
	PublicInputs [][]byte // each entry is a field element as big-endian bytes
}

// Prover is the interface a ZK backend must implement to prove and verify ACIR circuits.
type Prover interface {
	// Setup compiles the ACIR blob and runs the trusted setup (or loads SRS),
	// returning a ProvingKey and VerificationKey. Some backends (e.g. Groth16)
	// require a per-circuit setup; others (e.g. UltraPlonk) do not.
	Setup(ctx context.Context, acir ACIRBlob) (ProvingKey, VerificationKey, error)

	// Prove generates a proof for the given circuit and witness.
	// pk may be nil for backends that don't use a separate proving key.
	Prove(ctx context.Context, acir ACIRBlob, pk ProvingKey, witness WitnessMap) (ProofWithPublicInputs, error)

	// Verify checks a proof against the verification key and public inputs.
	Verify(ctx context.Context, vk VerificationKey, proof ProofWithPublicInputs) (bool, error)

	// SerializeVK encodes a VerificationKey into a portable byte format
	// (useful for on-chain or cross-backend serialization).
	SerializeVK(vk VerificationKey) ([]byte, error)

	// DeserializeVK decodes a VerificationKey from a portable byte format.
	DeserializeVK(data []byte) (VerificationKey, error)

	// Backend returns a string identifier for the proving system (e.g. "gnark-groth16", "gnark-plonk").
	Backend() string
}
