package daemon

import (
	"os"
	"testing"
	"time"
)

func TestMetadataStore(t *testing.T) {
	// Setup
	tmpFile := "test_metadata.db"
	key := []byte("12345678901234567890123456789012")

	// Clean up after test
	defer os.Remove(tmpFile)

	t.Run("New store is empty", func(t *testing.T) {
		ms := NewMetadataStore()
		if len(ms.Metadata) != 0 {
			t.Errorf("New metadata store should be empty, got %d items", len(ms.Metadata))
		}
	})

	t.Run("Save and load metadata", func(t *testing.T) {
		ms := NewMetadataStore()

		// Add test metadata
		testMeta := FileMetadata{
			EncryptedName:   "encrypted.txt",
			OriginalPath:    "/path/to/file.txt",
			OriginalHash:    "original123",
			EncryptedHash:   "encrypted456",
			LastModified:    time.Now(),
			FileSize:        1234,
			EncryptionNonce: []byte("nonce1"),
			FileNonce:       []byte("nonce2"),
		}

		ms.Metadata["encrypted.txt"] = testMeta

		// Save to file
		err := ms.SaveToFile(tmpFile, key)
		if err != nil {
			t.Fatalf("Failed to save metadata: %v", err)
		}

		// Load from file
		loaded, err := LoadMetadataStore(tmpFile, key)
		if err != nil {
			t.Fatalf("Failed to load metadata: %v", err)
		}

		// Verify loaded data
		if len(loaded.Metadata) != 1 {
			t.Errorf("Expected 1 item in loaded metadata, got %d", len(loaded.Metadata))
		}

		loadedMeta := loaded.Metadata["encrypted.txt"]
		if loadedMeta.EncryptedName != testMeta.EncryptedName {
			t.Errorf("Expected encrypted name %s, got %s", testMeta.EncryptedName, loadedMeta.EncryptedName)
		}
		if loadedMeta.FileSize != testMeta.FileSize {
			t.Errorf("Expected file size %d, got %d", testMeta.FileSize, loadedMeta.FileSize)
		}
	})

	t.Run("Load non-existent file", func(t *testing.T) {
		ms, err := LoadMetadataStore("nonexistent.db", key)
		if err != nil {
			t.Fatalf("Expected no error for non-existent file, got: %v", err)
		}
		if len(ms.Metadata) != 0 {
			t.Errorf("Expected empty metadata store, got %d items", len(ms.Metadata))
		}
	})
}
