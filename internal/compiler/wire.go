// compiler/wire.go
package compiler

type wireCompileResult struct {
	FormatVersion         uint32   `msgpack:"format_version"`
	NoirVersion           string   `msgpack:"noir_version"`
	AbiJSON               string   `msgpack:"abi_json"`
	AcirString            string   `msgpack:"acir_string"`
	AcirBytes             []int    `msgpack:"acir_bytes"`
	Hash                  uint64   `msgpack:"hash"`
	PrivateParamWitnesses []uint32 `msgpack:"private_param_witnesses"`
	PublicParamWitnesses  []uint32 `msgpack:"public_param_witnesses"`
}
