package compiler

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/YaniXIV/noir-go/internal/wasm"
)

const circuit_wide_4k = "../../testdata/circuit_wide_4k"
const regtest = "noirtest"

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
	projectPath := filepath.Join(regtest)
	_, err = CompileProgram(context.Background(), w, projectPath)
	if err != nil {
		t.Fatalf("CompileProgram failed: %v", err)
	}
	t.Logf("CompileProgram took %s", time.Since(start))

	// --- Output ---
	//t.Logf("Test Output: %+v", out)

	t.Logf("TOTAL TEST TIME took %s", time.Since(totalStart))
}

func TestCompileProgramReturnsResolverError(t *testing.T) {
	tmp := t.TempDir()
	err := os.WriteFile(filepath.Join(tmp, "Nargo.toml"), []byte("[package]\nname = \"tmp\"\ntype = \"bin\"\n\n[dependencies]\n"), 0644)
	if err != nil {
		t.Fatalf("write Nargo.toml: %v", err)
	}

	_, err = CompileProgram(context.Background(), nil, tmp)
	if err == nil {
		t.Fatalf("expected resolver error")
	}
}
