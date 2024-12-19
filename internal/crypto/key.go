package crypto

import (
	"crypto/rand"
	"errors"

	"golang.org/x/crypto/scrypt"
)

// DeriveKey uses scrypt to derive a key from a password.
// The salt can be stored alongside encrypted files. It's not secret, but must be unique.
func DeriveKey(password string, salt []byte) ([]byte, error) {
	if len(salt) == 0 {
		return nil, errors.New("no salt provided")
	}
	key, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, KeySize)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// GenerateSalt for key derivation.
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}
