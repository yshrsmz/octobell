package notify

import "testing"

func TestSelectBackend(t *testing.T) {
	tests := []struct {
		name          string
		notifyEnabled bool
		mode          Mode
		ghostty       bool
		ttyAvailable  bool
		want          backend
	}{
		{"通知無効は Noop が最優先", false, ModeOSC777, true, true, backendNoop},
		{"auto + Ghostty 検出 → OSC 777", true, ModeAuto, true, true, backendOSC777},
		{"auto + 未検出 → Beeep", true, ModeAuto, false, true, backendBeeep},
		{"osc777 明示は検出によらず OSC 777", true, ModeOSC777, false, true, backendOSC777},
		{"osc9 明示は OSC 9", true, ModeOSC9, false, true, backendOSC9},
		{"off は常に Beeep", true, ModeOff, true, true, backendBeeep},
		{"OSC 希望でも tty なしは Beeep", true, ModeOSC777, true, false, backendBeeep},
		{"auto+検出でも tty なしは Beeep", true, ModeAuto, true, false, backendBeeep},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selectBackend(tt.notifyEnabled, tt.mode, tt.ghostty, tt.ttyAvailable)
			if got != tt.want {
				t.Fatalf("selectBackend(%v, %v, ghostty=%v, tty=%v) = %v, want %v",
					tt.notifyEnabled, tt.mode, tt.ghostty, tt.ttyAvailable, got, tt.want)
			}
		})
	}
}

func TestSelectReturnsNoopWhenDisabled(t *testing.T) {
	n := Select(false, ModeAuto, func(string) string { return "" })
	if _, ok := n.(Noop); !ok {
		t.Fatalf("Select(notify=false) = %T, want Noop", n)
	}
}

func TestSelectReturnsBeeepForUndetectedTerminal(t *testing.T) {
	// auto かつ Ghostty 未検出 → Beeep（tty を開きにいかない経路）。
	n := Select(true, ModeAuto, func(string) string { return "" })
	if _, ok := n.(Beeep); !ok {
		t.Fatalf("Select(auto, non-ghostty) = %T, want Beeep", n)
	}
}
