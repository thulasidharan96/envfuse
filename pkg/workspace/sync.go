package workspace

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	enc "github.com/thulasidharan96/envfuse/pkg/crypto"
	"github.com/thulasidharan96/envfuse/pkg/keymgmt"
)

type SyncAction string

const (
	SyncActionPush SyncAction = "push"
	SyncActionPull SyncAction = "pull"
	SyncActionNoop SyncAction = "noop"
)

type FileSyncStatus struct {
	Path      string
	StorePath string
	Action    SyncAction
	Message   string
	Err       error
}

type SyncReport struct {
	Statuses []FileSyncStatus
}

func SynchronizeWorkspace(configPath string) error {
	_, err := SynchronizeWorkspaceWithReport(configPath)
	return err
}

func SynchronizeWorkspaceWithReport(configPath string) (*SyncReport, error) {
	resolvedConfigPath, err := resolveBlueprintPath(configPath)
	if err != nil {
		return nil, err
	}

	blueprint, err := LoadBlueprint(resolvedConfigPath)
	if err != nil {
		return nil, err
	}

	rootDir := filepath.Dir(resolvedConfigPath)
	storeDir := filepath.Join(rootDir, ".envfuse", "store")
	if err = os.MkdirAll(storeDir, 0o700); err != nil {
		return nil, fmt.Errorf("create workspace store %q: %w", storeDir, err)
	}

	trackedFiles, err := collectTrackedFiles(rootDir, storeDir, blueprint.Manifest)
	if err != nil {
		return nil, err
	}

	recipientKeys := make([]string, 0, len(blueprint.Team))
	for _, member := range blueprint.Team {
		recipientKeys = append(recipientKeys, member.Recipient)
	}

	identityKey, identityErr := loadIdentityKey()
	report := &SyncReport{Statuses: make([]FileSyncStatus, 0, len(trackedFiles))}
	var statusErrs []string

	for _, tracked := range trackedFiles {
		status := FileSyncStatus{
			Path:      tracked.RelPath,
			StorePath: tracked.StorePath,
			Action:    SyncActionNoop,
			Message:   "up to date",
		}

		plainInfo, plainErr := os.Stat(tracked.AbsPath)
		storeInfo, storeErr := os.Stat(tracked.StorePath)

		plainExists := plainErr == nil
		storeExists := storeErr == nil

		if plainErr != nil && !os.IsNotExist(plainErr) {
			status.Err = fmt.Errorf("stat workspace file %q: %w", tracked.AbsPath, plainErr)
		}
		if storeErr != nil && !os.IsNotExist(storeErr) {
			status.Err = fmt.Errorf("stat store file %q: %w", tracked.StorePath, storeErr)
		}
		if status.Err != nil {
			status.Action = SyncActionNoop
			status.Message = "stat failure"
			report.Statuses = append(report.Statuses, status)
			statusErrs = append(statusErrs, fmt.Sprintf("%s: %v", tracked.RelPath, status.Err))
			continue
		}

		shouldPush := plainExists && (!storeExists || plainInfo.ModTime().After(storeInfo.ModTime()))
		shouldPull := storeExists && (!plainExists || storeInfo.ModTime().After(plainInfo.ModTime()))

		if shouldPush {
			payload, readErr := os.ReadFile(tracked.AbsPath)
			if readErr != nil {
				status.Action = SyncActionPush
				status.Message = "failed to read plaintext"
				status.Err = fmt.Errorf("read workspace file %q: %w", tracked.AbsPath, readErr)
				report.Statuses = append(report.Statuses, status)
				statusErrs = append(statusErrs, fmt.Sprintf("%s: %v", tracked.RelPath, status.Err))
				continue
			}

			encrypted, encryptErr := enc.EncryptBytes(payload, recipientKeys)
			zeroBytes(payload)
			if encryptErr != nil {
				status.Action = SyncActionPush
				status.Message = "failed to encrypt"
				status.Err = fmt.Errorf("encrypt workspace file %q: %w", tracked.AbsPath, encryptErr)
				report.Statuses = append(report.Statuses, status)
				statusErrs = append(statusErrs, fmt.Sprintf("%s: %v", tracked.RelPath, status.Err))
				continue
			}

			if writeErr := keymgmt.WriteFileAtomic(tracked.StorePath, encrypted, 0o600, true); writeErr != nil {
				zeroBytes(encrypted)
				status.Action = SyncActionPush
				status.Message = "failed to write encrypted bundle"
				status.Err = fmt.Errorf("write store file %q: %w", tracked.StorePath, writeErr)
				report.Statuses = append(report.Statuses, status)
				statusErrs = append(statusErrs, fmt.Sprintf("%s: %v", tracked.RelPath, status.Err))
				continue
			}
			zeroBytes(encrypted)

			status.Action = SyncActionPush
			status.Message = "encrypted to workspace store"
			report.Statuses = append(report.Statuses, status)
			continue
		}

		if shouldPull {
			if identityErr != nil {
				status.Action = SyncActionPull
				status.Message = "missing local identity"
				status.Err = fmt.Errorf("load identity key: %w", identityErr)
				report.Statuses = append(report.Statuses, status)
				statusErrs = append(statusErrs, fmt.Sprintf("%s: %v", tracked.RelPath, status.Err))
				continue
			}

			encrypted, readErr := os.ReadFile(tracked.StorePath)
			if readErr != nil {
				status.Action = SyncActionPull
				status.Message = "failed to read encrypted bundle"
				status.Err = fmt.Errorf("read store file %q: %w", tracked.StorePath, readErr)
				report.Statuses = append(report.Statuses, status)
				statusErrs = append(statusErrs, fmt.Sprintf("%s: %v", tracked.RelPath, status.Err))
				continue
			}

			decrypted, decryptErr := enc.DecryptBytes(encrypted, identityKey)
			zeroBytes(encrypted)
			if decryptErr != nil {
				status.Action = SyncActionPull
				status.Message = "failed to decrypt"
				status.Err = fmt.Errorf("decrypt store file %q: %w", tracked.StorePath, decryptErr)
				report.Statuses = append(report.Statuses, status)
				statusErrs = append(statusErrs, fmt.Sprintf("%s: %v", tracked.RelPath, status.Err))
				continue
			}

			if writeErr := keymgmt.WriteFileAtomic(tracked.AbsPath, decrypted, 0o600, true); writeErr != nil {
				zeroBytes(decrypted)
				status.Action = SyncActionPull
				status.Message = "failed to write plaintext"
				status.Err = fmt.Errorf("write workspace file %q: %w", tracked.AbsPath, writeErr)
				report.Statuses = append(report.Statuses, status)
				statusErrs = append(statusErrs, fmt.Sprintf("%s: %v", tracked.RelPath, status.Err))
				continue
			}
			zeroBytes(decrypted)

			status.Action = SyncActionPull
			status.Message = "decrypted from workspace store"
			report.Statuses = append(report.Statuses, status)
			continue
		}

		report.Statuses = append(report.Statuses, status)
	}

	if len(statusErrs) > 0 {
		return report, fmt.Errorf("workspace sync completed with %d error(s): %s", len(statusErrs), strings.Join(statusErrs, "; "))
	}

	return report, nil
}

