package tui

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/yshrsmz/octobell/internal/config"
	"github.com/yshrsmz/octobell/internal/github"
	"github.com/yshrsmz/octobell/internal/notify"
)

func sampleNotifs() []github.Notification {
	repo := github.Repository{FullName: "o/r", HTMLURL: "https://github.com/o/r"}
	return []github.Notification{
		{ID: "1", Unread: true, Reason: "mention", Repository: repo,
			Subject: github.Subject{Title: "PR one", Type: "PullRequest", URL: "https://api.github.com/repos/o/r/pulls/1"}},
		{ID: "2", Unread: true, Reason: "review_requested", Repository: repo,
			Subject: github.Subject{Title: "Issue two", Type: "Issue", URL: "https://api.github.com/repos/o/r/issues/2"}},
	}
}

// TestModelFlowHeadless は TTY 無しで Model のメッセージ処理を実行し、パニックしないこと・
// 楽観的更新が効くことを検証する（返り値の Cmd は実行しない＝ネットワーク不要）。
func TestModelFlowHeadless(t *testing.T) {
	m := newModel(nil, notify.Noop{}, config.Default())

	// 空リストでの初期 View でパニックしない
	_ = m.View()

	// ウィンドウサイズ → list.SetSize
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = tm.(Model)

	// 取得結果の反映（refreshItems / Differ / View を駆動）
	tm, _ = m.Update(fetchedMsg{res: github.ListResult{
		Notifications: sampleNotifs(), LastModified: "Thu, 01 Jan 2026 00:00:00 GMT", PollInterval: 60,
	}})
	m = tm.(Model)
	if len(m.notifs) != 2 {
		t.Fatalf("通知 2 件を期待, got %d", len(m.notifs))
	}
	if m.pollInterval != 60 {
		t.Errorf("pollInterval=60 を期待, got %d", m.pollInterval)
	}
	_ = m.View()

	// 選択既読（All=false の楽観的更新で一覧から除去される）
	m.list.Select(0)
	_ = m.markSelectedRead()
	if len(m.notifs) != 1 {
		t.Fatalf("選択既読で 1 件除去され 1 件残るべき, got %d", len(m.notifs))
	}
	if m.notifs[0].ID != "2" {
		t.Errorf("残るのは ID=2 のはず, got %s", m.notifs[0].ID)
	}

	// 開く（既読化しない経路）でパニックしない
	m.list.Select(0)
	_ = m.openSelected(false)

	// 全既読 → All=false なら一覧が空になる
	tm, _ = m.Update(markedAllMsg{})
	m = tm.(Model)
	if len(m.notifs) != 0 {
		t.Errorf("全既読後（All=false）一覧は空になるべき, got %d", len(m.notifs))
	}

	// 各種メッセージでパニックしない
	for _, msg := range []tea.Msg{
		tickMsg{}, statusMsg("テスト"), clearStatusMsg{},
		markedMsg{err: nil}, markedMsg{err: errors.New("boom")},
		fetchedMsg{err: errors.New("取得失敗")},
		fetchedMsg{res: github.ListResult{NotModified: true, PollInterval: 90}},
	} {
		tm, _ = m.Update(msg)
		m = tm.(Model)
		_ = m.View()
	}
}

// loadedModel は指定 All 設定で通知 2 件を読み込んだ Model を返す。
func loadedModel(all bool) Model {
	cfg := config.Default()
	cfg.All = all
	m := newModel(nil, notify.Noop{}, cfg)
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = tm.(Model)
	tm, _ = m.Update(fetchedMsg{res: github.ListResult{Notifications: sampleNotifs()}})
	return tm.(Model)
}

// TestMarkReadRemovesWhenUnreadOnly は All=false で単一既読すると一覧から除去されることを検証する。
func TestMarkReadRemovesWhenUnreadOnly(t *testing.T) {
	m := loadedModel(false)
	m.list.Select(0)
	_ = m.markSelectedRead()
	if len(m.notifs) != 1 {
		t.Fatalf("1 件除去され 1 件残るべき, got %d", len(m.notifs))
	}
	if m.notifs[0].ID != "2" {
		t.Errorf("残るのは ID=2, got %s", m.notifs[0].ID)
	}
}

