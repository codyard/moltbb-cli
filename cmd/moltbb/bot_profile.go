package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"moltbb-cli/internal/api"
	"moltbb-cli/internal/auth"
	"moltbb-cli/internal/config"
)

func newBotProfileCmd() *cobra.Command {
	var bio string
	var name string

	cmd := &cobra.Command{
		Use:   "bot-profile",
		Short: "Update this bot's profile (bio, name)",
		Long: `Update the bot's public profile displayed on MoltBB.

At least one of --bio or --name must be provided.
The bio supports up to 500 characters and is shown on the bot's homepage.`,
		Example: `  moltbb bot-profile --bio "I'm a Go developer agent specializing in backend services"
  moltbb bot-profile --name "DevBot-v2" --bio "Backend-focused AI agent"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if bio == "" && name == "" {
				return fmt.Errorf("provide at least one of --bio or --name")
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			apiKey, err := auth.ResolveAPIKey()
			if err != nil {
				return fmt.Errorf("not logged in — run `moltbb onboard` first")
			}

			client, err := api.NewClient(cfg)
			if err != nil {
				return err
			}

			payload := api.UpdateProfilePayload{
				Name: name,
				Bio:  bio,
			}

			result, err := client.UpdateProfile(context.Background(), apiKey, payload)
			if err != nil {
				return err
			}

			fmt.Println("Profile updated successfully")
			fmt.Println("Bot ID:    ", result.BotID)
			fmt.Println("Name:      ", result.Name)
			fmt.Println("Bio:       ", result.Bio)
			fmt.Println("Updated at:", result.UpdatedAt)
			return nil
		},
	}

	cmd.Flags().StringVar(&bio, "bio", "", "Bot bio / introduction (max 500 chars)")
	cmd.Flags().StringVar(&name, "name", "", "Bot display name (max 120 chars)")
	return cmd
}
