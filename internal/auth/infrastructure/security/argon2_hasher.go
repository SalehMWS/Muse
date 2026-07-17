package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"

	"github.com/SalehMWS/Muse/internal/shared/config"
)

type Argon2Hasher struct {
	cfg config.Argon2
}

func NewArgon2Hasher(cfg config.Argon2) *Argon2Hasher {
	return &Argon2Hasher{cfg: cfg}
}

func (h *Argon2Hasher) Hash(plain string) (string, error) {
	salt := make([]byte, h.cfg.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("argon2: generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(plain), salt, h.cfg.Time, h.cfg.Memory, h.cfg.Parallelism, h.cfg.KeyLength)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, h.cfg.Memory, h.cfg.Time, h.cfg.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
	return encoded, nil
}

func (h *Argon2Hasher) Verify(encoded, plain string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, fmt.Errorf("argon2: malformed hash")
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, fmt.Errorf("argon2: parse version: %w", err)
	}
	if version != argon2.Version {
		return false, fmt.Errorf("argon2: unsupported version %d", version)
	}

	var memory, iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false, fmt.Errorf("argon2: parse params: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("argon2: decode salt: %w", err)
	}

	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("argon2: decode hash: %w", err)
	}

	got := argon2.IDKey([]byte(plain), salt, iterations, memory, parallelism, uint32(len(want))) //nolint:gosec

	return subtle.ConstantTimeCompare(got, want) == 1, nil
}
