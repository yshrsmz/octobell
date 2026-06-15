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

func TestEnrichable(t *testing.T) {
	url := "https://api.github.com/repos/owner/repo/pulls/42"
	cases := []struct {
		name   string
		reason string
		sub    Subject
		want   bool
	}{
		{"state_change の PR は対象", "state_change", Subject{Type: "PullRequest", URL: url}, true},
		{"state_change の Issue は対象", "state_change", Subject{Type: "Issue", URL: "https://api.github.com/repos/owner/repo/issues/7"}, true},
		{"reason が他なら対象外", "mention", Subject{Type: "PullRequest", URL: url}, false},
		{"対象外種別は対象外", "state_change", Subject{Type: "Commit", URL: "https://api.github.com/repos/owner/repo/commits/abc"}, false},
		{"subject.url 空は対象外", "state_change", Subject{Type: "PullRequest", URL: ""}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			n := Notification{Reason: c.reason, Subject: c.sub}
			if got := n.Enrichable(); got != c.want {
				t.Errorf("Enrichable() = %v, want %v", got, c.want)
			}
		})
	}
}

func TestDeriveState(t *testing.T) {
	cases := []struct {
		name        string
		subjectType string
		detail      subjectDetail
		want        SubjectState
	}{
		{"PR merged", "PullRequest", subjectDetail{State: "closed", Merged: true}, StateMerged},
		{"PR closed 未マージ", "PullRequest", subjectDetail{State: "closed", Merged: false}, StateClosed},
		{"PR draft", "PullRequest", subjectDetail{State: "open", Draft: true}, StateDraft},
		{"PR open", "PullRequest", subjectDetail{State: "open"}, StateOpen},
		{"Issue completed", "Issue", subjectDetail{State: "closed", StateReason: "completed"}, StateClosedCompleted},
		{"Issue not_planned", "Issue", subjectDetail{State: "closed", StateReason: "not_planned"}, StateClosedNotPlanned},
		{"Issue closed で reason 空は completed 扱い", "Issue", subjectDetail{State: "closed"}, StateClosedCompleted},
		{"Issue open", "Issue", subjectDetail{State: "open", StateReason: "reopened"}, StateOpen},
		{"対象外種別は unknown", "Commit", subjectDetail{State: "open"}, StateUnknown},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := deriveState(c.subjectType, c.detail); got != c.want {
				t.Errorf("deriveState(%q, %+v) = %q, want %q", c.subjectType, c.detail, got, c.want)
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
