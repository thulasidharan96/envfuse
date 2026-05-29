package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	enc "github.com/thulasidharan96/envfuse/pkg/crypto"
)

var (
	decryptIdentityPath string
	decryptOutFile      string
)

var decryptCmd = &cobra.Command{
	Use:   "decrypt [target]",
	Short: "Decrypt an age-encrypted file",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		targetPath := filepath.Clean(args[0])
		configPath, err := EnsureUserConfigPath()
		if err != nil {
			return err
		}

		identityPath := filepath.Join(configPath, "identity.key")
		if decryptIdentityPath != "" {
			identityPath = filepath.Clean(decryptIdentityPath)
		}

		identityRaw, err := os.ReadFile(identityPath)
		if err != nil {
			return fmt.Errorf("read identity file %q: %w", identityPath, err)
		}

		identityKey := strings.TrimSpace(string(identityRaw))
		zeroBytes(identityRaw)
		if identityKey == "" {
			return fmt.Errorf("identity file %q is empty", identityPath)
		}

		encrypted, err := os.ReadFile(targetPath)
		if err != nil {
			return fmt.Errorf("read encrypted file %q: %w", targetPath, err)
		}

		decrypted, err := enc.DecryptBytes(encrypted, identityKey)
		if err != nil {
			return err
		}
		defer zeroBytes(decrypted)

		outPath := filepath.Join(configPath, "decrypted.env")
		if decryptOutFile != "" {
			outPath = filepath.Clean(decryptOutFile)
		}

		if err = os.WriteFile(outPath, decrypted, 0o600); err != nil {
			return fmt.Errorf("write decrypted file %q: %w", outPath, err)
		}

		return nil
	},
}

func init() {
	decryptCmd.Flags().StringVar(&decryptIdentityPath, "identity", "", "Path to local age X25519 identity private key")
	decryptCmd.Flags().StringVar(&decryptOutFile, "out", "", "Destination path for decrypted output")
	rootCmd.AddCommand(decryptCmd)
}
