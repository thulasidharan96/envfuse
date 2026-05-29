package keymgmt

import (
	"fmt"
	"os"
	"path/filepath"

	"filippo.io/age"
)

type GeneratedKeyPair struct {
	Identity  string
	Recipient string
}

func GenerateX25519KeyPair() (*GeneratedKeyPair, error) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, fmt.Errorf("generate x25519 identity: %w", err)
	}

	return &GeneratedKeyPair{
		Identity:  identity.String(),
		Recipient: identity.Recipient().String(),
	}, nil
}

func WriteFileAtomic(path string, payload []byte, mode os.FileMode, overwrite bool) error {
	if path == "" {
		return fmt.Errorf("path is required")
	}

	outPath := filepath.Clean(path)
	parent := filepath.Dir(outPath)
	if err := os.MkdirAll(parent, 0o700); err != nil {
		return fmt.Errorf("create parent directory %q: %w", parent, err)
	}

	if !overwrite {
		if _, err := os.Stat(outPath); err == nil {
			return fmt.Errorf("refusing to overwrite existing file %q", outPath)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("check output file %q: %w", outPath, err)
		}
	}

	tmpFile, err := os.CreateTemp(parent, ".envfuse-*")
	if err != nil {
		return fmt.Errorf("create temporary output file: %w", err)
	}
	tmpPath := tmpFile.Name()
	keepTemp := true
	defer func() {
		_ = tmpFile.Close()
		if keepTemp {
			_ = os.Remove(tmpPath)
		}
	}()

	if err = tmpFile.Chmod(mode); err != nil {
		return fmt.Errorf("set temporary output file mode: %w", err)
	}
	if _, err = tmpFile.Write(payload); err != nil {
		return fmt.Errorf("write temporary output file: %w", err)
	}
	if err = tmpFile.Sync(); err != nil {
		return fmt.Errorf("sync temporary output file: %w", err)
	}
	if err = tmpFile.Close(); err != nil {
		return fmt.Errorf("close temporary output file: %w", err)
	}

	if err = os.Rename(tmpPath, outPath); err != nil {
		return fmt.Errorf("move temporary file to %q: %w", outPath, err)
	}

	keepTemp = false
	return nil
}
