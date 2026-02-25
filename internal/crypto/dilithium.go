package crypto

import (
	"crypto/subtle"
	"io"
	"runtime"

	"github.com/cloudflare/circl/sign/dilithium/mode3"
)

// GenerateDilithiumKey generates a new Dilithium3 keypair.
func GenerateDilithiumKey(rand io.Reader) (publicKey []byte, secretKey []byte, err error) {
	pk, sk, err := mode3.GenerateKey(rand)
	if err != nil {
		return nil, nil, err
	}
	// Pack into bytes
	return pk.Bytes(), sk.Bytes(), nil
}

// SignDilithium signs a message using a Dilithium3 secret key.
func SignDilithium(skBytes []byte, msg []byte) []byte {
	sk := new(mode3.PrivateKey)
	if err := sk.UnmarshalBinary(skBytes); err != nil {
		return nil
	}
	sig := make([]byte, mode3.SignatureSize)
	mode3.SignTo(sk, msg, sig)
	return sig
}

// VerifyDilithium verifies a Dilithium3 signature.
func VerifyDilithium(pkBytes []byte, msg []byte, sig []byte) bool {
	pk := new(mode3.PublicKey)
	if err := pk.UnmarshalBinary(pkBytes); err != nil {
		return false
	}
	return mode3.Verify(pk, msg, sig)
}

// WipeSecret explicitly zeroes out the memory of a secret key.
func WipeSecret(sk []byte) {
	if len(sk) == 0 {
		return
	}
	// Overwrite with zeros
	zeros := make([]byte, len(sk))
	subtle.ConstantTimeCopy(1, sk, zeros)
	runtime.KeepAlive(sk)
}
