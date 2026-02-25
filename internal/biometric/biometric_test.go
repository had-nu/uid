package biometric

import (
	"bytes"
	"context"
	"testing"
)

func TestBioHashDeterministic(t *testing.T) {
	capture1 := BioCapture{
		FingerprintEmbedding: []byte("fp"),
		VoiceprintEmbedding:  []byte("voice"),
		VascularPattern:      []byte("vp"),
		LivenessNonce:        []byte("nonce"),
		LivenessSig:          []byte("sig"),
		DeltaT:               1000,
		AIFilterScore:        0.95,
		CaptureTimestamp:     123456,
	}

	// Changing timestamp should not affect biohash
	capture2 := capture1
	capture2.CaptureTimestamp = 999999

	hash1, _ := BioHash(capture1)
	hash2, _ := BioHash(capture2)

	if !bytes.Equal(hash1, hash2) {
		t.Fatal("bio_hash is not deterministic for the same underlying vectors")
	}

	capture3 := capture1
	capture3.AIFilterScore = 0.82
	hash3, _ := BioHash(capture3)

	if bytes.Equal(hash1, hash3) {
		t.Fatal("bio_hash collision with different AIFilterScore")
	}
}

func TestSimulatedCapture(t *testing.T) {
	sim := SimulatedCapture{FounderID: "A", Seed: []byte("seed")}
	ctx := context.Background()
	nonce := []byte("nonce1")

	fp, _ := sim.CaptureFingerprint(ctx, nonce)
	vo, _ := sim.CaptureVoice(ctx, nonce)
	va, _ := sim.CaptureVascular(ctx, nonce)

	if bytes.Equal(fp, vo) || bytes.Equal(vo, va) {
		t.Fatal("simulated captures must differ per modality")
	}
}
