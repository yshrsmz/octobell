// Package github は GitHub Notifications API（gh の認証を再利用）へのアクセスを提供する。
package github

import (
	"fmt"
	"regexp"
	"time"
)

// Notification は GitHub の通知スレッド 1 件を表す。
// 参考: https://docs.github.com/en/rest/activity/notifications
type Notification struct {
	ID         string     `json:"id"`
	Unread     bool       `json:"unread"`
	Reason     string     `json:"reason"`
	UpdatedAt  time.Time  `json:"updated_at"`
	LastReadAt *time.Time `json:"last_read_at"`
	URL        string     `json:"url"`
	Subject    Subject    `json:"subject"`
	Repository Repository `json:"repository"`
}

// Subject は通知の対象（PR / Issue / Commit など）。
type Subject struct {
	Title            string `json:"title"`
	URL              string `json:"url"`
	LatestCommentURL string `json:"latest_comment_url"`
	Type             string `json:"type"`
}

// Repository は通知が属するリポジトリ。
type Repository struct {
	FullName string `json:"full_name"`
	HTMLURL  string `json:"html_url"`
	Private  bool   `json:"private"`
}

var (
	// subject.url は API URL。例: https://api.github.com/repos/owner/repo/(issues|pulls)/42
	reIssueOrPull = regexp.MustCompile(`/repos/[^/]+/[^/]+/(issues|pulls)/(\d+)$`)
	// 例: https://api.github.com/repos/owner/repo/commits/<sha>
	reCommit = regexp.MustCompile(`/repos/[^/]+/[^/]+/commits/(.+)$`)
)

// BrowserURL は subject.url（API URL）から、ブラウザで開ける html_url を組み立てる。
// repository.html_url をベースにするためエンタープライズの host にも追従する。
// 解決できない種別（Release / Discussion / CheckSuite 等）は repository.html_url にフォールバックする。
func (n Notification) BrowserURL() string {
	base := n.Repository.HTMLURL
	u := n.Subject.URL
	if u == "" || base == "" {
		return base
	}
	if m := reIssueOrPull.FindStringSubmatch(u); m != nil {
		seg := "issues"
		if m[1] == "pulls" { // PR の html_url は単数形 /pull/N
			seg = "pull"
		}
		return fmt.Sprintf("%s/%s/%s", base, seg, m[2])
	}
	if m := reCommit.FindStringSubmatch(u); m != nil {
		return fmt.Sprintf("%s/commit/%s", base, m[1])
	}
	return base
}

// SubjectNumber は subject.url（API URL）から Issue / PullRequest の番号を返す。
// reIssueOrPull を BrowserURL と共有し group2（番号）を取り出す。
// issues/pulls 以外の種別、または subject.url が空の場合は "" を返す。
func (n Notification) SubjectNumber() string {
	if m := reIssueOrPull.FindStringSubmatch(n.Subject.URL); m != nil {
		return m[2]
	}
	return ""
}

// ReasonStateChange は GitHub Notifications の reason フィールドの「状態変化」値。
// 実状態のエンリッチ対象判定（Enrichable）と TUI の副行整形で共有する。
const ReasonStateChange = "state_change"

// Enrichable は、この通知が subject 詳細の追加取得（実状態のエンリッチ）対象かを返す。
// reason=state_change の Issue / PullRequest で subject.url を持つもののみが対象。
func (n Notification) Enrichable() bool {
	if n.Reason != ReasonStateChange || n.Subject.URL == "" {
		return false
	}
	return n.Subject.Type == "Issue" || n.Subject.Type == "PullRequest"
}

// SubjectState は Issue / PullRequest の実状態。Notifications API は返さないため
// subject 詳細を追加取得して導出する。
type SubjectState string

const (
	StateUnknown          SubjectState = ""                   // 未取得・対象外
	StateOpen             SubjectState = "open"               // Issue / PR ともに open
	StateDraft            SubjectState = "draft"              // PR の draft
	StateMerged           SubjectState = "merged"             // PR が merge 済み
	StateClosed           SubjectState = "closed"             // PR が未マージで close
	StateClosedCompleted  SubjectState = "closed-completed"   // Issue を completed で close
	StateClosedNotPlanned SubjectState = "closed-not_planned" // Issue を not_planned で close
)

// String は表示用の状態文字列を返す。
func (s SubjectState) String() string { return string(s) }

// subjectDetail は Issue / PullRequest 詳細レスポンスのうち状態導出に必要なフィールド。
type subjectDetail struct {
	State       string `json:"state"`        // open | closed（Issue / PR 共通）
	StateReason string `json:"state_reason"` // completed | not_planned | reopened（Issue のみ）
	Merged      bool   `json:"merged"`       // PR のみ
	Draft       bool   `json:"draft"`        // PR のみ
}

// deriveState は subject の種別と詳細から実状態を導出する。
// PR: merged → merged / closed（未マージ）→ closed / draft → draft / それ以外 → open
// Issue: closed のとき state_reason で分岐（not_planned 以外は completed 扱い）/ それ以外 → open
func deriveState(subjectType string, d subjectDetail) SubjectState {
	switch subjectType {
	case "PullRequest":
		switch {
		case d.Merged:
			return StateMerged
		case d.State == "closed":
			return StateClosed
		case d.Draft:
			return StateDraft
		default:
			return StateOpen
		}
	case "Issue":
		if d.State == "closed" {
			if d.StateReason == "not_planned" {
				return StateClosedNotPlanned
			}
			return StateClosedCompleted
		}
		return StateOpen
	default:
		return StateUnknown
	}
}
