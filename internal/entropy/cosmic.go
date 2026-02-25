package entropy

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/had-nu/uid/internal/crypto"
	"github.com/soniakeys/meeus/v3/julian"
	"github.com/soniakeys/meeus/v3/sidereal"
)

// EntropySource is the interface for obtaining cosmic entropy.
type EntropySource interface {
	CosmicEntropy(ctx context.Context) (data []byte, siderealTime string, err error)
}

// RealCosmicEntropy uses astronomical calculation for sidereal time.
type RealCosmicEntropy struct{}

// CosmicEntropy computes the apparent sidereal time at Greenwich
// and returns a hash of the current observation.
func (r RealCosmicEntropy) CosmicEntropy(ctx context.Context) ([]byte, string, error) {
	jd := julian.TimeToJD(time.Now())
	app := sidereal.Apparent(jd)

	stStr := fmt.Sprintf("%f", app.Rad())

	cEntropy := crypto.Hash([]byte(stStr))
	return cEntropy, stStr, nil
}

// SimulatedCosmicEntropy allows deterministic entropy generation for tests.
type SimulatedCosmicEntropy struct {
	Seed int64
}

// CosmicEntropy ignores actual astronomical calculations and uses a seed.
func (s SimulatedCosmicEntropy) CosmicEntropy(ctx context.Context) ([]byte, string, error) {
	if s.Seed == 0 {
		s.Seed = time.Now().UnixNano()
	}
	stStr := "SIMULATED"
	pid := int64(os.Getpid())
	hostname, _ := os.Hostname()

	cEntropy := crypto.Hash(
		[]byte(fmt.Sprintf("%d", s.Seed)),
		[]byte(hostname),
		[]byte(fmt.Sprintf("%d", pid)),
	)
	return cEntropy, stStr, nil
}
