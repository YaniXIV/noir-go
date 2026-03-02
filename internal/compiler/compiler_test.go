package compiler

import (
	"path/filepath"
	"testing"

	"fmt"
	"noir-go/internal/wasm"
)

func TestCompileProgram(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping wasm compile in short mode")
	}

	w, err := wasm.NewWasmManager()
	if err != nil {
		t.Fatalf("NewWasmManager failed: %v", err)
	}
	if w.GetRuntime() == nil {
		t.Fatalf("expected non-nil runtime")
	}

	projectPath := filepath.Join("noirtest")
	out, err := CompileProgram(w, projectPath)
	if err != nil {
		t.Fatalf("CompileProgram failed: %v", err)
	}

	fmt.Println(out)
	/**
	if out == nil || len(out) == 0 {
		t.Skip("CompileProgram does not return bytes yet; enable once implemented")
	}
	*/
}
