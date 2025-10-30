package api

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"
)

const passwordSaltLength = 16

// hashPassword generates a salted SHA-256 hash for the supplied password.
func hashPassword(password string) (string, error) {
	salt := make([]byte, passwordSaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}
	sum := sha256.Sum256(combineSaltAndPassword(salt, password))
	return base64.StdEncoding.EncodeToString(salt) + ":" + base64.StdEncoding.EncodeToString(sum[:]), nil
}

// verifyPassword checks whether the provided password matches the stored hash.
func verifyPassword(hashed, password string) bool {
	parts := strings.Split(hashed, ":")
	if len(parts) != 2 {
		return false
	}
	salt, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	digest, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	sum := sha256.Sum256(combineSaltAndPassword(salt, password))
	return subtle.ConstantTimeCompare(digest, sum[:]) == 1
}

func combineSaltAndPassword(salt []byte, password string) []byte {
	buf := make([]byte, len(salt)+len(password))
	copy(buf, salt)
	copy(buf[len(salt):], password)
	return buf
}

// HashPasswordForTests exposes hashing for unit tests.
func HashPasswordForTests(password string) (string, error) {
	return hashPassword(password)
}

// VerifyPasswordForTests exposes password verification for unit tests.
func VerifyPasswordForTests(hashed, password string) bool {
	return verifyPassword(hashed, password)
}
