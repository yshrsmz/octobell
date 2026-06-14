// Package tui は Bubble Tea v2 ベースの通知一覧 UI を提供する。
package tui

import (
	"context"
	"fmt"
	"io"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/cli/go-gh/v2/pkg/browser"

	"github.com/yshrsmz/octobell/internal/config"
	"github.com/yshrsmz/octobell/internal/github"
	"github.com/yshrsmz/octobell/internal/notify"
)

// --- メッセージ型 ---

type tickMsg struct{}
type fetchedMsg struct {
	res github.ListResult
	err error
}
type markedMsg struct{ err error }
type markedAllMsg struct{ err error }
type statusMsg string
type clearStatusMsg struct{}

// Model は TUI の状態。
type Model struct {
	client   *github.Client
	notifier notify.Notifier
	cfg      config.Config
	differ   *notify.Differ
	browser  *browser.Browser

	list    list.Model
	keys    keyMap
	spinner spinner.Model

	notifs       []github.Notification
	lastModified string
	pollInterval int // 直近の X-Poll-Interval（秒）
	loading           bool
	status            string
	lastSync          time.Time
	confirmingAllRead bool // ctrl+a 二度押し確認の途中か
}

func newModel(client *github.Client, notifier notify.Notifier, cfg config.Config) Model {
	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	l.Title = "octobell"
	km := newKeyMap()
	l.AdditionalShortHelpKeys = km.bindings
	l.AdditionalFullHelpKeys = km.bindings

	return Model{
		client:   client,
		notifier: notifier,
		cfg:      cfg,
		differ:   notify.NewDiffer(),
		browser:  browser.New("", io.Discard, io.Discard),
		list:     l,
		keys:     km,
		spinner:  spinner.New(spinner.WithSpinner(spinner.Dot)),
		loading:  true,
	}
}

// Init は初回フェッチ・スピナー・ポーリングタイマーを起動する。
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.fetchCmd(), m.tickCmd())
}

// pollSeconds は max(ユーザー設定, X-Poll-Interval) を返す（GitHub の下限を尊重）。
func (m Model) pollSeconds() int {
	d := m.pollInterval
	if m.cfg.PollSeconds > d {
		d = m.cfg.PollSeconds
	}
	if d < 1 {
		d = 60
	}
	return d
}

func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(time.Duration(m.pollSeconds())*time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m Model) fetchCmd() tea.Cmd {
	client := m.client
	opts := github.ListOptions{All: m.cfg.All, Participating: m.cfg.Participating, PerPage: m.cfg.PerPage}
	lm := m.lastModified
	return func() tea.Msg {
		res, err := client.List(context.Background(), opts, lm)
		return fetchedMsg{res: res, err: err}
	}
}

// Update はメッセージを処理する。
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height-1) // 最下部 1 行をステータスに確保
		return m, nil

	case tickMsg:
		var cmds []tea.Cmd
		if !m.loading {
			m.loading = true
			cmds = append(cmds, m.fetchCmd())
		}
		cmds = append(cmds, m.tickCmd())
		return m, tea.Batch(cmds...)

	case fetchedMsg:
		m.loading = false
		return m.handleFetched(msg)

	case markedMsg:
		if msg.err != nil {
			m.status = "既読化に失敗: " + msg.err.Error()
		} else {
			m.status = "既読にしました"
		}
		return m, clearStatusAfter()

	case markedAllMsg:
		if msg.err != nil {
			m.status = "全既読に失敗: " + msg.err.Error()
		} else {
			m.status = "すべて既読にしました"
			ids := make([]string, len(m.notifs))
			for i, n := range m.notifs {
				ids[i] = n.ID
			}
			m.setReadLocal(ids...)
		}
		return m, clearStatusAfter()

	case statusMsg:
		m.status = string(msg)
		return m, clearStatusAfter()

	case clearStatusMsg:
		m.status = ""
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyPressMsg:
		// フィルタ入力中は全キーを list に委ねる。
		if !m.list.SettingFilter() {
			// 全既読の二度押し確認中に別キーが来たらキャンセルする。
			if m.confirmingAllRead && !key.Matches(msg, m.keys.MarkAllRead) {
				m.confirmingAllRead = false
				m.status = ""
			}
			switch {
			case key.Matches(msg, m.keys.Quit):
				return m, tea.Quit
			case key.Matches(msg, m.keys.Open):
				return m, m.openSelected(true)
			case key.Matches(msg, m.keys.OpenOnly):
				return m, m.openSelected(false)
			case key.Matches(msg, m.keys.MarkRead):
				return m, m.markSelectedRead()
			case key.Matches(msg, m.keys.MarkAllRead):
				// 取り消し不能な一括操作のため二度押しで確認する。
				if m.confirmingAllRead {
					m.confirmingAllRead = false
					m.status = ""
					return m, markAllReadCmd(m.client)
				}
				m.confirmingAllRead = true
				m.status = "すべて既読にします: もう一度 ctrl+a で実行（他のキーでキャンセル）"
				return m, nil
			case key.Matches(msg, m.keys.Refresh):
				if !m.loading {
					m.loading = true
					return m, m.fetchCmd()
				}
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) handleFetched(msg fetchedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.status = "取得に失敗: " + msg.err.Error()
		return m, clearStatusAfter()
	}
	res := msg.res
	if res.PollInterval > 0 {
		m.pollInterval = res.PollInterval
	}
	m.lastModified = res.LastModified
	m.lastSync = time.Now()
	if res.NotModified {
		return m, nil
	}
	m.notifs = res.Notifications
	cmd := m.refreshItems()

	// 新着（未読の新規 ID）のみ OS 通知する。
	ids := make([]string, 0, len(m.notifs))
	for _, n := range m.notifs {
		if n.Unread {
			ids = append(ids, n.ID)
		}
	}
	fresh := m.differ.New(ids)
	var notifyCmd tea.Cmd
	if len(fresh) > 0 {
		notifyCmd = m.notifyCmd(fresh)
	}
	return m, tea.Batch(cmd, notifyCmd)
}

