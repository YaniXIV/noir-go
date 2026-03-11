package wasm_test

import (
	"context"
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
