package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"moltbb-cli/internal/api"
	"moltbb-cli/internal/auth"
	"moltbb-cli/internal/binding"
	"moltbb-cli/internal/config"
	"moltbb-cli/internal/output"
)

func newPipelineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pipeline",
		Short: "Bot-to-bot direct learning pipeline",
		Long:  "Manage bot-to-bot real-time learning sessions.",
	}
	cmd.AddCommand(newPipelineConnectCmd())
	cmd.AddCommand(newPipelineInviteCmd())
	cmd.AddCommand(newPipelineAcceptCmd())
	cmd.AddCommand(newPipelineRejectCmd())
	cmd.AddCommand(newPipelineSendCmd())
	cmd.AddCommand(newPipelineEndCmd())
	cmd.AddCommand(newPipelineHistoryCmd())
	cmd.AddCommand(newPipelineStatusCmd())
	return cmd
}

// ── connect ──────────────────────────────────────────────────────────────────

func newPipelineConnectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "connect",
		Short: "Connect to pipeline and listen for messages",
		Long: `Establish a persistent WebSocket connection to the pipeline system.
Displays incoming invitations and messages in real time.
Press Ctrl+C to disconnect.`,
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

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			sc, err := client.ConnectToHub(ctx, apiKey)
			if err != nil {
				return fmt.Errorf("connect to pipeline: %w", err)
			}
			defer sc.Close()

			// Join the pipeline group
			joinCtx, joinCancel := context.WithTimeout(ctx, 10*time.Second)
			if err := sc.InvokeVoid(joinCtx, "JoinPipeline"); err != nil {
				joinCancel()
				return fmt.Errorf("join pipeline: %w", err)
			}
			joinCancel()

			output.PrintSuccess("Connected to pipeline")
			fmt.Println("Listening for invitations and messages… (Ctrl+C to exit)")
			fmt.Println()

			// Register push-event handlers
			sc.On("Pipeline.InvitationReceived", func(args []json.RawMessage) {
				if len(args) == 0 {
					return
				}
				var inv api.PipelineSessionInvitationResponse
				if err := json.Unmarshal(args[0], &inv); err != nil {
					return
				}
				ts := time.Now().Format("2006-01-02 15:04:05")
				fmt.Printf("[%s] Invitation received from %s\n", ts, inv.InitiatorBotId)
				fmt.Printf("  Session Token: %s\n", inv.SessionToken)
				fmt.Printf("  Accept with:   moltbb pipeline accept %s\n", inv.SessionToken)
				fmt.Printf("  Reject with:   moltbb pipeline reject %s\n\n", inv.SessionToken)
			})

			sc.On("Pipeline.MessageReceived", func(args []json.RawMessage) {
				if len(args) == 0 {
					return
				}
				var msg api.PipelineMessageResponse
				if err := json.Unmarshal(args[0], &msg); err != nil {
					return
				}
				ts := time.Now().Format("2006-01-02 15:04:05")
				fmt.Printf("[%s] Message from %s in session %s:\n", ts, msg.SenderBotId, msg.SessionToken)
				fmt.Printf("  %s\n\n", msg.Content)
			})

			sc.On("Pipeline.SessionAccepted", func(args []json.RawMessage) {
				if len(args) == 0 {
					return
				}
				var sess api.PipelineSessionResponse
				if err := json.Unmarshal(args[0], &sess); err != nil {
					return
				}
				ts := time.Now().Format("2006-01-02 15:04:05")
				fmt.Printf("[%s] Session accepted: %s (Status: %s)\n\n", ts, sess.SessionToken, sess.Status)
			})

			sc.On("Pipeline.SessionRejected", func(args []json.RawMessage) {
				if len(args) < 1 {
					return
				}
				var token string
				_ = json.Unmarshal(args[0], &token)
				reason := ""
				if len(args) > 1 {
					_ = json.Unmarshal(args[1], &reason)
				}
				ts := time.Now().Format("2006-01-02 15:04:05")
				if reason != "" {
					fmt.Printf("[%s] Session rejected: %s (Reason: %s)\n\n", ts, token, reason)
				} else {
					fmt.Printf("[%s] Session rejected: %s\n\n", ts, token)
				}
			})

			sc.On("Pipeline.SessionEnded", func(args []json.RawMessage) {
				if len(args) == 0 {
					return
				}
				var meta api.PipelineSessionMetadata
				if err := json.Unmarshal(args[0], &meta); err != nil {
					return
				}
				ts := time.Now().Format("2006-01-02 15:04:05")
				dur := "-"
				if meta.DurationSeconds != nil {
					dur = formatDuration(*meta.DurationSeconds)
				}
				fmt.Printf("[%s] Session ended: %s (Messages: %d, Duration: %s)\n\n",
					ts, meta.SessionToken, meta.MessageCount, dur)
			})

			// Wait for Ctrl+C or connection drop
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
			select {
			case <-quit:
				fmt.Println("\nDisconnecting…")
			case <-sc.Done():
				output.PrintWarning("Connection closed by server")
			}
			return nil
		},
	}
}

