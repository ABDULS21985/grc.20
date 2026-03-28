package crypto_test

import (
	"testing"

	"github.com/complianceforge/platform/internal/pkg/crypto"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	encryptor, err := crypto.NewEncryptor("test-encryption-key-minimum-16-chars")
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	plaintext := "This is a secret MFA token: JBSWY3DPEHPK3PXP"

	encrypted, err := encryptor.EncryptString(plaintext)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	if encrypted == plaintext {
		t.Error("Encrypted text should differ from plaintext")
	}

	decrypted, err := encryptor.DecryptString(encrypted)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Decrypted text doesn't match: got '%s', expected '%s'", decrypted, plaintext)
	}
}

func TestEncryptProducesDifferentCiphertexts(t *testing.T) {
	encryptor, _ := crypto.NewEncryptor("test-encryption-key-minimum-16-chars")
	plaintext := "same input"

	enc1, _ := encryptor.EncryptString(plaintext)
	enc2, _ := encryptor.EncryptString(plaintext)

	if enc1 == enc2 {
		t.Error("Same plaintext should produce different ciphertexts (random nonce)")
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	enc1, _ := crypto.NewEncryptor("correct-key-that-is-long-enough")
	enc2, _ := crypto.NewEncryptor("wrong-key-that-is-also-long-enou")

	encrypted, _ := enc1.EncryptString("secret data")

	_, err := enc2.DecryptString(encrypted)
	if err == nil {
		t.Error("Decryption with wrong key should fail")
	}
}

func TestEncryptorShortKey(t *testing.T) {
	_, err := crypto.NewEncryptor("short")
	if err == nil {
		t.Error("Short encryption key should be rejected")
	}
}

func TestEncryptEmptyString(t *testing.T) {
	encryptor, _ := crypto.NewEncryptor("test-encryption-key-minimum-16-chars")

	encrypted, err := encryptor.EncryptString("")
	if err != nil {
		t.Fatalf("Encrypting empty string should work: %v", err)
	}

	decrypted, err := encryptor.DecryptString(encrypted)
	if err != nil {
		t.Fatalf("Decrypting empty string should work: %v", err)
	}

	if decrypted != "" {
		t.Errorf("Expected empty string, got '%s'", decrypted)
	}
}

func TestGenerateRandomToken(t *testing.T) {
	token1, err := crypto.GenerateRandomToken(32)
	if err != nil {
		t.Fatalf("Token generation failed: %v", err)
	}

	token2, err := crypto.GenerateRandomToken(32)
	if err != nil {
		t.Fatalf("Token generation failed: %v", err)
	}

	if token1 == "" || token2 == "" {
		t.Error("Generated tokens should not be empty")
	}

	if token1 == token2 {
		t.Error("Two generated tokens should be different")
	}

	if len(token1) < 40 {
		t.Errorf("Token seems too short: %d chars", len(token1))
	}
}

func TestHashSHA256(t *testing.T) {
	hash1 := crypto.HashSHA256("hello")
	hash2 := crypto.HashSHA256("hello")
	hash3 := crypto.HashSHA256("world")

	if hash1 != hash2 {
		t.Error("Same input should produce same hash")
	}
	if hash1 == hash3 {
		t.Error("Different inputs should produce different hashes")
	}
	if len(hash1) != 64 {
		t.Errorf("SHA-256 hash should be 64 hex chars, got %d", len(hash1))
	}
}

func TestEncryptBinaryData(t *testing.T) {
	encryptor, _ := crypto.NewEncryptor("test-encryption-key-minimum-16-chars")

	data := []byte{0x00, 0x01, 0xFF, 0xFE, 0x80, 0x7F}

	encrypted, err := encryptor.Encrypt(data)
	if err != nil {
		t.Fatalf("Encrypting binary data failed: %v", err)
	}

	decrypted, err := encryptor.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypting binary data failed: %v", err)
	}

	if len(decrypted) != len(data) {
		t.Errorf("Decrypted length mismatch: %d != %d", len(decrypted), len(data))
	}

	for i, b := range decrypted {
		if b != data[i] {
			t.Errorf("Byte %d mismatch: %x != %x", i, b, data[i])
		}
	}
}
