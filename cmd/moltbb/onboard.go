package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"moltbb-cli/internal/api"
	"moltbb-cli/internal/auth"
	"moltbb-cli/internal/binding"
	"moltbb-cli/internal/config"
	"moltbb-cli/internal/utils"
)

type onboardOptions struct {
	nonInteractive       bool
	apiBaseURL           string
	inputPaths           string
	outputDir            string
	template             string
	apiKey               string
	bind                 bool
	allowHTTP            bool
	scheduleOS           string
	generateScheduleFile bool
}

func newOnboardCmd() *cobra.Command {
	opts := onboardOptions{}

	cmd := &cobra.Command{
		Use:   "onboard",
		Short: "Guided onboarding for config, credentials and binding",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOnboard(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.nonInteractive, "non-interactive", false, "Run onboarding with flags only (CI-friendly)")
	cmd.Flags().StringVar(&opts.apiBaseURL, "api-base-url", "", "MoltBB API base URL")
	cmd.Flags().StringVar(&opts.inputPaths, "input-paths", "", "Comma-separated OpenClaw input log paths")
	cmd.Flags().StringVar(&opts.outputDir, "output-dir", "", "Diary output directory")
	cmd.Flags().StringVar(&opts.template, "template", "", "Optional diary template name")
	cmd.Flags().StringVar(&opts.apiKey, "apikey", "", "Bot API key")
	cmd.Flags().BoolVar(&opts.bind, "bind", false, "Bind/activate this machine after validating API key")
	cmd.Flags().BoolVar(&opts.allowHTTP, "allow-http", false, "Allow insecure http endpoint")
	cmd.Flags().StringVar(&opts.scheduleOS, "schedule-os", "", "Scheduling target OS: linux|macos|windows")
	cmd.Flags().BoolVar(&opts.generateScheduleFile, "generate-schedule-files", false, "Generate scheduling snippets into ~/.moltbb/examples")

	return cmd
}

func runOnboard(opts onboardOptions) error {
	cfgPath, _ := utils.ConfigPath()
	credPath, _ := utils.CredentialsPath()
	bindPath, _ := utils.BindingPath()

	existingCfg, cfgErr := config.Load()
	cfgExists := cfgErr == nil
	if cfgErr != nil {
		existingCfg = config.Default()
	}

	var existingCred *auth.Credentials
	if cred, err := auth.Load(); err == nil {
		existingCred = &cred
	}

	var existingBinding *binding.State
	if state, err := binding.Load(); err == nil {
		existingBinding = &state
	}

	fmt.Println("MoltBB onboard wizard")
	fmt.Println("Detected files:")
	fmt.Printf("- config: %s (%v)\n", cfgPath, cfgErr == nil)
	fmt.Printf("- credentials: %s (%v)\n", credPath, existingCred != nil)
	fmt.Printf("- binding: %s (%v)\n", bindPath, existingBinding != nil)

	if opts.nonInteractive {
		return runOnboardNonInteractive(opts, existingCfg, cfgExists, existingCred, existingBinding)
	}
	return runOnboardInteractive(opts, existingCfg, existingCred, existingBinding)
}

func runOnboardInteractive(opts onboardOptions, cfg config.Config, existingCred *auth.Credentials, existingBinding *binding.State) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\nStep A: Server endpoint")
	defaultEndpoint := cfg.APIBaseURL
	if strings.TrimSpace(opts.apiBaseURL) != "" {
		defaultEndpoint = strings.TrimSpace(opts.apiBaseURL)
	}

	var apiBaseURL string
	allowHTTP := cfg.AllowInsecureHTTP || opts.allowHTTP
	for {
		value, err := utils.PromptString(reader, "MoltBB API base URL", defaultEndpoint)
		if err != nil {
			return err
		}
		value = strings.TrimSpace(value)
		if strings.HasPrefix(value, "https://") {
			apiBaseURL = strings.TrimRight(value, "/")
			allowHTTP = false
			break
		}
		if strings.HasPrefix(value, "http://") {
			confirmed, err := utils.PromptYesNo(reader, "HTTP is insecure. Force continue with http?", false)
			if err != nil {
				return err
			}
			if confirmed {
				apiBaseURL = strings.TrimRight(value, "/")
				allowHTTP = true
				break
			}
			continue
		}
		fmt.Println("Invalid URL. Expected http(s)://...")
	}

	fmt.Println("\nStep B: Local diary settings")
	inputDefault := strings.Join(cfg.InputPaths, ",")
	if strings.TrimSpace(opts.inputPaths) != "" {
		inputDefault = opts.inputPaths
	}

	var inputPaths []string
	for {
		raw, err := utils.PromptString(reader, "input_paths (comma-separated)", inputDefault)
		if err != nil {
			return err
		}
		inputPaths = config.ParseInputPathsCSV(raw)
		if len(inputPaths) > 0 {
			break
		}
		fmt.Println("At least one input path is required.")
	}
	warnMissingInputPaths(inputPaths)

	outputDefault := cfg.OutputDir
	if strings.TrimSpace(opts.outputDir) != "" {
		outputDefault = strings.TrimSpace(opts.outputDir)
	}
	outputDir, err := utils.PromptString(reader, "output_dir", outputDefault)
	if err != nil {
		return err
	}
	warnOutputDir(outputDir)

	templateDefault := cfg.Template
	if strings.TrimSpace(opts.template) != "" {
		templateDefault = strings.TrimSpace(opts.template)
	}
	template, err := utils.PromptString(reader, "template (optional)", templateDefault)
	if err != nil {
		return err
	}

	cfg.APIBaseURL = apiBaseURL
	cfg.AllowInsecureHTTP = allowHTTP
	cfg.InputPaths = inputPaths
	cfg.OutputDir = outputDir
	cfg.Template = strings.TrimSpace(template)

	if err := config.Save(cfg); err != nil {
		return err
	}

	fmt.Println("\nStep C: Credentials (API key)")
	keyValidated := false
	validatedKey := ""
	if existingCred != nil {
		fmt.Printf("Current API key: %s\n", maskAPIKey(existingCred.APIKey))
	}

	setOrUpdate, err := utils.PromptYesNo(reader, "Do you want to set/update Bot API key now?", false)
	if err != nil {
		return err
	}

	client, err := api.NewClient(cfg)
	if err != nil {
		return err
	}

	if setOrUpdate {
		for {
			inputKey, promptErr := utils.PromptSecret(reader, "Bot API key")
			if promptErr != nil {
				return promptErr
			}
			if strings.TrimSpace(inputKey) == "" {
				fmt.Println("API key is empty.")
				continue
			}
			resp, validateErr := validateAPIKey(client, cfg, inputKey)
			if validateErr == nil && resp.Valid {
				if saveErr := auth.Save(inputKey, resp.Token); saveErr != nil {
					return saveErr
				}
				validatedKey = inputKey
				keyValidated = true
				fmt.Println("API key validated and stored.")
				break
			}

			fmt.Println("API key validation failed.")
			retry, retryErr := utils.PromptYesNo(reader, "Retry API key input?", true)
			if retryErr != nil {
				return retryErr
			}
			if !retry {
				break
			}
		}
	} else if existingCred != nil {
		validateExisting, promptErr := utils.PromptYesNo(reader, "Validate existing API key now (needed before bind)?", true)
		if promptErr != nil {
			return promptErr
		}
		if validateExisting {
			resp, validateErr := validateAPIKey(client, cfg, existingCred.APIKey)
			if validateErr == nil && resp.Valid {
				validatedKey = existingCred.APIKey
				keyValidated = true
				fmt.Println("Existing API key validated.")
			} else {
				fmt.Println("Existing API key validation failed, binding will be skipped.")
			}
		}
	}

	fmt.Println("\nStep D: Binding / Activation")
	bound := existingBinding != nil && existingBinding.Bound
	if keyValidated {
		bindNow, promptErr := utils.PromptYesNo(reader, "Bind/activate this bot on this machine now?", true)
		if promptErr != nil {
			return promptErr
		}
		if bindNow {
			state, bindErr := bindMachine(client, cfg, validatedKey)
			if bindErr != nil {
				return bindErr
			}
			if saveErr := binding.Save(state); saveErr != nil {
				return saveErr
			}
			bound = true
			fmt.Printf("Bound bot_id=%s status=%s\n", state.BotID, state.ActivationStatus)
		} else {
			if saveErr := binding.Save(binding.State{Bound: false, Version: version}); saveErr != nil {
				return saveErr
			}
			bound = false
			fmt.Println("Binding marked as not bound.")
		}
	} else {
		fmt.Println("Skipped binding because no validated API key is available.")
	}

	fmt.Println("\nStep E: Scheduling guidance")
	detectedOS := detectedScheduleOS()
	selectedOS, err := utils.PromptString(reader, "Scheduling OS (linux/macos/windows)", detectedOS)
	if err != nil {
		return err
	}
	selectedOS = normalizeScheduleOS(selectedOS)
	if selectedOS == "" {
		selectedOS = detectedOS
	}
	printScheduleSnippet(selectedOS)

	generateFiles, promptErr := utils.PromptYesNo(reader, "Generate scheduling example files in ~/.moltbb/examples?", false)
	if promptErr != nil {
		return promptErr
	}
	if generateFiles {
		path, genErr := generateScheduleExamples(selectedOS)
		if genErr != nil {
			return genErr
		}
		fmt.Println("Generated scheduling examples in:", path)
	}

	fmt.Println("\nStep F: Final summary")
	return printOnboardSummary(cfg, keyValidated || existingCred != nil, bound)
}

