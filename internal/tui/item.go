package tui

import "github.com/yshrsmz/octobell/internal/github"

// item は list.Item として通知 1 件をラップする。
type item struct{ n github.Notification }

// FilterValue は `/` フィルタの対象文字列（リポ名・種別・理由・タイトル）。
func (i item) FilterValue() string {
	return i.n.Repository.FullName + " " + i.n.Subject.Type + " " + i.n.Reason + " " + i.n.Subject.Title
}

// Title は一覧の主行（未読マーク + タイトル）。
func (i item) Title() string {
	if i.n.Unread {
		return "● " + i.n.Subject.Title
	}
	return "  " + i.n.Subject.Title
}

// Description は副行（リポ名 · 種別 · 理由）。
func (i item) Description() string {
	return i.n.Repository.FullName + " · " + i.n.Subject.Type + " · " + i.n.Reason
}
