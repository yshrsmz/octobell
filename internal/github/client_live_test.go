package github

import (
	"context"
	"os"
	"testing"
)

// TestLiveConditionalRequest は実 GitHub に対する条件付きリクエスト（If-Modified-Since → 304）を検証する。
// 認証とネットワークを要するため、OCTOBELL_LIVE=1 のときのみ実行する（CI ではスキップ）。
func TestLiveConditionalRequest(t *testing.T) {
	if os.Getenv("OCTOBELL_LIVE") == "" {
		t.Skip("ライブテスト: OCTOBELL_LIVE=1 で実行")
	}
	c, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	ctx := context.Background()

	r1, err := c.List(ctx, ListOptions{All: true}, "")
	if err != nil {
		t.Fatalf("初回 List: %v", err)
	}
	if r1.NotModified {
		t.Fatal("初回は 304 ではなく 200 を期待")
	}
	if r1.LastModified == "" {
		t.Fatal("初回レスポンスに Last-Modified を期待")
	}
	t.Logf("初回: 通知 %d 件, X-Poll-Interval=%ds, Last-Modified=%q",
		len(r1.Notifications), r1.PollInterval, r1.LastModified)

	r2, err := c.List(ctx, ListOptions{All: true}, r1.LastModified)
	if err != nil {
		t.Fatalf("条件付き List: %v", err)
	}
	if !r2.NotModified {
		t.Fatalf("2 回目は If-Modified-Since により 304 (NotModified) を期待したが、通知 %d 件が返った", len(r2.Notifications))
	}
	t.Logf("2 回目: 304 Not Modified（レート制限を消費しない）, X-Poll-Interval=%ds", r2.PollInterval)
}
