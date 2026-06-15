package tui

import (
	"image/color"

	"charm.land/lipgloss/v2"

	"github.com/yshrsmz/octobell/internal/github"
)

// item は list.Item として通知 1 件をラップする。
// state はエンリッチ済みの実状態（未取得・対象外は github.StateUnknown）。
type item struct {
	n     github.Notification
	state github.SubjectState
}

// FilterValue は `/` フィルタの対象文字列（リポ名・種別・理由・タイトル、Issue/PR 番号、実状態）。
// 番号は "#42" として含めるため、"42" でも "#42" でも部分一致でヒットする。
// 実状態を取得済みなら（例 merged）でも絞り込めるよう連結する。
func (i item) FilterValue() string {
	v := i.n.Repository.FullName + " " + i.n.Subject.Type + " " + i.n.Reason + " " + i.n.Subject.Title
	if num := i.n.SubjectNumber(); num != "" {
		v += " #" + num
	}
	if i.state != github.StateUnknown {
		v += " " + i.state.String()
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
// reason=state_change で実状態を取得済みの場合は理由を `state_change(<実状態>)` に整形し、
// `(<実状態>)` を括弧ごと状態色で着色する（state_change 本体は通常色）。
// 着色は文字列末尾に置く——着色が中間にあると lipgloss の reset で delegate の
// 副行色が打ち切られ、reset 後ろの文字（閉じ括弧）が既定色に化けるため。
func (i item) Description() string {
	base := i.n.Repository.FullName + " · " + i.n.Subject.Type + " · "
	if i.n.Reason == github.ReasonStateChange && i.state != github.StateUnknown {
		colored := lipgloss.NewStyle().Foreground(stateColor(i.state)).Render("(" + i.state.String() + ")")
		return base + "state_change" + colored
	}
	return base + i.n.Reason
}

// stateColor は実状態に対応する前景色（GitHub 標準カラーに倣う）を返す。
// ダーク端末で沈まないよう明るめの色を選ぶ。
func stateColor(s github.SubjectState) color.Color {
	switch s {
	case github.StateOpen:
		return lipgloss.Color("#3FB950") // 緑
	case github.StateMerged, github.StateClosedCompleted:
		return lipgloss.Color("#A371F7") // 紫
	case github.StateDraft, github.StateClosedNotPlanned:
		return lipgloss.Color("#8B949E") // 灰
	case github.StateClosed:
		return lipgloss.Color("#F85149") // 赤（PR 未マージ close）
	default:
		return lipgloss.Color("#B2B2B2")
	}
}
