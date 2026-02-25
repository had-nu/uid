package genesis

import (
	"bytes"
	"sort"

	"github.com/had-nu/uid/internal/crypto"
)

// BioLeaf represents a founder's biometric hash with its capture timestamp.
type BioLeaf struct {
	Hash      []byte
	Timestamp int64
}

// BuildMerkleTree constructs a 2-level merkle tree for 3 founders.
// It returns the merkle root, and 3 merkle proofs corresponding to the input founders order.
func BuildMerkleTree(founders [3]BioLeaf) (root []byte, proofs [3][]byte, err error) {
	sorted := make([]BioLeaf, 3)
	copy(sorted, founders[:])
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Timestamp == sorted[j].Timestamp {
			return bytes.Compare(sorted[i].Hash, sorted[j].Hash) < 0
		}
		return sorted[i].Timestamp < sorted[j].Timestamp
	})

	leaf0 := crypto.Hash(sorted[0].Hash)
	leaf1 := crypto.Hash(sorted[1].Hash)
	leaf2 := crypto.Hash(sorted[2].Hash)

	node01 := crypto.Hash(leaf0, leaf1)
	node22 := crypto.Hash(leaf2, leaf2)

	root = crypto.Hash(node01, node22)

	for i, orig := range founders {
		var idx int
		for j, s := range sorted {
			if bytes.Equal(orig.Hash, s.Hash) {
				idx = j
				break
			}
		}

		var proof []byte
		if idx == 0 {
			proof = append(proof, 1) // leaf1 is right
			proof = append(proof, leaf1...)
			proof = append(proof, 1) // node22 is right
			proof = append(proof, node22...)
		} else if idx == 1 {
			proof = append(proof, 0) // leaf0 is left
			proof = append(proof, leaf0...)
			proof = append(proof, 1) // node22 is right
			proof = append(proof, node22...)
		} else {
			proof = append(proof, 1) // leaf2 is right (duplicate)
			proof = append(proof, leaf2...)
			proof = append(proof, 0) // node01 is left
			proof = append(proof, node01...)
		}
		proofs[i] = proof
	}

	return root, proofs, nil
}

// VerifyProof verifies a single founder's proof against the root.
func VerifyProof(leafHash []byte, proof []byte, root []byte) bool {
	if len(proof) != 66 {
		return false
	}

	curr := crypto.Hash(leafHash)
	for i := 0; i < 2; i++ {
		dir := proof[i*33]
		sibling := proof[i*33+1 : i*33+33]
		if dir == 0 {
			curr = crypto.Hash(sibling, curr)
		} else {
			curr = crypto.Hash(curr, sibling)
		}
	}
	return bytes.Equal(curr, root)
}
