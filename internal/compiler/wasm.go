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
type WasmLoader func(*WasmManager) (*WasmObject, error)

//go:embed noir-compile.wasm
var noirCompilerWasm []byte

const (
	Compiler WasmType = iota
	Prover
	Verifier
)

type WasmObject struct {
	Type     WasmType
	Compiled wazero.CompiledModule
}
type wasmInstance struct {
	once   sync.Once
	object *WasmObject
	err    error
}

type WasmManager struct {
	mu        sync.Mutex
	runtime   wazero.Runtime
	instances map[WasmType]*wasmInstance
	loaders   map[WasmType]WasmLoader
}

func NewWasmManager() (*WasmManager, error) {
	ctx := context.Background()
	r := wazero.NewRuntime(ctx)
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		return nil, err
	}

	return &WasmManager{

		runtime:   r,
		instances: make(map[WasmType]*wasmInstance),
		loaders: map[WasmType]WasmLoader{
			Compiler: loadCompiler,
			Prover:   loadProver,
			Verifier: loadVerifier,
		},
	}, nil
}

func InitializeWasmInstance() *wasmInstance {
	return &wasmInstance{}
}

func (w *WasmManager) Warmup() error {
	_, errCompiler := w.Get(Compiler)
	if errCompiler != nil {
		panic(errCompiler)
	}
	_, errProver := w.Get(Prover)
	if errProver != nil {
		panic(errProver)
	}
	_, errVerifier := w.Get(Verifier)
	if errVerifier != nil {
		panic(errVerifier)
	}
	return nil

}

func (w *WasmManager) Get(t WasmType) (*WasmObject, error) {
	w.mu.Lock()
	inst, ok := w.instances[t]
	if !ok {
		//fmt.Printf("%v Does not yet exist in map", t)
		inst = &wasmInstance{}
		w.instances[t] = inst
		//fmt.Println("key: value")
		//fmt.Println(t, w.instances[t])
	}

	loader, ok := w.loaders[t]
	if !ok {
		//fmt.Println("Loader does not exist yet!")
		w.mu.Unlock()
		return nil, fmt.Errorf("no loader found for wasm type %v", t)
	}

	w.mu.Unlock()

	inst.once.Do(func() {
		inst.object, inst.err = loader(w)
		fmt.Println("Loader: ", loader)
	})
	return inst.object, nil
}

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
