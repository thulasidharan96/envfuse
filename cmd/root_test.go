package cmd

import (
	"path/filepath"
	"testing"
)

func TestGetUserConfigPathUsesUserConfigDir(t *testing.T) {
	origUserConfigDir := userConfigDir
	origConfigDirFlag := configDirFlag
	t.Cleanup(func() {
		userConfigDir = origUserConfigDir
		configDirFlag = origConfigDirFlag
	})

	userConfigDir = func() (string, error) {
		return filepath.Join("base", "config"), nil
	}
	configDirFlag = ""

	path, err := GetUserConfigPath()
	if err != nil {
		t.Fatalf("GetUserConfigPath() error = %v", err)
	}

	expected := filepath.Join("base", "config", "envfuse")
	if path != expected {
		t.Fatalf("GetUserConfigPath() = %q, want %q", path, expected)
	}
}
