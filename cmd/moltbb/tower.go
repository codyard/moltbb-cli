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
	"moltbb-cli/internal/output"
)

func newTowerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tower",
		Short: "Manage Lobster Tower room and heartbeat",
	}
	cmd.AddCommand(newTowerCheckinCmd())
	cmd.AddCommand(newTowerHeartbeatCmd())
	cmd.AddCommand(newTowerMyRoomCmd())
	cmd.AddCommand(newTowerStatsCmd())
	cmd.AddCommand(newTowerListCmd())
	cmd.AddCommand(newTowerRoomCmd())
	return cmd
}

func newTowerCheckinCmd() *cobra.Command {
	var jsonOutput bool
	var roomCode string

	cmd := &cobra.Command{
		Use:   "checkin",
		Short: "Check in to get a tower room assignment",
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

			resp, err := client.TowerCheckin(ctx, apiKey, strings.TrimSpace(roomCode))
			if err != nil {
				return err
			}

			if jsonOutput {
				joinTimeVal := int64(0)
				if resp.JoinTime != nil {
					joinTimeVal = *resp.JoinTime
				}
				fmt.Printf(`{"code":"%s","globalIndex":%d,"floor":%d,"roomNumber":%d,"joinTime":%d}%s`,
					resp.Code, resp.GlobalIndex, resp.Floor, resp.RoomNumber, joinTimeVal, "\n")
				return nil
			}

			output.PrintSuccess("Room assigned successfully!")
			fmt.Println()
			fmt.Println("Room Code:   ", resp.Code)
			fmt.Println("Floor:       ", resp.Floor)
			fmt.Println("Room Number: ", resp.RoomNumber)
			fmt.Println("Global Index:", resp.GlobalIndex)
			if resp.JoinTime != nil && *resp.JoinTime > 0 {
				joinTime := time.Unix(*resp.JoinTime, 0).UTC()
				fmt.Println("Join Time:   ", joinTime.Format("2006-01-02 15:04:05 UTC"))
			}
			fmt.Println()
			fmt.Println("You can now send heartbeats using:")
			fmt.Println("  moltbb tower heartbeat")
			return nil
		},
	}

	cmd.Flags().StringVarP(&roomCode, "room-code", "r", "", "Specific room code to check in (3-char HEX, optional)")
	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")
	return cmd
}

func newTowerHeartbeatCmd() *cobra.Command {
	var roomCode string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "heartbeat",
		Short: "Send heartbeat signal to your tower room",
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

			// Auto-detect room code if not provided
			code := strings.TrimSpace(roomCode)
			if code == "" {
				myRoom, err := client.TowerGetMyRoom(ctx, apiKey)
				if err != nil {
					return fmt.Errorf("auto-detect room failed (use --room-code to specify): %w", err)
				}
				code = myRoom.Code
			}

			resp, err := client.TowerSendHeartbeat(ctx, apiKey, code)
			if err != nil {
				return err
			}

			if jsonOutput {
				fmt.Printf(`{"success":%t,"timestamp":%d}%s`, resp.Success, resp.Timestamp, "\n")
				return nil
			}

			output.PrintSuccess("Heartbeat sent successfully!")
			fmt.Println("Room Code:", code)
			heartbeatTime := time.Unix(resp.Timestamp, 0).UTC()
			fmt.Println("Timestamp:", heartbeatTime.Format("2006-01-02 15:04:05 UTC"))
			return nil
		},
	}

	cmd.Flags().StringVarP(&roomCode, "room-code", "r", "", "Room code (3-char HEX, auto-detected if not specified)")
	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")
	return cmd
}

func newTowerMyRoomCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "my-room",
		Short: "Get your current tower room assignment",
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

			room, err := client.TowerGetMyRoom(ctx, apiKey)
			if err != nil {
				return err
			}

			if jsonOutput {
				lastHeartbeatVal := int64(0)
				if room.LastHeartbeat != nil {
					lastHeartbeatVal = *room.LastHeartbeat
				}
				fmt.Printf(`{"code":"%s","globalIndex":%d,"botId":"%s","botName":"%s","status":%d,"lastHeartbeat":%d}%s`,
					room.Code, room.GlobalIndex, room.BotId, room.BotName, room.Status, lastHeartbeatVal, "\n")
				return nil
			}

			output.PrintSection("Your Tower Room")
			fmt.Println("Room Code:   ", room.Code)
			fmt.Println("Global Index:", room.GlobalIndex)
			fmt.Println("Bot ID:      ", room.BotId)
			if room.BotName != "" {
				fmt.Println("Bot Name:    ", room.BotName)
			}
			fmt.Println("Status:      ", formatNodeStatus(room.Status))
			if room.LastHeartbeat != nil && *room.LastHeartbeat > 0 {
				lastTime := time.Unix(*room.LastHeartbeat, 0).UTC()
				fmt.Println("Last Heartbeat:", lastTime.Format("2006-01-02 15:04:05 UTC"))
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")
	return cmd
}

func newTowerStatsCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show tower-wide statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			stats, err := client.TowerGetStatistics(ctx)
			if err != nil {
				return err
			}

			if jsonOutput {
				fmt.Printf(`{"totalRooms":%d,"occupiedRooms":%d,"onlineRooms":%d,"occupancyRate":%.4f,"roomsJoinedToday":%d,"fullFloors":%d,"isFullTower":%t}%s`,
					stats.TotalRooms, stats.OccupiedRooms, stats.OnlineRooms, stats.OccupancyRate,
					stats.RoomsJoinedToday, stats.FullFloors, stats.IsFullTower, "\n")
				return nil
			}

			output.PrintSection("Lobster Tower Statistics")
			fmt.Printf("Total Rooms:        %d\n", stats.TotalRooms)
			fmt.Printf("Occupied Rooms:     %d   (%.1f%%)\n", stats.OccupiedRooms, float64(stats.OccupiedRooms)/float64(stats.TotalRooms)*100)
			fmt.Printf("Online Rooms:       %d   (%.1f%%)\n", stats.OnlineRooms, float64(stats.OnlineRooms)/float64(stats.TotalRooms)*100)
			fmt.Printf("Occupancy Rate:     %.1f%%\n", stats.OccupancyRate*100)
			fmt.Printf("Joined Today:       %d\n", stats.RoomsJoinedToday)
			fmt.Printf("Full Floors:        %d\n", stats.FullFloors)
			fullTowerStatus := "No"
			if stats.IsFullTower {
				fullTowerStatus = "Yes"
			}
			fmt.Printf("Full Tower:         %s\n", fullTowerStatus)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")
	return cmd
}

func newTowerListCmd() *cobra.Command {
	var floor string
	var statusFilter string
	var occupied bool
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all tower rooms",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			rooms, err := client.TowerGetAllRooms(ctx)
			if err != nil {
				return err
			}

			// Apply filters
			filtered := make([]api.TowerRoomState, 0, len(rooms))
			for _, room := range rooms {
				if floor != "" && room.Code[:2] != strings.ToUpper(floor) {
					continue
				}
				if statusFilter != "" {
					expectedStatus := parseStatusFilter(statusFilter)
					if expectedStatus >= 0 && room.Status != expectedStatus {
						continue
					}
				}
				if occupied && room.BotId == "" {
					continue
				}
				filtered = append(filtered, room)
			}

			if jsonOutput {
				fmt.Print("[")
				for i, room := range filtered {
					if i > 0 {
						fmt.Print(",")
					}
					lastHeartbeatVal := int64(0)
					if room.LastHeartbeat != nil {
						lastHeartbeatVal = *room.LastHeartbeat
					}
					fmt.Printf(`{"code":"%s","globalIndex":%d,"botId":"%s","botName":"%s","status":%d,"lastHeartbeat":%d}`,
						room.Code, room.GlobalIndex, room.BotId, room.BotName, room.Status, lastHeartbeatVal)
				}
				fmt.Println("]")
				return nil
			}

			if len(filtered) == 0 {
				fmt.Println("No rooms found matching filters")
				return nil
			}

			fmt.Println("ROOM    FLOOR  STATUS      BOT NAME           LAST HEARTBEAT")
			fmt.Println("------  -----  ----------  -----------------  ---------------")
			for _, room := range filtered {
				botName := "-"
				if room.BotName != "" {
					botName = room.BotName
					if len(botName) > 17 {
						botName = botName[:14] + "..."
					}
				}
				lastHeartbeat := "-"
				if room.LastHeartbeat != nil && *room.LastHeartbeat > 0 {
					lastTime := time.Unix(*room.LastHeartbeat, 0).UTC()
					lastHeartbeat = formatTimeAgo(lastTime)
				}
				floorNum := room.Code[:2]
				fmt.Printf("%-6s  %-5s  %-10s  %-17s  %s\n",
					room.Code, floorNum, formatNodeStatus(room.Status), botName, lastHeartbeat)
			}
			fmt.Println()
			fmt.Printf("Total: %d rooms | Occupied: %d | Online: %d\n",
				len(filtered),
				countOccupied(filtered),
				countOnline(filtered))
			return nil
		},
	}

	cmd.Flags().StringVarP(&floor, "floor", "f", "", "Filter by floor (01-40 in HEX)")
	cmd.Flags().StringVarP(&statusFilter, "status", "s", "", "Filter by status (offline|online|stable-7d|stable-30d)")
	cmd.Flags().BoolVarP(&occupied, "occupied", "o", false, "Show only occupied rooms")
	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")
	return cmd
}

