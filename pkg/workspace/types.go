package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const DefaultBlueprintFile = "envfuse.yaml"

type TeamMember struct {
	Name      string `yaml:"name"`
	Recipient string `yaml:"recipient"`
	PublicKey string `yaml:"public_key"`
	Key       string `yaml:"key"`
}

type Blueprint struct {
	Team     []TeamMember `yaml:"team"`
	Manifest []string     `yaml:"manifest"`
}

func (m TeamMember) RecipientKey() string {
	if key := normalizeScalar(m.Recipient); key != "" {
		return key
	}
	if key := normalizeScalar(m.PublicKey); key != "" {
		return key
	}
	return normalizeScalar(m.Key)
}

func (m TeamMember) DisplayName() string {
	return normalizeScalar(m.Name)
}

func LoadBlueprint(configPath string) (*Blueprint, error) {
	if configPath == "" {
		return nil, fmt.Errorf("blueprint path is required")
	}

	absPath, err := filepath.Abs(filepath.Clean(configPath))
	if err != nil {
		return nil, fmt.Errorf("resolve blueprint path: %w", err)
	}

	raw, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("read blueprint %q: %w", absPath, err)
	}

	decoder := yaml.NewDecoder(strings.NewReader(string(raw)))
	decoder.KnownFields(true)

	var blueprint Blueprint
	if err = decoder.Decode(&blueprint); err != nil {
		return nil, fmt.Errorf("parse blueprint %q: %w", absPath, err)
	}

	normalized, err := normalizeBlueprint(blueprint)
	if err != nil {
		return nil, err
	}
	return &normalized, nil
}

func normalizeBlueprint(blueprint Blueprint) (Blueprint, error) {
	if len(blueprint.Team) == 0 {
		return Blueprint{}, fmt.Errorf("blueprint validation failed: team must include at least one member")
	}
	if len(blueprint.Manifest) == 0 {
		return Blueprint{}, fmt.Errorf("blueprint validation failed: manifest must include at least one path")
	}

	normalized := Blueprint{
		Team:     make([]TeamMember, 0, len(blueprint.Team)),
		Manifest: make([]string, 0, len(blueprint.Manifest)),
	}

	for index, member := range blueprint.Team {
		name := member.DisplayName()
		if name == "" {
			return Blueprint{}, fmt.Errorf("blueprint validation failed: team[%d].name is required", index)
		}

		recipient := member.RecipientKey()
		if recipient == "" {
			return Blueprint{}, fmt.Errorf("blueprint validation failed: team[%d] requires recipient/public_key/key", index)
		}

		normalized.Team = append(normalized.Team, TeamMember{
			Name:      name,
			Recipient: recipient,
		})
	}

	for index, entry := range blueprint.Manifest {
		clean, err := sanitizeManifestPath(entry)
		if err != nil {
			return Blueprint{}, fmt.Errorf("blueprint validation failed for manifest[%d]: %w", index, err)
		}
		normalized.Manifest = append(normalized.Manifest, clean)
	}

	return normalized, nil
}

func sanitizeManifestPath(entry string) (string, error) {
	trimmed := normalizeScalar(entry)
	if trimmed == "" {
		return "", fmt.Errorf("path must not be empty")
	}
	if filepath.IsAbs(trimmed) {
		return "", fmt.Errorf("path %q must be relative", trimmed)
	}

	clean := filepath.Clean(trimmed)
	if clean == "." {
		return "", fmt.Errorf("path %q is not a valid target", trimmed)
	}
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q escapes repository root", trimmed)
	}

	return clean, nil
}

func normalizeScalar(value string) string {
	value = strings.ReplaceAll(value, "\r", "")
	return strings.TrimSpace(value)
}