func runOnboardNonInteractive(opts onboardOptions, cfg config.Config, cfgExists bool, existingCred *auth.Credentials, existingBinding *binding.State) error {
	if !cfgExists {
		if strings.TrimSpace(opts.apiBaseURL) == "" {
			return errors.New("non-interactive mode requires --api-base-url when config does not exist")
		}
		if strings.TrimSpace(opts.inputPaths) == "" {
			return errors.New("non-interactive mode requires --input-paths when config does not exist")
		}
		if strings.TrimSpace(opts.outputDir) == "" {
			return errors.New("non-interactive mode requires --output-dir when config does not exist")
		}
	}

	apiBaseURL := strings.TrimSpace(opts.apiBaseURL)
	if apiBaseURL == "" {
		if strings.TrimSpace(cfg.APIBaseURL) == "" {
			return errors.New("non-interactive mode requires --api-base-url (or existing config)")
		}
		apiBaseURL = cfg.APIBaseURL
	}

	allowHTTP := opts.allowHTTP || cfg.AllowInsecureHTTP
	if strings.HasPrefix(apiBaseURL, "http://") && !allowHTTP {
		return errors.New("http endpoint requires --allow-http in non-interactive mode")
	}
	if !strings.HasPrefix(apiBaseURL, "https://") && !strings.HasPrefix(apiBaseURL, "http://") {
		return errors.New("--api-base-url must be http(s)")
	}

	inputPaths := config.ParseInputPathsCSV(opts.inputPaths)
	if len(inputPaths) == 0 {
		inputPaths = cfg.InputPaths
	}
	if len(inputPaths) == 0 {
		return errors.New("non-interactive mode requires --input-paths (or existing config input_paths)")
	}

	outputDir := strings.TrimSpace(opts.outputDir)
	if outputDir == "" {
		outputDir = cfg.OutputDir
	}
	if outputDir == "" {
		return errors.New("non-interactive mode requires --output-dir (or existing config output_dir)")
	}

	if strings.TrimSpace(opts.template) != "" {
		cfg.Template = strings.TrimSpace(opts.template)
	}

	cfg.APIBaseURL = strings.TrimRight(apiBaseURL, "/")
	cfg.AllowInsecureHTTP = allowHTTP
	cfg.InputPaths = inputPaths
	cfg.OutputDir = outputDir

	if err := config.Save(cfg); err != nil {
		return err
	}

	warnMissingInputPaths(inputPaths)
	warnOutputDir(outputDir)

	client, err := api.NewClient(cfg)
	if err != nil {
		return err
	}

	apiKey := strings.TrimSpace(opts.apiKey)
	if apiKey == "" && existingCred != nil {
		apiKey = existingCred.APIKey
	}

	keyReady := false
	if apiKey != "" {
		resp, validateErr := validateAPIKey(client, cfg, apiKey)
		if validateErr != nil || !resp.Valid {
			return fmt.Errorf("api key validation failed in non-interactive mode")
		}
		if saveErr := auth.Save(apiKey, resp.Token); saveErr != nil {
			return saveErr
		}
		keyReady = true
	}

	bound := existingBinding != nil && existingBinding.Bound
	if opts.bind {
		if !keyReady {
			return errors.New("--bind requires a valid API key (--apikey or existing credentials)")
		}
		state, bindErr := bindMachine(client, cfg, apiKey)
		if bindErr != nil {
			return bindErr
		}
		if saveErr := binding.Save(state); saveErr != nil {
			return saveErr
		}
		bound = true
	}

	selectedOS := normalizeScheduleOS(opts.scheduleOS)
	if selectedOS == "" {
		selectedOS = detectedScheduleOS()
	}
	printScheduleSnippet(selectedOS)
	if opts.generateScheduleFile {
		path, genErr := generateScheduleExamples(selectedOS)
		if genErr != nil {
			return genErr
		}
		fmt.Println("Generated scheduling examples in:", path)
	}

	return printOnboardSummary(cfg, keyReady || existingCred != nil, bound)
}

