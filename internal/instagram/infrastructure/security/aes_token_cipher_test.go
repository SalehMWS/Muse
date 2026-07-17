package security

import (
	"testing"
)

func TestAESTokenCipher_RoundTrip(t *testing.T) {
	cipher, err := NewAESTokenCipher("test-secret")
	if err != nil {
		t.Fatalf("NewAESTokenCipher() unexpected error: %v", err)
	}

	plaintext := "IGQVJ...long-lived-access-token"
	encrypted, err := cipher.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() unexpected error: %v", err)
	}
	if encrypted == plaintext {
		t.Fatal("Encrypt() returned plaintext")
	}

	decrypted, err := cipher.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt() unexpected error: %v", err)
	}
	if decrypted != plaintext {
		t.Fatalf("Decrypt() = %q, want %q", decrypted, plaintext)
	}
}

func TestAESTokenCipher_NonceIsRandom(t *testing.T) {
	cipher, _ := NewAESTokenCipher("test-secret")
	first, _ := cipher.Encrypt("same")
	second, _ := cipher.Encrypt("same")
	if first == second {
		t.Fatal("Encrypt() produced identical ciphertext for identical input")
	}
}

func TestAESTokenCipher_DecryptErrors(t *testing.T) {
	cipher, _ := NewAESTokenCipher("test-secret")

	if _, err := cipher.Decrypt("not-base64!!!"); err == nil {
		t.Fatal("Decrypt() invalid base64 = nil error, want error")
	}
	if _, err := cipher.Decrypt("QUJD"); err != ErrInvalidCiphertext {
		t.Fatalf("Decrypt() short ciphertext error = %v, want %v", err, ErrInvalidCiphertext)
	}

	other, _ := NewAESTokenCipher("different-secret")
	encrypted, _ := cipher.Encrypt("secret")
	if _, err := other.Decrypt(encrypted); err == nil {
		t.Fatal("Decrypt() with wrong key = nil error, want authentication failure")
	}
}
