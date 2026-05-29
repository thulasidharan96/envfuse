package crypto

import (
	"strings"
	"testing"

	"filippo.io/age"
)

func TestEncryptDecryptBytesRoundTrip(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("generate keypair: %v", err)
	}

	payload := []byte("DATABASE_URL=postgres://localhost/dev")
	encrypted, err := EncryptBytes(payload, []string{identity.Recipient().String()})
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if len(encrypted) == 0 {
		t.Fatal("expected encrypted data")
	}

	decrypted, err := DecryptBytes(encrypted, identity.String())
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if string(decrypted) != string(payload) {
		t.Fatalf("unexpected decrypted payload: got %q want %q", decrypted, payload)
	}
}

func TestEncryptBytesInvalidRecipient(t *testing.T) {
	_, err := EncryptBytes([]byte("secret"), []string{"not-a-recipient"})
	if err == nil {
		t.Fatal("expected error for invalid recipient")
	}
	if !strings.Contains(err.Error(), "parse recipient key") {
		t.Fatalf("unexpected error: %v", err)
	}
}
