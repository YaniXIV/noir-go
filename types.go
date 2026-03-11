package noirgo

type ABI struct {
	Parameters []ABIParameter     `json:"parameters"`
	ReturnType *ABIType           `json:"return_type"`
	ErrorTypes map[string]ABIType `json:"error_types"`

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

type ACIR struct {
	Bytes []byte
	JSON  string
}
type Compilation struct {
	ACIR ACIR
	ABI  ABI

	NoirVersion string
	Hash        uint64

	PrivateParamWitnesses []uint32
	PublicParamWitnesses  []uint32
}
