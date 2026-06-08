package cmd

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// generateCodeVerifier generates a code_verifier compliant with RFC 7636.
// Uses crypto/rand to generate 32 bytes of random data, then encodes them
// with base64url (no padding). The result is 43 characters long (RFC 7636
// requires 43~128 characters), containing only unreserved URI characters
// [A-Za-z0-9\-._~].
func generateCodeVerifier() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate code_verifier failed: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// generateCodeChallenge computes the S256 code_challenge for a given code_verifier.
// Algorithm: BASE64URL(SHA256(code_verifier)), without padding characters.
func generateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// generateState generates a random UUID v4 string for use as the OAuth state parameter.
func generateState() (string, error) {
	var uuid [16]byte
	if _, err := rand.Read(uuid[:]); err != nil {
		return "", fmt.Errorf("generate state failed: %w", err)
	}

	// Set UUID v4 version and variant bits.
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // variant 10

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4],
		uuid[4:6],
		uuid[6:8],
		uuid[8:10],
		uuid[10:16],
	), nil
}
