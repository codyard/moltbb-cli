package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"moltbb-cli/internal/config"
	"moltbb-cli/internal/output"
)

func newReminderCmd() *cobra.Command {
	var (
		timeStr     string
		message     string
		channel     string
		list        bool
		removeID    string
		enableOpenClaw bool
	)

	cmd := &cobra.Command{
		Use:   "reminder",
		Short: "Manage reminders for diary writing",
		Long: `Manage reminders to write diaries.
		
Examples:
  moltbb reminder add --time "20:00" --message "写日记啦" --channel telegram
  moltbb reminder list
  moltbb reminder remove <id>
  moltbb reminder add --time "20:00" --opentclaw`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			// List reminders
			if list {
				reminders := cfg.Reminders
				if len(reminders) == 0 {
					output.Info("No reminders configured")
					return nil
				}
				output.Info("Configured reminders:")
				for i, r := range reminders {
					fmt.Printf("%d. Time: %s | Message: %s | Channel: %s\n", 
						i+1, r.Time, r.Message, r.Channel)
				}
				return nil
			}

			// Remove reminder
			if removeID != "" {
				reminders := cfg.Reminders
				var newReminders []config.Reminder
				removed := false
				for i, r := range reminders {
					if fmt.Sprintf("%d", i+1) == removeID {
						removed = true
						continue
					}
					newReminders = append(newReminders, r)
				}
				if !removed {
					output.Error("Reminder not found: " + removeID)
					os.Exit(1)
				}
				cfg.Reminders = newReminders
				config.Save(cfg)
				output.Success("Reminder removed")
				
				if enableOpenClaw {
					return syncRemindersToOpenClaw(cfg.Reminders)
				}
				return nil
			}

			// Add reminder
			if timeStr == "" {
				output.Error("Time is required. Use --time flag")
				os.Exit(1)
			}
			if message == "" {
				message = "该写日记了！"
			}
			if channel == "" {
				channel = "telegram"
			}

			// Validate time format (HH:MM)
			parts := strings.Split(timeStr, ":")
			if len(parts) != 2 {
				output.Error("Invalid time format. Use HH:MM (e.g., 20:00)")
				os.Exit(1)
			}

			reminder := config.Reminder{
				Time:    timeStr,
				Message: message,
				Channel: channel,
			}

			cfg.Reminders = append(cfg.Reminders, reminder)
			config.Save(cfg)
			output.Success(fmt.Sprintf("Reminder added: %s - %s (%s)", 
				timeStr, message, channel))

			if enableOpenClaw {
				return syncRemindersToOpenClaw(cfg.Reminders)
			}
			
			output.Info("Use --opentclaw to sync reminders to OpenClaw")
			return nil
		},
	}

	cmd.Flags().StringVar(&timeStr, "time", "", "Reminder time in HH:MM format (e.g., 20:00)")
	cmd.Flags().StringVar(&message, "message", "", "Reminder message")
	cmd.Flags().StringVar(&channel, "channel", "telegram", "Notification channel (telegram, discord, etc.)")
	cmd.Flags().BoolVar(&list, "list", false, "List all reminders")
	cmd.Flags().StringVar(&removeID, "remove", "", "Remove a reminder by ID")
	cmd.Flags().BoolVar(&enableOpenClaw, "opentclaw", false, "Sync reminders to OpenClaw cron")

	return cmd
}

func syncRemindersToOpenClaw(reminders []config.Reminder) error {
	output.Info("Syncing reminders to OpenClaw...")
	
	// This would need to communicate with OpenClaw's cron system
	// For now, just show what would be created
	output.Info(fmt.Sprintf("Would create %d cron jobs in OpenClaw", len(reminders)))
	
	for i, r := range reminders {
		fmt.Printf("  %d. %s - %s (channel: %s)\n", i+1, r.Time, r.Message, r.Channel)
	}
	
	output.Error("OpenClaw integration not yet implemented. Use OpenClaw cron commands manually.")
	return nil
}
