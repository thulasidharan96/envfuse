package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	enc "github.com/thulasidharan96/envfuse/pkg/crypto"
)

var (
	encryptKeysFile string
	encryptOutFile  string
)

var encryptCmd = &cobra.Command{
	Use:   "encrypt [target]",
	Short: "Encrypt a file for one or more age recipients",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		targetPath := filepath.Clean(args[0])
		if encryptKeysFile == "" {
			return fmt.Errorf("--keys is required")
		}

		recipients, err := parseRecipientFile(filepath.Clean(encryptKeysFile))
		if err != nil {
			return err
		}

		payload, err := os.ReadFile(targetPath)
		if err != nil {
			return fmt.Errorf("read target file %q: %w", targetPath, err)
		}

		encrypted, err := enc.EncryptBytes(payload, recipients)
		zeroBytes(payload)
		if err != nil {
			return err
		}

		outPath := targetPath + ".age"
		if encryptOutFile != "" {
			outPath = filepath.Clean(encryptOutFile)
		}

		if err = os.WriteFile(outPath, encrypted, 0o600); err != nil {
			return fmt.Errorf("write encrypted file %q: %w", outPath, err)
		}

		return nil
	},
}

func parseRecipientFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open key file %q: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	keys := make([]string, 0)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		keys = append(keys, line)
	}

	if err = scanner.Err(); err != nil {
		return nil, fmt.Errorf("read key file %q: %w", path, err)
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("no recipient keys found in %q", path)
	}

	return keys, nil
}

func init() {
	encryptCmd.Flags().StringVar(&encryptKeysFile, "keys", "", "Path to file containing one age recipient key per line")
	encryptCmd.Flags().StringVar(&encryptOutFile, "out", "", "Destination path for encrypted output")
	rootCmd.AddCommand(encryptCmd)
}
