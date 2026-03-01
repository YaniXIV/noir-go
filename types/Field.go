package types

import (
	"fmt"
	"math/big"
)

var Modulus, _ = new(big.Int).SetString(
	"21888242871839275222246405745257275088548364400416034343698204186575808495617",
	10,
)

type Field struct {
	n *big.Int
}

func NewField(x *big.Int) (*Field, error) {
	if x.Sign() < 0 {
		return nil, fmt.Errorf("field element cannot be negative")
	}

	// reduce mod p
	n := new(big.Int).Mod(x, Modulus)

	return &Field{n: n}, nil
}

func (f *Field) BigInt() *big.Int {
	return new(big.Int).Set(f.n)
}

func (f *Field) String() string {
	return f.n.String()
}
