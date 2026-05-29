package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildManifestCommandOutput(t *testing.T) {
	origUserConfigDir := userConfigDir
	origConfigDirFlag := configDirFlag
	origGetenv := getenv
	origOut := manifestOutFile
	t.Cleanup(func() {
		userConfigDir = origUserConfigDir
		configDirFlag = origConfigDirFlag
		getenv = origGetenv
		manifestOutFile = origOut
	})

	base := t.TempDir()
	userConfigDir = func() (string, error) { return base, nil }
	getenv = func(string) string { return "" }
	configDirFlag = ""
	manifestOutFile = filepath.Join(t.TempDir(), "manifest.json")

	if err := manifestCmd.RunE(manifestCmd, nil); err != nil {
		t.Fatalf("manifestCmd.RunE() error = %v", err)
	}

	var parsed runtimeManifest
	raw, err := os.ReadFile(manifestOutFile)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", manifestOutFile, err)
	}
	if err = json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("Unmarshal manifest error = %v", err)
	}

	expectedConfig := filepath.Join(base, "envfuse")
	if parsed.ConfigDir != expectedConfig {
		t.Fatalf("ConfigDir = %q, want %q", parsed.ConfigDir, expectedConfig)
	}
	if parsed.ConfigDirEnvVar != configDirEnvVar {
		t.Fatalf("ConfigDirEnvVar = %q, want %q", parsed.ConfigDirEnvVar, configDirEnvVar)
	}
}
