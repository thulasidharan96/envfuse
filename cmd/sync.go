package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/thulasidharan96/envfuse/pkg/workspace"
)

var syncConfigPath string
var syncPlainOutput bool

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize workspace files and encrypted bundles",
	RunE: func(_ *cobra.Command, _ []string) error {
		blueprintPath := syncConfigPath
		if blueprintPath != "" {
			blueprintPath = filepath.Clean(blueprintPath)
		}

		report, err := workspace.SynchronizeWorkspaceWithReport(blueprintPath)
		if report != nil {
			renderSyncReport(report, syncPlainOutput)
		}
		return err
	},
}

func renderSyncReport(report *workspace.SyncReport, plain bool) {
	var pushed int
	var pulled int
	var unchanged int
	var failed int

	for _, status := range report.Statuses {
		symbol := "•"
		summary := "up to date"

		switch status.Action {
		case workspace.SyncActionPush:
			symbol = "→"
			summary = "encrypted"
			pushed++
		case workspace.SyncActionPull:
			symbol = "✓"
			summary = "decrypted"
			pulled++
		default:
			unchanged++
		}

		if status.Err != nil {
			symbol = "❌"
			summary = status.Err.Error()
			failed++
		} else if status.Message != "" {
			summary = status.Message
		}

		if plain {
			symbol = plainSymbol(status.Action, status.Err != nil)
		}

		fmt.Fprintf(rootCmd.OutOrStdout(), "%s %s (%s)\n", symbol, status.Path, summary)
	}

	fmt.Fprintf(
		rootCmd.OutOrStdout(),
		"sync summary: pushed=%d pulled=%d unchanged=%d failed=%d\n",
		pushed,
		pulled,
		unchanged,
		failed,
	)
}

func plainSymbol(action workspace.SyncAction, hasError bool) string {
	if hasError {
		return "[ERROR]"
	}
	switch action {
	case workspace.SyncActionPush:
		return "[PUSH]"
	case workspace.SyncActionPull:
		return "[PULL]"
	default:
		return "[OK]"
	}
}

func init() {
	syncCmd.Flags().StringVar(&syncConfigPath, "config", "", "Path to workspace blueprint file (default: ./envfuse.yaml)")
	syncCmd.Flags().BoolVar(&syncPlainOutput, "plain", false, "Use ASCII-only status markers")
	rootCmd.AddCommand(syncCmd)
}
