package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
	"os"
)

// EncryptFile encrypts the file at `plaintextPath` and writes the result to `ciphertextPath`.
func EncryptFile(key []byte, plaintextPath string, ciphertextPath string) error {
	plaintext, err := os.ReadFile(plaintextPath)
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	return os.WriteFile(ciphertextPath, ciphertext, 0644)
}

var ErrInvalidCiphertext = errors.New("invalid ciphertext")

// DecryptFile decrypts the file at `ciphertextPath` and writes the plaintext to `plaintextPath`.
func DecryptFile(key []byte, ciphertextPath string, plaintextPath string) error {
	ciphertext, err := os.ReadFile(ciphertextPath)
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return ErrInvalidCiphertext
	}
	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return err
	}

	return os.WriteFile(plaintextPath, plaintext, 0644)
}
