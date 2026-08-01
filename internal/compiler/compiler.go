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

const maxStderrSnippet = 4 << 10 // 4KB

// wireEnvelopeVersion must match the "v" field write_envelope stamps on
// every response in tools/noir-compile/src/main.rs.
const wireEnvelopeVersion = 1

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

	var stderr bytes.Buffer
	mod, err := w.Instantiate(ctx, obj.Compiled, wazero.NewModuleConfig().WithStderr(&stderr))
	if err != nil {
		return nil, err
	}
	defer mod.Close(context.Background())

	// Call Bridge
	resultPayload, err := callWasmCompile(ctx, mod, projectData)
	if err != nil {
		return nil, err
	}

	payload, err := decodeEnvelope(resultPayload, "noir compiler", &stderr)
	if err != nil {
		return nil, err
	}

	var wire WireCompileResult
	if err := msgpack.Unmarshal(payload, &wire); err != nil {
		return nil, fmt.Errorf("msgpack unmarshal failed: %w", err)
	}
	comp, err := wire.processCompilation()
	if err != nil {
		return nil, err
	}

	return comp, nil
}

// decodeEnvelope unwraps a WireMessage envelope (see
// tools/noir-compile/src/main.rs's write_envelope): on success it returns
// the raw payload bytes for the caller to unmarshal further; on failure it
// decodes the WireError payload and returns it as a Go error, with a
// truncated capture of the wasm module's stderr appended for extra context
// when available.
func decodeEnvelope(resultPayload []byte, context string, stderr *bytes.Buffer) ([]byte, error) {
	if len(resultPayload) == 0 {
		return nil, fmt.Errorf("%s: wasm returned an empty result (allocation failure?)%s", context, stderrSnippet(stderr))
	}

	var envelope WireEnvelope
	if err := msgpack.Unmarshal(resultPayload, &envelope); err != nil {
		return nil, fmt.Errorf("%s: malformed wasm response: %w", context, err)
	}
	if envelope.Version != wireEnvelopeVersion {
		return nil, fmt.Errorf("%s: unsupported wasm response version %d (expected %d) -- the embedded wasm binary may be out of sync with this build", context, envelope.Version, wireEnvelopeVersion)
	}

	if !envelope.Ok {
		var wireErr WireError
		if err := msgpack.Unmarshal(envelope.Payload, &wireErr); err != nil {
			return nil, fmt.Errorf("%s: failed and the error payload itself was malformed: %w", context, err)
		}
		return nil, fmt.Errorf("%s: %s%s", context, wireErr.Message, stderrSnippet(stderr))
	}

	return envelope.Payload, nil
}

// stderrSnippet returns a " (stderr: ...)" suffix with a truncated capture
// of the wasm module's stderr, or "" if nothing was captured.
func stderrSnippet(stderr *bytes.Buffer) string {
	if stderr == nil || stderr.Len() == 0 {
		return ""
	}
	s := stderr.String()
	if len(s) > maxStderrSnippet {
		s = s[:maxStderrSnippet] + "... (truncated)"
	}
	return fmt.Sprintf(" (stderr: %s)", s)
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

	var stderr bytes.Buffer
	mod, err := w.Instantiate(ctx, obj.Compiled, wazero.NewModuleConfig().WithStderr(&stderr))
	if err != nil {
		return nil, err
	}
	defer mod.Close(context.Background())

	resultPayload, err := callWasmExecute(ctx, mod, packed)
	if err != nil {
		return nil, fmt.Errorf("callWasmExecute: %w", err)
	}

	payload, err := decodeEnvelope(resultPayload, "noir execute", &stderr)
	if err != nil {
		return nil, err
	}

	var raw map[string]interface{}
	if err := msgpack.Unmarshal(payload, &raw); err != nil {
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
