// Package notify は OS デスクトップ通知の送信を抽象化する。
// 実機・headless 環境差を吸収するため interface で分離している。
package notify

// Notifier はデスクトップ通知の送信を抽象化する。
type Notifier interface {
	Notify(title, message string) error
}

// Noop は通知を送らない実装（OS 通知無効時 / headless 環境向け）。
type Noop struct{}

// Notify は何もしない。
func (Noop) Notify(string, string) error { return nil }
