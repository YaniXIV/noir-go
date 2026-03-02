package types

import (
	"fmt"
	"math/big"
)

// Modulus is the base field modulus used by Noir.
var Modulus, _ = new(big.Int).SetString(
	"21888242871839275222246405745257275088548364400416034343698204186575808495617",
	10,
)

// Field represents an element of the Noir base field.
type Field struct {
	n *big.Int
}

// NewField constructs a Field element, reducing x modulo Modulus.
// Returns an error if x is negative.
func NewField(x *big.Int) (*Field, error) {
	if x.Sign() < 0 {
		return nil, fmt.Errorf("field element cannot be negative")
	}

	// reduce mod p
	n := new(big.Int).Mod(x, Modulus)

	return &Field{n: n}, nil
}

// BigInt returns a copy of the underlying value as a big.Int.
func (f *Field) BigInt() *big.Int {
	return new(big.Int).Set(f.n)
}

// String returns the decimal representation of the field element.
func (f *Field) String() string {
	return f.n.String()
}
