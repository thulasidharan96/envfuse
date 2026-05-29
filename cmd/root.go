package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	version       = "0.1.0"
	configDirFlag string
	userConfigDir = os.UserConfigDir
)

var rootCmd = &cobra.Command{
	Use:           "envfuse",
	Short:         "Securely sync encrypted env and binary assets with Git",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func GetUserConfigPath() (string, error) {
	if configDirFlag != "" {
		return filepath.Clean(configDirFlag), nil
	}

	dir, err := userConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config directory: %w", err)
	}

	return filepath.Join(dir, "envfuse"), nil
}

func init() {
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("{{.Version}}\n")
	rootCmd.PersistentFlags().StringVar(&configDirFlag, "config-dir", "", "Override envfuse configuration directory")
}
