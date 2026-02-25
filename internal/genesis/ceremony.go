package genesis

import (
	"context"
	"fmt"
	"time"

	"github.com/had-nu/uid/internal/biometric"
	"github.com/had-nu/uid/internal/crypto"
	"github.com/had-nu/uid/internal/entropy"
)

// CeremonyConfig holds protocol thresholds.
type CeremonyConfig struct {
	MinEntropy    float64
	MinReputation uint64
	AIFilterMin   float64
	DeltaTMax     int64
	Simulate      bool
}

// CeremonyResult holds the outputs of a successful ceremony.
type CeremonyResult struct {
	FEntropy      []byte
	GenesisHash   []byte
	FounderProofs [3][]byte
	TriadSigs     [3][]byte
	Founders      [3]biometric.BioCapture
	RootID        []byte
	SiderealTime  string
	ReputationSum uint64
	GeneratedAt   int64
	Simulated     bool
	TriadMembers  [3]TriadMember
}

// RunCeremony executes the Genesis state machine.
func RunCeremony(
	ctx context.Context,
	cfg CeremonyConfig,
	entropySrc entropy.EntropySource,
	captures [3]biometric.BiometricCapture,
	triad TriadSigner,
) (*CeremonyResult, error) {
	// [1/5] Check Preconditions
	if triad.ReputationSum() < cfg.MinReputation {
		return nil, fmt.Errorf("precondition failed: triad reputation %d < %d", triad.ReputationSum(), cfg.MinReputation)
	}

	cEntropy, stStr, err := entropySrc.CosmicEntropy(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cosmic entropy: %w", err)
	}

	fEntropy, bits, err := entropy.FundamentalEntropy(cEntropy, entropy.SymbolicEntropy())
	if err != nil {
		return nil, fmt.Errorf("fundamental entropy failed: %w", err)
	}
	if bits < cfg.MinEntropy {
		return nil, fmt.Errorf("precondition failed: entropy %f bits < %f", bits, cfg.MinEntropy)
	}

	// [2/5] Biometric Ceremony
	var bioCaptures [3]biometric.BioCapture
	var bioLeaves [3]BioLeaf

	for i, capSrc := range captures {
		start := time.Now()
		nonce, err := biometric.GenerateLivenessNonce()
		if err != nil {
			return nil, err
		}

		fp, err := capSrc.CaptureFingerprint(ctx, nonce)
		if err != nil {
			return nil, err
		}
		vp, err := capSrc.CaptureVoice(ctx, nonce)
		if err != nil {
			return nil, err
		}
		va, err := capSrc.CaptureVascular(ctx, nonce)
		if err != nil {
			return nil, err
		}

		deltaT := time.Since(start).Nanoseconds()
		if deltaT > cfg.DeltaTMax {
			return nil, fmt.Errorf("founder %d: delta_t %d > %d", i, deltaT, cfg.DeltaTMax)
		}

		capture := biometric.BioCapture{
			FingerprintEmbedding: fp,
			VoiceprintEmbedding:  vp,
			VascularPattern:      va,
			LivenessNonce:        nonce,
			LivenessSig:          []byte("simulated_sig"), // Placeholder, signature logic usually from device
			DeltaT:               deltaT,
			CaptureTimestamp:     time.Now().UnixNano(),
		}

		score, err := capSrc.AIFilterScore(capture)
		if err != nil {
			return nil, err
		}
		capture.AIFilterScore = score

		if score < cfg.AIFilterMin {
			return nil, fmt.Errorf("founder %d: ai_score %f < %f", i, score, cfg.AIFilterMin)
		}

		bioCaptures[i] = capture

		hash, err := biometric.BioHash(capture)
		if err != nil {
			return nil, fmt.Errorf("failed to hash founder %d: %w", i, err)
		}
		bioLeaves[i] = BioLeaf{Hash: hash, Timestamp: capture.CaptureTimestamp}
	}

	genesisHash, founderProofs, err := BuildMerkleTree(bioLeaves)
	if err != nil {
		return nil, fmt.Errorf("failed to build merkle tree: %w", err)
	}

	// [3/5] Triple Digest Mutual
	tCtx := TriadContext{
		FEntropy:      fEntropy,
		GenesisHash:   genesisHash,
		ReputationSum: triad.ReputationSum(),
		SiderealTime:  stStr,
		CycleIndex:    0,
	}

	sigR, sigU, sigA, err := triad.Sign(ctx, tCtx)
	if err != nil {
		return nil, fmt.Errorf("triple digest mutual failed: %w", err)
	}

	// [4/5] RootID logic
	rootID := crypto.KyberDigest(
		tCtx.Bytes(),
		sigR,
		sigU,
		sigA,
	)

	return &CeremonyResult{
		FEntropy:      fEntropy,
		GenesisHash:   genesisHash,
		FounderProofs: founderProofs,
		TriadSigs:     [3][]byte{sigR, sigU, sigA},
		Founders:      bioCaptures,
		RootID:        rootID,
		SiderealTime:  stStr,
		ReputationSum: triad.ReputationSum(),
		GeneratedAt:   time.Now().UnixNano(),
		Simulated:     cfg.Simulate,
		TriadMembers:  triad.Members(),
	}, nil
}
