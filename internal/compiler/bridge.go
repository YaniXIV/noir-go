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
	if alloc == nil {
		return nil, fmt.Errorf("wasm export %q not found", "alloc")
	}
	if dealloc == nil {
		return nil, fmt.Errorf("wasm export %q not found", "dealloc")
	}
	if compileFn == nil {
		return nil, fmt.Errorf("wasm export %q not found", "compile_wasm")
	}

	// Allocate and Write Input
	inputSize := uint64(len(input))
	res, err := alloc.Call(ctx, inputSize)
	if err != nil {
		return nil, fmt.Errorf("alloc input failed: %w", err)
	}
	inPtr := uint32(res[0])
	// Defer the deallocation
	defer dealloc.Call(ctx, uint64(inPtr), inputSize)

	if !mod.Memory().Write(inPtr, input) {
		return nil, fmt.Errorf("failed to write input to wasm memory")
	}

	// Allocate space for the Result Pointer (8 bytes for ptr+len)
	resPtrAlloc, err := alloc.Call(ctx, 8)
	if err != nil {
		return nil, fmt.Errorf("alloc res holder failed: %w", err)
	}
	outStructPtr := uint32(resPtrAlloc[0])
	// Defer the deallocation
	defer dealloc.Call(ctx, uint64(outStructPtr), 8)

	// Execute
	_, err = compileFn.Call(ctx, uint64(outStructPtr), uint64(inPtr), inputSize)
	if err != nil {
		return nil, fmt.Errorf("compile_wasm execution failed: %w", err)
	}

	// Read the pointer and length from the output struct
	buf, ok := mod.Memory().Read(outStructPtr, 8)
	if !ok {
		return nil, fmt.Errorf("failed to read output header")
	}
	retPtr := binary.LittleEndian.Uint32(buf[0:4])
	retLen := binary.LittleEndian.Uint32(buf[4:8])
	if retLen > 0 {
		defer dealloc.Call(ctx, uint64(retPtr), uint64(retLen))
	}

	// Read the actual result bytes
	resultBytes, ok := mod.Memory().Read(retPtr, retLen)
	if !ok {
		return nil, fmt.Errorf("failed to read result data")
	}

	// Copy data out of WASM memory so it persists after module close
	out := make([]byte, len(resultBytes))
	copy(out, resultBytes)

	return out, nil
}

func callWasmExecute(ctx context.Context, mod api.Module, input []byte) ([]byte, error) {
	alloc := mod.ExportedFunction("alloc")
	dealloc := mod.ExportedFunction("dealloc")
	executeFn := mod.ExportedFunction("execute_wasm")

	if alloc == nil {
		return nil, fmt.Errorf("wasm export %q not found", "alloc")
	}
	if dealloc == nil {
		return nil, fmt.Errorf("wasm export %q not found", "dealloc")
	}
	if executeFn == nil {
		return nil, fmt.Errorf("wasm export %q not found", "execute_wasm")
	}

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

	resPtrAlloc, err := alloc.Call(ctx, 8)
	if err != nil {
		return nil, fmt.Errorf("alloc res holder failed: %w", err)
	}
	outStructPtr := uint32(resPtrAlloc[0])
	defer dealloc.Call(ctx, uint64(outStructPtr), 8)

	_, err = executeFn.Call(ctx, uint64(outStructPtr), uint64(inPtr), inputSize)
	if err != nil {
		return nil, fmt.Errorf("execute_wasm execution failed: %w", err)
	}

	buf, ok := mod.Memory().Read(outStructPtr, 8)
	if !ok {
		return nil, fmt.Errorf("failed to read output header")
	}
	retPtr := binary.LittleEndian.Uint32(buf[0:4])
	retLen := binary.LittleEndian.Uint32(buf[4:8])
	if retLen > 0 {
		defer dealloc.Call(ctx, uint64(retPtr), uint64(retLen))
	}

	resultBytes, ok := mod.Memory().Read(retPtr, retLen)
	if !ok {
		return nil, fmt.Errorf("failed to read result data")
	}

	out := make([]byte, len(resultBytes))
	copy(out, resultBytes)
	return out, nil
}
