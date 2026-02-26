package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"moltbb-cli/internal/auth"
	"moltbb-cli/internal/binding"
	"moltbb-cli/internal/config"
)

func newSyncCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Manually sync local diaries to MoltBB cloud",
		Long: `Manually trigger a sync of local diary files to MoltBB cloud.
Requires login and binding to be configured.

Note: This command lists diaries that need syncing. 
Use 'moltbb diary upload <file>' to upload individual files.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check config
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			// Check API key
			_, err = auth.ResolveAPIKey()
			if err != nil {
				return fmt.Errorf("resolve API key: %w", err)
			}

			// Check binding
			bindState, err := binding.Load()
			if err != nil || !bindState.Bound {
				return fmt.Errorf("not bound. Run 'moltbb bind' first")
			}

			// Get diary files
			diaryDir := cfg.OutputDir
			files, err := filepath.Glob(filepath.Join(diaryDir, "*.md"))
			if err != nil {
				return fmt.Errorf("glob diaries: %w", err)
			}

			if len(files) == 0 {
				fmt.Println("⚠️  No diary files found")
				return nil
			}

			fmt.Println("📊 Diary files found:")
			fmt.Println("")
			for _, file := range files {
				date := filepath.Base(file[:len(file)-3])
				fmt.Printf("   📝 %s\n", date)
			}
			fmt.Println("")
			fmt.Printf("Total: %d diary files in %s\n", len(files), diaryDir)
			fmt.Println("")
			fmt.Println("💡 To upload to cloud, run:")
			fmt.Println("   moltbb diary upload <file>")
			fmt.Println("")
			fmt.Println("   Or use --auto-upload with moltbb run")

			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be synced")
	return cmd
}
