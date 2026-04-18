package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	defaultArgonMemory      = 64 * 1024
	defaultArgonIterations  = 3
	defaultArgonParallelism = 1
	defaultArgonSaltLength  = 16
	defaultArgonKeyLength   = 32
)

// 密码散列配置单独收敛，方便后续替换参数和扩展校验策略。
type PasswordHasher struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

func DefaultPasswordHasher() PasswordHasher {
	return PasswordHasher{
		Memory:      defaultArgonMemory,
		Iterations:  defaultArgonIterations,
		Parallelism: defaultArgonParallelism,
		SaltLength:  defaultArgonSaltLength,
		KeyLength:   defaultArgonKeyLength,
	}
}

func (h PasswordHasher) Hash(password string) (string, error) {
	config := h.withDefaults()

	salt := make([]byte, config.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("read argon salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		config.Iterations,
		config.Memory,
		config.Parallelism,
		config.KeyLength,
	)

	return fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		config.Memory,
		config.Iterations,
		config.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func (h PasswordHasher) Verify(encodedHash string, password string) (bool, error) {
	config, salt, expectedHash, err := parseArgon2IDHash(encodedHash)
	if err != nil {
		return false, err
	}

	computedHash := argon2.IDKey(
		[]byte(password),
		salt,
		config.Iterations,
		config.Memory,
		config.Parallelism,
		uint32(len(expectedHash)),
	)
	return subtle.ConstantTimeCompare(expectedHash, computedHash) == 1, nil
}

func parseArgon2IDHash(encodedHash string) (PasswordHasher, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
		return PasswordHasher{}, nil, nil, ErrInvalidPasswordHash
	}

	var memory uint64
	var iterations uint64
	var parallelism uint64
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return PasswordHasher{}, nil, nil, ErrInvalidPasswordHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return PasswordHasher{}, nil, nil, ErrInvalidPasswordHash
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return PasswordHasher{}, nil, nil, ErrInvalidPasswordHash
	}

	if len(salt) == 0 || len(hash) == 0 {
		return PasswordHasher{}, nil, nil, ErrInvalidPasswordHash
	}

	return PasswordHasher{
		Memory:      uint32(memory),
		Iterations:  uint32(iterations),
		Parallelism: uint8(parallelism),
		SaltLength:  uint32(len(salt)),
		KeyLength:   uint32(len(hash)),
	}, salt, hash, nil
}

func (h PasswordHasher) withDefaults() PasswordHasher {
	if h.Memory == 0 {
		h.Memory = defaultArgonMemory
	}
	if h.Iterations == 0 {
		h.Iterations = defaultArgonIterations
	}
	if h.Parallelism == 0 {
		h.Parallelism = defaultArgonParallelism
	}
	if h.SaltLength == 0 {
		h.SaltLength = defaultArgonSaltLength
	}
	if h.KeyLength == 0 {
		h.KeyLength = defaultArgonKeyLength
	}
	return h
}
