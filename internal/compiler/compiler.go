package compiler

import (
	"bytes"
	"context"
	//"encoding/binary"
	//"encoding/json"
	//"encoding/base64"
	"fmt"
	"github.com/YaniXIV/noir-go/internal/fs"
	"github.com/YaniXIV/noir-go/internal/wasm"
	"github.com/YaniXIV/noir-go/result"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/tetratelabs/wazero"
	"os"
)

func CompileProgram(ctx context.Context, w *wasm.WasmManager, projectPath string) (*result.Compilation, error) {
	// make sure the filepath actually exists, if not exit before doing expensive calls.
	_, err := os.Stat(projectPath)
	if err != nil {
		return nil, err
	}
	// Resolve and Serialize Project
	r := fs.NewResolver()
	if err := r.Resolve(projectPath); err != nil {
		return nil, fmt.Errorf("resolve noir project %q: %w", projectPath, err)
	}
	projectData, err := r.Serialize()
	if err != nil {
		return nil, fmt.Errorf("project serialization failed: %w", err)
	}

	// Setup WASM Instance
	obj, err := w.Get(wasm.Compiler)
	if err != nil || obj == nil {
		return nil, fmt.Errorf("failed to get compiler wasm: %v", err)
	}

	mod, err := w.Instantiate(ctx, obj.Compiled, wazero.NewModuleConfig())
	if err != nil {
		return nil, err
	}
	defer mod.Close(context.Background())

	// Call Bridge
	resultPayload, err := callWasmCompile(ctx, mod, projectData)
	if err != nil {
		return nil, err
	}

	// Unmarshal
	var wire WireCompileResult
	if err := msgpack.Unmarshal(resultPayload, &wire); err != nil {
		return nil, fmt.Errorf("msgpack unmarshal failed: %w", err)
	}
	comp, err := wire.processCompilation()
	if err != nil {
		return nil, err
	}

	return comp, nil
}

func printLogs(outBuf *bytes.Buffer) {
	fmt.Println("--- Start of wasm logs ---")
	fmt.Println(outBuf.String())
	fmt.Println("--- End of wasm logs ---")
}
