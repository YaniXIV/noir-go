package compiler

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/tetratelabs/wazero/api"
)

// callWasmCompile handles the low-level memory dance with the WASM module
func callWasmCompile(ctx context.Context, mod api.Module, input []byte) ([]byte, error) {
	alloc := mod.ExportedFunction("alloc")
	dealloc := mod.ExportedFunction("dealloc")
	compileFn := mod.ExportedFunction("compile_wasm")

	// 1. Allocate and Write Input
	inputSize := uint64(len(input))
	res, err := alloc.Call(ctx, inputSize)
	if err != nil {
		return nil, fmt.Errorf("alloc input failed: %w", err)
	}
	inPtr := uint32(res[0])
	defer dealloc.Call(ctx, uint64(inPtr), inputSize)

	if !mod.Memory().Write(inPtr, input) {
		return nil, fmt.Errorf("failed to write input to wasm memory")
	}

	// 2. Allocate space for the Result Pointer (8 bytes for ptr+len)
	resPtrAlloc, err := alloc.Call(ctx, 8)
	if err != nil {
		return nil, fmt.Errorf("alloc res holder failed: %w", err)
	}
	outStructPtr := uint32(resPtrAlloc[0])

	// 3. Execute
	_, err = compileFn.Call(ctx, uint64(outStructPtr), uint64(inPtr), inputSize)
	if err != nil {
		return nil, fmt.Errorf("compile_wasm execution failed: %w", err)
	}

	// 4. Read the pointer and length from the output struct
	buf, ok := mod.Memory().Read(outStructPtr, 8)
	if !ok {
		return nil, fmt.Errorf("failed to read output header")
	}
	retPtr := binary.LittleEndian.Uint32(buf[0:4])
	retLen := binary.LittleEndian.Uint32(buf[4:8])

	// 5. Read the actual result bytes
	resultBytes, ok := mod.Memory().Read(retPtr, retLen)
	if !ok {
		return nil, fmt.Errorf("failed to read result data")
	}

	// Copy data out of WASM memory so it persists after module close
	out := make([]byte, len(resultBytes))
	copy(out, resultBytes)

	return out, nil
}
