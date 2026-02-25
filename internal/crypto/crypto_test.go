package crypto

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestBlake3(t *testing.T) {
	d1 := Hash([]byte("prana"), []byte("genesis"))
	d2 := Hash([]byte("prana"), []byte("genesis"))
	if !bytes.Equal(d1, d2) {
		t.Fatal("Hash implementation is not deterministic")
	}

	derived := Derive("prana:genesis:test:v1", d1, 32)
	if len(derived) != 32 {
		t.Fatalf("expected 32 bytes, got %d", len(derived))
	}
}

func TestKyberDigest(t *testing.T) {
	d1 := KyberDigest([]byte("a"), []byte("b"))
	d2 := KyberDigest([]byte("a"), []byte("b"))
	d3 := KyberDigest([]byte("a"), []byte("c"))

	if !bytes.Equal(d1, d2) {
		t.Fatal("KyberDigest is not deterministic for same inputs")
	}
	if bytes.Equal(d1, d3) {
		t.Fatal("KyberDigest produced same output for different inputs")
	}
}

func TestDilithium(t *testing.T) {
	pk, sk, err := GenerateDilithiumKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	msg := []byte("hello prana")
	sig := SignDilithium(sk, msg)

	if !VerifyDilithium(pk, msg, sig) {
		t.Fatal("signature verification failed")
	}

	if VerifyDilithium(pk, []byte("tampered block"), sig) {
		t.Fatal("tampered message passed verification")
	}

	WipeSecret(sk)
	for _, b := range sk {
		if b != 0 {
			t.Fatal("secret key was not wiped properly")
		}
	}
}
