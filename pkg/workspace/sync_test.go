package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"filippo.io/age"
	"github.com/thulasidharan96/envfuse/pkg/keymgmt"
)

func TestSynchronizeWorkspacePushPullRoundTrip(t *testing.T) {
	workspaceDir := t.TempDir()
	configDir := filepath.Join(t.TempDir(), "envfuse")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatalf("create config directory: %v", err)
	}
	t.Setenv("ENVFUSE_CONFIG_DIR", configDir)

	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("generate local identity: %v", err)
	}
	teammate, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("generate teammate identity: %v", err)
	}

	identityPath := filepath.Join(configDir, "identity.key")
	if err = keymgmt.WriteFileAtomic(identityPath, []byte(identity.String()+"\n"), 0o600, true); err != nil {
		t.Fatalf("write identity file: %v", err)
	}

	blueprintPath := filepath.Join(workspaceDir, DefaultBlueprintFile)
	blueprint := strings.Join([]string{
		"team:",
		"  - name: alice",
		"    recipient: " + identity.Recipient().String(),
		"  - name: bob",
		"    recipient: " + teammate.Recipient().String(),
		"manifest:",
		"  - .env.local",
	}, "\n") + "\n"
	if err = os.WriteFile(blueprintPath, []byte(blueprint), 0o600); err != nil {
		t.Fatalf("write blueprint: %v", err)
	}

	plainPath := filepath.Join(workspaceDir, ".env.local")
	if err = os.WriteFile(plainPath, []byte("TOKEN=local\n"), 0o600); err != nil {
		t.Fatalf("write plaintext: %v", err)
	}

	report, err := SynchronizeWorkspaceWithReport(blueprintPath)
	if err != nil {
		t.Fatalf("push sync failed: %v", err)
	}
	if report == nil || len(report.Statuses) == 0 {
		t.Fatalf("expected statuses from push sync")
	}

	storePath := filepath.Join(workspaceDir, ".envfuse", "store", ".env.local.enc")
	if _, err = os.Stat(storePath); err != nil {
		t.Fatalf("expected encrypted store file: %v", err)
	}

	if err = os.Remove(plainPath); err != nil {
		t.Fatalf("remove plaintext: %v", err)
	}

	report, err = SynchronizeWorkspaceWithReport(blueprintPath)
	if err != nil {
		t.Fatalf("pull sync failed: %v", err)
	}
	if report == nil || len(report.Statuses) == 0 {
		t.Fatalf("expected statuses from pull sync")
	}

	restored, err := os.ReadFile(plainPath)
	if err != nil {
		t.Fatalf("read restored plaintext: %v", err)
	}
	if string(restored) != "TOKEN=local\n" {
		t.Fatalf("unexpected restored payload: got %q", restored)
	}

	if err = os.WriteFile(plainPath, []byte("TOKEN=local-updated\n"), 0o600); err != nil {
		t.Fatalf("update plaintext: %v", err)
	}
	now := time.Now().UTC().Add(2 * time.Second)
	if err = os.Chtimes(plainPath, now, now); err != nil {
		t.Fatalf("set plaintext mtime: %v", err)
	}

	if _, err = SynchronizeWorkspaceWithReport(blueprintPath); err != nil {
		t.Fatalf("second push sync failed: %v", err)
	}

	if err = os.Remove(plainPath); err != nil {
		t.Fatalf("remove plaintext before second pull: %v", err)
	}

	if _, err = SynchronizeWorkspaceWithReport(blueprintPath); err != nil {
		t.Fatalf("second pull sync failed: %v", err)
	}

	restoredUpdated, err := os.ReadFile(plainPath)
	if err != nil {
		t.Fatalf("read restored updated plaintext: %v", err)
	}
	if string(restoredUpdated) != "TOKEN=local-updated\n" {
		t.Fatalf("unexpected updated payload: got %q", restoredUpdated)
	}
}

func TestLoadBlueprintNormalizesCarriageReturns(t *testing.T) {
	workspaceDir := t.TempDir()
	configPath := filepath.Join(workspaceDir, DefaultBlueprintFile)

	blueprint := "team:\r\n  - name: windows-user\r\n    recipient: age1example\r\nmanifest:\r\n  - .env.local\r\n"
	if err := os.WriteFile(configPath, []byte(blueprint), 0o600); err != nil {
		t.Fatalf("write blueprint: %v", err)
	}

	parsed, err := LoadBlueprint(configPath)
	if err != nil {
		t.Fatalf("load blueprint: %v", err)
	}

	if parsed.Team[0].Name != "windows-user" {
		t.Fatalf("unexpected normalized name: %q", parsed.Team[0].Name)
	}
	if strings.Contains(parsed.Team[0].Recipient, "\r") {
		t.Fatalf("recipient still contains carriage return: %q", parsed.Team[0].Recipient)
	}
}
