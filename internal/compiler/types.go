package compiler

type ABI struct {
	//gonna do some abi stuff here.

}
type AcirBlob []byte

type Compilation struct {
	ACIR []byte
	//ABI  abi.ABI

	NoirVersion string
	Hash        uint64

	PrivateParamWitnesses []uint32
	PublicParamWitnesses  []uint32
}

type WireCompileResult struct {
	FormatVersion         uint32   `msgpack:"format_version"`
	NoirVersion           string   `msgpack:"noir_version"`
	AbiJSON               string   `msgpack:"abi_json"`
	AcirString            string   `msgpack:"acir_string"`
	AcirBytes             []int    `msgpack:"acir_bytes"`
	Hash                  uint64   `msgpack:"hash"`
	PrivateParamWitnesses []uint32 `msgpack:"private_param_witnesses"`
	PublicParamWitnesses  []uint32 `msgpack:"public_param_witnesses"`
}

type wasmBuf struct {
	ptr uint32
	len uint32
}
