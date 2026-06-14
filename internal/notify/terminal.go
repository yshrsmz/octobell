package notify

// detectGhostty は環境変数から Ghostty 端末を判定する。
// Ghostty は desktop-notifications が既定 ON で、OSC 9 / OSC 777 を解釈してデスクトップ通知を出す。
//
// 検出は `TERM=xterm-ghostty`（Ghostty 既定の terminfo）または `TERM_PROGRAM=ghostty` を見る。
// SSH 越しでは TERM_PROGRAM がローカル変数のため失われ、TERM もリモートに ghostty terminfo が
// 無いと潰れる。検出できないときは安全側に倒して false を返す（呼び出し側は Beeep へフォールバックする）。
//
// getenv は os.Getenv を差し替え可能にしてテストするための引数。
func detectGhostty(getenv func(string) string) bool {
	return getenv("TERM") == "xterm-ghostty" || getenv("TERM_PROGRAM") == "ghostty"
}
