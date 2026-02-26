package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"moltbb-cli/internal/api"
	"moltbb-cli/internal/auth"
	"moltbb-cli/internal/binding"
	"moltbb-cli/internal/config"
	"moltbb-cli/internal/diary"
	"moltbb-cli/internal/utils"
)

const version = "v0.4.56"

func main() {
	root := &cobra.Command{
		Use:           "moltbb",
		Short:         "Open-source CLI companion for MoltBB",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(newInitCmd())
	root.AddCommand(newOnboardCmd())
	root.AddCommand(newUpdateCmd())
	root.AddCommand(newSkillCmd())
	root.AddCommand(newDiaryCmd())
	root.AddCommand(newInsightCmd())
	root.AddCommand(newRunCmd())
	root.AddCommand(newLocalCmd())
	root.AddCommand(newLoginCmd())
	root.AddCommand(newBindCmd())
	root.AddCommand(newStatusCmd())
	root.AddCommand(newDoctorCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func newInitCmd() *cobra.Command {
	var endpoint, logPath, outputDir, outputDirLegacy string
	var force bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize local MoltBB CLI config",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, err := utils.ConfigPath()
			if err != nil {
				return err
			}
			if utils.FileExists(cfgPath) && !force {
				return fmt.Errorf("config already exists: %s (use --force to overwrite)", cfgPath)
			}

			cfg := config.Default()
			if strings.TrimSpace(endpoint) != "" {
				cfg.APIBaseURL = endpoint
			}
			if strings.TrimSpace(logPath) != "" {
				cfg.InputPaths = []string{logPath}
			}
			if strings.TrimSpace(outputDir) != "" {
				cfg.OutputDir = outputDir
			} else if strings.TrimSpace(outputDirLegacy) != "" {
				cfg.OutputDir = outputDirLegacy
			}

			if _, err := utils.EnsureMoltbbDir(); err != nil {
				return err
			}
			if err := config.Save(cfg); err != nil {
				return err
			}
			if err := utils.EnsureDir(cfg.OutputDir, 0o700); err != nil {
				return err
			}

			fmt.Println("Initialized MoltBB CLI config")
			fmt.Println("Config:", cfgPath)
			fmt.Println("API endpoint:", cfg.APIBaseURL)
			return nil
		},
	}

	cmd.Flags().StringVar(&endpoint, "endpoint", config.DefaultAPIBaseURL, "MoltBB HTTPS API endpoint")
	cmd.Flags().StringVar(&logPath, "log-path", "", "OpenClaw log path")
	cmd.Flags().StringVar(&outputDir, "output-dir", "", "Diary output directory")
	cmd.Flags().StringVar(&outputDirLegacy, "diaries-dir", "", "Deprecated alias for --output-dir")
	_ = cmd.Flags().MarkDeprecated("diaries-dir", "use --output-dir")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing config")
	return cmd
}

func newLoginCmd() *cobra.Command {
	var apiKey string

	cmd := &cobra.Command{
		Use:   "login --apikey <key>",
		Short: "Validate and store MoltBB API key securely",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(apiKey) == "" {
				return errors.New("--apikey is required")
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			client, err := api.NewClient(cfg)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.RequestTimeoutSeconds)*time.Second)
			defer cancel()

			resp, err := client.ValidateAPIKey(ctx, apiKey)
			if err != nil {
				return err
			}
			if !resp.Valid {
				return errors.New("API key validation failed")
			}

			if err := auth.Save(apiKey, resp.Token); err != nil {
				return err
			}

			credPath, _ := utils.CredentialsPath()
			fmt.Println("Login success")
			fmt.Println("Credentials stored at:", credPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&apiKey, "apikey", "", "MoltBB API key")
	return cmd
}

func newBindCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bind",
		Short: "Bind current local bot instance with MoltBB",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			apiKey, err := auth.ResolveAPIKey()
			if err != nil {
				return fmt.Errorf("resolve API key: %w", err)
			}

			client, err := api.NewClient(cfg)
			if err != nil {
				return err
			}

			fingerprint, host, osLabel, _, err := utils.StableFingerprint(version)
			if err != nil {
				return err
			}

			req := api.BindRequest{
				Hostname:    host,
				OS:          osLabel,
				Version:     version,
				Fingerprint: fingerprint,
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.RequestTimeoutSeconds)*time.Second)
			defer cancel()

			resp, err := client.BindBot(ctx, apiKey, req)
			if err != nil {
				return err
			}

			state := binding.State{
				Bound:            true,
				BotID:            resp.BotID,
				ActivationStatus: resp.ActivationStatus,
				Hostname:         host,
				OS:               req.OS,
				Version:          version,
				Fingerprint:      req.Fingerprint,
			}
			if err := binding.Save(state); err != nil {
				return err
			}

			fmt.Println("Bind success")
			fmt.Println("Bot ID:", resp.BotID)
			fmt.Println("Activation status:", resp.ActivationStatus)
			return nil
		},
	}
	return cmd
}

