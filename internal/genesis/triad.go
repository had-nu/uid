package genesis

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"

	"github.com/had-nu/uid/internal/crypto"
)

type TriadMember struct {
	Name      string
	PublicKey []byte
	secretKey []byte
}

type TriadContext struct {
	FEntropy      []byte
	GenesisHash   []byte
	ReputationSum uint64
	SiderealTime  string
	CycleIndex    uint64 // sempre 0
}

// Bytes serializes the TriadContext for signing.
func (tc TriadContext) Bytes() []byte {
	buf := bytes.NewBuffer(nil)
	buf.Write(tc.FEntropy)
	buf.Write(tc.GenesisHash)

	repBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(repBytes, tc.ReputationSum)
	buf.Write(repBytes)

	buf.WriteString(tc.SiderealTime)

	idxBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(idxBytes, tc.CycleIndex)
	buf.Write(idxBytes)

	return buf.Bytes()
}

// TriadSigner abstracts the Algorithmic Triad components.
type TriadSigner interface {
	Members() [3]TriadMember
	ReputationSum() uint64
	Sign(ctx context.Context, tc TriadContext) (sigR, sigU, sigA []byte, err error)
	Destroy()
}

// LocalTriad is the default local implementation using Dilithium3.
type LocalTriad struct {
	members    [3]TriadMember
	reputation uint64
}

// NewLocalTriad initializes a triad with ephemeral keys.
func NewLocalTriad(reputation uint64) (*LocalTriad, error) {
	names := []string{"Recovery", "Update", "Audit"}
	triad := &LocalTriad{reputation: reputation}

	for i, name := range names {
		pk, sk, err := crypto.GenerateDilithiumKey(rand.Reader)
		if err != nil {
			return nil, err
		}
		triad.members[i] = TriadMember{Name: name, PublicKey: pk, secretKey: sk}
	}
	return triad, nil
}

func (lt *LocalTriad) Members() [3]TriadMember {
	return lt.members
}

func (lt *LocalTriad) ReputationSum() uint64 {
	return lt.reputation
}

func (lt *LocalTriad) Sign(ctx context.Context, tc TriadContext) ([]byte, []byte, []byte, error) {
	if lt.reputation < 150 {
		return nil, nil, nil, errors.New("reputation_sum < 150")
	}
	if tc.CycleIndex != 0 {
		return nil, nil, nil, errors.New("cycle_index must be 0")
	}

	msg := tc.Bytes()

	// Each member signs the context
	sig0 := crypto.SignDilithium(lt.members[0].secretKey, msg)
	sig1 := crypto.SignDilithium(lt.members[1].secretKey, msg)
	sig2 := crypto.SignDilithium(lt.members[2].secretKey, msg)

	if sig0 == nil || sig1 == nil || sig2 == nil {
		return nil, nil, nil, errors.New("failed to generate signatures")
	}

	// Mutual verification cross-check
	if !crypto.VerifyDilithium(lt.members[0].PublicKey, msg, sig0) ||
		!crypto.VerifyDilithium(lt.members[1].PublicKey, msg, sig1) ||
		!crypto.VerifyDilithium(lt.members[2].PublicKey, msg, sig2) {
		return nil, nil, nil, errors.New("mutual verification failed")
	}

	return sig0, sig1, sig2, nil
}

func (lt *LocalTriad) Destroy() {
	for i := range lt.members {
		if lt.members[i].secretKey != nil {
			crypto.WipeSecret(lt.members[i].secretKey)
			lt.members[i].secretKey = nil
		}
	}
}
