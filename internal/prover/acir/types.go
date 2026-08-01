// Package acir decodes the ACIR JSON produced by the noir-go compiler bridge
// into Go types suitable for building a gnark constraint system.
//
// Only AssertZero opcodes carry algebraic constraints; everything else
// (e.g. BrilligCall) is either a witness-generation hint with no constraint
// of its own, or unsupported and must be rejected explicitly rather than
// silently dropped.
package acir

import (
	"encoding/json"
	"fmt"
	"math/big"
)

type Program struct {
	Functions              []Function        `json:"functions"`
	UnconstrainedFunctions []json.RawMessage `json:"unconstrained_functions"`
}

type Function struct {
	FunctionName        string            `json:"function_name"`
	CurrentWitnessIndex uint32            `json:"current_witness_index"`
	Opcodes             []Opcode          `json:"opcodes"`
	PrivateParameters   []uint32          `json:"private_parameters"`
	PublicParameters    []uint32          `json:"public_parameters"`
	ReturnValues        []uint32          `json:"return_values"`
	AssertMessages      []json.RawMessage `json:"assert_messages"`
}

// Opcode is a tagged union over ACIR opcode kinds. Kind holds the single
// top-level JSON key of the opcode object (e.g. "AssertZero", "BrilligCall",
// "BlackBoxFuncCall", "MemoryOp") so callers can tell "no constraint, safe
// to skip" apart from "unrecognized, must fail loudly".
type Opcode struct {
	Kind       string
	AssertZero *AssertZero
}

func (o *Opcode) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("opcode: %w", err)
	}
	if len(raw) != 1 {
		return fmt.Errorf("opcode: expected exactly one variant key, got %d", len(raw))
	}
	for k, v := range raw {
		o.Kind = k
		if k == "AssertZero" {
			var az AssertZero
			if err := json.Unmarshal(v, &az); err != nil {
				return fmt.Errorf("opcode AssertZero: %w", err)
			}
			o.AssertZero = &az
		}
	}
	return nil
}

type AssertZero struct {
	MulTerms           []MulTerm           `json:"mul_terms"`
	LinearCombinations []LinearCombination `json:"linear_combinations"`
	QC                 FieldElement        `json:"q_c"`
}

// FieldElement is a 32-byte big-endian field element. encoding/json decodes
// a JSON array of integers into a fixed-size byte array element-by-element,
// so no custom UnmarshalJSON is needed.
type FieldElement [32]byte

// BigInt returns the field element as a big-endian unsigned integer.
func (f FieldElement) BigInt() *big.Int {
	return new(big.Int).SetBytes(f[:])
}

type MulTerm struct {
	Coeff FieldElement
	LHS   uint32
	RHS   uint32
}

func (m *MulTerm) UnmarshalJSON(data []byte) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw) != 3 {
		return fmt.Errorf("MulTerm: expected 3 elements, got %d", len(raw))
	}
	if err := json.Unmarshal(raw[0], &m.Coeff); err != nil {
		return fmt.Errorf("MulTerm coeff: %w", err)
	}
	if err := json.Unmarshal(raw[1], &m.LHS); err != nil {
		return fmt.Errorf("MulTerm lhs: %w", err)
	}
	if err := json.Unmarshal(raw[2], &m.RHS); err != nil {
		return fmt.Errorf("MulTerm rhs: %w", err)
	}
	return nil
}

type LinearCombination struct {
	Coeff   FieldElement
	Witness uint32
}

func (l *LinearCombination) UnmarshalJSON(data []byte) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw) != 2 {
		return fmt.Errorf("LinearCombination: expected 2 elements, got %d", len(raw))
	}
	if err := json.Unmarshal(raw[0], &l.Coeff); err != nil {
		return fmt.Errorf("LinearCombination coeff: %w", err)
	}
	if err := json.Unmarshal(raw[1], &l.Witness); err != nil {
		return fmt.Errorf("LinearCombination witness: %w", err)
	}
	return nil
}