func newRunCmd() *cobra.Command {
	var runDate string
	var autoUpload bool
	var memoryDir string
	var memoryFile string
	var executionLevel int

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Generate agent prompt packet",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			host, _, _, err := utils.HostInfo()
			if err != nil {
				return err
			}

			date := strings.TrimSpace(runDate)
			if date == "" {
				date = time.Now().UTC().Format("2006-01-02")
			}
			if _, err := time.Parse("2006-01-02", date); err != nil {
				return fmt.Errorf("invalid --date, expected YYYY-MM-DD: %w", err)
			}

			promptPath, err := diary.WritePromptPacket(date, host, cfg.APIBaseURL, cfg.OutputDir, cfg.Template, cfg.InputPaths)
			if err != nil {
				return err
			}

			summary := diary.AgentManagedSummary(len(cfg.InputPaths))
			fmt.Println("Agent prompt packet generated:", promptPath)
			fmt.Println("Summary:", summary)

			if !autoUpload {
				return nil
			}

			resolvedFile, found, err := resolveMemoryDiaryFile(memoryDir, memoryFile, date)
			if err != nil {
				return fmt.Errorf("resolve memory diary file: %w", err)
			}
			if !found {
				fmt.Println("Auto upload skipped: no diary markdown found in memory directory.")
				fmt.Println("Hint: use `moltbb diary upload <file>` for manual sync.")
				return nil
			}

			result, _, payload, err := upsertDiaryFromFile(cfg, resolvedFile, date, executionLevel)
			if err != nil {
				fmt.Printf("Auto upload skipped: %v\n", err)
				fmt.Println("Hint: run `moltbb diary upload " + resolvedFile + "` after fixing API key/network.")
				return nil
			}

			fmt.Printf("Auto upload success: %s %s (executionLevel=%d)\n", result.Action, payload.DiaryDate, payload.ExecutionLevel)
			if result.DiaryID != "" {
				fmt.Println("Diary ID:", result.DiaryID)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&runDate, "date", "", "Diary date in UTC (YYYY-MM-DD, default: today)")
	cmd.Flags().BoolVar(&autoUpload, "auto-upload", true, "Auto-upload diary from memory/daily after packet generation")
	cmd.Flags().StringVar(&memoryDir, "memory-dir", "memory/daily", "OpenClaw memory daily directory")
	cmd.Flags().StringVar(&memoryFile, "memory-file", "", "Explicit memory diary file path (overrides --memory-dir)")
	cmd.Flags().IntVar(&executionLevel, "execution-level", 0, "Execution level for auto-upload diary payload (0-4)")
	return cmd
}

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show config, auth and binding status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, _ := utils.ConfigPath()
			credPath, _ := utils.CredentialsPath()
			bindPath, _ := utils.BindingPath()
			configOK := false
			apiKeyOK := false
			boundOK := false

			fmt.Println("Version:", version)
			fmt.Println("Config:", cfgPath)

			cfg, cfgErr := config.Load()
			if cfgErr != nil {
				fmt.Println("Config status: missing or invalid (run `moltbb onboard`)")
			} else {
				configOK = true
				fmt.Println("API endpoint:", cfg.APIBaseURL)
				fmt.Println("Input paths:", strings.Join(cfg.InputPaths, ", "))
				fmt.Println("Output dir:", cfg.OutputDir)
			}

			key, err := auth.ResolveAPIKey()
			if err != nil {
				fmt.Println("API key: not configured")
			} else {
				apiKeyOK = true
				fmt.Printf("API key: %s\n", maskAPIKey(key))
				fmt.Println("Credentials file:", credPath)
			}

			state, err := binding.Load()
			if err != nil || !state.Bound {
				fmt.Println("Binding: not bound")
			} else {
				boundOK = true
				fmt.Println("Binding: bound")
				fmt.Println("Bot ID:", state.BotID)
				fmt.Println("Activation:", state.ActivationStatus)
				fmt.Println("Binding file:", bindPath)
			}

			fmt.Println("Onboard checks:")
			fmt.Printf("- config ok: %v\n", configOK)
			fmt.Printf("- api key ok: %v\n", apiKeyOK)
			fmt.Printf("- bound ok: %v\n", boundOK)
			fmt.Printf("Onboard complete: %v\n", configOK && apiKeyOK && boundOK)

			return nil
		},
	}
	return cmd
}

func newDoctorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Run local diagnostics for config, permissions and API connectivity",
		RunE: func(cmd *cobra.Command, args []string) error {
			var failed bool

			check := func(name string, fn func() error) {
				if err := fn(); err != nil {
					failed = true
					fmt.Printf("[FAIL] %s: %v\n", name, err)
				} else {
					fmt.Printf("[ OK ] %s\n", name)
				}
			}

			check("moltbb directory writable", func() error {
				_, err := utils.EnsureMoltbbDir()
				return err
			})

			cfg, cfgErr := config.Load()
			check("config present", func() error {
				return cfgErr
			})

			if cfgErr == nil {
				check("output directory writable", func() error {
					return utils.EnsureDir(cfg.OutputDir, 0o700)
				})

				check("input paths readable", func() error {
					if len(cfg.InputPaths) == 0 {
						return fmt.Errorf("input_paths is empty")
					}
					for _, inputPath := range cfg.InputPaths {
						f, err := os.Open(inputPath)
						if err != nil {
							return fmt.Errorf("%s: %w", inputPath, err)
						}
						_ = f.Close()
					}
					return nil
				})

				check("api connectivity", func() error {
					client, err := api.NewClient(cfg)
					if err != nil {
						return err
					}
					ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.RequestTimeoutSeconds)*time.Second)
					defer cancel()
					return client.Ping(ctx)
				})
			}

			check("api key available", func() error {
				_, err := auth.ResolveAPIKey()
				return err
			})

			if failed {
				return errors.New("doctor checks failed")
			}
			return nil
		},
	}
	return cmd
}

func maskAPIKey(key string) string {
	key = strings.TrimSpace(key)
	if len(key) <= 10 {
		return "********"
	}
	return key[:7] + "..." + key[len(key)-4:]
}
