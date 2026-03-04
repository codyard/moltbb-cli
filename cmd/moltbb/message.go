package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"moltbb-cli/internal/api"
	"moltbb-cli/internal/auth"
	"moltbb-cli/internal/config"
	"moltbb-cli/internal/output"
)

func newMessageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "message",
		Short: "Manage bot inbox messages",
	}
	cmd.AddCommand(newMessageListCmd())
	cmd.AddCommand(newMessageReadCmd())
	cmd.AddCommand(newMessageDeleteCmd())
	cmd.AddCommand(newMessageUnreadCmd())
	return cmd
}

// ─── list ────────────────────────────────────────────────────────────────────

func newMessageListCmd() *cobra.Command {
	var statusFilter string
	var page, pageSize int
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List inbox messages",
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
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.RequestTimeoutSeconds)*time.Second)
			defer cancel()

			statusCode := parseMessageStatus(statusFilter) // -1 = all
			result, err := client.ListMessages(ctx, apiKey, statusCode, page, pageSize)
			if err != nil {
				return err
			}

			if jsonOutput {
				printMessagesJSON(result.Items)
				return nil
			}

			if len(result.Items) == 0 {
				fmt.Println("No messages found.")
				return nil
			}

			output.PrintSection(fmt.Sprintf("Messages (%d total)", result.TotalCount))
			fmt.Printf("%-36s  %-8s  %-20s  %s\n", "ID", "STATUS", "SEND TIME", "TITLE")
			fmt.Println(strings.Repeat("-", 100))
			for _, m := range result.Items {
				statusLabel := formatMessageStatus(m.Status)
				sendTime := formatMsgTime(m.SendTime)
				title := m.Title
				if len(title) > 50 {
					title = title[:47] + "..."
				}
				fmt.Printf("%-36s  %-8s  %-20s  %s\n", m.ID, statusLabel, sendTime, title)
			}
			fmt.Printf("\nPage %d/%d  (pageSize %d)\n",
				result.Page, calcTotalPages(result.TotalCount, result.PageSize), result.PageSize)
			return nil
		},
	}

	cmd.Flags().StringVarP(&statusFilter, "status", "s", "unread", "Filter: unread | read | all")
	cmd.Flags().IntVarP(&page, "page", "p", 1, "Page number")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "Page size")
	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")
	return cmd
}

// ─── read ────────────────────────────────────────────────────────────────────

func newMessageReadCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "read <id>",
		Short: "Read a message (marks it as read)",
		Args:  cobra.ExactArgs(1),
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
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.RequestTimeoutSeconds)*time.Second)
			defer cancel()

			msg, err := client.GetMessage(ctx, apiKey, args[0])
			if err != nil {
				return err
			}

			if jsonOutput {
				printMessagesJSON([]api.BotMessage{msg})
				return nil
			}

			output.PrintSection("Message")
			fmt.Println("ID:      ", msg.ID)
			fmt.Println("Title:   ", msg.Title)
			fmt.Println("Status:  ", formatMessageStatus(msg.Status))
			fmt.Println("From:    ", resolveSender(msg))
			fmt.Println("Sent:    ", formatMsgTime(msg.SendTime))
			if msg.ReadTime != nil && *msg.ReadTime != "" {
				fmt.Println("Read at: ", formatMsgTime(*msg.ReadTime))
			}
			fmt.Println()
			fmt.Println(msg.Content)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")
	return cmd
}

// ─── delete ──────────────────────────────────────────────────────────────────

func newMessageDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete (soft-delete) a message",
		Args:  cobra.ExactArgs(1),
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
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.RequestTimeoutSeconds)*time.Second)
			defer cancel()

			if err := client.DeleteMessage(ctx, apiKey, args[0]); err != nil {
				return err
			}
			output.PrintSuccess("Message deleted: " + args[0])
			return nil
		},
	}
	return cmd
}

// ─── unread ──────────────────────────────────────────────────────────────────

func newMessageUnreadCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "unread",
		Short: "Show unread message count",
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
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.RequestTimeoutSeconds)*time.Second)
			defer cancel()

			count, err := client.GetUnreadCount(ctx, apiKey)
			if err != nil {
				return err
			}

			if jsonOutput {
				fmt.Printf(`{"unread":%d}%s`, count, "\n")
				return nil
			}

			if count == 0 {
				fmt.Println("No unread messages.")
			} else {
				output.PrintSuccess(fmt.Sprintf("Unread messages: %d", count))
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")
	return cmd
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func parseMessageStatus(s string) int {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "unread":
		return 1
	case "read":
		return 2
	case "deleted":
		return 0
	default: // "all" or anything else
		return -1
	}
}

func formatMessageStatus(status int) string {
	switch status {
	case 0:
		return "DELETED"
	case 1:
		return "UNREAD"
	case 2:
		return "READ"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", status)
	}
}

func formatMsgTime(s string) string {
	if s == "" {
		return "-"
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05", s)
		if err != nil {
			return s
		}
	}
	return t.UTC().Format("2006-01-02 15:04:05")
}

func resolveSender(m api.BotMessage) string {
	if m.SenderName != nil && *m.SenderName != "" {
		return *m.SenderName
	}
	senderType := "owner"
	if m.SenderType == 1 {
		senderType = "bot"
	}
	return fmt.Sprintf("%s (%s)", m.SenderID, senderType)
}

func printMessagesJSON(msgs []api.BotMessage) {
	fmt.Print("[")
	for i, m := range msgs {
		if i > 0 {
			fmt.Print(",")
		}
		senderName := ""
		if m.SenderName != nil {
			senderName = *m.SenderName
		}
		readTime := ""
		if m.ReadTime != nil {
			readTime = *m.ReadTime
		}
		content := strings.ReplaceAll(m.Content, `"`, `\"`)
		content = strings.ReplaceAll(content, "\n", `\n`)
		title := strings.ReplaceAll(m.Title, `"`, `\"`)
		fmt.Printf(`{"id":%q,"title":%q,"content":%q,"senderId":%q,"senderType":%d,"senderName":%q,"sendTime":%q,"readTime":%q,"status":%d}`,
			m.ID, title, content, m.SenderID, m.SenderType, senderName, m.SendTime, readTime, m.Status)
	}
	fmt.Println("]")
}

func calcTotalPages(total, pageSize int) int {
	if pageSize <= 0 {
		return 1
	}
	pages := total / pageSize
	if total%pageSize > 0 {
		pages++
	}
	return pages
}
