package security_test

import (
	"testing"

	"github.com/SalehMWS/Muse/internal/auth/infrastructure/security"
	"github.com/SalehMWS/Muse/internal/shared/config"
)

func testArgon2Config() config.Argon2 {
	return config.Argon2{Memory: 8192, Time: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32}
}

func TestArgon2Hasher_HashAndVerify(t *testing.T) {
	hasher := security.NewArgon2Hasher(testArgon2Config())

	encoded, err := hasher.Hash("Str0ng!Passw0rd")
	if err != nil {
		t.Fatalf("Hash() unexpected error: %v", err)
	}
	if encoded == "Str0ng!Passw0rd" {
		t.Fatal("Hash() returned the plaintext password unchanged")
	}

	match, err := hasher.Verify(encoded, "Str0ng!Passw0rd")
	if err != nil {
		t.Fatalf("Verify() unexpected error: %v", err)
	}
	if !match {
		t.Fatal("Verify() = false for the correct password")
	}

	match, err = hasher.Verify(encoded, "wrong-password")
	if err != nil {
		t.Fatalf("Verify() unexpected error: %v", err)
	}
	if match {
		t.Fatal("Verify() = true for the wrong password")
	}
}

func TestArgon2Hasher_Verify_MalformedHash(t *testing.T) {
	hasher := security.NewArgon2Hasher(testArgon2Config())

	if _, err := hasher.Verify("not-a-valid-argon2-hash", "whatever"); err == nil {
		t.Fatal("Verify() expected an error for a malformed encoded hash")
	}
}

func TestArgon2Hasher_Hash_DistinctSaltPerCall(t *testing.T) {
	hasher := security.NewArgon2Hasher(testArgon2Config())

	first, err := hasher.Hash("Str0ng!Passw0rd")
	if err != nil {
		t.Fatalf("Hash() unexpected error: %v", err)
	}
	second, err := hasher.Hash("Str0ng!Passw0rd")
	if err != nil {
		t.Fatalf("Hash() unexpected error: %v", err)
	}

	if first == second {
		t.Fatal("Hash() produced identical output for two calls with the same password")
	}
}
