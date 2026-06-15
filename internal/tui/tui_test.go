package tui

import (
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/yshrsmz/octobell/internal/config"
	"github.com/yshrsmz/octobell/internal/github"
	"github.com/yshrsmz/octobell/internal/notify"
)

// reANSI は表示検証のため ANSI エスケープ（色コード）を除去する。
var reANSI = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string { return reANSI.ReplaceAllString(s, "") }

func sampleNotifs() []github.Notification {
	repo := github.Repository{FullName: "o/r", HTMLURL: "https://github.com/o/r"}
	return []github.Notification{
		{ID: "1", Unread: true, Reason: "mention", Repository: repo,
			Subject: github.Subject{Title: "PR one", Type: "PullRequest", URL: "https://api.github.com/repos/o/r/pulls/1"}},
		{ID: "2", Unread: true, Reason: "review_requested", Repository: repo,
			Subject: github.Subject{Title: "Issue two", Type: "Issue", URL: "https://api.github.com/repos/o/r/issues/2"}},
	}
}

// TestItemTitleShowsNumber は Issue/PR の主行に #<番号> が付き、番号を持たない種別では
// 付かないこと、および FilterValue に番号が含まれることを検証する。
func TestItemTitleShowsNumber(t *testing.T) {
	repo := github.Repository{FullName: "o/r", HTMLURL: "https://github.com/o/r"}

	pr := item{n: github.Notification{Unread: true, Subject: github.Subject{
		Title: "PR one", Type: "PullRequest", URL: "https://api.github.com/repos/o/r/pulls/42"}, Repository: repo}}
	if got := pr.Title(); got != "● #42 PR one" {
		t.Errorf("PR の Title = %q, want %q", got, "● #42 PR one")
	}
	if fv := pr.FilterValue(); !strings.Contains(fv, "#42") {
		t.Errorf("FilterValue に #42 を含むべき, got %q", fv)
	}

	// 既読 Issue は未読マークなしで番号が付く。
	iss := item{n: github.Notification{Unread: false, Subject: github.Subject{
		Title: "Issue two", Type: "Issue", URL: "https://api.github.com/repos/o/r/issues/7"}, Repository: repo}}
	if got := iss.Title(); got != "  #7 Issue two" {
		t.Errorf("Issue の Title = %q, want %q", got, "  #7 Issue two")
	}

	// 番号を持たない種別（Commit）は # を付さない。
	commit := item{n: github.Notification{Unread: true, Reason: "subscribed", Subject: github.Subject{
		Title: "Build", Type: "Commit", URL: "https://api.github.com/repos/o/r/commits/abc123"}, Repository: repo}}
	if got := commit.Title(); got != "● Build" {
		t.Errorf("Commit の Title = %q, want %q", got, "● Build")
	}
	// 番号が付与されないこと（4 フィールド連結のまま）を厳密に検証する。
	// title 等に "#" が含まれても誤検知しないよう、期待値と完全一致で確認する。
	if got, want := commit.FilterValue(), "o/r Commit subscribed Build"; got != want {
		t.Errorf("番号なし種別の FilterValue = %q, want %q（番号を付さない）", got, want)
	}
}

// stateChangeNotif は reason=state_change の PR 通知を返す。
func stateChangeNotif(id string, updated time.Time) github.Notification {
	return github.Notification{
		ID: id, Unread: true, Reason: "state_change", UpdatedAt: updated,
		Repository: github.Repository{FullName: "o/r", HTMLURL: "https://github.com/o/r"},
		Subject:    github.Subject{Title: "PR x", Type: "PullRequest", URL: "https://api.github.com/repos/o/r/pulls/9"},
	}
}

// TestDescriptionEnrichment はエンリッチ前は reason のみ、エンリッチ後は state_change(<状態>)
// が副行に出ることを検証する（色コードは除去して比較）。
func TestDescriptionEnrichment(t *testing.T) {
	before := item{n: stateChangeNotif("9", time.Unix(100, 0))}
	if got := stripANSI(before.Description()); got != "o/r · PullRequest · state_change" {
		t.Errorf("エンリッチ前の副行 = %q, want %q", got, "o/r · PullRequest · state_change")
	}

	after := item{n: stateChangeNotif("9", time.Unix(100, 0)), state: github.StateMerged}
	if got := stripANSI(after.Description()); got != "o/r · PullRequest · state_change(merged)" {
		t.Errorf("エンリッチ後の副行 = %q, want %q", got, "o/r · PullRequest · state_change(merged)")
	}
	// 着色されている（生文字列に ANSI を含む）こと。
	if after.Description() == stripANSI(after.Description()) {
		t.Error("(merged) 部分は色付けされ ANSI を含むべき")
	}
	// FilterValue に実状態が含まれる。
	if !strings.Contains(after.FilterValue(), "merged") {
		t.Errorf("FilterValue に merged を含むべき, got %q", after.FilterValue())
	}
}

