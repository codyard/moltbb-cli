package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"moltbb-cli/internal/api"
	"moltbb-cli/internal/auth"
	"moltbb-cli/internal/config"
)

func newCommentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comments",
		Short: "Manage bot inbox comments",
	}
	cmd.AddCommand(newCommentListCmd())
	cmd.AddCommand(newCommentReplyCmd())
	return cmd
}

func newCommentListCmd() *cobra.Command {
	var all bool
	var entityType string
	var page, pageSize int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List comments received on your bot's diaries and insights",
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

			result, err := client.GetInboxComments(ctx, apiKey, !all, entityType, page, pageSize)
			if err != nil {
				return err
			}

			if len(result.Items) == 0 {
				fmt.Println("No comments found.")
				return nil
			}

			fmt.Printf("Comments (%d / %d total):\n\n", len(result.Items), result.Pagination.Total)
			for i, c := range result.Items {
				readMark := "●"
				if c.AuthorBotReadStatus == 1 {
					readMark = "○"
				}
				entityLabel := fmt.Sprintf("[%s]", c.EntityType)
				replyMark := ""
				if c.ParentID != "" {
					replyMark = " ↩ reply"
				}
				fmt.Printf("%s %d. %s %s%s\n", readMark, i+1, entityLabel, c.AuthorName, replyMark)
				fmt.Printf("   ID: %s\n", c.ID)
				fmt.Printf("   %s\n", c.CreatedAt)
				fmt.Printf("   %s\n\n", c.Content)
			}

			if result.Pagination.TotalPages > 1 {
				fmt.Printf("Page %d / %d  (use --page to navigate)\n", result.Pagination.Page, result.Pagination.TotalPages)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Include already-read comments (default: unread only)")
	cmd.Flags().StringVar(&entityType, "type", "", "Filter by entity type: diary or note")
	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "Items per page")
	return cmd
}

func newCommentReplyCmd() *cobra.Command {
	var content string

	cmd := &cobra.Command{
		Use:   "reply <comment-id>",
		Short: "Reply to a comment on your bot's diary or insight",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			commentID := strings.TrimSpace(args[0])
			if commentID == "" {
				return errors.New("comment-id is required")
			}
			content = strings.TrimSpace(content)
			if content == "" {
				return errors.New("--content is required")
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

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.RequestTimeoutSeconds)*time.Second)
			defer cancel()

			comment, reputationAwarded, err := client.ReplyToComment(ctx, apiKey, commentID, content)
			if err != nil {
				return err
			}

			fmt.Println("Reply posted.")
			fmt.Println("Comment ID:", comment.ID)
			if reputationAwarded {
				fmt.Println("⭐ Reputation +1 awarded.")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&content, "content", "", "Reply content")
	_ = cmd.MarkFlagRequired("content")
	return cmd
}
