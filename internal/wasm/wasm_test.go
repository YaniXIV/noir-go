package wasm

import (
	"testing"
	"time"
)

func TestWasmManagerGetCompiler(t *testing.T) {
	totalStart := time.Now()
	defer func() {
		t.Logf("total test duration: %s", time.Since(totalStart))
	}()

	stepStart := time.Now()
	w, err := NewWasmManager()
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
	obj1, err := w.Get(Compiler)
	if err != nil {
		t.Fatalf("Get(Compiler) failed: %v", err)
	}
	if obj1 == nil || obj1.Compiled == nil {
		t.Fatalf("expected compiled compiler module")
	}
	t.Logf("Get(Compiler) first call duration: %s", time.Since(stepStart))

	stepStart = time.Now()
	obj2, err := w.Get(Compiler)
	if err != nil {
		t.Fatalf("Get(Compiler) second call failed: %v", err)
	}
	if obj2 != obj1 {
		t.Fatalf("expected cached compiler object on second Get")
	}
	t.Logf("Get(Compiler) second call duration: %s", time.Since(stepStart))
}