func validateAPIKey(client *api.Client, cfg config.Config, apiKey string) (api.ValidateResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.RequestTimeoutSeconds)*time.Second)
	defer cancel()
	return client.ValidateAPIKey(ctx, apiKey)
}

func bindMachine(client *api.Client, cfg config.Config, apiKey string) (binding.State, error) {
	fingerprint, hostname, osLabel, _, err := utils.StableFingerprint(version)
	if err != nil {
		return binding.State{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.RequestTimeoutSeconds)*time.Second)
	defer cancel()

	resp, err := client.BindBot(ctx, apiKey, api.BindRequest{
		Hostname:    hostname,
		OS:          osLabel,
		Version:     version,
		Fingerprint: fingerprint,
	})
	if err != nil {
		return binding.State{}, err
	}

	return binding.State{
		Bound:            true,
		BotID:            resp.BotID,
		ActivationStatus: resp.ActivationStatus,
		Hostname:         hostname,
		OS:               osLabel,
		Version:          version,
		Fingerprint:      fingerprint,
	}, nil
}

func warnMissingInputPaths(paths []string) {
	for _, p := range paths {
		expanded, err := utils.ExpandPath(p)
		if err != nil {
			fmt.Printf("[WARN] input path invalid: %s (%v)\n", p, err)
			continue
		}
		if _, err := os.Stat(expanded); err != nil {
			fmt.Printf("[WARN] input path missing: %s\n", expanded)
		}
	}
}

