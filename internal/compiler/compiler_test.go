package compiler

import (
	"fmt"
	"log"
	"noir-go/internal/wasm"
	"testing"
)

func TestCompiler(t *testing.T) {

	//runWasmCompiler(noirCompilerWasm)
	RawCompilerTest()
}

func RawCompilerTest() {
	fmt.Println("starting program")
	w, err := wasm.NewWasmManager()

	if err != nil || w.GetRuntime() == nil {
		fmt.Println("error with wasm manager")
		panic(err)
	}
	//fmt.Println("wasmManger instantiated", w.instances[Compiler])
	AcirBlob, errCompile := CompileProgram(w, "noirtest")
	fmt.Println("Compiler instantiated", w.GetInstance(wasm.Compiler))
	if errCompile != nil {

		fmt.Println("Compiler failed to get", w.GetInstance(wasm.Compiler))
		panic(errCompile)
	}
	if AcirBlob == nil {
		//panic("AcirBlob is nil.")
		//return
		log.Println("Acir blob is nil")
	}
	log.Println("Test completes")

}
