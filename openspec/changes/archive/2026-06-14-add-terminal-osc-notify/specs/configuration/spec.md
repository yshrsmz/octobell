## MODIFIED Requirements

### Requirement: 既定値
システムは以下の既定値を持たなければならない（MUST）: `poll_seconds=60`, `all=false`, `participating=false`, `per_page=50`, `mark_read_on_open=true`, `notify=true`, `terminal_notify=auto`。

#### Scenario: 既定設定の値
- **WHEN** 既定設定を生成する
- **THEN** 上記の各値を持つ

## ADDED Requirements

### Requirement: 端末ネイティブ通知の設定
システムは設定キー `terminal_notify` を持たなければならない（MUST）。値は `auto | osc777 | osc9 | off` のいずれかであり、通知バックエンドの選択を制御する。既存のキー上書き規則（存在するキーのみ既定値へ上書き、欠損キーは既定値を維持）に従わなければならない（MUST）。

#### Scenario: 既定は auto
- **WHEN** 設定ファイルに `terminal_notify` が無い
- **THEN** `terminal_notify` は既定値 `auto` になる

#### Scenario: 明示値で上書き
- **WHEN** 設定ファイルに `terminal_notify` が `osc777` / `osc9` / `off` のいずれかで記載されている
- **THEN** その値で既定を上書きする

#### Scenario: 不正値は既定にフォールバック
- **WHEN** `terminal_notify` に未知の値が指定されている
- **THEN** 既定 `auto` として扱い、アプリの動作を妨げない
