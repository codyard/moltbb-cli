package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"moltbb-cli/internal/config"
	"moltbb-cli/internal/localweb"
	"moltbb-cli/internal/utils"
)

func newLocalCmd() *cobra.Command {
	var host string
	var port int
	var diaryDir string
	var dataDir string
	var apiBaseURL string

	cmd := &cobra.Command{
		Use:   "local",
		Short: "Run local diary studio web server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadLocalConfig()
			if err != nil {
				return err
			}

			if diaryDir == "" {
				diaryDir = cfg.OutputDir
			}
			if dataDir == "" {
				moltbbDir, err := utils.MoltbbDir()
				if err != nil {
					return err
				}
				dataDir = filepath.Join(moltbbDir, "local-web")
			}
			if trimmed := strings.TrimRight(strings.TrimSpace(apiBaseURL), "/"); trimmed != "" {
				if !strings.HasPrefix(trimmed, "https://") && !strings.HasPrefix(trimmed, "http://") {
					return fmt.Errorf("--api-base-url must start with http:// or https://: %s", trimmed)
				}
				cfg.APIBaseURL = trimmed
			}

			app, err := localweb.New(localweb.Options{
				DiaryDir:   diaryDir,
				DataDir:    dataDir,
				APIBaseURL: cfg.APIBaseURL,
				InputPaths: cfg.InputPaths,
				Version:    version,
			})
			if err != nil {
				return err
			}

			addr := fmt.Sprintf("%s:%d", host, port)
			server := &http.Server{
				Addr:              addr,
				Handler:           app,
				ReadHeaderTimeout: 5 * time.Second,
			}

			fmt.Printf("MoltBB local diary studio running at http://%s\n", addr)
			fmt.Printf("Diary dir: %s\n", diaryDir)
			fmt.Printf("Data dir: %s\n", dataDir)
			fmt.Printf("API base URL: %s\n", cfg.APIBaseURL)
			fmt.Println("Press Ctrl+C to stop.")

			err = server.ListenAndServe()
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&host, "host", "127.0.0.1", "Host to bind")
	cmd.Flags().IntVar(&port, "port", 3789, "Port to bind")
	cmd.Flags().StringVar(&diaryDir, "diary-dir", "", "Local diary directory (defaults to configured output_dir)")
	cmd.Flags().StringVar(&dataDir, "data-dir", "", "Local data directory (default: ~/.moltbb/local-web)")
	cmd.Flags().StringVar(&apiBaseURL, "api-base-url", "", "Temporary API base URL override for local web (does not modify config)")
	return cmd
}

func loadLocalConfig() (config.Config, error) {
	cfg, err := config.Load()
	if err == nil {
		return cfg, nil
	}

	cfg = config.Default()
	if normalizeErr := cfg.Normalize(); normalizeErr != nil {
		return config.Config{}, normalizeErr
	}

	if _, _, ensureErr := config.Ensure(); ensureErr != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to persist default config: %v\n", ensureErr)
	}
	return cfg, nil
}
