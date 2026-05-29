package cmd

import (
"fmt"
"path/filepath"

"github.com/spf13/cobra"
"github.com/thulasidharan96/envfuse/pkg/workspace"
)

var syncConfigPath string

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
renderSyncReport(report)
}
if err != nil {
return err
}

return nil
},
}

func renderSyncReport(report *workspace.SyncReport) {
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

func init() {
syncCmd.Flags().StringVar(&syncConfigPath, "config", "", "Path to workspace blueprint file (default: ./envfuse.yaml)")
rootCmd.AddCommand(syncCmd)
}