func warnOutputDir(outputDir string) {
	expanded, err := utils.ExpandPath(outputDir)
	if err != nil {
		fmt.Printf("[WARN] output_dir invalid: %v\n", err)
		return
	}
	if err := utils.EnsureDir(expanded, 0o700); err != nil {
		fmt.Printf("[WARN] output_dir not writable: %s (%v)\n", expanded, err)
	}
}

func printOnboardSummary(cfg config.Config, hasKey, bound bool) error {
	cfgPath, _ := utils.ConfigPath()
	fmt.Println("Onboard summary:")
	fmt.Println("- config path:", cfgPath)
	fmt.Println("- output_dir:", cfg.OutputDir)
	fmt.Printf("- api key configured: %v\n", hasKey)
	fmt.Printf("- bound: %v\n", bound)
	fmt.Println("- next: moltbb run")
	return nil
}

func detectedScheduleOS() string {
	switch runtime.GOOS {
	case "darwin":
		return "macos"
	case "windows":
		return "windows"
	default:
		return "linux"
	}
}

func normalizeScheduleOS(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "linux":
		return "linux"
	case "mac", "macos", "darwin":
		return "macos"
	case "windows", "win":
		return "windows"
	default:
		return ""
	}
}

func printScheduleSnippet(osType string) {
	fmt.Println("Scheduling snippets:")
	fmt.Println(scheduleSnippet(osType))
}

func scheduleSnippet(osType string) string {
	switch osType {
	case "macos":
		return `macOS launchd (~/Library/LaunchAgents/com.moltbb.run.plist):
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key><string>com.moltbb.run</string>
  <key>ProgramArguments</key>
  <array><string>/usr/local/bin/moltbb</string><string>run</string></array>
  <key>StartCalendarInterval</key>
  <dict><key>Hour</key><integer>21</integer><key>Minute</key><integer>0</integer></dict>
  <key>StandardOutPath</key><string>~/.moltbb/logs/launchd.out.log</string>
  <key>StandardErrorPath</key><string>~/.moltbb/logs/launchd.err.log</string>
</dict>
</plist>

Load command:
launchctl load ~/Library/LaunchAgents/com.moltbb.run.plist`
	case "windows":
		return `Windows Task Scheduler (PowerShell):
$Action = New-ScheduledTaskAction -Execute "moltbb" -Argument "run"
$Trigger = New-ScheduledTaskTrigger -Daily -At 9:00PM
Register-ScheduledTask -TaskName "MoltBBDiary" -Action $Action -Trigger $Trigger`
	default:
		return `Linux cron:
0 21 * * * /usr/local/bin/moltbb run >> ~/.moltbb/logs/cron.log 2>&1`
	}
}

func generateScheduleExamples(osType string) (string, error) {
	base, err := utils.MoltbbDir()
	if err != nil {
		return "", err
	}
	examplesDir := filepath.Join(base, "examples")
	if err := utils.EnsureDir(examplesDir, 0o700); err != nil {
		return "", err
	}

	files := map[string]string{}
	switch osType {
	case "macos":
		files["launchd.plist"] = scheduleSnippet("macos")
	case "windows":
		files["task-scheduler.ps1"] = scheduleSnippet("windows")
	default:
		files["cron.txt"] = scheduleSnippet("linux")
	}

	for name, content := range files {
		if err := utils.SecureWriteFile(filepath.Join(examplesDir, name), []byte(content+"\n"), 0o600); err != nil {
			return "", err
		}
	}

	return examplesDir, nil
}
