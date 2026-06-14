package notify

import (
	"bytes"
	"testing"
)

func TestTerminalOSC777Sequence(t *testing.T) {
	var buf bytes.Buffer
	n := newTerminalOSC(oscMode777, &buf)
	if err := n.Notify("GitHub: octocat/Hello-World", "PR がマージされました"); err != nil {
		t.Fatalf("Notify: %v", err)
	}
	want := "\x1b]777;notify;GitHub: octocat/Hello-World;PR がマージされました\x1b\\"
	if got := buf.String(); got != want {
		t.Fatalf("OSC 777 sequence\n got=%q\nwant=%q", got, want)
	}
}

func TestTerminalOSC9SequenceJoinsTitleAndBody(t *testing.T) {
	var buf bytes.Buffer
	n := newTerminalOSC(oscMode9, &buf)
	if err := n.Notify("GitHub: octocat/Hello-World", "PR がマージされました"); err != nil {
		t.Fatalf("Notify: %v", err)
	}
	want := "\x1b]9;GitHub: octocat/Hello-World: PR がマージされました\x1b\\"
	if got := buf.String(); got != want {
		t.Fatalf("OSC 9 sequence\n got=%q\nwant=%q", got, want)
	}
}

func TestTerminalOSC777SanitizesControlChars(t *testing.T) {
	var buf bytes.Buffer
	n := newTerminalOSC(oscMode777, &buf)
	// 本文に複数件の改行・件名に ; を含むケース。改行は空白へ、件名の ; は除去される。
	if err := n.Notify("GitHub: 新着通知 2 件", "• A\n• B\n"); err != nil {
		t.Fatalf("Notify: %v", err)
	}
	got := buf.String()
	// 制御文字（改行・ESC）が残っていないこと。
	if bytes.ContainsAny([]byte(got[len("\x1b]777;"):]), "\n\r") {
		t.Fatalf("sanitized body still contains control chars: %q", got)
	}
	want := "\x1b]777;notify;GitHub: 新着通知 2 件;• A • B\x1b\\"
	if got != want {
		t.Fatalf("OSC 777 sanitized\n got=%q\nwant=%q", got, want)
	}
}

func TestTerminalOSC9AvoidsLeadingDigitSemicolon(t *testing.T) {
	var buf bytes.Buffer
	n := newTerminalOSC(oscMode9, &buf)
	// 件名が「数字;」で始まると ConEmu 拡張と衝突するため、先頭に空白を補う。
	if err := n.Notify("4;something", ""); err != nil {
		t.Fatalf("Notify: %v", err)
	}
	want := "\x1b]9; 4;something\x1b\\"
	if got := buf.String(); got != want {
		t.Fatalf("OSC 9 leading-digit guard\n got=%q\nwant=%q", got, want)
	}
}
