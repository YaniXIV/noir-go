package noirgo_test

import (
	"testing"

	"fmt"
	noirgo "github.com/YaniXIV/noir-go"
)

func TestNoirGo(t *testing.T) {
	c, err := noirgo.Compile("internal/compiler/noirtest")
	if err != nil {
		panic(err)
	}

	fmt.Println("Compilation succesful!")
	fmt.Println("here is the Acir!", c.ACIR)
	fmt.Println("here is the ABI Params!", c.ABI.Parameters)
}
