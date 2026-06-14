## Context

`internal/tui/tui.go` の `newModel()` は `list.New(nil, list.NewDefaultDelegate(), 0, 0)` をそのまま使い、空表示・ヘルプのスタイルを上書きしていない。色は bubbles v2.1.0 のデフォルト（`list/style.go`・`help/help.go`）由来で、ダークターミナルでは `lightDark()` がダーク側の値を選ぶ。その値が極端に暗く、視認性を損なっている。

ダーク側のデフォルト前景色（数値が小さいほど暗い）:

| 要素 | スタイル | デフォルト（dark） |
|---|---|---|
| 空表示「No items.」 | `list.Styles.NoItems` | `#626262` |
| ヘルプ key（`enter` 等） | help `keyStyle` | `#626262` |
| ヘルプ desc（`開く+既読` 等） | help `descStyle` | `#4A4A4A` |
| ヘルプ区切り（`•`） | help `sepStyle` | `#3C3C3C` |

開発環境は headless のため色の自動 render 検証はできず、実機確認に頼る（プロジェクトのテスト方針）。

## Goals / Non-Goals

**Goals:**
- ダークターミナルで空表示とヘルプ（short / full）が読める前景色に上書きする。
- 変更を `newModel()` 内のスタイル差し替えに閉じ、ロジック・キーバインド・テスト対象の振る舞いを変えない。

**Non-Goals:**
- ライトターミナルへの最適化（今回はダーク前提。`lightDark()` を完全に再現しない）。
- 色のユーザー設定化（config capability への拡張）。
- ステータス行や list アイテム本体（タイトル・副行）の配色変更。

## Decisions

### 決定 1: `newModel()` で `list.New(...)` 直後にスタイルを上書きする
`list.Model` は `Styles` フィールド（`NoItems` 等）と `Help`（`bubbles/help` の `Model`、`Styles` を持つ）を公開している。`list.New` 後にこれらを代入して上書きする。デリゲートやキーマップには触れない。

- 代替案: フォークやテーマパッケージの導入 → 最小修正の方針に反するため不採用。

### 決定 2: ダーク前提の固定グレー 1 系統に寄せる
`lightDark()` を使わず、ダークで読める明るめグレーを固定値で与える。bubbles の他スタイル（list アイテムの subdued 系 `#9B9B9B` 等）と調和する範囲で、デフォルトより明確に明るい値にする。具体値の目安:

- 空表示「No items.」: `#9B9B9B` 程度（subdued 相当の明るさ）
- ヘルプ key: `#909090` 程度（light 側相当の明るさをダークでも使う）
- ヘルプ desc: `#B2B2B2` 程度（light 側相当）
- ヘルプ区切り: `#626262` 程度（key/desc より一段暗く、区切りとして機能する範囲）

最終値は実装時に実機で見て微調整してよい（spec はあくまで「視認可能であること」を要件とする）。

- 代替案: `lightDark()` でダーク側だけ持ち上げる → ライト対応のコストが増え、今回の Non-Goal。

## Risks / Trade-offs

- [ライトターミナルでの見え方が最適化されない] → 今回はダーク前提と明記。ライト対応が必要になれば別 change で `lightDark()` 化する。
- [bubbles のバージョン更新でデフォルト構造が変わる可能性] → 上書きはフィールド代入のみで、API が変わればコンパイルエラーで検知できる。
- [色の良し悪しが主観的で自動検証できない] → `docs/manual-verification.md` のチェック項目でカバーする。