func newTowerRoomCmd() *cobra.Command {
	var code string
	var mine bool
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "room",
		Short: "Get details for a specific room",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !mine && code == "" && len(args) == 0 {
				return errors.New("either --code, --mine, or room code as argument is required")
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

			roomCode := strings.TrimSpace(code)
			if roomCode == "" && len(args) > 0 {
				roomCode = strings.TrimSpace(args[0])
			}

			if mine {
				apiKey, err := auth.ResolveAPIKey()
				if err != nil {
					return fmt.Errorf("resolve API key: %w", err)
				}
				myRoom, err := client.TowerGetMyRoom(ctx, apiKey)
				if err != nil {
					return err
				}
				roomCode = myRoom.Code
			}

			if roomCode == "" {
				return errors.New("room code is required")
			}

			room, err := client.TowerGetRoomDetail(ctx, roomCode)
			if err != nil {
				return err
			}

			if jsonOutput {
				lastHeartbeatVal := int64(0)
				if room.LastHeartbeat != nil {
					lastHeartbeatVal = *room.LastHeartbeat
				}
				joinTimeVal := int64(0)
				if room.JoinTime != nil {
					joinTimeVal = *room.JoinTime
				}
				fmt.Printf(`{"code":"%s","floor":%d,"roomNumber":%d,"globalIndex":%d,"botId":"%s","botName":"%s","status":%d,"lastHeartbeat":%d,"joinTime":%d,"totalHeartbeats":%d}%s`,
					room.Code, room.Floor, room.RoomNumber, room.GlobalIndex, room.BotId, room.BotName,
					room.Status, lastHeartbeatVal, joinTimeVal, room.TotalHeartbeats, "\n")
				return nil
			}

			output.PrintSection("Room Details")
			fmt.Println("Code:              ", room.Code)
			fmt.Println("Global Index:      ", room.GlobalIndex)
			fmt.Println("Floor:             ", room.Floor)
			fmt.Println("Room:              ", room.RoomNumber)
			fmt.Println("Status:            ", formatNodeStatus(room.Status))
			if room.BotId != "" {
				fmt.Println("Bot:               ", room.BotName)
				fmt.Println("Bot ID:            ", room.BotId)
			} else {
				fmt.Println("Bot:                -")
			}
			if room.LastHeartbeat != nil && *room.LastHeartbeat > 0 {
				lastTime := time.Unix(*room.LastHeartbeat, 0).UTC()
				fmt.Println("Last Heartbeat:    ", lastTime.Format("2006-01-02 15:04:05 UTC"))
			} else {
				fmt.Println("Last Heartbeat:     -")
			}
			if room.JoinTime != nil && *room.JoinTime > 0 {
				joinTime := time.Unix(*room.JoinTime, 0).UTC()
				fmt.Println("Join Time:         ", joinTime.Format("2006-01-02 15:04:05 UTC"))
			}
			fmt.Println("Total Heartbeats:  ", room.TotalHeartbeats)
			return nil
		},
	}

	cmd.Flags().StringVarP(&code, "code", "c", "", "Room code (3-char HEX)")
	cmd.Flags().BoolVarP(&mine, "mine", "m", false, "Show your assigned room")
	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")
	return cmd
}

func formatNodeStatus(status int) string {
	switch status {
	case 0:
		return "OFFLINE"
	case 1:
		return "ONLINE"
	case 2:
		return "STABLE_7D"
	case 3:
		return "STABLE_30D"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", status)
	}
}

func parseStatusFilter(filter string) int {
	switch strings.ToLower(strings.TrimSpace(filter)) {
	case "offline":
		return 0
	case "online":
		return 1
	case "stable-7d", "stable_7d":
		return 2
	case "stable-30d", "stable_30d":
		return 3
	default:
		return -1
	}
}

func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)
	if duration < time.Minute {
		return "just now"
	}
	if duration < time.Hour {
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%d min ago", minutes)
	}
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		return fmt.Sprintf("%d hr ago", hours)
	}
	days := int(duration.Hours() / 24)
	return fmt.Sprintf("%d days ago", days)
}

func countOccupied(rooms []api.TowerRoomState) int {
	count := 0
	for _, room := range rooms {
		if room.BotId != "" {
			count++
		}
	}
	return count
}

func countOnline(rooms []api.TowerRoomState) int {
	count := 0
	for _, room := range rooms {
		if room.Status == 1 || room.Status == 2 || room.Status == 3 {
			count++
		}
	}
	return count
}
