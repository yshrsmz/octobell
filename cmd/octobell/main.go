// Command octobell は GitHub の通知を定期取得して未読管理する TUI アプリ。
// gh の認証を再利用し、新着を OS 通知で知らせる。
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/yshrsmz/octobell/internal/config"
	"github.com/yshrsmz/octobell/internal/github"
	"github.com/yshrsmz/octobell/internal/notify"
	"github.com/yshrsmz/octobell/internal/tui"
)

// version はビルド時に -ldflags "-X main.version=..." で差し込む。
var version = "dev"

func main() {
	var (
		once     = flag.Bool("once", false, "通知を一度だけ取得して一覧表示し、終了する（TUI を起動しない）")
		showVer  = flag.Bool("version", false, "バージョンを表示して終了する")
		noNotify = flag.Bool("no-notify", false, "OS 通知を無効化する")
		all      = flag.Bool("all", false, "既読も含めてすべての通知を取得する")
	)
	flag.Parse()

	if *showVer {
		fmt.Println("octobell", version)
		return
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "設定の読み込みに失敗:", err)
		os.Exit(1)
	}
	if *all {
		cfg.All = true
	}
	if *noNotify {
		cfg.Notify = false
	}

	client, err := github.NewClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, "GitHub クライアントの初期化に失敗（`gh auth login` 済みか確認してください）:", err)
		os.Exit(1)
	}

	ctx := context.Background()

	if *once {
		if err := runOnce(ctx, client, cfg); err != nil {
			fmt.Fprintln(os.Stderr, "エラー:", err)
			os.Exit(1)
		}
		return
	}

	// 設定・端末検出・tty 有無から単一の Notifier を選ぶ（Beeep と OSC は排他）。
	notifier := notify.Select(cfg.Notify, terminalNotifyMode(cfg.TerminalNotify), os.Getenv)

	if err := tui.Run(ctx, client, notifier, cfg); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(1)
	}
}

// terminalNotifyMode は config の terminal_notify 文字列を notify.Mode へ変換する。
// 文字列の語彙は config が一元管理し、ここが notify の型付き enum への唯一の橋渡し点。
// 未知値（config 側で正規化済みのはずだが念のため）は auto 扱い。
func terminalNotifyMode(v string) notify.Mode {
	switch v {
	case config.TerminalNotifyOSC777:
		return notify.ModeOSC777
	case config.TerminalNotifyOSC9:
		return notify.ModeOSC9
	case config.TerminalNotifyOff:
		return notify.ModeOff
	default: // config.TerminalNotifyAuto およびその他
		return notify.ModeAuto
	}
}

// runOnce は非 TUI のスモーク。API スパイン（認証・取得・URL 変換）の検証に使う。
func runOnce(ctx context.Context, client *github.Client, cfg config.Config) error {
	res, err := client.List(ctx, github.ListOptions{
		All:           cfg.All,
		Participating: cfg.Participating,
		PerPage:       cfg.PerPage,
	}, "")
	if err != nil {
		return err
	}
	if len(res.Notifications) == 0 {
		fmt.Println("通知はありません。")
		return nil
	}
	fmt.Printf("通知 %d 件 (X-Poll-Interval=%ds):\n\n", len(res.Notifications), res.PollInterval)
	for _, n := range res.Notifications {
		mark := " "
		if n.Unread {
			mark = "●"
		}
		fmt.Printf("%s [%s] %s\n    %s · %s\n    %s\n\n",
			mark, n.Subject.Type, n.Subject.Title, n.Repository.FullName, n.Reason, n.BrowserURL())
	}
	return nil
}
