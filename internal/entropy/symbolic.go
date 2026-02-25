package entropy

import (
	"github.com/had-nu/uid/internal/crypto"
)

// SymbolicText is the foundational text used to generate the static entropy layer.
const SymbolicText = "A luz da consciência precede o tempo. Na quietude do vazio, a primeira vibração..."

// SymbolicEntropy generates s_entropy using BLAKE3 of the foundational text.
func SymbolicEntropy() []byte {
	return crypto.Hash([]byte(SymbolicText))
}