func (m *Model) refreshItems() tea.Cmd {
	idx := m.list.Index()
	items := make([]list.Item, len(m.notifs))
	for i, n := range m.notifs {
		items[i] = item{n: n}
	}
	cmd := m.list.SetItems(items)
	// 項目除去で件数が減ったとき末尾にあふれないようクランプする。
	if len(items) > 0 {
		if idx < 0 {
			idx = 0
		}
		if idx >= len(items) {
			idx = len(items) - 1
		}
		m.list.Select(idx)
	}
	return cmd
}

// setReadLocal は既読化した ID 群をローカルに反映する（楽観的更新）。
// 既読のみ表示（All=false）なら一覧から除去し、全件表示（All=true）なら未読フラグのみ落とす。
func (m *Model) setReadLocal(ids ...string) {
	if len(ids) == 0 {
		return
	}
	idset := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		idset[id] = struct{}{}
	}
	if m.cfg.All {
		for i := range m.notifs {
			if _, ok := idset[m.notifs[i].ID]; ok {
				m.notifs[i].Unread = false
			}
		}
	} else {
		kept := m.notifs[:0]
		for _, n := range m.notifs {
			if _, ok := idset[n.ID]; !ok {
				kept = append(kept, n)
			}
		}
		m.notifs = kept
	}
	m.refreshItems()
}

// openSelected は選択中の通知をブラウザで開く。markRead かつ設定が有効なら既読化（楽観的更新）。
func (m *Model) openSelected(markRead bool) tea.Cmd {
	it, ok := m.list.SelectedItem().(item)
	if !ok {
		return nil
	}
	url := it.n.BrowserURL()
	b := m.browser
	cmds := []tea.Cmd{func() tea.Msg {
		if err := b.Browse(url); err != nil {
			return statusMsg("ブラウザ起動に失敗: " + err.Error())
		}
		return statusMsg("ブラウザで開きました")
	}}
	if markRead && m.cfg.MarkReadOnOpen && it.n.Unread {
		m.setReadLocal(it.n.ID) // 楽観的更新（再ポーリングは待たない）
		cmds = append(cmds, markReadCmd(m.client, it.n.ID))
	}
	return tea.Batch(cmds...)
}

func (m *Model) markSelectedRead() tea.Cmd {
	it, ok := m.list.SelectedItem().(item)
	if !ok || !it.n.Unread {
		return nil
	}
	m.setReadLocal(it.n.ID)
	return markReadCmd(m.client, it.n.ID)
}

func (m Model) notifyCmd(ids []string) tea.Cmd {
	byID := make(map[string]github.Notification, len(m.notifs))
	for _, x := range m.notifs {
		byID[x.ID] = x
	}
	notifs := make([]github.Notification, 0, len(ids))
	for _, id := range ids {
		if x, ok := byID[id]; ok {
			notifs = append(notifs, x)
		}
	}
	n := m.notifier
	return func() tea.Msg {
		if len(notifs) == 0 {
			return nil
		}
		var title, body string
		if len(notifs) == 1 {
			title = "GitHub: " + notifs[0].Repository.FullName
			body = notifs[0].Subject.Title
		} else {
			title = fmt.Sprintf("GitHub: 新着通知 %d 件", len(notifs))
			for i, x := range notifs {
				if i >= 3 {
					body += fmt.Sprintf("…ほか %d 件", len(notifs)-3)
					break
				}
				body += "• " + x.Subject.Title + "\n"
			}
		}
		_ = n.Notify(title, body)
		return nil
	}
}

func markReadCmd(c *github.Client, id string) tea.Cmd {
	return func() tea.Msg {
		return markedMsg{err: c.MarkThreadRead(context.Background(), id)}
	}
}

func markAllReadCmd(c *github.Client) tea.Cmd {
	return func() tea.Msg {
		return markedAllMsg{err: c.MarkAllRead(context.Background())}
	}
}

func clearStatusAfter() tea.Cmd {
	return tea.Tick(4*time.Second, func(time.Time) tea.Msg { return clearStatusMsg{} })
}

// View は一覧 + 最下部のステータス行を描画する。
func (m Model) View() tea.View {
	v := tea.NewView(m.list.View() + "\n" + m.statusLine())
	v.AltScreen = true
	return v
}

func (m Model) statusLine() string {
	if m.status != "" {
		return m.status
	}
	if m.loading {
		return m.spinner.View() + " 更新中…"
	}
	last := "—"
	if !m.lastSync.IsZero() {
		last = m.lastSync.Format("15:04:05")
	}
	return fmt.Sprintf("最終更新 %s · %ds 間隔 · ?: ヘルプ", last, m.pollSeconds())
}

// Run は TUI を起動する。
func Run(ctx context.Context, client *github.Client, notifier notify.Notifier, cfg config.Config) error {
	p := tea.NewProgram(newModel(client, notifier, cfg), tea.WithContext(ctx))
	_, err := p.Run()
	return err
}
