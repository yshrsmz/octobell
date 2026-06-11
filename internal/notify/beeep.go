package notify

import "github.com/gen2brain/beeep"

// Beeep は gen2brain/beeep を用いたクロスプラットフォーム通知実装。
// macOS は osascript/terminal-notifier、Linux は D-Bus/notify-send を内部利用する。
type Beeep struct{}

// Notify はデスクトップ通知を送る（fire-and-forget。クリックアクションは非対応）。
func (Beeep) Notify(title, message string) error {
	return beeep.Notify(title, message, "")
}
