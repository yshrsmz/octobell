package github

import "testing"

func TestBrowserURL(t *testing.T) {
	repo := Repository{FullName: "owner/repo", HTMLURL: "https://github.com/owner/repo"}
	cases := []struct {
		name string
		sub  Subject
		want string
	}{
		{"PR は単数形 pull に変換", Subject{Type: "PullRequest", URL: "https://api.github.com/repos/owner/repo/pulls/42"}, "https://github.com/owner/repo/pull/42"},
		{"Issue は issues のまま", Subject{Type: "Issue", URL: "https://api.github.com/repos/owner/repo/issues/7"}, "https://github.com/owner/repo/issues/7"},
		{"Commit は commit に変換", Subject{Type: "Commit", URL: "https://api.github.com/repos/owner/repo/commits/abc123"}, "https://github.com/owner/repo/commit/abc123"},
		{"解決不能はリポにフォールバック", Subject{Type: "Release", URL: "https://api.github.com/repos/owner/repo/releases/99"}, "https://github.com/owner/repo"},
		{"subject.url 空はリポにフォールバック", Subject{Type: "Discussion", URL: ""}, "https://github.com/owner/repo"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			n := Notification{Subject: c.sub, Repository: repo}
			if got := n.BrowserURL(); got != c.want {
				t.Errorf("BrowserURL() = %q, want %q", got, c.want)
			}
		})
	}
}

func TestSubjectNumber(t *testing.T) {
	cases := []struct {
		name string
		sub  Subject
		want string
	}{
		{"PR は番号を返す", Subject{Type: "PullRequest", URL: "https://api.github.com/repos/owner/repo/pulls/42"}, "42"},
		{"Issue は番号を返す", Subject{Type: "Issue", URL: "https://api.github.com/repos/owner/repo/issues/7"}, "7"},
		{"Commit は空", Subject{Type: "Commit", URL: "https://api.github.com/repos/owner/repo/commits/abc123"}, ""},
		{"Release は空", Subject{Type: "Release", URL: "https://api.github.com/repos/owner/repo/releases/99"}, ""},
		{"subject.url 空は空", Subject{Type: "Discussion", URL: ""}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			n := Notification{Subject: c.sub}
			if got := n.SubjectNumber(); got != c.want {
				t.Errorf("SubjectNumber() = %q, want %q", got, c.want)
			}
		})
	}
}