// TestMarkAllReadEmptiesWhenUnreadOnly は All=false で全件既読すると一覧が空になることを検証する。
func TestMarkAllReadEmptiesWhenUnreadOnly(t *testing.T) {
	m := loadedModel(false)
	tm, _ := m.Update(markedAllMsg{})
	m = tm.(Model)
	if len(m.notifs) != 0 {
		t.Errorf("一覧は空になるべき, got %d", len(m.notifs))
	}
}

// TestMarkReadKeepsWhenAll は All=true で既読化しても一覧に残り Unread=false になることを検証する。
func TestMarkReadKeepsWhenAll(t *testing.T) {
	m := loadedModel(true)
	m.list.Select(0)
	_ = m.markSelectedRead()
	if len(m.notifs) != 2 {
		t.Fatalf("All=true では除去されず 2 件のまま, got %d", len(m.notifs))
	}
	if m.notifs[0].Unread {
		t.Error("既読化した項目は Unread=false になるべき")
	}
	// 全件既読でも残る
	tm, _ := m.Update(markedAllMsg{})
	m = tm.(Model)
	if len(m.notifs) != 2 {
		t.Fatalf("All=true の全既読でも 2 件のまま, got %d", len(m.notifs))
	}
	for i, n := range m.notifs {
		if n.Unread {
			t.Errorf("全既読後 notifs[%d] が未読のまま", i)
		}
	}
}

// TestRemoveLastKeepsSelectionValid は末尾項目を既読化で除去しても選択位置が範囲内に保たれることを検証する。
func TestRemoveLastKeepsSelectionValid(t *testing.T) {
	m := loadedModel(false)
	m.list.Select(1) // 末尾を選択
	_ = m.markSelectedRead()
	if len(m.notifs) != 1 {
		t.Fatalf("1 件残るべき, got %d", len(m.notifs))
	}
	if idx := m.list.Index(); idx < 0 || idx >= len(m.notifs) {
		t.Errorf("選択インデックスが範囲内であるべき, got %d (len=%d)", idx, len(m.notifs))
	}
}

// recordingNotifier は Notify の呼び出し回数を記録するテスト用 Notifier。
type recordingNotifier struct{ calls int }

func (r *recordingNotifier) Notify(string, string) error { r.calls++; return nil }

// TestManualRefreshForceSkipsConditional は手動更新（force=true）が
// If-Modified-Since を送らない（無条件取得になる）ことを検証する。
func TestManualRefreshForceSkipsConditional(t *testing.T) {
	m := newModel(nil, notify.Noop{}, config.Default())
	m.lastModified = "Thu, 01 Jan 2026 00:00:00 GMT"
	if since := m.conditionalSince(true); since != "" {
		t.Errorf("手動更新は If-Modified-Since を送らない（空文字）べき, got %q", since)
	}
}

// TestAutoPollKeepsConditional は自動ポーリング（force=false）が
// 直近の Last-Modified を条件付きリクエストに使うことを検証する。
func TestAutoPollKeepsConditional(t *testing.T) {
	m := newModel(nil, notify.Noop{}, config.Default())
	m.lastModified = "Thu, 01 Jan 2026 00:00:00 GMT"
	if since := m.conditionalSince(false); since != m.lastModified {
		t.Errorf("自動ポーリングは直近の Last-Modified を送るべき, got %q want %q", since, m.lastModified)
	}
}

// TestForceFetchUpdatesLastModified は強制取得の 200 応答後に m.lastModified が
// 更新され（handleFetched は NotModified チェック前に無条件更新）、次回の自動ポーリングが
// 条件付きに戻ることを検証する。
func TestForceFetchUpdatesLastModified(t *testing.T) {
	m := loadedModel(false)
	tm, _ := m.Update(fetchedMsg{res: github.ListResult{
		Notifications: sampleNotifs(), LastModified: "Fri, 02 Jan 2026 12:00:00 GMT",
	}})
	m = tm.(Model)
	if m.lastModified != "Fri, 02 Jan 2026 12:00:00 GMT" {
		t.Fatalf("200 応答で lastModified が更新されるべき, got %q", m.lastModified)
	}
	if since := m.conditionalSince(false); since != "Fri, 02 Jan 2026 12:00:00 GMT" {
		t.Errorf("次回の自動ポーリングは更新後の Last-Modified を条件付きに使うべき, got %q", since)
	}
}