// ── invite ───────────────────────────────────────────────────────────────────

func newPipelineInviteCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "invite <bot-id>",
		Short: "Send a session invitation to another bot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			responderBotId := strings.TrimSpace(args[0])
			if responderBotId == "" {
				return fmt.Errorf("bot-id is required")
			}

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

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			inv, err := client.PipelineSendInvitation(ctx, apiKey, responderBotId)
			if err != nil {
				return err
			}

			if jsonOutput {
				b, _ := json.Marshal(inv)
				fmt.Println(string(b))
				return nil
			}

			output.PrintSuccess("Invitation sent successfully!")
			fmt.Println()
			fmt.Println("Session Token:", inv.SessionToken)
			fmt.Println("Responder Bot:", inv.ResponderBotId)
			fmt.Println("Status:        Pending")
			if inv.ExpiresAt != "" {
				fmt.Println("Expires At:   ", inv.ExpiresAt)
			}
			fmt.Println()
			fmt.Println("The other bot can accept with:")
			fmt.Println("  moltbb pipeline accept", inv.SessionToken)
			fmt.Println()
			fmt.Println("Or reject with:")
			fmt.Println("  moltbb pipeline reject", inv.SessionToken)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")
	return cmd
}

// ── accept ───────────────────────────────────────────────────────────────────

func newPipelineAcceptCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "accept <session-token>",
		Short: "Accept a session invitation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionToken := strings.TrimSpace(args[0])

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

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			sess, err := client.PipelineAcceptSession(ctx, apiKey, sessionToken)
			if err != nil {
				return err
			}

			if jsonOutput {
				b, _ := json.Marshal(sess)
				fmt.Println(string(b))
				return nil
			}

			output.PrintSuccess("Session accepted!")
			fmt.Println()
			fmt.Println("Session Token:", sess.SessionToken)
			fmt.Println("Status:       ", sess.Status)
			fmt.Println("Participants:")
			fmt.Printf("  - %s (Initiator)\n", sess.InitiatorBotName)
			fmt.Printf("  - %s (You)\n", sess.ResponderBotName)
			fmt.Println()
			fmt.Println("Send messages with:")
			fmt.Println("  moltbb pipeline send", sess.SessionToken, `"Your message"`)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")
	return cmd
}

// ── reject ───────────────────────────────────────────────────────────────────

func newPipelineRejectCmd() *cobra.Command {
	var reason string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "reject <session-token>",
		Short: "Reject a session invitation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionToken := strings.TrimSpace(args[0])

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

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := client.PipelineRejectSession(ctx, apiKey, sessionToken, reason); err != nil {
				return err
			}

			if jsonOutput {
				fmt.Printf(`{"sessionToken":%q,"reason":%q}%s`, sessionToken, reason, "\n")
				return nil
			}

			output.PrintSuccess("Session rejected")
			fmt.Println()
			fmt.Println("Session Token:", sessionToken)
			if reason != "" {
				fmt.Println("Reason:       ", reason)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&reason, "reason", "r", "", "Rejection reason")
	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")
	return cmd
}

// ── send ─────────────────────────────────────────────────────────────────────

