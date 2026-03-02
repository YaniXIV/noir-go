package compiler

import (
	"bytes"
	"context"
	//"encoding/binary"
	//"encoding/json"
	//"encoding/base64"
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
	"noir-go/internal/fs"
	"noir-go/internal/wasm"
	"unsafe"

	"github.com/tetratelabs/wazero"
)

func CompileProgram(ctx context.Context, w *wasm.WasmManager, projectPath string) (*Compilation, error) {

	// Resolve and Serialize Project
	r := fs.NewResolver()
	r.Resolve(projectPath)
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
	defer mod.Close(ctx)

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

func readBytes(addr uintptr, size int) ([]byte, error) {
	if size < 1 {
		return nil, fmt.Errorf("data size cannot be < 1")
	}
	var AcirBlob []byte = make([]byte, size)
	data := unsafe.Slice((*byte)(unsafe.Pointer(addr)), size)
	copy(data, AcirBlob)
	if len(AcirBlob) < 1 {
		return nil, fmt.Errorf("Error writting data")
	}
	return AcirBlob, nil

}

func writeBytes(addr uintptr, data []byte) {
	dst := unsafe.Slice((*byte)(unsafe.Pointer(addr)), len(data))
	copy(dst, data)
}
