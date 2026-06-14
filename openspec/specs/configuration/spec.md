# configuration Specification

## Purpose
TBD - created by archiving change document-existing-behavior. Update Purpose after archive.
## Requirements
### Requirement: XDG 準拠の設定読み込み
システムは設定を `XDG_CONFIG_HOME`（未設定なら `~/.config`）配下の `octobell/config.json` から読み込まなければならない（MUST）。ファイルが存在しない場合は既定値で動作する。存在するキーのみを既定値に上書きし、欠損キーは既定値を維持しなければならない（MUST）。

#### Scenario: ファイルなしは既定値
- **WHEN** 設定ファイルが存在しない
- **THEN** 既定設定で動作する（エラーにしない）

#### Scenario: 一部キーのみの部分上書き
- **WHEN** 設定ファイルに一部のキーだけが存在する
- **THEN** そのキーのみ上書きし、未記載のキーは既定値を維持する

#### Scenario: XDG_CONFIG_HOME を優先
- **WHEN** `XDG_CONFIG_HOME` が設定されている
- **THEN** そのディレクトリ配下の `octobell/config.json` を参照する

### Requirement: 既定値
システムは以下の既定値を持たなければならない（MUST）: `poll_seconds=60`, `all=false`, `participating=false`, `per_page=50`, `mark_read_on_open=true`, `notify=true`, `terminal_notify=auto`。

#### Scenario: 既定設定の値
- **WHEN** 既定設定を生成する
- **THEN** 上記の各値を持つ

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

### Requirement: CLI フラグによる上書き
システムは CLI フラグで設定を上書きできなければならない（MUST）。`--all` は取得対象を全件にし、`--no-notify` は通知を無効化する。

#### Scenario: --all で全件取得
- **WHEN** `--all` を指定して起動する
- **THEN** 設定の `All` が true に上書きされる

#### Scenario: --no-notify で通知無効
- **WHEN** `--no-notify` を指定して起動する
- **THEN** 設定の `Notify` が false に上書きされる

### Requirement: 実行モード
システムは TUI 起動を既定とし、`--once` で TUI を起動せず通知を一度だけ取得して一覧表示してから終了し、`--version` でバージョンを表示して終了しなければならない（MUST）。

#### Scenario: --once は一覧表示して終了
- **WHEN** `--once` を指定して起動する
- **THEN** 通知を一度取得して標準出力に一覧表示し、TUI を起動せずに終了する

#### Scenario: --version はバージョン表示
- **WHEN** `--version` を指定して起動する
- **THEN** バージョンを表示して終了する

#### Scenario: 既定は TUI 起動
- **WHEN** モード指定なしで起動する
- **THEN** 定期ポーリングと OS 通知を伴う TUI を起動する

