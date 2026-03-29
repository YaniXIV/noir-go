package witness

import (
	"fmt"
	"math/big"
)

// SolvedWitness holds the fully solved witness map from execution.
// Keys are Noir witness indices, values are 32-byte big-endian field elements.
type SolvedWitness map[uint32][32]byte

// ToBigInt converts a single witness value to *big.Int.
func (s SolvedWitness) ToBigInt(idx uint32) (*big.Int, error) {
	val, ok := s[idx]
	if !ok {
		return nil, fmt.Errorf("witness index %d not found in solved witness", idx)
	}
	return new(big.Int).SetBytes(val[:]), nil
}

// ToGnarkValues returns a map of Noir witness index → *big.Int for all entries.
func (s SolvedWitness) ToGnarkValues() map[uint32]*big.Int {
	out := make(map[uint32]*big.Int, len(s))
	for idx, val := range s {
		v := val
		out[idx] = new(big.Int).SetBytes(v[:])
	}
	return out
}