// TestDescriptionOtherReasonNoBadge は state があっても reason が state_change 以外なら
// 付記しないことを検証する（item.state は通常 state_change 以外には付かないが防御的に確認）。
func TestDescriptionOtherReasonNoBadge(t *testing.T) {
	n := github.Notification{
		Reason: "mention", Repository: github.Repository{FullName: "o/r"},
		Subject: github.Subject{Title: "x", Type: "PullRequest", URL: "https://api.github.com/repos/o/r/pulls/9"},
	}
	it := item{n: n, state: github.StateMerged}
	if got := stripANSI(it.Description()); got != "o/r · PullRequest · mention" {
		t.Errorf("他 reason は付記しない, got %q", got)
	}
}

// TestEnrichDisabledNoCmd は enrich_state=false のとき handleFetched が
// エンリッチ Cmd を発行しないことを検証する。
func TestEnrichDisabledNoCmd(t *testing.T) {
	cfg := config.Default()
	cfg.EnrichState = false
	m := newModel(nil, notify.Noop{}, cfg)
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = tm.(Model)
	m.notifs = []github.Notification{stateChangeNotif("9", time.Unix(100, 0))}
	if cmd := m.enrichCmds(); cmd != nil {
		t.Error("enrich_state=false ではエンリッチ Cmd を発行しないべき")
	}
}

// TestEnrichCacheHitSkips は updated_at 一致のキャッシュがあれば再取得 Cmd を出さず、
// 不一致なら出すことを検証する。
func TestEnrichCacheHitSkips(t *testing.T) {
	m := newModel(nil, notify.Noop{}, config.Default())
	updated := time.Unix(100, 0)
	m.notifs = []github.Notification{stateChangeNotif("9", updated)}

	// キャッシュ無し → Cmd あり
	if cmd := m.enrichCmds(); cmd == nil {
		t.Fatal("キャッシュ無しではエンリッチ Cmd を出すべき")
	}
	// updated_at 一致キャッシュ → Cmd なし
	m.enrichCache["9"] = enrichEntry{updatedAt: updated, state: github.StateMerged}
	if cmd := m.enrichCmds(); cmd != nil {
		t.Error("updated_at 一致のキャッシュがあれば再取得しないべき")
	}
	// updated_at 不一致 → Cmd あり
	m.notifs = []github.Notification{stateChangeNotif("9", time.Unix(200, 0))}
	if cmd := m.enrichCmds(); cmd == nil {
		t.Error("updated_at が変われば再取得するべき")
	}
}

// TestHandleEnrichedUpdatesDisplay は enrichedMsg 成功でキャッシュが更新され、
// 副行が state_change(<状態>) に切り替わることを検証する。失敗は据え置く。
func TestHandleEnrichedUpdatesDisplay(t *testing.T) {
	m := loadedModel(false)
	updated := time.Unix(100, 0)
	m.notifs = []github.Notification{stateChangeNotif("9", updated)}
	_ = m.refreshItems()

	// 成功メッセージ
	tm, _ := m.Update(enrichedMsg{id: "9", updatedAt: updated, state: github.StateMerged})
	m = tm.(Model)
	if e, ok := m.enrichCache["9"]; !ok || e.state != github.StateMerged {
		t.Fatalf("キャッシュが merged で更新されるべき, got %+v ok=%v", e, ok)
	}
	it := m.list.Items()[0].(item)
	if it.state != github.StateMerged {
		t.Errorf("item.state が merged に反映されるべき, got %q", it.state)
	}

	// 失敗メッセージは据え置き（キャッシュを壊さない）
	tm, _ = m.Update(enrichedMsg{id: "9", updatedAt: updated, err: errors.New("boom")})
	m = tm.(Model)
	if m.enrichCache["9"].state != github.StateMerged {
		t.Error("失敗メッセージは既存キャッシュを据え置くべき")
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
		enrichedMsg{id: "1", state: github.StateOpen}, enrichedMsg{id: "1", err: errors.New("boom")},
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
