// compiler/wire.go
package compiler

import (
	"encoding/json"
	"fmt"
	"github.com/YaniXIV/noir-go/result"
)

// ProcessCompilation converts the wire format coming back from wasm/msgpack
// into an internal Compilation object + parsed ABI with ordered param witness indices.
func (w *WireCompileResult) processCompilation() (*result.Compilation, error) {
	// 1) Parse ABI JSON
	var abi result.ABI
	if w.AbiJSON != "" {
		if err := json.Unmarshal([]byte(w.AbiJSON), &abi); err != nil {
			return nil, fmt.Errorf("parse abi_json: %w", err)
		}
	} else {
		// If AbiJSON is empty, keep abi as zero-value but still proceed.
		abi.Parameters = nil
		abi.ErrorTypes = nil
		abi.ReturnType = nil
	}

	// Convert ACIR bytes from []int -> []byte (range check)
	acirBytes, err := intsToBytes(w.AcirBytes)
	if err != nil {
		return nil, fmt.Errorf("convert acir_bytes: %w", err)
	}

	// Build ordered witness indices in the ABI parameter order
	/*ordered, err := orderedParamWitnessIndices(abi.Parameters, w.PrivateParamWitnesses, w.PublicParamWitnesses)
	if err != nil {
		return nil, err
	}
	abi.ParameterWitnessIndices = ordered
	*/

	// Build Compilation
	comp := &result.Compilation{
		ACIR:                  result.ACIR{acirBytes, w.AcirJSON},
		ABI:                   abi,
		NoirVersion:           w.NoirVersion,
		Hash:                  w.Hash,
		PrivateParamWitnesses: append([]uint32(nil), w.PrivateParamWitnesses...),
		PublicParamWitnesses:  append([]uint32(nil), w.PublicParamWitnesses...),
	}
	//fmt.Println(acirBytes)
	//fmt.Println(w.PrivateParamWitnesses)

	return comp, nil
}

func intsToBytes(src []int) ([]byte, error) {
	out := make([]byte, len(src))
	for i, v := range src {
		if v < 0 || v > 255 {
			return nil, fmt.Errorf("acir_bytes[%d]=%d out of byte range", i, v)
		}
		out[i] = byte(v)
	}
	return out, nil
}

func orderedParamWitnessIndices(params []result.ABIParameter, priv []uint32, pub []uint32) ([]uint32, error) {
	out := make([]uint32, 0, len(params))

	pi, pu := 0, 0
	for i, p := range params {
		switch p.Visibility {
		case "private":
			if pi >= len(priv) {
				return nil, fmt.Errorf("abi param %d (%s) is private but private witnesses exhausted (need %d, have %d)", i, p.Name, pi+1, len(priv))
			}
			out = append(out, priv[pi])
			pi++
		case "public":
			if pu >= len(pub) {
				return nil, fmt.Errorf("abi param %d (%s) is public but public witnesses exhausted (need %d, have %d)", i, p.Name, pu+1, len(pub))
			}
			out = append(out, pub[pu])
			pu++
		default:
			return nil, fmt.Errorf("abi param %d (%s) has unknown visibility %q", i, p.Name, p.Visibility)
		}
	}

	// Optional sanity checks: did we consume exactly all of them?
	// You can relax these if Noir sometimes includes extras (it *shouldn't* for params).
	if pi != len(priv) {
		return nil, fmt.Errorf("private witnesses mismatch: abi consumed %d but have %d", pi, len(priv))
	}
	if pu != len(pub) {
		return nil, fmt.Errorf("public witnesses mismatch: abi consumed %d but have %d", pu, len(pub))
	}

	return out, nil
}
