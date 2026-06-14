package notify

import (
	"io"
	"os"
	"strings"
	"unicode"
)

// oscMode は OSC 通知のシーケンス種別。
type oscMode int

const (
	oscMode777 oscMode = iota // ESC ] 777 ; notify ; title ; body ST（件名・本文を分離）
	oscMode9                  // ESC ] 9 ; message ST（単一文字列）
)

// TerminalOSC は OSC エスケープシーケンスでデスクトップ通知を出す Notifier 実装。
// 対応端末（Ghostty 等）が端末自身で通知を出すため、osascript のような外部プロセスを起動しない。
//
// 注意: OSC は戻り値を持たず、配信成否を呼び出し側に伝えられない（端末側で通知が無効でも黙って出ない）。
// Notify が返すエラーは「シーケンスの書き込み失敗」のみで、「バナーが出たか」は保証しない。
type TerminalOSC struct {
	w    io.Writer
	mode oscMode
}

// newTerminalOSC は出力先 w と mode を指定して TerminalOSC を生成する。
func newTerminalOSC(mode oscMode, w io.Writer) TerminalOSC {
	return TerminalOSC{w: w, mode: mode}
}

const (
	esc = "\x1b"     // ESC
	st  = "\x1b\\"   // ST（String Terminator）= ESC \
)

// Notify は OSC 通知シーケンスを組み立て、シーケンス全体を 1 回の Write で出力する。
// 単一 Write にすることで、alt-screen 描画中でも frame 出力とバイト単位で混ざらない。
func (t TerminalOSC) Notify(title, message string) error {
	var seq string
	switch t.mode {
	case oscMode9:
		// OSC 9 は単一文字列。件名と本文を結合する。
		msg := sanitizeOSC(title)
		if b := sanitizeOSC(message); b != "" {
			msg += ": " + b
		}
		// message を「数字 + ;」で始めない（ConEmu 拡張 / OSC 9;4 進捗表示との衝突回避）。
		msg = avoidLeadingDigitSemicolon(msg)
		seq = esc + "]9;" + msg + st
	default: // oscMode777
		// OSC 777 は ; 区切りのフィールドを持つ。件名に ; が混ざるとフィールドがずれるため除去する
		//（本文は最終フィールドなので ; を残してよい）。
		ttl := strings.ReplaceAll(sanitizeOSC(title), ";", " ")
		body := sanitizeOSC(message)
		seq = esc + "]777;notify;" + ttl + ";" + body + st
	}
	_, err := io.WriteString(t.w, seq)
	return err
}

// sanitizeOSC は OSC シーケンスを壊す制御文字（ESC・BEL・改行など C0 制御と DEL）と空白を、
// 単一の空白へ畳みつつ前後をトリムする。1 パスで処理し中間文字列を作らない。
func sanitizeOSC(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	pendingSpace := false
	for _, r := range s {
		// 制御文字（非空白の C0・DEL 含む）と各種空白はまとめて空白扱いにする。
		if r < 0x20 || r == 0x7f || unicode.IsSpace(r) {
			pendingSpace = b.Len() > 0 // 先頭の空白は出力しない（前トリム）。
			continue
		}
		if pendingSpace {
			b.WriteByte(' ')
			pendingSpace = false
		}
		b.WriteRune(r)
	}
	// pendingSpace を最後に書き出さないことで後トリムも兼ねる。
	return b.String()
}

// avoidLeadingDigitSemicolon は「数字 + ;」で始まる文字列の先頭に空白を付けて衝突を避ける。
func avoidLeadingDigitSemicolon(s string) string {
	if len(s) >= 2 && s[0] >= '0' && s[0] <= '9' && s[1] == ';' {
		return " " + s
	}
	return s
}

// openTTY は controlling terminal（/dev/tty）を書き込み用に開く。
// 開けない場合（パイプ経由・--once 等）は nil とエラーを返し、呼び出し側は Beeep にフォールバックする。
func openTTY() (*os.File, error) {
	return os.OpenFile("/dev/tty", os.O_WRONLY, 0)
}
