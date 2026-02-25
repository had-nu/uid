package genesis

import (
	"context"
	"testing"

	"github.com/had-nu/uid/internal/biometric"
	"github.com/had-nu/uid/internal/entropy"
)

func TestMerkleTree(t *testing.T) {
	leaves := [3]BioLeaf{
		{Hash: []byte("A"), Timestamp: 10},
		{Hash: []byte("B"), Timestamp: 20},
		{Hash: []byte("C"), Timestamp: 30},
	}

	root, proofs, err := BuildMerkleTree(leaves)
	if err != nil {
		t.Fatal(err)
	}

	for i, leaf := range leaves {
		if !VerifyProof(leaf.Hash, proofs[i], root) {
			t.Fatalf("verification failed for leaf %d", i)
		}
	}

	// Test invalid verification
	if VerifyProof([]byte("X"), proofs[0], root) {
		t.Fatal("verification succeeded for incorrect leaf hash")
	}
}

func TestTriadMutualSign(t *testing.T) {
	triad, err := NewLocalTriad(150)
	if err != nil {
		t.Fatal(err)
	}
	defer triad.Destroy()

	ctx := TriadContext{
		FEntropy:      []byte("fent"),
		GenesisHash:   []byte("gen"),
		ReputationSum: 150,
		SiderealTime:  "SIMULATED",
		CycleIndex:    0,
	}

	sR, sU, sA, err := triad.Sign(context.Background(), ctx)
	if err != nil {
		t.Fatalf("triple digest failed: %v", err)
	}

	if len(sR) == 0 || len(sU) == 0 || len(sA) == 0 {
		t.Fatal("empty signatures")
	}

	// test low reputation
	triadLow, _ := NewLocalTriad(100)
	_, _, _, err = triadLow.Sign(context.Background(), ctx)
	if err == nil {
		t.Fatal("expected failure due to low reputation")
	}
}

func TestCeremonyHappyPath(t *testing.T) {
	triad, _ := NewLocalTriad(150)
	defer triad.Destroy()

	cfg := CeremonyConfig{
		MinEntropy:    50.0,
		MinReputation: 150,
		AIFilterMin:   0.80,
		// Huge delta T max so simulated code delay doesn't fail
		DeltaTMax: 5000000000,
		Simulate:  true,
	}

	cap := biometric.SimulatedCapture{FounderID: "Test", Seed: []byte("seed")}
	captures := [3]biometric.BiometricCapture{cap, cap, cap}

	res, err := RunCeremony(
		context.Background(),
		cfg,
		entropy.SimulatedCosmicEntropy{Seed: 42},
		captures,
		triad,
	)
	if err != nil {
		t.Fatalf("ceremony failed: %v", err)
	}

	if len(res.RootID) == 0 {
		t.Fatal("RootID not generated")
	}
}