type trackedFile struct {
	RelPath   string
	AbsPath   string
	StorePath string
}

func collectTrackedFiles(rootDir, storeDir string, manifest []string) ([]trackedFile, error) {
	tracked := make(map[string]trackedFile)

	for _, manifestPath := range manifest {
		manifestAbs := filepath.Join(rootDir, manifestPath)
		info, err := os.Stat(manifestAbs)
		if err == nil {
			if info.IsDir() {
				if err = collectFilesFromDirectory(rootDir, storeDir, manifestAbs, tracked); err != nil {
					return nil, err
				}
				if err = collectStoreEntriesForPrefix(storeDir, manifestPath, rootDir, tracked); err != nil {
					return nil, err
				}
				continue
			}

			if err = addTrackedFile(rootDir, storeDir, manifestAbs, tracked); err != nil {
				return nil, err
			}
			continue
		}

		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("stat manifest path %q: %w", manifestAbs, err)
		}

		missingAbs := filepath.Join(rootDir, manifestPath)
		if err = addTrackedFile(rootDir, storeDir, missingAbs, tracked); err != nil {
			return nil, err
		}

		if err = collectStoreEntriesForPrefix(storeDir, manifestPath, rootDir, tracked); err != nil {
			return nil, err
		}
	}

	results := make([]trackedFile, 0, len(tracked))
	for _, item := range tracked {
		results = append(results, item)
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].RelPath < results[j].RelPath
	})

	return results, nil
}

