## Why

bubbles の list / help コンポーネントのデフォルトスタイルは、ダークターミナルで `lightDark()` がダーク側に極端に暗いグレーを返すため、空表示「No items.」（`#626262`）とヘルプ（キーバインド一覧）の説明文（desc `#4A4A4A`）・キー（`#626262`）・区切り（`#3C3C3C`）が背景に沈んで読みづらい。octobell 側はこれらのスタイルを上書きしておらず、操作のヒントである一覧が視認できないのは実用上の不便。

## What Changes

- ダークターミナル前提で、`list` の空表示と `help`（short / full）の前景色を、bubbles デフォルトの低コントラストなグレーから読みやすい明るめグレーに上書きする。
- 対象スタイルは `list.Styles.NoItems` と help の key / desc / separator。`internal/tui/tui.go` の `newModel()` で `list.New(...)` 直後に差し替える最小修正。
- `keys.go` のキーバインド説明文言は変更しない（色のみ）。
- headless 環境では render 検証ができないため、`docs/manual-verification.md` に short help / `?` の full help の視認性チェック項目を追加する。

## Capabilities

### New Capabilities
（なし）

### Modified Capabilities
- `notification-tui`: 通知一覧の表示要件に、空表示とヘルプ（キーバインド一覧）をダークターミナルで視認可能な前景色で描画する要件を追加する。

## Impact

- 実装: `internal/tui/tui.go`（`newModel()` でのスタイル上書き）。
- ドキュメント: `docs/manual-verification.md`（視認性チェック項目追加）。
- 依存・API への影響なし。ロジック・キーバインド・テスト対象の振る舞いは不変（描画スタイルのみ）。
