package tui

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

type MonitorStatsMsg struct {
	System struct {
		Goroutines int    `json:"goroutines"`
		HeapAlloc  uint64 `json:"heap_alloc"`
		HeapSys    uint64 `json:"heap_sys"`
	} `json:"system"`
	Connections int `json:"connections"`
	Maps        []struct {
		ID    int `json:"id"`
		Users int `json:"users"`
	} `json:"maps"`
	Err error
}

func fetchMonitorStatsCmd() tea.Cmd {
	return func() tea.Msg {
		resp, err := http.Get("http://localhost:7667/monitor/stats")
		if err != nil {
			return MonitorStatsMsg{Err: err}
		}
		defer resp.Body.Close()

		var stats MonitorStatsMsg
		if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
			return MonitorStatsMsg{Err: err}
		}
		return stats
	}
}

type LogMsg struct {
	Lines     []string
	NewOffset int64
}

func readLogsCmd(offset int64) tea.Cmd {
	return func() tea.Msg {
		file, err := os.Open("server.log")
		if err != nil {
			return LogMsg{NewOffset: offset} // Retry later
		}
		defer file.Close()

		stat, err := file.Stat()
		if err != nil {
			return LogMsg{NewOffset: offset}
		}

		if stat.Size() < offset {
			offset = 0 // File truncated
		}

		if stat.Size() == offset {
			return LogMsg{NewOffset: offset} // No new data
		}

		_, err = file.Seek(offset, 0)
		if err != nil {
			return LogMsg{NewOffset: offset}
		}

		var lines []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		// Update offset
		newOffset, _ := file.Seek(0, 1) // Get current position

		return LogMsg{
			Lines:     lines,
			NewOffset: newOffset,
		}
	}
}

type UserListMsg struct {
	Users []struct {
		Addr string `json:"addr"`
		User string `json:"user"`
	}
	Err error
}

func fetchUserListCmd() tea.Cmd {
	return func() tea.Msg {
		resp, err := http.Get("http://localhost:7667/conn/list")
		if err != nil {
			return UserListMsg{Err: err}
		}
		defer resp.Body.Close()

		var users []struct {
			Addr string `json:"addr"`
			User string `json:"user"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
			return UserListMsg{Err: err}
		}
		return UserListMsg{Users: users}
	}
}

func kickUserCmd(name string) tea.Cmd {
	return func() tea.Msg {
		resp, err := http.Get(fmt.Sprintf("http://localhost:7667/conn/kick?name=%s", url.QueryEscape(name)))
		if err != nil {
			return ActionMsg{Err: err}
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return ActionMsg{Output: string(body)}
	}
}

func banUserCmd(name string) tea.Cmd {
	return func() tea.Msg {
		resp, err := http.Get(fmt.Sprintf("http://localhost:7667/conn/ban?nick=%s", url.QueryEscape(name)))
		if err != nil {
			return ActionMsg{Err: err}
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return ActionMsg{Output: string(body)}
	}
}

func inspectUserCmd(name string) tea.Cmd {
	return func() tea.Msg {
		resp, err := http.Get(fmt.Sprintf("http://localhost:7667/player/info?nick=%s", url.QueryEscape(name)))
		if err != nil {
			return ActionMsg{Err: err}
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			return ActionMsg{Err: fmt.Errorf("player not found or offline")}
		}

		// Just dumping JSON for now as inspection result
		body, _ := io.ReadAll(resp.Body)
		var prettyJSON map[string]interface{}
		json.Unmarshal(body, &prettyJSON)
		formatted, _ := json.MarshalIndent(prettyJSON, "", "  ")
		
		return ActionMsg{Output: string(formatted)}
	}
}

type MapListMsg struct {
	Maps []struct {
		ID    int `json:"id"`
		Users int `json:"users"`
		NPCs  int `json:"npcs"`
	}
	Err error
}

func fetchMapsCmd() tea.Cmd {
	return func() tea.Msg {
		resp, err := http.Get("http://localhost:7667/world/list")
		if err != nil {
			return MapListMsg{Err: err}
		}
		defer resp.Body.Close()

		var maps []struct {
			ID    int `json:"id"`
			Users int `json:"users"`
			NPCs  int `json:"npcs"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&maps); err != nil {
			return MapListMsg{Err: err}
		}
		return MapListMsg{Maps: maps}
	}
}

func reloadMapCmd(id int) tea.Cmd {
	return func() tea.Msg {
		resp, err := http.Get(fmt.Sprintf("http://localhost:7667/world/reload?id=%d", id))
		if err != nil {
			return ActionMsg{Err: err}
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return ActionMsg{Output: string(body)}
	}
}

func unloadMapCmd(id int) tea.Cmd {
	return func() tea.Msg {
		resp, err := http.Get(fmt.Sprintf("http://localhost:7667/world/unload?id=%d", id))
		if err != nil {
			return ActionMsg{Err: err}
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return ActionMsg{Output: string(body)}
	}
}

type ConfigItem struct {
	Key         string      `json:"key"`
	Description string      `json:"description"`
	Type        string      `json:"type"`
	Value       interface{} `json:"value"`
}

type ConfigListMsg struct {
	Items []ConfigItem
	Err   error
}

func fetchConfigListCmd() tea.Cmd {
	return func() tea.Msg {
		resp, err := http.Get("http://localhost:7667/config/list")
		if err != nil {
			return ConfigListMsg{Err: err}
		}
		defer resp.Body.Close()

		var items []ConfigItem
		if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
			return ConfigListMsg{Err: err}
		}
		return ConfigListMsg{Items: items}
	}
}

func setConfigCmd(key string, value string) tea.Cmd {
	return func() tea.Msg {
		resp, err := http.Get(fmt.Sprintf("http://localhost:7667/config/set?key=%s&value=%s", url.QueryEscape(key), url.QueryEscape(value)))
		if err != nil {
			return ActionMsg{Err: err}
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return ActionMsg{Output: string(body)}
	}
}
