package notify

import "testing"

func TestDetectGhostty(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want bool
	}{
		{"TERM=xterm-ghostty", map[string]string{"TERM": "xterm-ghostty"}, true},
		{"TERM_PROGRAM=ghostty", map[string]string{"TERM_PROGRAM": "ghostty"}, true},
		{"両方一致", map[string]string{"TERM": "xterm-ghostty", "TERM_PROGRAM": "ghostty"}, true},
		{"TERM=xterm-256color のみ", map[string]string{"TERM": "xterm-256color"}, false},
		{"TERM_PROGRAM=vscode", map[string]string{"TERM": "xterm-256color", "TERM_PROGRAM": "vscode"}, false},
		{"未設定", map[string]string{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getenv := func(k string) string { return tt.env[k] }
			if got := detectGhostty(getenv); got != tt.want {
				t.Fatalf("detectGhostty() = %v, want %v", got, tt.want)
			}
		})
	}
}
