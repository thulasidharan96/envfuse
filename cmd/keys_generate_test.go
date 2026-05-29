package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateKeyFilesWritesIdentityAndRecipient(t *testing.T) {
	identityPath := filepath.Join(t.TempDir(), "identity.key")
	recipientsPath := filepath.Join(t.TempDir(), "recipients.txt")

	if err := generateKeyFiles(identityPath, recipientsPath, false); err != nil {
		t.Fatalf("generateKeyFiles() error = %v", err)
	}

	identityRaw, err := os.ReadFile(identityPath)
	if err != nil {
		t.Fatalf("ReadFile(identity) error = %v", err)
	}
	recipientRaw, err := os.ReadFile(recipientsPath)
	if err != nil {
		t.Fatalf("ReadFile(recipients) error = %v", err)
	}

	identity := strings.TrimSpace(string(identityRaw))
	recipient := strings.TrimSpace(string(recipientRaw))
	if !strings.HasPrefix(identity, "AGE-SECRET-KEY-") {
		t.Fatalf("identity has unexpected format: %q", identity)
	}
	if !strings.HasPrefix(recipient, "age1") {
		t.Fatalf("recipient has unexpected format: %q", recipient)
	}
}

func TestGenerateKeyFilesNoOverwrite(t *testing.T) {
	identityPath := filepath.Join(t.TempDir(), "identity.key")
	recipientsPath := filepath.Join(t.TempDir(), "recipients.txt")

	if err := os.WriteFile(identityPath, []byte("existing\n"), 0o600); err != nil {
		t.Fatalf("seed identity file error = %v", err)
	}

	err := generateKeyFiles(identityPath, recipientsPath, false)
	if err == nil {
		t.Fatal("expected overwrite protection error")
	}
	if !strings.Contains(err.Error(), "write identity key") {
		t.Fatalf("unexpected error: %v", err)
	}
}

