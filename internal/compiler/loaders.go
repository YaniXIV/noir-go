package compiler

import (
	"context"
	"fmt"
)

func loadCompiler(w *WasmManager) (*WasmObject, error) {

	ctx := context.Background()
	compiled, err := w.runtime.CompileModule(ctx, noirCompilerWasm)
	if err != nil {
		return nil, err
	}
	return &WasmObject{Compiler, compiled}, nil
}

func loadProver(w *WasmManager) (*WasmObject, error) {
	return &WasmObject{Prover, nil}, nil
}

func loadVerifier(w *WasmManager) (*WasmObject, error) {
	return &WasmObject{Verifier, nil}, nil
}

func loadError(w *WasmManager) (*WasmObject, error) {
	return nil, fmt.Errorf("Unknown wasm type")
}
