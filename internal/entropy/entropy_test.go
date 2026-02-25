package entropy

import (
	"bytes"
	"context"
	"testing"
)

func TestSymbolicEntropy(t *testing.T) {
	s1 := SymbolicEntropy()
	s2 := SymbolicEntropy()
	if !bytes.Equal(s1, s2) {
		t.Fatal("Symbolic entropy must be deterministic")
	}
}

func TestCosmicEntropySimulated(t *testing.T) {
	sim1 := SimulatedCosmicEntropy{Seed: 42}
	sim2 := SimulatedCosmicEntropy{Seed: 42}
	sim3 := SimulatedCosmicEntropy{Seed: 99}

	c1, _, _ := sim1.CosmicEntropy(context.Background())
	c2, _, _ := sim2.CosmicEntropy(context.Background())
	c3, _, _ := sim3.CosmicEntropy(context.Background())

	if !bytes.Equal(c1, c2) {
		t.Fatal("Simulated cosmic entropy must be deterministic for the same seed")
	}
	if bytes.Equal(c1, c3) {
		t.Fatal("Simulated cosmic entropy must differ for different seeds")
	}
}

func TestFundamentalEntropy(t *testing.T) {
	sim := SimulatedCosmicEntropy{Seed: 12345}
	cEntropy, _, _ := sim.CosmicEntropy(context.Background())
	sEntropy := SymbolicEntropy()

	fEntropy, bits, err := FundamentalEntropy(cEntropy, sEntropy)
	if err != nil {
		t.Fatalf("fundamental entropy failed: %v", err)
	}

	if bits < 50.0 {
		t.Fatalf("expected >= 50 bits of entropy, got %f", bits)
	}

	if len(fEntropy) != 32 { // Kyber shared secret size
		t.Fatalf("expected 32 bytes of output, got %d", len(fEntropy))
	}
}

func TestShannonEntropy(t *testing.T) {
	// Constant string should have ~0 entropy
	constBytes := []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	bits := ShannonEntropy(constBytes)
	if bits > 5.0 {
		t.Fatalf("constant byte string should have near zero entropy, got %f", bits)
	}
}
