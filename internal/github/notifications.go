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
