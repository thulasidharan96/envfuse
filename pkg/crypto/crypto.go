package crypto

import (
	"bytes"
	"fmt"
	"io"

	"filippo.io/age"
)

func EncryptBytes(data []byte, recipientKeys []string) ([]byte, error) {
	if len(recipientKeys) == 0 {
		return nil, fmt.Errorf("at least one recipient key is required")
	}

	recipients := make([]age.Recipient, 0, len(recipientKeys))
	for _, key := range recipientKeys {
		recipient, err := age.ParseX25519Recipient(key)
		if err != nil {
			return nil, fmt.Errorf("parse recipient key: %w", err)
		}
		recipients = append(recipients, recipient)
	}

	var out bytes.Buffer
	writer, err := age.Encrypt(&out, recipients...)
	if err != nil {
		return nil, fmt.Errorf("create age encrypt writer: %w", err)
	}

	if _, err = writer.Write(data); err != nil {
		return nil, fmt.Errorf("encrypt payload: %w", err)
	}

	if err = writer.Close(); err != nil {
		return nil, fmt.Errorf("finalize encrypted payload: %w", err)
	}

	return out.Bytes(), nil
}

func DecryptBytes(encData []byte, identityKey string) ([]byte, error) {
	identity, err := age.ParseX25519Identity(identityKey)
	if err != nil {
		return nil, fmt.Errorf("parse identity key: %w", err)
	}

	reader, err := age.Decrypt(bytes.NewReader(encData), identity)
	if err != nil {
		return nil, fmt.Errorf("create age decrypt reader: %w", err)
	}

	plain, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("decrypt payload: %w", err)
	}

	return plain, nil
}
