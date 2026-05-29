package keymgmt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateX25519KeyPair(t *testing.T) {
	pair, err := GenerateX25519KeyPair()
	if err != nil {
		t.Fatalf("GenerateX25519KeyPair() error = %v", err)
	}
	if !strings.HasPrefix(pair.Identity, "AGE-SECRET-KEY-") {
		t.Fatalf("Identity key has unexpected format: %q", pair.Identity)
	}
	if !strings.HasPrefix(pair.Recipient, "age1") {
		t.Fatalf("Recipient key has unexpected format: %q", pair.Recipient)
	}
}

func TestWriteFileAtomic(t *testing.T) {
	outPath := filepath.Join(t.TempDir(), "keys", "identity.key")
	content := []byte("secret\n")

	if err := WriteFileAtomic(outPath, content, 0o600, false); err != nil {
		t.Fatalf("WriteFileAtomic() error = %v", err)
	}

	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", outPath, err)
	}
	if string(got) != string(content) {
		t.Fatalf("file content = %q, want %q", got, content)
	}
}

func TestWriteFileAtomicNoOverwrite(t *testing.T) {
	outPath := filepath.Join(t.TempDir(), "identity.key")
	if err := os.WriteFile(outPath, []byte("first\n"), 0o600); err != nil {
		t.Fatalf("seed file error = %v", err)
	}

	err := WriteFileAtomic(outPath, []byte("second\n"), 0o600, false)
	if err == nil {
		t.Fatal("expected overwrite protection error")
	}
	if !strings.Contains(err.Error(), "refusing to overwrite existing file") {
		t.Fatalf("unexpected error: %v", err)
	}
}

