package compiler

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/vmihailenco/msgpack/v5"

	"noir-go/internal/fs"
	"noir-go/internal/wasm"
)

type benchProject struct {
	name string
	path string
}

type benchTotals struct {
	newManager  time.Duration
	resolve     time.Duration
	getCompiler time.Duration
	instantiate time.Duration
	callCompile time.Duration
	unmarshal   time.Duration
	process     time.Duration
}

func BenchmarkCompileProgram(b *testing.B) {
	projects := []benchProject{
		{name: "noirtest", path: filepath.Join("noirtest")},
		{name: "circuit_small", path: filepath.Join("..", "..", "testdata", "circuit_small")},
		{name: "circuit_wide_4k", path: filepath.Join("..", "..", "testdata", "circuit_wide_4k")},
		{name: "circuit_int_heavy_5k", path: filepath.Join("..", "..", "testdata", "circuit_int_heavy_5k")},
		{name: "circuit_constraints_20k", path: filepath.Join("..", "..", "testdata", "circuit_constraints_20k")},
	}

	for _, p := range projects {
		p := p
		b.Run("cold/"+p.name, func(b *testing.B) {
			benchCompileProject(b, p.path, true)
		})
		b.Run("warm/"+p.name, func(b *testing.B) {
			benchCompileProject(b, p.path, false)
		})
	}
}

func benchCompileProject(b *testing.B, projectPath string, cold bool) {
	ctx := context.Background()
	var w *wasm.WasmManager
	var err error
	var totals benchTotals

	if !cold {
		w, err = wasm.NewWasmManager()
		if err != nil {
			b.Fatalf("NewWasmManager failed: %v", err)
		}
		if _, err := w.Get(wasm.Compiler); err != nil {
			b.Fatalf("Get(Compiler) warmup failed: %v", err)
		}
		b.Cleanup(func() {
			_ = w.Close(ctx)
		})
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if cold {
			stepStart := time.Now()
			w, err = wasm.NewWasmManager()
			if err != nil {
				b.Fatalf("NewWasmManager failed: %v", err)
			}
			totals.newManager += time.Since(stepStart)
		}

		stepStart := time.Now()
		r := fs.NewResolver()
		r.Resolve(projectPath)
		projectData, err := r.Serialize()
		if err != nil {
			b.Fatalf("project serialization failed: %v", err)
		}
		totals.resolve += time.Since(stepStart)

		stepStart = time.Now()
		obj, err := w.Get(wasm.Compiler)
		if err != nil || obj == nil || obj.Compiled == nil {
			b.Fatalf("Get(Compiler) failed: %v", err)
		}
		totals.getCompiler += time.Since(stepStart)

		stepStart = time.Now()
		mod, err := w.Instantiate(ctx, obj.Compiled, wazero.NewModuleConfig())
		if err != nil {
			b.Fatalf("Instantiate failed: %v", err)
		}
		totals.instantiate += time.Since(stepStart)

		stepStart = time.Now()
		resultPayload, err := callWasmCompile(ctx, mod, projectData)
		if err != nil {
			b.Fatalf("callWasmCompile failed: %v", err)
		}
		totals.callCompile += time.Since(stepStart)

		stepStart = time.Now()
		var wire WireCompileResult
		if err := msgpack.Unmarshal(resultPayload, &wire); err != nil {
			b.Fatalf("msgpack unmarshal failed: %v", err)
		}
		totals.unmarshal += time.Since(stepStart)

		stepStart = time.Now()
		comp, err := wire.processCompilation()
		if err != nil {
			b.Fatalf("processCompilation failed: %v", err)
		}
		if len(comp.ACIR) == 0 {
			b.Fatalf("expected non-empty ACIR result")
		}
		totals.process += time.Since(stepStart)

		if err := mod.Close(ctx); err != nil {
			b.Fatalf("module close failed: %v", err)
		}
		if cold {
			if err := w.Close(ctx); err != nil {
				b.Fatalf("runtime close failed: %v", err)
			}
		}
	}
	b.StopTimer()

	reportTotals(b, totals)
}

func reportTotals(b *testing.B, totals benchTotals) {
	perOp := func(d time.Duration) float64 {
		if b.N == 0 {
			return 0
		}
		return float64(d.Microseconds()) / float64(b.N)
	}

	if totals.newManager > 0 {
		b.ReportMetric(perOp(totals.newManager), "new_manager_us/op")
	}
	b.ReportMetric(perOp(totals.resolve), "resolve_us/op")
	b.ReportMetric(perOp(totals.getCompiler), "get_compiler_us/op")
	b.ReportMetric(perOp(totals.instantiate), "instantiate_us/op")
	b.ReportMetric(perOp(totals.callCompile), "compile_us/op")
	b.ReportMetric(perOp(totals.unmarshal), "unmarshal_us/op")
	b.ReportMetric(perOp(totals.process), "process_us/op")
}
