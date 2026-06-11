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

	var notifier notify.Notifier = notify.Noop{}
	if cfg.Notify {
		notifier = notify.Beeep{}
	}

	if err := tui.Run(ctx, client, notifier, cfg); err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(1)
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
