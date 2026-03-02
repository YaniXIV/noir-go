package compiler

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"noir-go/internal/wasm"
)

func TestCompileProgram(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping wasm compile in short mode")
	}

	totalStart := time.Now()

	// --- NewWasmManager ---
	start := time.Now()
	w, err := wasm.NewWasmManager()
	if err != nil {
		t.Fatalf("NewWasmManager failed: %v", err)
	}
	t.Logf("NewWasmManager took %s", time.Since(start))

	// --- Runtime check ---
	start = time.Now()
	if w.GetRuntime() == nil {
		t.Fatalf("expected non-nil runtime")
	}
	t.Logf("Runtime check took %s", time.Since(start))

	// --- CompileProgram ---
	start = time.Now()
	projectPath := filepath.Join("../../testdata/circuit_constraints_20k")
	_, err = CompileProgram(context.Background(), w, projectPath)
	if err != nil {
		t.Fatalf("CompileProgram failed: %v", err)
	}
	t.Logf("CompileProgram took %s", time.Since(start))

	// --- Output ---
	//t.Logf("Test Output: %+v", out)

	t.Logf("TOTAL TEST TIME took %s", time.Since(totalStart))
}
