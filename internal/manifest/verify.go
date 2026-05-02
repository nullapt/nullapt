package manifest

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
)

// Verify checks the manifest's embedded Ed25519 signature against its payload.
// Returns nil on success, a descriptive error on any failure.
func (s *Skill) Verify() error {
	if s.Signature.Algorithm != "ed25519" {
		return fmt.Errorf("unsupported signature algorithm %q; only ed25519 is accepted", s.Signature.Algorithm)
	}

	pubKeyBytes, err := base64.StdEncoding.DecodeString(s.Signature.PublicKey)
	if err != nil {
		return fmt.Errorf("decoding public key: %w", err)
	}
	if len(pubKeyBytes) != ed25519.PublicKeySize {
		return fmt.Errorf("public key has wrong length (%d bytes, want %d)", len(pubKeyBytes), ed25519.PublicKeySize)
	}

	sigBytes, err := base64.StdEncoding.DecodeString(s.Signature.Value)
	if err != nil {
		return fmt.Errorf("decoding signature: %w", err)
	}

	payload, err := s.PayloadBytes()
	if err != nil {
		return fmt.Errorf("serialising payload: %w", err)
	}

	if !ed25519.Verify(ed25519.PublicKey(pubKeyBytes), payload, sigBytes) {
		return fmt.Errorf("signature verification failed: manifest may be tampered")
	}
	return nil
}

// VerifyAgainstTrustedKey is the strict path used during `nullapt get`.
// It verifies the manifest signature AND checks the public key against
// a known-good key pinned in the transparency log (future: registry call).
// For now it delegates to the embedded-key path and notes the log check is a stub.
func (s *Skill) VerifyAgainstTrustedKey(trustedPubKey ed25519.PublicKey) error {
	pubKeyBytes, err := base64.StdEncoding.DecodeString(s.Signature.PublicKey)
	if err != nil {
		return fmt.Errorf("decoding embedded public key: %w", err)
	}
	if !trustedPubKey.Equal(ed25519.PublicKey(pubKeyBytes)) {
		return fmt.Errorf("public key mismatch: manifest key does not match the transparency log entry for %q", s.Author)
	}
	return s.Verify()
}
