package compiler

// WireEnvelope wraps every compile_wasm/execute_wasm response. Payload holds
// the msgpack-encoded success struct when Ok is true, or a WireError when
// Ok is false -- see tools/noir-compile/src/main.rs's write_envelope.
type WireEnvelope struct {
	Version uint16 `msgpack:"v"`
	Ok      bool   `msgpack:"k"`
	Payload []byte `msgpack:"p"`
}

// WireError is the envelope payload on failure.
type WireError struct {
	Message string `msgpack:"message"`
}

type WireCompileResult struct {
	FormatVersion         uint32   `msgpack:"format_version"`
	NoirVersion           string   `msgpack:"noir_version"`
	AbiJSON               string   `msgpack:"abi_json"`
	AcirString            string   `msgpack:"acir_string"`
	AcirBytes             []int    `msgpack:"acir_bytes"`
	AcirJSON              string   `msgpack:"acir_json"`
	Hash                  uint64   `msgpack:"hash"`
	PrivateParamWitnesses []uint32 `msgpack:"private_param_witnesses"`
	PublicParamWitnesses  []uint32 `msgpack:"public_param_witnesses"`
}

type wasmBuf struct {
	ptr uint32
	len uint32
}
