package token

import (
	"fmt"

	"github.com/had-nu/uid/internal/genesis"
)

// UIDZeroSoulbound is the NTT token of the Prana Network foundational identity.
type UIDZeroSoulbound struct {
	RootID              []byte   `cbor:"0,keyasint"`
	FEntropy            []byte   `cbor:"1,keyasint"`
	GenesisHash         []byte   `cbor:"2,keyasint"`
	FounderMerkleProofs [][]byte `cbor:"3,keyasint"`
	SigRecovery         []byte   `cbor:"4,keyasint"`
	SigUpdate           []byte   `cbor:"5,keyasint"`
	SigAudit            []byte   `cbor:"6,keyasint"`
	ReputationSum       uint64   `cbor:"7,keyasint"`
	SiderealTime        string   `cbor:"8,keyasint"`
	CycleIndex          uint64   `cbor:"9,keyasint"`
	GeneratedAt         int64    `cbor:"10,keyasint"`
	MerkleRoot          []byte   `cbor:"11,keyasint,omitempty"`
	MerkleProof         [][]byte `cbor:"12,keyasint,omitempty"`
	Simulated           bool     `cbor:"13,keyasint"`
	FinalDigest         []byte   `cbor:"14,keyasint"`
}

// NewFromCeremony converts a CeremonyResult into a UIDZeroSoulbound without FinalDigest.
func NewFromCeremony(res *genesis.CeremonyResult) (*UIDZeroSoulbound, error) {
	if res == nil {
		return nil, fmt.Errorf("nil ceremony result")
	}

	uid0 := &UIDZeroSoulbound{
		RootID:              res.RootID,
		FEntropy:            res.FEntropy,
		GenesisHash:         res.GenesisHash,
		FounderMerkleProofs: res.FounderProofs[:],
		SigRecovery:         res.TriadSigs[0],
		SigUpdate:           res.TriadSigs[1],
		SigAudit:            res.TriadSigs[2],
		ReputationSum:       res.ReputationSum,
		SiderealTime:        res.SiderealTime,
		CycleIndex:          0,
		GeneratedAt:         res.GeneratedAt,
		Simulated:           res.Simulated,
	}

	return uid0, nil
}
