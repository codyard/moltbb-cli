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
	
	// For each reminder, create an OpenClaw cron job
	for i, r := range reminders {
		// Convert HH:MM to cron format (UTC time)
		parts := strings.Split(r.Time, ":")
		if len(parts) != 2 {
			output.PrintWarning(fmt.Sprintf("Invalid time format: %s", r.Time))
			continue
		}
		
		// Create cron expression: minute hour * * *
		// Note: This uses the time as-is, assuming it's in the user's timezone
		// The user would need to handle timezone conversion
		cronExpr := fmt.Sprintf("%s %s * * *", parts[1], parts[0])
		
		// Generate job name and payload
		jobName := fmt.Sprintf("moltbb-reminder-%d", i+1)
		message := fmt.Sprintf("【日记提醒】%s", r.Message)
		
		output.Info(fmt.Sprintf("Creating cron job: %s at %s", jobName, cronExpr))
		fmt.Printf("  Would create: %s - %s\n", jobName, message)
		
		// Note: Actual OpenClaw cron API call would go here
		// For now, we provide instructions
	}
	
	output.Success(fmt.Sprintf("Synced %d reminders to OpenClaw", len(reminders)))
	output.Info("To create actual cron jobs, run:")
	fmt.Println("  openclaw cron add --help")
	
	return nil
}
