package biometric

import (
	"context"

	"github.com/had-nu/uid/internal/crypto"
)

// SimulatedCapture provides deterministic stubs for testing and simulation mode.
type SimulatedCapture struct {
	FounderID string
	Seed      []byte
}

func (s SimulatedCapture) CaptureFingerprint(ctx context.Context, nonce []byte) ([]byte, error) {
	return crypto.Hash([]byte("fingerprint"), []byte(s.FounderID), s.Seed, nonce), nil
}

func (s SimulatedCapture) CaptureVoice(ctx context.Context, nonce []byte) ([]byte, error) {
	return crypto.Hash([]byte("voice"), []byte(s.FounderID), s.Seed, nonce), nil
}

func (s SimulatedCapture) CaptureVascular(ctx context.Context, nonce []byte) ([]byte, error) {
	return crypto.Hash([]byte("vascular"), []byte(s.FounderID), s.Seed, nonce), nil
}

func (s SimulatedCapture) AIFilterScore(capture BioCapture) (float64, error) {
	// Always passes with flying colors in simulation.
	return 0.99, nil
}
