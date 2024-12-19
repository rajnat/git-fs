package daemon

import (
	"encoding/json"
	"git-fs/internal/crypto"
	"os"
	"sync"
	"time"
)

// FileMetadata stores encryption and integrity information
type FileMetadata struct {
	EncryptedName   string    `json:"encrypted_name"`
	OriginalPath    string    `json:"original_path"`  // Stored encrypted
	OriginalHash    string    `json:"original_hash"`  // SHA-256 of original file
	EncryptedHash   string    `json:"encrypted_hash"` // SHA-256 of encrypted file
	LastModified    time.Time `json:"last_modified"`
	FileSize        int64     `json:"file_size"`
	EncryptionNonce []byte    `json:"encryption_nonce"` // For filename encryption
	FileNonce       []byte    `json:"file_nonce"`       // For file content encryption
}

type MetadataStore struct {
	Mu       sync.RWMutex
	Metadata map[string]FileMetadata `json:"metadata"` // Maps encrypted filename to metadata
}

type ChangeSet struct {
	Mu    sync.Mutex
	Files map[string]struct{}
}

func NewMetadataStore() *MetadataStore {
	return &MetadataStore{
		Metadata: make(map[string]FileMetadata),
	}
}

func (ms *MetadataStore) SaveToFile(path string, key []byte) error {
	ms.Mu.RLock()
	defer ms.Mu.RUnlock()

	data, err := json.Marshal(ms)
	if err != nil {
		return err
	}

	// Encrypt the metadata before saving
	encryptedData, err := crypto.Encrypt(key, data)
	if err != nil {
		return err
	}

	return os.WriteFile(path, encryptedData, 0600)
}

func LoadMetadataStore(path string, key []byte) (*MetadataStore, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewMetadataStore(), nil
		}
		return nil, err
	}

	// Decrypt the metadata
	decryptedData, err := crypto.Decrypt(key, data)
	if err != nil {
		return nil, err
	}

	ms := NewMetadataStore()
	if err := json.Unmarshal(decryptedData, ms); err != nil {
		return nil, err
	}

	return ms, nil
}
