## MODIFIED Requirements

### Requirement: 既定値
システムは以下の既定値を持たなければならない（MUST）: `poll_seconds=60`, `all=false`, `participating=false`, `per_page=50`, `mark_read_on_open=true`, `notify=true`, `terminal_notify=auto`, `enrich_state=true`。

#### Scenario: 既定設定の値
- **WHEN** 既定設定を生成する
- **THEN** 上記の各値を持つ

## ADDED Requirements

### Requirement: 実状態エンリッチの設定
システムは設定キー `enrich_state`（真偽値、既定 `true`）を持たなければならない（MUST）。`true` のとき Issue/PR の subject 詳細を追加取得して実状態を表示し、`false` のとき追加取得を一切行わず副行は従来どおり `reason` を表示する。既存のキー上書き規則（存在するキーのみ既定値へ上書き、欠損キーは既定値を維持）に従わなければならない（MUST）。

#### Scenario: 既定は有効
- **WHEN** 設定ファイルに `enrich_state` が無い
- **THEN** `enrich_state` は既定値 `true` になる

#### Scenario: 明示的に無効化
- **WHEN** 設定ファイルに `enrich_state` が `false` で記載されている
- **THEN** その値で既定を上書きし、subject 詳細の追加取得を行わない

#### Scenario: 部分設定で他キーを維持
- **WHEN** 設定ファイルに `enrich_state` のみが記載されている
- **THEN** `enrich_state` のみ上書きし、未記載の他キーは既定値を維持する
