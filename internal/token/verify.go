package token

import (
	"bytes"
	"fmt"

	"github.com/had-nu/uid/internal/crypto"
	"github.com/had-nu/uid/internal/entropy"
	"github.com/had-nu/uid/internal/genesis"
)

type VerifyOptions struct {
	PublicKeys      [3][]byte // Recovery, Update, Audit
	RejectSimulated bool
	Strict          bool // Additional validity checks
}

type VerificationResult struct {
	Valid       bool
	FailedField string
	Err         error
}

// Verify performs isolated verification of the token using external public keys.
func Verify(u *UIDZeroSoulbound, opts VerifyOptions) VerificationResult {
	// 1. FinalDigest
	recalcDigest, err := u.CalculateFinalDigest()
	if err != nil {
		return VerificationResult{false, "FinalDigest", fmt.Errorf("failed to recalculate FinalDigest: %w", err)}
	}
	if !bytes.Equal(recalcDigest, u.FinalDigest) {
		return VerificationResult{false, "FinalDigest", fmt.Errorf("token tampered: FinalDigest mismatch")}
	}

	// 2. Simulated Flag
	if opts.RejectSimulated && u.Simulated {
		return VerificationResult{false, "Simulated", fmt.Errorf("simulated token rejected in production mode")}
	}

	// 3. CycleIndex
	if u.CycleIndex != 0 {
		return VerificationResult{false, "CycleIndex", fmt.Errorf("CycleIndex %d != 0", u.CycleIndex)}
	}

	// 4. Reputation
	if u.ReputationSum < 150 {
		return VerificationResult{false, "ReputationSum", fmt.Errorf("reputation too low: %d", u.ReputationSum)}
	}

	// 5. Entropy
	bits := entropy.ShannonEntropy(u.FEntropy)
	if bits < 50.0 {
		return VerificationResult{false, "FEntropy", fmt.Errorf("insufficient entropy: %f bits", bits)}
	}

	// 6. Triad Mutual Signatures
	tCtx := genesis.TriadContext{
		FEntropy:      u.FEntropy,
		GenesisHash:   u.GenesisHash,
		ReputationSum: u.ReputationSum,
		SiderealTime:  u.SiderealTime,
		CycleIndex:    u.CycleIndex,
	}
	msg := tCtx.Bytes()

	if !crypto.VerifyDilithium(opts.PublicKeys[0], msg, u.SigRecovery) {
		return VerificationResult{false, "SigRecovery", fmt.Errorf("invalid Recovery signature")}
	}
	if !crypto.VerifyDilithium(opts.PublicKeys[1], msg, u.SigUpdate) {
		return VerificationResult{false, "SigUpdate", fmt.Errorf("invalid Update signature")}
	}
	if !crypto.VerifyDilithium(opts.PublicKeys[2], msg, u.SigAudit) {
		return VerificationResult{false, "SigAudit", fmt.Errorf("invalid Audit signature")}
	}

	// 7. RootID check
	recalcRootID := crypto.KyberDigest(
		msg,
		u.SigRecovery,
		u.SigUpdate,
		u.SigAudit,
	)
	if !bytes.Equal(recalcRootID, u.RootID) {
		return VerificationResult{false, "RootID", fmt.Errorf("RootID derivation mismatch")}
	}

	return VerificationResult{Valid: true}
}
