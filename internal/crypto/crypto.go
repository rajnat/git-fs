package crypto

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

const (
	NonceSize = 12 // Size for GCM nonce
	KeySize   = 32 // Size for AES-256
)

func Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func Decompress(data []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	return io.ReadAll(gz)
}

// EncryptFileName encrypts a filename and returns the encrypted name plus nonces
func EncryptFileName(key []byte, name string) (string, []byte, []byte, error) {
	if len(key) != KeySize {
		return "", nil, nil, errors.New("invalid key size")
	}

	// Generate nonces for filename and file content
	fileNonce := make([]byte, NonceSize)
	nameNonce := make([]byte, NonceSize)

	if _, err := io.ReadFull(rand.Reader, fileNonce); err != nil {
		return "", nil, nil, err
	}
	if _, err := io.ReadFull(rand.Reader, nameNonce); err != nil {
		return "", nil, nil, err
	}

	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", nil, nil, err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", nil, nil, err
	}

	// Encrypt the filename
	encrypted := gcm.Seal(nil, nameNonce, []byte(name), nil)

	// Encode to base64 for safe filename usage
	encodedName := base64.URLEncoding.EncodeToString(encrypted)

	return encodedName, fileNonce, nameNonce, nil
}

// EncryptFile encrypts file content using the provided key and nonce
func EncryptFile(key []byte, content []byte, nonce []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, errors.New("invalid key size")
	}
	if len(nonce) != NonceSize {
		return nil, errors.New("invalid nonce size")
	}

	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	// Compress data
	compressed, err := Compress(content)
	if err != nil {
		return nil, err
	}

	// Encrypt the content
	encrypted := gcm.Seal(nil, nonce, compressed, nil)
	return encrypted, nil
}

// Encrypt encrypts arbitrary data using AES-GCM
func Encrypt(key []byte, data []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, errors.New("invalid key size")
	}

	// Generate nonce
	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Encrypt the data
	encrypted := gcm.Seal(nil, nonce, data, nil)

	// Prepend nonce to encrypted data
	return append(nonce, encrypted...), nil
}

// DecryptFile decrypts a file and writes it to the destination path
func DecryptFile(key []byte, sourcePath, destPath string) error {
	// Read encrypted file
	encryptedData, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}

	if len(encryptedData) < NonceSize {
		return errors.New("encrypted file too short")
	}

	// Extract nonce from the start of file
	nonce := encryptedData[:NonceSize]
	ciphertext := encryptedData[NonceSize:]

	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	// Decrypt the file content
	decrypted, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return err
	}

	// Decompress data
	decompressed, err := Decompress(decrypted)
	if err != nil {
		return err
	}
	// Write decrypted content to destination
	return os.WriteFile(destPath, decompressed, 0600)
}

// Decrypt decrypts data that was encrypted with Encrypt()
func Decrypt(key []byte, data []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, errors.New("invalid key size")
	}
	if len(data) < NonceSize {
		return nil, errors.New("data too short")
	}

	// Extract nonce from the start of data
	nonce := data[:NonceSize]
	ciphertext := data[NonceSize:]

	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
