package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/thulasidharan96/envfuse/pkg/keymgmt"
)

var (
	generateIdentityOutPath   string
	generateRecipientsOutPath string
	generateForceOverwrite    bool
)

var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "Native key management for age identity and recipients",
}

var keysGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate local age identity and recipient files",
	RunE: func(_ *cobra.Command, _ []string) error {
		configPath, err := EnsureUserConfigPath()
		if err != nil {
			return err
		}

		identityPath := filepath.Join(configPath, "identity.key")
		if generateIdentityOutPath != "" {
			identityPath = filepath.Clean(generateIdentityOutPath)
		}

		recipientsPath := filepath.Join(configPath, "recipients.txt")
		if generateRecipientsOutPath != "" {
			recipientsPath = filepath.Clean(generateRecipientsOutPath)
		}

		if err = generateKeyFiles(identityPath, recipientsPath, generateForceOverwrite); err != nil {
			return err
		}

		fmt.Fprintf(
			rootCmd.OutOrStdout(),
			"generated key material:\n  identity: %s\n  recipients: %s\n",
			identityPath,
			recipientsPath,
		)
		return nil
	},
}

func generateKeyFiles(identityPath, recipientsPath string, force bool) error {
	keyPair, err := keymgmt.GenerateX25519KeyPair()
	if err != nil {
		return err
	}

	if err = keymgmt.WriteFileAtomic(identityPath, []byte(keyPair.Identity+"\n"), 0o600, force); err != nil {
		return fmt.Errorf("write identity key: %w", err)
	}
	if err = keymgmt.WriteFileAtomic(recipientsPath, []byte(keyPair.Recipient+"\n"), 0o600, force); err != nil {
		return fmt.Errorf("write recipients file: %w", err)
	}

	return nil
}

func init() {
	keysGenerateCmd.Flags().StringVar(&generateIdentityOutPath, "identity-out", "", "Destination path for generated private identity key")
	keysGenerateCmd.Flags().StringVar(&generateRecipientsOutPath, "recipients-out", "", "Destination path for generated recipient public key list")
	keysGenerateCmd.Flags().BoolVar(&generateForceOverwrite, "force", false, "Overwrite existing output files")

	keysCmd.AddCommand(keysGenerateCmd)
	rootCmd.AddCommand(keysCmd)
}
