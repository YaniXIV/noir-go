package wasm

import "testing"

func TestWasmManagerGetCompiler(t *testing.T) {
	w, err := NewWasmManager()
	if err != nil {
		t.Fatalf("NewWasmManager failed: %v", err)
	}
	if w.GetRuntime() == nil {
		t.Fatalf("expected non-nil runtime")
	}

	obj1, err := w.Get(Compiler)
	if err != nil {
		t.Fatalf("Get(Compiler) failed: %v", err)
	}
	if obj1 == nil || obj1.Compiled == nil {
		t.Fatalf("expected compiled compiler module")
	}

	obj2, err := w.Get(Compiler)
	if err != nil {
		t.Fatalf("Get(Compiler) second call failed: %v", err)
	}
	if obj2 != obj1 {
		t.Fatalf("expected cached compiler object on second Get")
	}
}
