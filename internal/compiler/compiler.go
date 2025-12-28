package compiler

import (
	"fmt"
)

func Compile(Input any) {
	fmt.Println(Input)
	runWasmCompiler(noirWasm)

}