// TestForceFetchSameSetNoNotification は強制取得で同じ未読セットが返っても
// Differ が空を返し OS 通知が発火しないこと（304 スキップが無くなっても通知スパムしない）を検証する。
func TestForceFetchSameSetNoNotification(t *testing.T) {
	rec := &recordingNotifier{}
	cfg := config.Default()
	m := newModel(nil, rec, cfg)
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = tm.(Model)

	// 1 回目の取得（初回は Differ がバックログ全件を抑制 → 通知なし）。
	tm, cmd := m.Update(fetchedMsg{res: github.ListResult{Notifications: sampleNotifs()}})
	m = tm.(Model)
	runCmd(cmd)

	// 2 回目（強制取得相当）: 同じ未読セットなので Differ は空 → 通知なし。
	tm, cmd = m.Update(fetchedMsg{res: github.ListResult{Notifications: sampleNotifs()}})
	m = tm.(Model)
	runCmd(cmd)

	if rec.calls != 0 {
		t.Errorf("同じ未読セットの再取得では OS 通知が発火しないべき, got calls=%d", rec.calls)
	}
}

// runCmd は Cmd（および tea.Batch が返す BatchMsg 内の子 Cmd）を再帰的に実行する。
// client(nil) に触れない notify/refresh 系の検証に使う。
func runCmd(cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			runCmd(c)
		}
	}
}

// TestQuitKey は q キーで Quit コマンドが返ることを検証する。
func TestQuitKey(t *testing.T) {
	m := newModel(nil, notify.Noop{}, config.Default())
	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: 'q', Text: "q"}))
	if cmd == nil {
		t.Fatal("q で Quit コマンドを期待したが nil")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("q の Cmd は QuitMsg を期待, got %T", cmd())
	}
}

// TestEmptyListActions は空リストでアクションを呼んでもパニック/誤動作しないことを検証する。
func TestEmptyListActions(t *testing.T) {
	m := newModel(nil, notify.Noop{}, config.Default())
	if cmd := m.markSelectedRead(); cmd != nil {
		t.Error("空リストの既読は nil Cmd を期待")
	}
	if cmd := m.openSelected(true); cmd != nil {
		t.Error("空リストの open は nil Cmd を期待")
	}
}

// TestMarkAllReadConfirm は ctrl+a の二度押し確認を検証する。
func TestMarkAllReadConfirm(t *testing.T) {
	ctrlA := tea.KeyPressMsg(tea.Key{Code: 'a', Mod: tea.ModCtrl})

	m := newModel(nil, notify.Noop{}, config.Default())
	// 1 回目: 確認待ちになり、まだ実行コマンドは出ない
	tm, cmd := m.Update(ctrlA)
	m = tm.(Model)
	if !m.confirmingAllRead {
		t.Fatal("1 回目の ctrl+a で確認待ちになるべき")
	}
	if cmd != nil {
		t.Fatal("1 回目はまだ全既読を実行しない")
	}
	// 2 回目: 実行コマンドが返る
	tm, cmd = m.Update(ctrlA)
	m = tm.(Model)
	if m.confirmingAllRead {
		t.Error("2 回目で確認状態が解除されるべき")
	}
	if cmd == nil {
		t.Fatal("2 回目で全既読コマンドを期待")
	}
	// cmd() の実行は client(nil) を呼ぶため行わない（コマンドが返ることのみ検証）。

	// 1 回目の後に別キーが来たら確認はキャンセルされる
	m2 := newModel(nil, notify.Noop{}, config.Default())
	tm, _ = m2.Update(ctrlA)
	m2 = tm.(Model)
	tm, _ = m2.Update(tea.KeyPressMsg(tea.Key{Code: 'j', Text: "j"}))
	m2 = tm.(Model)
	if m2.confirmingAllRead {
		t.Error("別キーで確認がキャンセルされるべき")
	}
}