func collectFilesFromDirectory(rootDir, storeDir, directory string, tracked map[string]trackedFile) error {
	return filepath.WalkDir(directory, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk directory %q: %w", directory, walkErr)
		}
		if entry.IsDir() {
			return nil
		}
		return addTrackedFile(rootDir, storeDir, path, tracked)
	})
}

func collectStoreEntriesForPrefix(storeDir, manifestPath, rootDir string, tracked map[string]trackedFile) error {
	prefixStorePath := filepath.Join(storeDir, manifestPath)
	info, err := os.Stat(prefixStorePath)
	if err == nil && info.IsDir() {
		return filepath.WalkDir(prefixStorePath, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return fmt.Errorf("walk store directory %q: %w", prefixStorePath, walkErr)
			}
			if entry.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".enc") {
				return nil
			}

			relStorePath, relErr := filepath.Rel(storeDir, path)
			if relErr != nil {
				return fmt.Errorf("resolve store relative path %q: %w", path, relErr)
			}
			workspaceRelPath := strings.TrimSuffix(relStorePath, ".enc")
			workspaceAbsPath := filepath.Join(rootDir, workspaceRelPath)
			return addTrackedFile(rootDir, storeDir, workspaceAbsPath, tracked)
		})
	}
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("stat store prefix %q: %w", prefixStorePath, err)
	}

	directStorePath := filepath.Join(storeDir, manifestPath+".enc")
	if _, err = os.Stat(directStorePath); err == nil {
		workspaceAbsPath := filepath.Join(rootDir, manifestPath)
		if err = addTrackedFile(rootDir, storeDir, workspaceAbsPath, tracked); err != nil {
			return err
		}
	}
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("stat store file %q: %w", directStorePath, err)
	}

	return nil
}

func addTrackedFile(rootDir, storeDir, absPath string, tracked map[string]trackedFile) error {
	cleanAbs := filepath.Clean(absPath)
	relPath, err := filepath.Rel(rootDir, cleanAbs)
	if err != nil {
		return fmt.Errorf("resolve path %q relative to root %q: %w", cleanAbs, rootDir, err)
	}
	if relPath == "." {
		return fmt.Errorf("invalid workspace file path %q", cleanAbs)
	}
	if relPath == ".." || strings.HasPrefix(relPath, ".."+string(filepath.Separator)) {
		return fmt.Errorf("workspace file path %q escapes repository root", cleanAbs)
	}

	tracked[relPath] = trackedFile{
		RelPath:   relPath,
		AbsPath:   filepath.Join(rootDir, relPath),
		StorePath: filepath.Join(storeDir, relPath+".enc"),
	}
	return nil
}

func resolveBlueprintPath(configPath string) (string, error) {
	if configPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("resolve working directory: %w", err)
		}
		configPath = filepath.Join(cwd, DefaultBlueprintFile)
	}

	absPath, err := filepath.Abs(filepath.Clean(configPath))
	if err != nil {
		return "", fmt.Errorf("resolve blueprint path %q: %w", configPath, err)
	}
	return absPath, nil
}

func loadIdentityKey() (string, error) {
	identityPath, err := resolveIdentityPath()
	if err != nil {
		return "", err
	}

	raw, err := os.ReadFile(identityPath)
	if err != nil {
		return "", fmt.Errorf("read identity file %q: %w", identityPath, err)
	}
	identity := normalizeScalar(string(raw))
	zeroBytes(raw)

	if identity == "" {
		return "", fmt.Errorf("identity file %q is empty", identityPath)
	}
	return identity, nil
}

func resolveIdentityPath() (string, error) {
	if configDir := normalizeScalar(os.Getenv("ENVFUSE_CONFIG_DIR")); configDir != "" {
		return filepath.Join(filepath.Clean(configDir), "identity.key"), nil
	}

	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config directory: %w", err)
	}
	return filepath.Join(userConfigDir, "envfuse", "identity.key"), nil
}

func zeroBytes(data []byte) {
	for index := range data {
		data[index] = 0
	}
}
