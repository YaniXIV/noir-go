package compiler

type ABI struct {
	Parameters []ABIParameter     `json:"parameters"`
	ReturnType *ABIType           `json:"return_type"`
	ErrorTypes map[string]ABIType `json:"error_types"`

	// 👇 Your custom field (not part of JSON)
	ParameterWitnessIndices []uint32 `json:"-"`
}

type ABIParameter struct {
	Name       string  `json:"name"`
	Type       ABIType `json:"type"`
	Visibility string  `json:"visibility"` // "public" | "private"
}

type ABIType struct {
	Kind string `json:"kind"`
}

type AcirBlob []byte

type ACIR struct {
	AcirBytes []byte
	AcirJson  string
}
type Compilation struct {
	ACIR ACIR
	ABI  ABI

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
	AcirJson              string   `msgpack:"acir_json"`
	Hash                  uint64   `msgpack:"hash"`
	PrivateParamWitnesses []uint32 `msgpack:"private_param_witnesses"`
	PublicParamWitnesses  []uint32 `msgpack:"public_param_witnesses"`
}

type wasmBuf struct {
	ptr uint32
	len uint32
}
