// Package config は octobell の設定（XDG Base Directory 準拠の JSON ファイル）を扱う。
package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Config はユーザー設定。フィールドが設定ファイルに存在しない場合は Default の値が維持される。
type Config struct {
	// PollSeconds は希望ポーリング間隔（秒）。実際の間隔は GitHub の X-Poll-Interval を下限に強制される。
	PollSeconds int `json:"poll_seconds"`
	// All は既読も含めて取得するか。
	All bool `json:"all"`
	// Participating は参加（mention / review 依頼など）通知のみに絞るか。
	Participating bool `json:"participating"`
	// PerPage は 1 ページあたりの取得件数（最大 50）。
	PerPage int `json:"per_page"`
	// MarkReadOnOpen は enter でブラウザを開いたときに既読化するか。
	MarkReadOnOpen bool `json:"mark_read_on_open"`
	// Notify は OS デスクトップ通知を有効にするか。
	Notify bool `json:"notify"`
	// TerminalNotify は通知バックエンドの選択を制御する（auto | osc777 | osc9 | off）。
	// auto は対応端末（Ghostty）を検出したら OSC 777 を使い、未検出なら Beeep にフォールバックする。
	TerminalNotify string `json:"terminal_notify"`
}

// terminal_notify の取り得る値。
const (
	TerminalNotifyAuto   = "auto"
	TerminalNotifyOSC777 = "osc777"
	TerminalNotifyOSC9   = "osc9"
	TerminalNotifyOff    = "off"
)

// normalizeTerminalNotify は未知の terminal_notify 値を既定（auto）に丸める。
func normalizeTerminalNotify(v string) string {
	switch v {
	case TerminalNotifyAuto, TerminalNotifyOSC777, TerminalNotifyOSC9, TerminalNotifyOff:
		return v
	default:
		return TerminalNotifyAuto
	}
}

// Default は既定設定を返す。
func Default() Config {
	return Config{
		PollSeconds:    60,
		All:            false,
		Participating:  false,
		PerPage:        50,
		MarkReadOnOpen: true,
		Notify:         true,
		TerminalNotify: TerminalNotifyAuto,
	}
}

// Path は設定ファイルのパス（XDG_CONFIG_HOME or ~/.config 配下）を返す。
func Path() (string, error) {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "octobell", "config.json"), nil
}

// Load は設定ファイルを読み込む。ファイルが存在しない場合は Default を返す。
func Load() (Config, error) {
	cfg := Default()
	path, err := Path()
	if err != nil {
		return cfg, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}
	// 既定値を保持したまま、ファイルに存在するキーのみ上書きする。
	if err := json.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}
	// 不正な terminal_notify 値は既定（auto）に丸める。
	cfg.TerminalNotify = normalizeTerminalNotify(cfg.TerminalNotify)
	return cfg, nil
}
