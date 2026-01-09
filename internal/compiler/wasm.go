package compiler

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"log"

	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// probably not optimal abstraction yet, work on this later.
type WasmType int

const (
	Compiler WasmType = iota
	Prover
	Verifier
)

type WasmManager struct {
	Once    sync.Once
	Object  WasmObject
	wasmErr error
}

type WasmObject struct {
	Type      WasmType
	wasmBytes []byte
}

func NewWasmManager() *WasmManager {
	return &WasmManager{}
}

func (m *WasmManager) get(*WasmObject) {
	m.Once.Do(func() {
		// Add wasm initial ization function  here.
	})

}

//go:embed noir-compile.wasm
var noirWasm []byte

func runWasmCompiler(wasmBytes []byte) {
	ctx := context.Background()

	// 1. Create a new runtime
	runtime := wazero.NewRuntime(ctx)
	defer runtime.Close(ctx)

	// 2. Instantiate WASI so the WASM module can use stdout, stderr, etc.
	_, err := wasi_snapshot_preview1.Instantiate(ctx, runtime)
	if err != nil {
		log.Fatalf("failed to instantiate WASI: %v", err)
	}

	// 3. Compile the WASM module
	compiled, err := runtime.CompileModule(ctx, wasmBytes)
	if err != nil {
		log.Fatalf("failed to compile WASM module: %v", err)
	}

	// 4. Prepare an output buffer to capture stdout/stderr
	outputBuf := new(bytes.Buffer)

	// 5. Configure the module
	config := wazero.NewModuleConfig().
		WithStdout(outputBuf).
		WithStderr(outputBuf)

	// 6. Instantiate the module
	mod, err := runtime.InstantiateModule(ctx, compiled, config)
	if err != nil {
		log.Fatalf("failed to instantiate module: %v", err)
	}
	defer mod.Close(ctx)

	// 7. Optionally call an exported function
	fn := mod.ExportedFunction("test_compile_wasm_go")
	if fn != nil {
		_, err := fn.Call(ctx)
		if err != nil {
			fmt.Printf("function call error: %v\n", err)
		}
	}

	// 8. Print output
	fmt.Println("--- start of wasm logs ---")
	fmt.Print(outputBuf.String())
	fmt.Println("--- end of wasm logs ---")
}