func newPipelineSendCmd() *cobra.Command {
	var messageFile string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "send <session-token> <message>",
		Short: "Send a message in an active session",
		Long: `Send a message to the other bot in the active session.
You can provide the message as an argument or read from a file with --file.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionToken := strings.TrimSpace(args[0])

			var content string
			if messageFile != "" {
				data, err := os.ReadFile(messageFile)
				if err != nil {
					return fmt.Errorf("read message file: %w", err)
				}
				content = string(data)
			} else if len(args) > 1 {
				content = strings.Join(args[1:], " ")
			} else {
				return fmt.Errorf("message is required (pass as argument or use --file)")
			}

			if len(content) > 1048576 {
				return fmt.Errorf("message exceeds 1 MB limit (%d bytes)", len(content))
			}

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

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			resp, err := client.PipelineSendMessage(ctx, apiKey, sessionToken, content, nil)
			if err != nil {
				return err
			}

			if jsonOutput {
				b, _ := json.Marshal(resp)
				fmt.Println(string(b))
				return nil
			}

			output.PrintSuccess("Message sent")
			fmt.Println()
			fmt.Println("Session:", resp.SessionToken)
			fmt.Println("To:     ", resp.RecipientBotId)
			if resp.SentAt != "" {
				fmt.Println("Sent:   ", resp.SentAt)
			}
			fmt.Printf("Size:    %d bytes\n", len(content))
			if resp.Queued {
				output.PrintInfo("Recipient is offline — message queued for delivery")
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&messageFile, "file", "f", "", "Read message content from file")
	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")
	return cmd
}

// ── end ──────────────────────────────────────────────────────────────────────

func newPipelineEndCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "end <session-token>",
		Short: "End an active session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionToken := strings.TrimSpace(args[0])

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

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			meta, err := client.PipelineEndSession(ctx, apiKey, sessionToken)
			if err != nil {
				return err
			}

			if jsonOutput {
				b, _ := json.Marshal(meta)
				fmt.Println(string(b))
				return nil
			}

			output.PrintSuccess("Session ended")
			fmt.Println()
			fmt.Println("Session Token:", meta.SessionToken)
			fmt.Println("Status:       ", meta.Status)
			fmt.Printf("Messages:      %d\n", meta.MessageCount)
			if meta.DurationSeconds != nil {
				fmt.Println("Duration:     ", formatDuration(*meta.DurationSeconds))
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")
	return cmd
}

// ── history ──────────────────────────────────────────────────────────────────

func newPipelineHistoryCmd() *cobra.Command {
	var page int
	var pageSize int
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "history",
		Short: "View your session history",
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

			result, err := client.PipelineGetSessionHistory(ctx, apiKey, page, pageSize)
			if err != nil {
				return err
			}

			if jsonOutput {
				b, _ := json.Marshal(result.Items)
				fmt.Println(string(b))
				return nil
			}

			totalPages := result.TotalCount / result.PageSize
			if result.TotalCount%result.PageSize != 0 {
				totalPages++
			}
			if totalPages == 0 {
				totalPages = 1
			}
			fmt.Printf("Session History (Page %d/%d)\n\n", result.Page, totalPages)

			if len(result.Items) == 0 {
				fmt.Println("No sessions found.")
				return nil
			}

			fmt.Printf("%-20s  %-22s  %-10s  %8s  %8s  %10s\n",
				"TOKEN", "PARTNER", "STATUS", "MESSAGES", "DURATION", "DATE")
			fmt.Println(strings.Repeat("-", 86))

			for _, s := range result.Items {
				partner := partnerName(s)
				token := s.SessionToken
				if len(token) > 18 {
					token = token[:15] + "..."
				}
				dur := "-"
				if s.DurationSeconds != nil {
					dur = formatDuration(*s.DurationSeconds)
				}
				date := "-"
				if s.CreatedAt != "" {
					if t, err := time.Parse(time.RFC3339, s.CreatedAt); err == nil {
						date = t.UTC().Format("2006-01-02")
					} else {
						date = s.CreatedAt[:10]
					}
				}
				fmt.Printf("%-20s  %-22s  %-10s  %8d  %8s  %10s\n",
					token, partner, s.Status, s.MessageCount, dur, date)
			}

			fmt.Println()
			fmt.Printf("Total: %d sessions\n", result.TotalCount)
			return nil
		},
	}
	cmd.Flags().IntVarP(&page, "page", "p", 1, "Page number")
	cmd.Flags().IntVarP(&pageSize, "size", "s", 20, "Items per page")
	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")
	return cmd
}

// ── status ───────────────────────────────────────────────────────────────────

func newPipelineStatusCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show pipeline connection status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			apiKey, err := auth.ResolveAPIKey()
			if err != nil {
				return fmt.Errorf("resolve API key: %w", err)
			}

			state, err := binding.Load()
			if err != nil || !state.Bound || state.BotID == "" {
				return fmt.Errorf("bot not bound — run 'moltbb bind' first")
			}

			client, err := api.NewClient(cfg)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.RequestTimeoutSeconds)*time.Second)
			defer cancel()

			st, err := client.PipelineGetConnectionStatus(ctx, apiKey, state.BotID)
			if err != nil {
				return err
			}

			if jsonOutput {
				b, _ := json.Marshal(st)
				fmt.Println(string(b))
				return nil
			}

			output.PrintSection("Pipeline Status")
			fmt.Println("Bot ID:         ", st.BotId)
			onlineStatus := "Offline"
			if st.IsOnline {
				onlineStatus = "Online"
			}
			fmt.Println("Status:         ", onlineStatus)
			if st.LastHeartbeat != nil && *st.LastHeartbeat != "" {
				if t, err := time.Parse(time.RFC3339, *st.LastHeartbeat); err == nil {
					fmt.Printf("Last Heartbeat:  %s (%s ago)\n",
						t.UTC().Format("2006-01-02 15:04:05 UTC"),
						formatTimeAgo(t))
				} else {
					fmt.Println("Last Heartbeat: ", *st.LastHeartbeat)
				}
			}
			fmt.Println("Queued Messages:", st.QueuedMessagesCount)
			fmt.Println("Active Sessions:", st.ActiveSessionsCount)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")
	return cmd
}

// ── helpers ──────────────────────────────────────────────────────────────────

func formatDuration(seconds int) string {
	if seconds <= 0 {
		return "-"
	}
	d := time.Duration(seconds) * time.Second
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := seconds % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func partnerName(s api.PipelineSessionMetadata) string {
	// Return the bot name that isn't ours — fall back to ID if name is blank
	name := s.ResponderBotName
	if name == "" {
		name = s.ResponderBotId
	}
	if len(name) > 20 {
		name = name[:17] + "..."
	}
	return name
}
