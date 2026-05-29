package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var manifestOutFile string

type runtimeManifest struct {
	Version              string `json:"version"`
	ConfigDir            string `json:"config_dir"`
	ConfigDirEnvVar      string `json:"config_dir_env_var"`
	DefaultIdentityPath  string `json:"default_identity_path"`
	DefaultDecryptOutput string `json:"default_decrypt_output"`
	ExportCommand        string `json:"export_command"`
}

var manifestCmd = &cobra.Command{
	Use:   "manifest",
	Short: "Generate runtime manifest for production setup",
	RunE: func(_ *cobra.Command, _ []string) error {
		configPath, err := EnsureUserConfigPath()
		if err != nil {
			return err
		}

		manifest := runtimeManifest{
			Version:              version,
			ConfigDir:            configPath,
			ConfigDirEnvVar:      configDirEnvVar,
			DefaultIdentityPath:  filepath.Join(configPath, "identity.key"),
			DefaultDecryptOutput: filepath.Join(configPath, "decrypted.env"),
			ExportCommand:        fmt.Sprintf("export %s=%q", configDirEnvVar, configPath),
		}

		body, err := json.MarshalIndent(manifest, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal manifest: %w", err)
		}
		body = append(body, '\n')

		if manifestOutFile == "" {
			_, err = os.Stdout.Write(body)
			return err
		}

		outPath := filepath.Clean(manifestOutFile)
		if err = os.WriteFile(outPath, body, 0o600); err != nil {
			return fmt.Errorf("write manifest file %q: %w", outPath, err)
		}
		return nil
	},
}

func init() {
	manifestCmd.Flags().StringVar(&manifestOutFile, "out", "", "Destination path for generated manifest JSON")
	rootCmd.AddCommand(manifestCmd)
}
