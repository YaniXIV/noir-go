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
const circuit_simple = "../../testdata/circuit_simple"

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

func TestExecuteProgram_SmallCircuit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping wasm execute in short mode")
	}

	ctx := context.Background()

	w, err := wasm.NewWasmManager()
	if err != nil {
		t.Fatalf("NewWasmManager: %v", err)
	}
	defer w.Close(ctx)

	comp, err := CompileProgram(ctx, w, circuit_simple)
	if err != nil {
		t.Fatalf("CompileProgram: %v", err)
	}

	t.Logf("private witnesses: %v", comp.PrivateParamWitnesses)
	t.Logf("public witnesses: %v", comp.PublicParamWitnesses)

	inputs := make(map[uint32][32]byte)
	for _, idx := range comp.PrivateParamWitnesses {
		var val [32]byte
		val[31] = 3
		inputs[idx] = val
	}
	for _, idx := range comp.PublicParamWitnesses {
		var val [32]byte
		val[31] = 3
		inputs[idx] = val
	}

	solved, err := ExecuteProgram(ctx, w, comp, inputs)
	if err != nil {
		t.Fatalf("ExecuteProgram: %v", err)
	}

	if len(solved) == 0 {
		t.Fatal("solved witness is empty")
	}

	t.Logf("solved witness entries: %d", len(solved))
	for idx, val := range solved {
		t.Logf("  witness[%d] = %x", idx, val)
	}
}

func buildInitialWitnessRaw(inputs map[uint32][32]byte) []interface{} {
	var out []interface{}
	for idx, val := range inputs {
		v := val
		out = append(out, []interface{}{idx, v[:]})
	}
	return out
}
