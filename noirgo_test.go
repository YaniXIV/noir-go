package noirgo_test

import (
	"fmt"
	"testing"

	noirgo "github.com/YaniXIV/noir-go"
)

func TestNoirGo(t *testing.T) {
	c, err := noirgo.Compile("internal/compiler/noirtest")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	fmt.Println("Compilation succesful!")
	fmt.Println("here is the Acir!", c.ACIR)
	fmt.Println("here is the ABI Params!", c.ABI.Parameters)
}

func TestExecute(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping wasm execute in short mode")
	}

	comp, err := noirgo.Compile("testdata/circuit_simple")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	inputs := make(map[uint32][32]byte)
	for _, idx := range comp.PrivateParamWitnesses {
		var val [32]byte
		val[31] = 3
		inputs[idx] = val
	}
	for _, idx := range comp.PublicParamWitnesses {
		var val [32]byte
		val[31] = 3
		inputs[idx] = val
	}

	solved, err := noirgo.Execute(comp, inputs)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if len(solved) == 0 {
		t.Fatal("solved witness is empty")
	}
}
