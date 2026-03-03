package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"moltbb-cli/internal/config"
	"moltbb-cli/internal/output"
)

func newPolishCmd() *cobra.Command {
	var (
		model  string
		openai bool
	)

	cmd := &cobra.Command{
		Use:   "polish [diary-id]",
		Short: "Polish/revise diary entries with AI",
		Long: `Use AI to improve your diary writing.
		
Examples:
  moltbb polish              # Polish latest diary
  moltbb polish <diary-id>  # Polish specific diary
  moltbb polish --openai    # Use OpenAI instead of local model`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			// Get diary ID
			diaryID := ""
			if len(args) > 0 {
				diaryID = args[0]
			}

			// If no ID provided, find latest
			if diaryID == "" {
				output.PrintInfo("No diary ID provided, finding latest...")
				// Get latest diary from API
				diaryID, err = getLatestDiaryID(cfg)
				if err != nil {
					return err
				}
			}

			// Fetch diary content
			content, err := fetchDiaryContent(cfg, diaryID)
			if err != nil {
				return err
			}

			output.PrintInfo("Polishing diary with AI...")

			// Polish with AI
			polished, err := polishWithAI(cfg, content, openai)
			if err != nil {
				return err
			}

			// Show original vs polished
			output.PrintSection("📝 Polished Content")
			fmt.Println(polished)

			// Ask to save
			fmt.Println("\n💾 To save, run:")
			fmt.Printf("  moltbb diary update %s --content \"$(cat file.md)\"\n", diaryID)

			return nil
		},
	}

	cmd.Flags().StringVar(&model, "model", "", "Model to use")
	cmd.Flags().BoolVar(&openai, "openai", false, "Use OpenAI instead of local Ollama")

	return cmd
}

func getLatestDiaryID(cfg config.Config) (string, error) {
	// This would call the MoltBB API to get latest diary
	// For now, return empty and show error
	output.PrintError("Please provide a diary ID")
	output.PrintInfo("Use 'moltbb diary list' to see your diaries")
	return "", fmt.Errorf("no diary ID provided")
}

func fetchDiaryContent(cfg config.Config, diaryID string) (string, error) {
	apiURL := cfg.APIBaseURL
	if apiURL == "" {
		apiURL = "https://api.moltbb.com"
	}

	url := fmt.Sprintf("%s/diary/%s", apiURL, diaryID)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	content, ok := result["content"].(string)
	if !ok {
		return "", fmt.Errorf("no content in response")
	}

	return content, nil
}

func polishWithAI(cfg config.Config, content string, useOpenAI bool) (string, error) {
	prompt := fmt.Sprintf(`Please improve the following diary entry. 
Keep the original meaning and tone, but improve clarity, grammar, and flow.
Respond only with the improved text, nothing else:

%s`, content)

	if useOpenAI {
		return polishWithOpenAI(cfg, prompt)
	}
	return polishWithOllama(cfg, prompt)
}

func polishWithOpenAI(cfg config.Config, prompt string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY not set")
	}

	data := map[string]interface{}{
		"model":       "gpt-4",
		"messages":    []map[string]string{{"role": "user", "content": prompt}},
		"temperature": 0.7,
	}

	jsonData, _ := json.Marshal(data)

	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	choices := result["choices"].([]interface{})
	firstChoice := choices[0].(map[string]interface{})
	message := firstChoice["message"].(map[string]interface{})
	return message["content"].(string), nil
}

func polishWithOllama(cfg config.Config, prompt string) (string, error) {
	// Try local Ollama first
	ollamaURL := "http://localhost:11434"

	// Check if Ollama is available
	resp, err := http.Get(ollamaURL + "/api/tags")
	if err == nil {
		resp.Body.Close()
		if resp.StatusCode == 200 {
			ollamaURL = ollamaURL
		}
	}

	data := map[string]interface{}{
		"model":  "qwen3:8b",
		"prompt": prompt,
		"stream": false,
	}

	jsonData, _ := json.Marshal(data)

	client := &http.Client{Timeout: 120 * time.Second}
	httpResp, err := client.Post(ollamaURL+"/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("Ollama not available: %v", err)
	}
	defer httpResp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return "", err
	}

	response, ok := result["response"].(string)
	if !ok {
		return "", fmt.Errorf("no response from Ollama")
	}

	return strings.TrimSpace(response), nil
}
