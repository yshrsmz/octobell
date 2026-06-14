## 1. スタイル上書きの実装

- [x] 1.1 `internal/tui/tui.go` の `newModel()` で `list.New(...)` 直後に `list.Styles.NoItems` の前景色を明るめグレーに上書きする
- [x] 1.2 同じく `list.Help`（`bubbles/help`）の key / desc / separator の前景色を明るめグレーに上書きする（desc が最も暗いので確実に持ち上げる）
- [x] 1.3 `keys.go` のキーバインド説明文言は変更しない（色のみの変更であることを確認する）

## 2. 検証

- [x] 2.1 `go build ./cmd/octobell` と `go test ./...` が通ることを確認する（振る舞い・テスト対象は不変）
- [x] 2.2 `go vet ./...` が通ることを確認する
- [x] 2.3 `docs/manual-verification.md` に「ダークターミナルで空表示『No items.』と short help / `?` の full help が視認できる」チェック項目を追加する
