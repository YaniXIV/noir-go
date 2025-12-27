package compiler

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

/*
//go:embed noir_compiler.wasm
var noirWasm []byte
*/

func loadWasm() {
	ctx := context.background()
	r := wazero.newruntime(ctx)
	defer r.close(ctx)

	// 1. instantiate wasi so println! has a host function to call
	wasi_snapshot_preview1.mustinstantiate(ctx, r)

	// 2. load the wasm
	wasmbytes, err := os.readfile("noir-compile.wasm")
	if err != nil {
		log.Fatalf("failed to read wasm: %v", err)
	}

	// 3. compile the module first
	compiled, err := r.compilemodule(ctx, wasmbytes)
	if err != nil {
		log.Fatalf("failed to compile: %v", err)
	}

	// 4. create a buffer to collect all output
	// this will grow automatically to fit all println! calls.
	outputbuf := new(bytes.buffer)

	// 5. build the configuration
	// we point stdout and stderr to the same buffer to catch logs and crashes
	config := wazero.newmoduleconfig().
		withstdout(outputbuf).
		withstderr(outputbuf).
		withsysnanotime().
		withname("rotate").
		withargs("rotate", "angle=90", "dir=cw")

	// 6. instantiate the module
	mod, err := r.instantiatemodule(ctx, compiled, config)
	if err != nil {
		log.Fatalf("failed to instantiate: %v", err)
	}
	defer mod.close(ctx)

	// 7. call your specific function
	fn := mod.exportedfunction("test_compile_wasm_go")
	if fn == nil {
		log.Fatalf("function not found")
	}

	_, err = fn.call(ctx)
	if err != nil {
		// even if it fails, the buffer might have captured the "why"
		fmt.Printf("execution error: %v\n", err)
	}

	// 8. print everything at the end
	fmt.Println("--- start of wasm logs ---")
	if outputbuf.len() == 0 {
		fmt.Println("(no output captured)")
	} else {
		fmt.Print(outputbuf.string())
	}
	fmt.Println("--- end of wasm logs ---")
}

func runWasmCompiler(noirWasm []byte, Input any) {

}
