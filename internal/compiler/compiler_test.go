package compiler

import (
	"fmt"
	"testing"
)

func TestCompiler(t *testing.T) {

	runWasmCompiler(noirCompilerWasm)
	RawCompilerTest()
}

func RawCompilerTest() {
	fmt.Println("starting program")
	w, err := NewWasmManager()
	if err != nil || w.runtime == nil {
		fmt.Println("error with wasm manager")
		panic(err)
	}
	fmt.Println("wasmManger instantiated", w.instances[Compiler])
	errCompile := w.CompileProgram(".")
	fmt.Println("Compiler instantiated", w.instances[Compiler])
	if errCompile != nil {

		fmt.Println("Compiler failed to get", w.instances[Compiler])
		panic(errCompile)
	}

}
