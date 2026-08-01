package acir

import (
	"encoding/json"
	"fmt"
)

// DecodeAcir parses ACIR JSON (as produced by result.Compilation.ACIR.JSON)
// into a Program.
func DecodeAcir(data []byte) (*Program, error) {
	var program Program
	if err := json.Unmarshal(data, &program); err != nil {
		return nil, fmt.Errorf("decode acir json: %w", err)
	}
	return &program, nil
}
