package biometric

import (
	"context"
	"encoding/binary"
	"math"

	"github.com/had-nu/uid/internal/crypto"
)

// BioCapture aggregates raw biometric vectors from a founder.
// The raw data is not persisted, only the bio_hash.
type BioCapture struct {
	FingerprintEmbedding []byte
	VoiceprintEmbedding  []byte
	VascularPattern      []byte
	LivenessNonce        []byte
	LivenessSig          []byte
	DeltaT               int64
	AIFilterScore        float64
	CaptureTimestamp     int64
}

// BiometricCapture abstracts hardware sensors.
type BiometricCapture interface {
	CaptureFingerprint(ctx context.Context, nonce []byte) ([]byte, error)
	CaptureVoice(ctx context.Context, nonce []byte) ([]byte, error)
	CaptureVascular(ctx context.Context, nonce []byte) ([]byte, error)
	AIFilterScore(capture BioCapture) (float64, error)
}

// BioHash computes the final digest of a founder's capture.
func BioHash(c BioCapture) ([]byte, error) {
	deltaTBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(deltaTBytes, uint64(c.DeltaT))

	aiScoreBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(aiScoreBytes, math.Float64bits(c.AIFilterScore))

	return crypto.KyberDigest(
		c.FingerprintEmbedding,
		c.VoiceprintEmbedding,
		c.VascularPattern,
		c.LivenessNonce,
		c.LivenessSig,
		deltaTBytes,
		aiScoreBytes,
	), nil
}
