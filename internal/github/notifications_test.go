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
