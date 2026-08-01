package prover_test

import (
	"testing"

	noirgo "github.com/YaniXIV/noir-go"
)

// TestGroth16RoundTrip compiles a small Noir circuit (x*y == pub return),
// solves a witness for it, and drives the full gnark Groth16 pipeline:
// Setup -> Prove -> Verify. It also checks that a corrupted proof and a
// corrupted public input are both correctly rejected.
func TestGroth16RoundTrip(t *testing.T) {
	comp, err := noirgo.Compile("../../testdata/circuit_simple")
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	inputs := make(map[uint32][32]byte)
	for _, idx := range comp.PrivateParamWitnesses {
		var v [32]byte
		v[31] = 3
		inputs[idx] = v
	}
	for _, idx := range comp.PublicParamWitnesses {
		var v [32]byte
		v[31] = 3
		inputs[idx] = v
	}

	solved, err := noirgo.Execute(comp, inputs)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	pk, vk, err := noirgo.Setup(comp)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	proof, err := noirgo.Prove(comp, pk, solved)
	if err != nil {
		t.Fatalf("prove: %v", err)
	}

	ok, err := noirgo.Verify(vk, proof)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !ok {
		t.Fatal("expected valid proof to verify")
	}

	t.Run("tampered proof bytes are rejected", func(t *testing.T) {
		tampered := proof
		tampered.Proof = append([]byte(nil), proof.Proof...)
		tampered.Proof[len(tampered.Proof)-1] ^= 0xFF

		ok, err := noirgo.Verify(vk, tampered)
		if ok {
			t.Fatal("expected tampered proof bytes to fail verification")
		}
		_ = err // a corrupted proof may surface as an error or a false result
	})

	t.Run("tampered public input is rejected", func(t *testing.T) {
		if len(proof.PublicInputs) == 0 {
			t.Skip("circuit has no public inputs to tamper with")
		}
		tampered := proof
		tampered.PublicInputs = make([][]byte, len(proof.PublicInputs))
		for i, v := range proof.PublicInputs {
			tampered.PublicInputs[i] = append([]byte(nil), v...)
		}
		tampered.PublicInputs[0][len(tampered.PublicInputs[0])-1] ^= 0xFF

		ok, err := noirgo.Verify(vk, tampered)
		if ok {
			t.Fatal("expected tampered public input to fail verification")
		}
		_ = err
	})
}

// TestGroth16RoundTripWide exercises a circuit large enough to contain
// BrilligCall opcodes (unconstrained witness-generation hints), confirming
// the R1CS builder correctly skips them rather than choking on them.
func TestGroth16RoundTripWide(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping wide-circuit gnark round trip in short mode")
	}

	comp, err := noirgo.Compile("../../testdata/circuit_wide_4k")
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	inputs := make(map[uint32][32]byte)
	for _, idx := range comp.PrivateParamWitnesses {
		var v [32]byte
		v[31] = 1
		inputs[idx] = v
	}
	for _, idx := range comp.PublicParamWitnesses {
		var v [32]byte
		v[31] = 1
		inputs[idx] = v
	}

	solved, err := noirgo.Execute(comp, inputs)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	pk, vk, err := noirgo.Setup(comp)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	proof, err := noirgo.Prove(comp, pk, solved)
	if err != nil {
		t.Fatalf("prove: %v", err)
	}

	ok, err := noirgo.Verify(vk, proof)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !ok {
		t.Fatal("expected valid proof to verify")
	}
}
