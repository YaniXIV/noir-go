// Command proveverify is a minimal, runnable demonstration of noir-go's
// full pipeline: compile a Noir circuit, solve a witness for it, run the
// gnark Groth16 trusted setup, generate a proof, and verify it.
//
// Run it from the module root:
//
//	go run ./examples/proveverify
package main

import (
	"fmt"
	"log"

	noirgo "github.com/YaniXIV/noir-go"
)

func main() {
	// 1. Compile: testdata/circuit_simple is `fn main(x: Field, y: Field) -> pub Field { x * y }`.
	// No local Noir/Nargo install required -- the compiler runs as an embedded WASM module.
	comp, err := noirgo.Compile("testdata/circuit_simple")
	if err != nil {
		log.Fatalf("compile failed: %v", err)
	}
	fmt.Printf("1. compiled: %d bytes of ACIR, noir version %s\n", len(comp.ACIR.Bytes), comp.NoirVersion)

	// 2. Execute: solve the full witness for x=3, y=4.
	inputs := make(map[uint32][32]byte)
	values := []byte{3, 4} // x, y in ABI parameter order
	for i, idx := range comp.PrivateParamWitnesses {
		var v [32]byte
		v[31] = values[i]
		inputs[idx] = v
	}
	solved, err := noirgo.Execute(comp, inputs)
	if err != nil {
		log.Fatalf("execute failed: %v", err)
	}
	fmt.Printf("2. executed: solved %d witness values (expect x*y = 12)\n", len(solved))

	// 3. Setup: gnark Groth16 trusted setup for this circuit (BN254).
	pk, vk, err := noirgo.Setup(comp)
	if err != nil {
		log.Fatalf("setup failed: %v", err)
	}
	fmt.Printf("3. setup: proving key %d bytes, verification key %d bytes\n", len(pk), len(vk))

	// 4. Prove: generate a Groth16 proof for the solved witness.
	proof, err := noirgo.Prove(comp, pk, solved)
	if err != nil {
		log.Fatalf("prove failed: %v", err)
	}
	fmt.Printf("4. proved: %d byte proof, %d public input(s)\n", len(proof.Proof), len(proof.PublicInputs))

	// 5. Verify: a valid proof should pass.
	ok, err := noirgo.Verify(vk, proof)
	if err != nil {
		log.Fatalf("verify failed: %v", err)
	}
	fmt.Printf("5. verify (valid proof):     ok=%v\n", ok)

	// 6. Verify: a tampered proof should be rejected, not silently accepted.
	tampered := proof
	tampered.Proof = append([]byte(nil), proof.Proof...)
	tampered.Proof[len(tampered.Proof)-1] ^= 0xFF
	ok, err = noirgo.Verify(vk, tampered)
	fmt.Printf("6. verify (tampered proof):  ok=%v err=%v\n", ok, err)
}
