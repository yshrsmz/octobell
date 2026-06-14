## Context

現状、既読化操作は楽観的にローカル状態を更新する（`internal/tui/tui.go`）。

- 単一既読: `setReadLocal(id)` が `m.notifs` 内の該当 ID の `Unread = false` にして `refreshItems()` を呼ぶ。
- 全件既読: `markedAllMsg` ハンドラが全 `m.notifs[i].Unread = false` にして `refreshItems()` を呼ぶ。

`refreshItems()` は `m.notifs` 全件から `list.Item` を再構築するため、既読化した項目は `item.Title()` の `●` マークが消えるだけで一覧に残り続ける。デフォルト設定 `All=false`（未読のみ取得）では、次回ポーリングで初めて API が当該項目を返さなくなり一覧から消える。

「未読管理 TUI」という性質上、ユーザーは既読化した瞬間に項目が一覧から消えることを期待する。サーバ既読化（`PUT /notifications` の 202、`PATCH /notifications/threads/{id}` の 205）は正しく成功しているが、視覚フィードバックが伴わないため「本当に既読化されたのか」という不信を生んでいる。

## Goals / Non-Goals

**Goals:**
- 既読のみ表示モード（`All=false`）で、既読化操作の直後に対象項目を一覧から除去する（単一・全件の双方）。
- 全件表示モード（`All=true`）では従来挙動（既読化しても残す）を維持する。
- 項目除去後も選択位置が破綻しないようにする。
- headless テストで挙動を検証できるようにする。

**Non-Goals:**
- サーバ既読化 API（`internal/github`）の変更。
- 既読化のロールバック（失敗時のローカル状態の復元）。既存方針どおり失敗はステータス通知のみ。
- フィルタ入力中の挙動変更。
- 全件表示モードでの既読項目の見せ方変更（ソート・グルーピング等）。

## Decisions

### 決定1: モデルの真実は `m.notifs`、ビュー除去は `m.notifs` からの削除で表現する

`refreshItems()` が `m.notifs` 全件をそのままビューに反映する構造を維持し、「一覧から消す」＝「`m.notifs` から該当要素を取り除く」とする。これにより一覧（list.Item）とモデル（`m.notifs`）の二重管理を避け、既存の `refreshItems()`／選択インデックス保持ロジックをそのまま活かせる。

- **代替案**: `m.notifs` は全件保持したまま `refreshItems()` で `All=false` のとき未読のみフィルタしてビューに流す。→ モデルとビューの件数がずれ、`differ`（新着判定）や将来のロジックが `m.notifs` を全件前提で扱うと不整合になりうる。`All=false` では「未読だけを扱う」モデルの方が単純なため採用しない。

### 決定2: `All` による分岐を楽観的更新の中心に置く

- 単一既読: `m.cfg.All` が false なら `m.notifs` から該当 ID を除去、true なら従来どおり `Unread = false`。
- 全件既読: `m.cfg.All` が false なら `m.notifs` を空にする、true なら従来どおり全 `Unread = false`。

分岐の共通化のため、`setReadLocal`（単一）と `markedAllMsg`（全件）の双方が参照する小さなヘルパー、または各所での `if m.cfg.All` 分岐とする。実装時に重複が出るなら「既読化した ID 集合を除去 or 未読落とし」する一関数に寄せる。

### 決定3: 選択インデックスの保持は `refreshItems()` 既存ロジックに委ねる

`refreshItems()` は除去後の `len(items)` に対して `idx` をクランプ（`idx >= 0 && idx < len(items)` のときのみ `Select(idx)`）する。除去で件数が減ると、bubbles の list が末尾に丸めるか、`refreshItems` のクランプで妥当な位置に収まる。テストで「残り項目があれば有効な選択」を確認する。実装で末尾除去時にインデックスがあふれるなら `min(idx, len-1)` でクランプを補強する。

## Risks / Trade-offs

- [`All=false` で既読化した項目を二度と見られなくなる] → 仕様どおり（GitHub の通知 UX と同じく、未読管理では既読は一覧から外れる）。必要なら別途 `All=true` 設定で全件表示できる。
- [`m.notifs` を破壊的に縮めることで `differ` 等が影響を受ける] → `differ` は `handleFetched` 内で取得結果から ID 集合を作るため、既読化での `m.notifs` 縮小とは独立。影響なしと判断。次回ポーリングで `m.notifs` は取得結果に置き換わる。
- [選択インデックスのクランプ漏れ] → headless テストで末尾項目除去・全件除去のケースを検証して回避。
