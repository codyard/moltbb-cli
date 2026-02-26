package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"moltbb-cli/internal/utils"
)

func newDaemonCmd() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "daemon [start|stop|status|restart]",
		Short: "Run MoltBB as a background daemon service",
		Long: `Manage MoltBB local web server as a background daemon.
Commands:
  start   - Start the daemon
  stop    - Stop the daemon
  status  - Show daemon status
  restart - Restart the daemon`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			command := "status"
			if len(args) > 0 {
				command = args[0]
			}

			moltbbDir, err := utils.MoltbbDir()
			if err != nil {
				return err
			}

			pidFile := filepath.Join(moltbbDir, "daemon.pid")
			logFile := filepath.Join(moltbbDir, "daemon.log")

			switch command {
			case "start":
				return daemonStart(moltbbDir, pidFile, logFile, port)
			case "stop":
				return daemonStop(pidFile)
			case "status":
				return daemonStatus(pidFile, logFile)
			case "restart":
				if err := daemonStop(pidFile); err != nil {
					fmt.Println("Stop warning:", err)
				}
				time.Sleep(500 * time.Millisecond)
				return daemonStart(moltbbDir, pidFile, logFile, port)
			default:
				return fmt.Errorf("unknown command: %s (use: start, stop, status, restart)", command)
			}
		},
	}

	cmd.Flags().IntVar(&port, "port", 3789, "Port for local web server")
	return cmd
}

func daemonStart(moltbbDir, pidFile, logFile string, port int) error {
	// Check if already running
	if _, err := os.Stat(pidFile); err == nil {
		pidData, err := os.ReadFile(pidFile)
		if err == nil {
			var pid int
			if _, err := fmt.Sscanf(strings.TrimSpace(string(pidData)), "%d", &pid); err == nil {
				// Try to find process (won't work reliably on all platforms)
				fmt.Printf("⚠️  Daemon may already be running (PID: %d)\n", pid)
				fmt.Println("   Use 'moltbb daemon stop' first or check status")
				return nil
			}
			os.Remove(pidFile)
		}
	}

	// Find moltbb binary
	moltbbPath, err := exec.LookPath("moltbb")
	if err != nil {
		return fmt.Errorf("moltbb not found in PATH: %w", err)
	}

	// Use nohup to start in background
	log, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer log.Close()

	// Start process with nohup
	runCmd := exec.Command("nohup", moltbbPath, "local", "--port", fmt.Sprintf("%d", port))
	runCmd.Stdout = log
	runCmd.Stderr = log

	if err := runCmd.Start(); err != nil {
		return fmt.Errorf("start daemon: %w", err)
	}

	// Note: nohup makes the process detached, but we can't get the real PID easily
	// Just note that it's started
	fmt.Printf("✅ Daemon started\n")
	fmt.Printf("🌐 Running at: http://127.0.0.1:%d\n", port)
	fmt.Printf("📝 Log file: %s\n", logFile)
	fmt.Println("💡 Use 'moltbb daemon status' to check")
	return nil
}

func daemonStop(pidFile string) error {
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		// Try to kill any running moltbb local process
		cmd := exec.Command("pkill", "-f", "moltbb local")
		err := cmd.Run()
		if err != nil {
			fmt.Println("⚠️  No daemon process found")
		} else {
			fmt.Println("✅ Daemon stopped")
		}
		return nil
	}

	pidData, err := os.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("read PID file: %w", err)
	}

	var pid int
	if _, err := fmt.Sscanf(strings.TrimSpace(string(pidData)), "%d", &pid); err != nil {
		os.Remove(pidFile)
		return fmt.Errorf("invalid PID file: %w", err)
	}

	// Try to kill by PID
	cmd := exec.Command("kill", fmt.Sprintf("%d", pid))
	if err := cmd.Run(); err != nil {
		fmt.Println("⚠️  Could not kill process, trying pkill")
	}

	// Also try pkill as fallback
	cmd = exec.Command("pkill", "-f", "moltbb local")
	cmd.Run()

	os.Remove(pidFile)
	fmt.Println("✅ Daemon stopped")
	return nil
}

func daemonStatus(pidFile, logFile string) error {
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		// Check if process is running anyway
		cmd := exec.Command("pgrep", "-f", "moltbb local")
		err := cmd.Run()
		if err == nil {
			fmt.Println("✅ Daemon is running (found via pgrep)")
			return nil
		}
		fmt.Println("📴 Daemon is not running")
		fmt.Println("💡 Use 'moltbb daemon start' to start")
		return nil
	}

	pidData, err := os.ReadFile(pidFile)
	if err != nil {
		fmt.Println("⚠️  PID file corrupted")
		return nil
	}

	var pid int
	if _, err := fmt.Sscanf(strings.TrimSpace(string(pidData)), "%d", &pid); err != nil {
		fmt.Println("⚠️  Invalid PID")
		return nil
	}

	fmt.Printf("✅ Daemon is running (PID: %d)\n", pid)
	fmt.Printf("📝 Log file: %s\n", logFile)

	// Show recent logs
	if _, err := os.Stat(logFile); err == nil {
		fmt.Println("")
		fmt.Println("--- Recent logs ---")
		cmd := exec.Command("tail", "-5", logFile)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}

	return nil
}
