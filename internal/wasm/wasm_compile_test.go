package wasm_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/YaniXIV/noir-go/internal/compiler"
	"github.com/YaniXIV/noir-go/internal/wasm"
)

func TestCompileProgramTimingsFromWasm(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping wasm compile in short mode")
	}

	totalStart := time.Now()
	defer func() {
		t.Logf("total test duration: %s", time.Since(totalStart))
	}()

	stepStart := time.Now()
	w, err := wasm.NewWasmManager()
	if err != nil {
		t.Fatalf("NewWasmManager failed: %v", err)
	}
	t.Logf("NewWasmManager duration: %s", time.Since(stepStart))

	stepStart = time.Now()
	if w.GetRuntime() == nil {
		t.Fatalf("expected non-nil runtime")
	}
	t.Logf("GetRuntime duration: %s", time.Since(stepStart))

	stepStart = time.Now()
	obj, err := w.Get(wasm.Compiler)
	if err != nil {
		t.Fatalf("Get(Compiler) failed: %v", err)
	}
	if obj == nil || obj.Compiled == nil {
		t.Fatalf("expected compiled compiler module")
	}
	t.Logf("Get(Compiler) duration: %s", time.Since(stepStart))

	stepStart = time.Now()
	compilation, err := compiler.CompileProgram(context.Background(), w, "../../testdata/circuit_constraints_20k")
	if err != nil {
		t.Fatalf("CompileProgram failed: %v", err)
	}
	if compilation == nil {
		t.Fatalf("expected non-nil compilation result")
	}
	if len(compilation.ACIR.Bytes) == 0 {
		t.Fatalf("expected non-empty ACIR result")
	}
	t.Logf("CompileProgram duration: %s", time.Since(stepStart))
}

func TestCompileProgramTimingsFromWarmup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping wasm compile in short mode")
	}

	totalStart := time.Now()
	defer func() {
		t.Logf("total test duration: %s", time.Since(totalStart))
	}()

	stepStart := time.Now()
	w, err := wasm.NewWasmManager()
	if err != nil {
		t.Fatalf("NewWasmManager failed: %v", err)
	}
	t.Logf("NewWasmManager duration: %s", time.Since(stepStart))

	stepStart = time.Now()
	w.Warmup()

	t.Logf("Warmup duration: %s", time.Since(stepStart))
	stepStart = time.Now()
	if w.GetRuntime() == nil {
		t.Fatalf("expected non-nil runtime")
	}
	t.Logf("GetRuntime duration: %s", time.Since(stepStart))

	stepStart = time.Now()
	obj, err := w.Get(wasm.Compiler)
	if err != nil {
		t.Fatalf("Get(Compiler) failed: %v", err)
	}
	if obj == nil || obj.Compiled == nil {
		t.Fatalf("expected compiled compiler module")
	}
	t.Logf("Get(Compiler) duration: %s", time.Since(stepStart))

	stepStart = time.Now()
	compilation, err := compiler.CompileProgram(context.Background(), w, "../../testdata/circuit_constraints_20k")
	if err != nil {
		t.Fatalf("CompileProgram failed: %v", err)
	}
	if compilation == nil {
		t.Fatalf("expected non-nil compilation result")
	}
	if len(compilation.ACIR.Bytes) == 0 {
		t.Fatalf("expected non-empty ACIR result")
	}
	t.Logf("CompileProgram duration: %s", time.Since(stepStart))
}

func compileProgramTimings(t *testing.T, projectPath string) {
	t.Helper()
	totalStart := time.Now()
	defer func() {
		t.Logf("[%s] total duration: %s", projectPath, time.Since(totalStart))
	}()

	stepStart := time.Now()
	w, err := wasm.NewWasmManager()
	if err != nil {
		t.Fatalf("NewWasmManager failed: %v", err)
	}
	t.Logf("[%s] NewWasmManager: %s", projectPath, time.Since(stepStart))

	stepStart = time.Now()
	if w.GetRuntime() == nil {
		t.Fatalf("expected non-nil runtime")
	}
	t.Logf("[%s] GetRuntime: %s", projectPath, time.Since(stepStart))

	stepStart = time.Now()
	obj, err := w.Get(wasm.Compiler)
	if err != nil {
		t.Fatalf("Get(Compiler) failed: %v", err)
	}
	if obj == nil || obj.Compiled == nil {
		t.Fatalf("expected compiled compiler module")
	}
	t.Logf("[%s] Get(Compiler): %s", projectPath, time.Since(stepStart))

	stepStart = time.Now()
	compilation, err := compiler.CompileProgram(context.Background(), w, projectPath)
	if err != nil {
		t.Fatalf("CompileProgram failed: %v", err)
	}
	if compilation == nil {
		t.Fatalf("expected non-nil compilation result")
	}
	if len(compilation.ACIR.Bytes) == 0 {
		t.Fatalf("expected non-empty ACIR result")
	}
	t.Logf("[%s] CompileProgram: %s", projectPath, time.Since(stepStart))
}

// actual test that loops through a directory
func TestCompileProgramTimingsMultipleCircuits(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping wasm compile in short mode")
	}

	circuitsDir := "../../testdata"

	entries, err := os.ReadDir(circuitsDir)
	if err != nil {
		t.Fatalf("failed to read circuits dir: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectPath := filepath.Join(circuitsDir, entry.Name())
		t.Run(entry.Name(), func(t *testing.T) {
			compileProgramTimings(t, projectPath)
		})
	}
}
