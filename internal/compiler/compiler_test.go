package compiler

import (
	"context"
	"os"
	"path/filepath"
	"strings"
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

func writeNoirProject(t *testing.T, name, source string) string {
	t.Helper()
	tmp := t.TempDir()
	manifest := "[package]\nname = \"" + name + "\"\ntype = \"bin\"\n\n[dependencies]\n"
	if err := os.WriteFile(filepath.Join(tmp, "Nargo.toml"), []byte(manifest), 0644); err != nil {
		t.Fatalf("write Nargo.toml: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, "src"), 0755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "src", "main.nr"), []byte(source), 0644); err != nil {
		t.Fatalf("write main.nr: %v", err)
	}
	return tmp
}

// TestCompileProgramReturnsCompilerError checks that a genuine Noir compile
// error (an undefined identifier) surfaces as a real diagnostic from the
// Rust compiler, not the generic "msgpack unmarshal failed" that used to
// come back when compile_wasm's error paths zeroed the output buffer.
func TestCompileProgramReturnsCompilerError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping wasm compile in short mode")
	}

	tmp := writeNoirProject(t, "broken", "fn main(x: Field) -> pub Field {\n    x + this_identifier_does_not_exist\n}\n")

	w, err := wasm.NewWasmManager()
	if err != nil {
		t.Fatalf("NewWasmManager: %v", err)
	}
	defer w.Close(context.Background())

	_, err = CompileProgram(context.Background(), w, tmp)
	if err == nil {
		t.Fatal("expected a compile error for an undefined identifier")
	}
	t.Logf("got error: %v", err)

	msg := err.Error()
	if strings.Contains(msg, "msgpack unmarshal failed") || strings.Contains(msg, "malformed wasm response") {
		t.Fatalf("expected the real compiler diagnostic, got a generic decode failure: %v", err)
	}
	if !strings.Contains(msg, "this_identifier_does_not_exist") {
		t.Fatalf("expected error to reference the undefined identifier, got: %v", err)
	}
}

// TestExecuteProgramReturnsSolverError checks that an unsatisfiable
// constraint (assert(x == y) with x != y) surfaces as a real ACVM solver
// diagnostic, not the generic "execute_wasm returned empty payload" that
// used to come back when execute_wasm's error paths zeroed the output
// buffer.
func TestExecuteProgramReturnsSolverError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping wasm execute in short mode")
	}

	tmp := writeNoirProject(t, "assert_fail", "fn main(x: Field, y: Field) {\n    assert(x == y);\n}\n")

	ctx := context.Background()
	w, err := wasm.NewWasmManager()
	if err != nil {
		t.Fatalf("NewWasmManager: %v", err)
	}
	defer w.Close(ctx)

	comp, err := CompileProgram(ctx, w, tmp)
	if err != nil {
		t.Fatalf("CompileProgram: %v", err)
	}
	if len(comp.PrivateParamWitnesses) != 2 {
		t.Fatalf("expected 2 private witnesses (x, y), got %v", comp.PrivateParamWitnesses)
	}

	// x = 1, y = 2: deliberately violates assert(x == y).
	inputs := make(map[uint32][32]byte)
	for i, idx := range comp.PrivateParamWitnesses {
		var v [32]byte
		v[31] = byte(i + 1)
		inputs[idx] = v
	}

	_, err = ExecuteProgram(ctx, w, comp, inputs)
	if err == nil {
		t.Fatal("expected a solver failure for x != y")
	}
	t.Logf("got error: %v", err)

	msg := err.Error()
	if strings.Contains(msg, "empty payload") || strings.Contains(msg, "malformed wasm response") {
		t.Fatalf("expected the real solver diagnostic, got a generic decode failure: %v", err)
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
