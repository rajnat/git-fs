package crypto

import (
	"bytes"
	"testing"
)

func TestEncryptionRoundtrip(t *testing.T) {
	// Create a proper 32-byte key for AES-256
	key := make([]byte, KeySize)
	for i := range key {
		key[i] = byte(i)
	}

	t.Run("Encrypt and Decrypt", func(t *testing.T) {
		originalData := []byte("Hello, World!")

		// Encrypt
		encrypted, err := Encrypt(key, originalData)
		if err != nil {
			t.Fatalf("Encryption failed: %v", err)
		}

		// Decrypt
		decrypted, err := Decrypt(key, encrypted)
		if err != nil {
			t.Fatalf("Decryption failed: %v", err)
		}

		// Compare
		if !bytes.Equal(originalData, decrypted) {
			t.Errorf("Decrypted data doesn't match original\nExpected: %v\nGot: %v", originalData, decrypted)
		}
	})

	t.Run("Invalid Key Size", func(t *testing.T) {
		shortKey := []byte("too short")
		_, err := Encrypt(shortKey, []byte("test"))
		if err == nil {
			t.Error("Expected error for invalid key size")
		}
	})

	t.Run("Filename Encryption", func(t *testing.T) {
		filename := "test.txt"
		encName, fileNonce, nameNonce, err := EncryptFileName(key, filename)
		if err != nil {
			t.Fatalf("Failed to encrypt filename: %v", err)
		}

		if len(fileNonce) != NonceSize {
			t.Errorf("Expected file nonce size %d, got %d", NonceSize, len(fileNonce))
		}

		if len(nameNonce) != NonceSize {
			t.Errorf("Expected name nonce size %d, got %d", NonceSize, len(nameNonce))
		}

		if encName == filename {
			t.Error("Encrypted filename should be different from original")
		}
	})

	t.Run("Compression Roundtrip", func(t *testing.T) {
		original := []byte("This is some test data that should be compressible")

		compressed, err := Compress(original)
		if err != nil {
			t.Fatalf("Compression failed: %v", err)
		}

		decompressed, err := Decompress(compressed)
		if err != nil {
			t.Fatalf("Decompression failed: %v", err)
		}

		if !bytes.Equal(original, decompressed) {
			t.Error("Decompressed data doesn't match original")
		}
	})

	t.Run("File Content Encryption", func(t *testing.T) {
		content := []byte("This is file content")
		nonce := make([]byte, NonceSize)

		encrypted, err := EncryptFile(key, content, nonce)
		if err != nil {
			t.Fatalf("File encryption failed: %v", err)
		}

		if len(encrypted) <= len(content) {
			t.Error("Encrypted content should be longer than original due to auth tag")
		}
	})
}
