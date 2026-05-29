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
	getenv        = os.Getenv
)

const configDirEnvVar = "ENVFUSE_CONFIG_DIR"

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
	if envDir := getenv(configDirEnvVar); envDir != "" {
		return filepath.Clean(envDir), nil
	}

	dir, err := userConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config directory: %w", err)
	}

	return filepath.Join(dir, "envfuse"), nil
}

func EnsureUserConfigPath() (string, error) {
	path, err := GetUserConfigPath()
	if err != nil {
		return "", err
	}
	if err = os.MkdirAll(path, 0o700); err != nil {
		return "", fmt.Errorf("create config directory %q: %w", path, err)
	}
	return path, nil
}

func zeroBytes(data []byte) {
	for i := range data {
		data[i] = 0
	}
}

func init() {
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("{{.Version}}\n")
	rootCmd.PersistentFlags().StringVar(&configDirFlag, "config-dir", "", "Override envfuse configuration directory")
}
