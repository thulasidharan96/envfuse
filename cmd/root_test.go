package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetUserConfigPathUsesUserConfigDir(t *testing.T) {
	origUserConfigDir := userConfigDir
	origConfigDirFlag := configDirFlag
	origGetenv := getenv
	t.Cleanup(func() {
		userConfigDir = origUserConfigDir
		configDirFlag = origConfigDirFlag
		getenv = origGetenv
	})

	userConfigDir = func() (string, error) {
		return filepath.Join("base", "config"), nil
	}
	getenv = func(string) string { return "" }
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

func TestGetUserConfigPathUsesEnvVar(t *testing.T) {
	origUserConfigDir := userConfigDir
	origConfigDirFlag := configDirFlag
	origGetenv := getenv
	t.Cleanup(func() {
		userConfigDir = origUserConfigDir
		configDirFlag = origConfigDirFlag
		getenv = origGetenv
	})

	userConfigDir = func() (string, error) {
		return filepath.Join("ignored", "config"), nil
	}
	getenv = func(key string) string {
		if key == configDirEnvVar {
			return filepath.Join("account", "envfuse")
		}
		return ""
	}
	configDirFlag = ""

	path, err := GetUserConfigPath()
	if err != nil {
		t.Fatalf("GetUserConfigPath() error = %v", err)
	}

	expected := filepath.Join("account", "envfuse")
	if path != expected {
		t.Fatalf("GetUserConfigPath() = %q, want %q", path, expected)
	}
}

func TestEnsureUserConfigPathCreatesDirectory(t *testing.T) {
	origUserConfigDir := userConfigDir
	origConfigDirFlag := configDirFlag
	origGetenv := getenv
	t.Cleanup(func() {
		userConfigDir = origUserConfigDir
		configDirFlag = origConfigDirFlag
		getenv = origGetenv
	})

	base := t.TempDir()
	userConfigDir = func() (string, error) { return base, nil }
	getenv = func(string) string { return "" }
	configDirFlag = ""

	path, err := EnsureUserConfigPath()
	if err != nil {
		t.Fatalf("EnsureUserConfigPath() error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat(%q) error = %v", path, err)
	}
	if !info.IsDir() {
		t.Fatalf("expected %q to be a directory", path)
	}
}
