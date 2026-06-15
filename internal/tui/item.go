package tui

import "github.com/yshrsmz/octobell/internal/github"

// item は list.Item として通知 1 件をラップする。
type item struct{ n github.Notification }

// FilterValue は `/` フィルタの対象文字列（リポ名・種別・理由・タイトル、および Issue/PR 番号）。
// 番号は "#42" として含めるため、"42" でも "#42" でも部分一致でヒットする。
func (i item) FilterValue() string {
	v := i.n.Repository.FullName + " " + i.n.Subject.Type + " " + i.n.Reason + " " + i.n.Subject.Title
	if num := i.n.SubjectNumber(); num != "" {
		v += " #" + num
	}
	return v
}

// Title は一覧の主行（未読マーク + Issue/PR 番号 + タイトル）。
// 番号を持たない種別（Commit / Release など）では番号を付さない。
func (i item) Title() string {
	mark := "  "
	if i.n.Unread {
		mark = "● "
	}
	if num := i.n.SubjectNumber(); num != "" {
		return mark + "#" + num + " " + i.n.Subject.Title
	}
	return mark + i.n.Subject.Title
}

// Description は副行（リポ名 · 種別 · 理由）。
func (i item) Description() string {
	return i.n.Repository.FullName + " · " + i.n.Subject.Type + " · " + i.n.Reason
}
