package config

import (
	"os"
	"path/filepath"
	"testing"
)

// writeConfig はテスト用の設定ファイルを XDG_CONFIG_HOME 配下に書き、その dir を返す。
func writeConfig(t *testing.T, content string) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	path := filepath.Join(dir, "octobell")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(path, "config.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func TestDefaultTerminalNotify(t *testing.T) {
	if got := Default().TerminalNotify; got != TerminalNotifyAuto {
		t.Fatalf("Default().TerminalNotify = %q, want %q", got, TerminalNotifyAuto)
	}
}

func TestLoadPartialKeepsTerminalNotifyDefault(t *testing.T) {
	// terminal_notify を含まない部分設定でも既定 auto を維持する。
	writeConfig(t, `{"poll_seconds": 30}`)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.PollSeconds != 30 {
		t.Fatalf("PollSeconds = %d, want 30", cfg.PollSeconds)
	}
	if cfg.TerminalNotify != TerminalNotifyAuto {
		t.Fatalf("TerminalNotify = %q, want %q", cfg.TerminalNotify, TerminalNotifyAuto)
	}
}

func TestLoadTerminalNotifyOverride(t *testing.T) {
	writeConfig(t, `{"terminal_notify": "osc777"}`)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.TerminalNotify != TerminalNotifyOSC777 {
		t.Fatalf("TerminalNotify = %q, want %q", cfg.TerminalNotify, TerminalNotifyOSC777)
	}
}

func TestLoadTerminalNotifyUnknownFallsBackToAuto(t *testing.T) {
	writeConfig(t, `{"terminal_notify": "bogus"}`)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.TerminalNotify != TerminalNotifyAuto {
		t.Fatalf("TerminalNotify = %q, want %q (unknown should fall back)", cfg.TerminalNotify, TerminalNotifyAuto)
	}
}
