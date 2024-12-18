package fileutils

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// WriteFileAtomic writes data to a file atomically.
func WriteFileAtomic(filename string, data []byte, perm os.FileMode) error {
	tmpfile, err := ioutil.TempFile(filepath.Dir(filename), "tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(data); err != nil {
		tmpfile.Close()
		return err
	}
	tmpfile.Close()

	return os.Rename(tmpfile.Name(), filename)
}
