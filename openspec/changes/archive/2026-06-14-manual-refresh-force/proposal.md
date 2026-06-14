## Why

ユーザーが `ctrl+r` で手動更新しても一覧が更新されないことがある。手動更新も自動ポーリングと同じく `If-Modified-Since`（条件付きリクエスト）を送るため、サーバが `304 Not Modified` を返すと `m.notifs` を一切更新せず、既読化済みなどで実際には変化した一覧が古いまま残る。手動更新はユーザーが明示的に最新化を求める操作であり、`304` で「何もしない」のは意図に反する。

## What Changes

- `ctrl+r` による手動更新を**強制無条件取得**にする。`If-Modified-Since` を付けず常に最新の一覧を取得し、`m.notifs` を置き換える。
- 自動ポーリング（tick 経由）は従来どおり条件付きリクエストを維持し、レート制限を節約する。
- `fetchCmd` に「条件付き / 強制」を切り替える引数を導入し、Refresh ハンドラと tick で使い分ける。

## Capabilities

### New Capabilities
<!-- なし -->

### Modified Capabilities
- `notification-polling`: 「条件付きリクエストによるレート節約」は自動ポーリングに適用する旨を明確化し、新たに「手動更新は条件付きリクエストを行わず強制取得する」要件を追加する。

## Impact

- `internal/tui/tui.go`: `fetchCmd` に強制フラグを追加。`Refresh`（`ctrl+r`）ハンドラは強制取得、`tickMsg` ハンドラは従来どおり条件付き。強制取得時は `lastModified` を空で渡す。
- `internal/tui/tui_test.go`: 手動更新で条件付きリクエストを行わない（`lastModified` を送らない）こと、tick では送ることの headless テストを追加。
- `internal/github` の `List` 自体は変更不要（`lastModified` 空で従来どおり無条件 GET になる）。
- `docs/manual-verification.md`: 手動更新で一覧が最新化されるチェック項目を追加。
