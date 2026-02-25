package entropy

import (
	"fmt"
	"math"

	"github.com/had-nu/uid/internal/crypto"
)

// FundamentalEntropy calculates f_entropy = Kyber1024(c_entropy || s_entropy)
// and verifies that the measured Shannon entropy is >= 50.0 bits.
func FundamentalEntropy(cosmic []byte, symbolic []byte) (fEntropy []byte, bits float64, err error) {
	fEntropy = crypto.KyberDigest(cosmic, symbolic)
	bits = ShannonEntropy(fEntropy)
	if bits < 50.0 {
		return nil, bits, fmt.Errorf("insufficient entropy: %f bits (minimum 50.0)", bits)
	}
	return fEntropy, bits, nil
}

// ShannonEntropy calculates the theoretical entropy in bits of the slice.
func ShannonEntropy(data []byte) float64 {
	if len(data) == 0 {
		return 0.0
	}
	counts := make(map[byte]float64)
	for _, b := range data {
		counts[b]++
	}

	var entropy float64
	total := float64(len(data))
	for _, count := range counts {
		prob := count / total
		if prob > 0 {
			entropy -= prob * math.Log2(prob)
		}
	}
	// Total bits = entropy component per byte * sequence length
	return entropy * float64(len(data))
}
