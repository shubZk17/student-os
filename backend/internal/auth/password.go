package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

var DefaultArgon2Params = &Argon2Params{
	Memory:      64 * 1024, // 64 MB
	Iterations:  3,
	Parallelism: 2,
	SaltLength:  16,
	KeyLength:   32,
}

// Each Argon2 hash allocates Params.Memory (64 MB); cap concurrency so a login burst can't exhaust RAM.
var hashSlots = make(chan struct{}, 4)

func argon2Key(password, salt []byte, time, memory uint32, threads uint8, keyLen uint32) []byte {
	hashSlots <- struct{}{}
	defer func() { <-hashSlots }()
	return argon2.IDKey(password, salt, time, memory, threads, keyLen)
}

// HashPassword hashes a plain password using Argon2id
func HashPassword(password string) (string, error) {
	salt := make([]byte, DefaultArgon2Params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2Key(
		[]byte(password),
		salt,
		DefaultArgon2Params.Iterations,
		DefaultArgon2Params.Memory,
		DefaultArgon2Params.Parallelism,
		DefaultArgon2Params.KeyLength,
	)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		DefaultArgon2Params.Memory,
		DefaultArgon2Params.Iterations,
		DefaultArgon2Params.Parallelism,
		b64Salt,
		b64Hash,
	)

	return encodedHash, nil
}

// CheckPasswordHash compares a hashed password against a plain password.
// Supports both Argon2id and fallback Bcrypt.
func CheckPasswordHash(password, encodedHash string) (bool, error) {
	if strings.HasPrefix(encodedHash, "$argon2id$") {
		parts := strings.Split(encodedHash, "$")
		if len(parts) < 6 {
			return false, errors.New("invalid argon2 hash format")
		}

		var version int
		_, err := fmt.Sscanf(parts[2], "v=%d", &version)
		if err != nil {
			return false, err
		}

		var memory, iterations uint32
		var parallelism uint8
		_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
		if err != nil {
			return false, err
		}

		salt, err := base64.RawStdEncoding.DecodeString(parts[4])
		if err != nil {
			return false, err
		}

		decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
		if err != nil {
			return false, err
		}

		comparisonHash := argon2Key(
			[]byte(password),
			salt,
			iterations,
			memory,
			parallelism,
			uint32(len(decodedHash)),
		)

		if subtle.ConstantTimeCompare(decodedHash, comparisonHash) == 1 {
			return true, nil
		}
		return false, nil
	}

	// Fallback to bcrypt
	err := bcrypt.CompareHashAndPassword([]byte(encodedHash), []byte(password))
	if err == nil {
		return true, nil
	}
	return false, nil
}
