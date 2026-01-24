package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbletea"
)

type ServerStatusMsg struct {
	Status    string
	StartTime time.Time
	Running   bool
}

type ActionMsg struct {
	Output string
	Err    error
}

func checkServerStatusCmd() tea.Cmd {
	return func() tea.Msg {
		pidData, err := os.ReadFile("server.pid")
		if err != nil {
			return ServerStatusMsg{Status: "STOPPED", Running: false}
		}

		pid, err := strconv.Atoi(string(pidData))
		if err != nil {
			return ServerStatusMsg{Status: "ERROR (PID)", Running: false}
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			return ServerStatusMsg{Status: "STOPPED", Running: false}
		}

		// Check if process is alive (Unix specific)
		if err := process.Signal(syscall.Signal(0)); err != nil {
			return ServerStatusMsg{Status: "STOPPED", Running: false}
		}

		// Get uptime from PID file mod time as a proxy
		fileInfo, err := os.Stat("server.pid")
		startTime := time.Now()
		if err == nil {
			startTime = fileInfo.ModTime()
		}

		return ServerStatusMsg{
			Status:    "RUNNING",
			StartTime: startTime,
			Running:   true,
		}
	}
}

func execServerActionCmd(action string) tea.Cmd {
	return func() tea.Msg {
		// Get current executable to call launcher commands
		exe, err := os.Executable()
		if err != nil {
			return ActionMsg{Err: err}
		}

		var args []string
		switch action {
		case "start":
			// Start server in background
			cmd := exec.Command("bash", "-c", fmt.Sprintf("%s start > server.log 2>&1 &", exe))
			err := cmd.Run()
			return ActionMsg{Output: "Start initiated", Err: err}

		case "restart_graceful":
			// For restart, we need to run it in a way that it doesn't block the TUI
			// and persists.
			// We'll use "start" after "stop" manually or use restart command if it supports detach.
			// Given existing commands, "restart" is foreground.
			// Strategy: Spawn a shell command that runs restart and redirects output
			cmd := exec.Command("bash", "-c", fmt.Sprintf("%s restart --graceful > server.log 2>&1 &", exe))
			err := cmd.Run()
			return ActionMsg{Output: "Restart initiated", Err: err}

		case "stop_graceful":
			args = []string{"stop", "--graceful"}
		case "stop_force":
			args = []string{"stop"}
		case "reload_config":
			args = []string{"config", "reload"}
		}

		if len(args) > 0 {
			cmd := exec.Command(exe, args...)
			out, err := cmd.CombinedOutput()
			return ActionMsg{Output: string(out), Err: err}
		}

		return ActionMsg{Output: "Unknown action", Err: nil}
	}
}
