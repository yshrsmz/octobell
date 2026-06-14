package notify

// Mode は通知バックエンドの選択方針（terminal_notify 設定に対応）。
// 文字列ではなく型付き enum とし、設定文字列との対応付けは呼び出し側（cmd）が 1 箇所で行う。
// これにより notify は config に依存せず、文字列リテラルの定義を二重化しない。
type Mode int

const (
	ModeAuto   Mode = iota // 対応端末（Ghostty）を検出したら OSC 777、未検出なら Beeep
	ModeOSC777             // 検出によらず OSC 777 を強制
	ModeOSC9               // 検出によらず OSC 9 を強制
	ModeOff                // 常に Beeep
)

// backend は選択された通知バックエンド種別。
type backend int

const (
	backendNoop   backend = iota // 通知無効
	backendBeeep                 // gen2brain/beeep（osascript / D-Bus 等）
	backendOSC777                // OSC 777
	backendOSC9                  // OSC 9
)

// selectBackend は設定・端末検出・tty 有無から通知バックエンドを決める純粋関数。
//
// 優先順位:
//  1. notifyEnabled=false なら Noop（最優先）。
//  2. terminalNotify と端末検出で OSC を使うか決める（auto は Ghostty 検出時のみ OSC 777、
//     osc777/osc9 は検出によらず強制、off は常に Beeep）。
//  3. OSC を使う場合でも tty を開けなければ Beeep にフォールバックする。
func selectBackend(notifyEnabled bool, mode Mode, ghostty, ttyAvailable bool) backend {
	if !notifyEnabled {
		return backendNoop
	}
	want := backendBeeep
	switch mode {
	case ModeOff:
		want = backendBeeep
	case ModeOSC777:
		want = backendOSC777
	case ModeOSC9:
		want = backendOSC9
	case ModeAuto:
		if ghostty {
			want = backendOSC777
		}
	}
	// OSC を選んでも controlling terminal を開けなければ Beeep に落とす。
	if (want == backendOSC777 || want == backendOSC9) && !ttyAvailable {
		return backendBeeep
	}
	return want
}

// Select は設定と環境から単一の Notifier を組み立てる。
// Beeep と TerminalOSC は同時に使わない（二重通知を避けるため、必ず 1 つだけ返す）。
//
// notifyEnabled は cfg.Notify（--no-notify で false）、mode は cfg.TerminalNotify を変換した値を渡す。
func Select(notifyEnabled bool, mode Mode, getenv func(string) string) Notifier {
	ghostty := detectGhostty(getenv)

	// tty を開くのは OSC を使う場合だけにしたいので、まず「tty はある」前提で種別を決め、
	// OSC を選んだときに初めて実際の tty を開く（開けなければ Beeep にフォールバック）。
	switch selectBackend(notifyEnabled, mode, ghostty, true) {
	case backendNoop:
		return Noop{}
	case backendOSC777:
		return oscOrBeeep(oscMode777)
	case backendOSC9:
		return oscOrBeeep(oscMode9)
	default:
		return Beeep{}
	}
}

// oscOrBeeep は controlling terminal を開けたら TerminalOSC、開けなければ Beeep を返す。
func oscOrBeeep(om oscMode) Notifier {
	if f, err := openTTY(); err == nil {
		return newTerminalOSC(om, f)
	}
	return Beeep{}
}
