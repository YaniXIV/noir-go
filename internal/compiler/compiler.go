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

func ExecuteProgram(ctx context.Context, w *wasm.WasmManager, comp *result.Compilation, inputs map[uint32][32]byte) (map[uint32][32]byte, error) {
	var initialWitness []interface{}
	for idx, val := range inputs {
		v := val
		initialWitness = append(initialWitness, []interface{}{idx, v[:]})
	}

	input := map[string]interface{}{
		"acir_bytes":      comp.ACIR.Bytes,
		"initial_witness": initialWitness,
	}

	packed, err := msgpack.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshal execute input: %w", err)
	}

	obj, err := w.Get(wasm.Compiler)
	if err != nil || obj == nil {
		return nil, fmt.Errorf("failed to get compiler wasm: %w", err)
	}

	mod, err := w.Instantiate(ctx, obj.Compiled, wazero.NewModuleConfig())
	if err != nil {
		return nil, err
	}
	defer mod.Close(context.Background())

	resultPayload, err := callWasmExecute(ctx, mod, packed)
	if err != nil {
		return nil, fmt.Errorf("callWasmExecute: %w", err)
	}

	if len(resultPayload) == 0 {
		return nil, fmt.Errorf("execute_wasm returned empty payload")
	}

	var raw map[string]interface{}
	if err := msgpack.Unmarshal(resultPayload, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal execute result: %w", err)
	}

	witnessOut := make(map[uint32][32]byte)
	entries, _ := raw["witness"].([]interface{})
	for _, e := range entries {
		pair, ok := e.([]interface{})
		if !ok || len(pair) != 2 {
			continue
		}

		var idx uint32
		switch v := pair[0].(type) {
		case int8:
			idx = uint32(v)
		case uint8:
			idx = uint32(v)
		case int16:
			idx = uint32(v)
		case uint16:
			idx = uint32(v)
		case int32:
			idx = uint32(v)
		case uint32:
			idx = v
		case int64:
			idx = uint32(v)
		case uint64:
			idx = uint32(v)
		default:
			continue
		}

		rawBytes, ok := pair[1].([]interface{})
		if !ok {
			continue
		}

		if len(rawBytes) != 32 {
			continue
		}

		var arr [32]byte
		for i, b := range rawBytes {
			switch v := b.(type) {
			case int8:
				arr[i] = byte(v)
			case uint8:
				arr[i] = v
			case int64:
				arr[i] = byte(v)
			case uint64:
				arr[i] = byte(v)
			default:
				arr[i] = 0
			}
		}
		witnessOut[idx] = arr
	}

	return witnessOut, nil
}
