# notification-alerting Specification

## Purpose
TBD - created by archiving change document-existing-behavior. Update Purpose after archive.
## Requirements
### Requirement: 新着のみの差分判定
システムは前回ポーリングで見たスレッド ID を記憶し、新規かつ未読のスレッドのみを通知対象としなければならない（MUST）。初回ポーリング（記憶が空）では既存バックログを通知対象としてはならない（MUST NOT）。

#### Scenario: 初回はバックログを通知しない
- **WHEN** 起動後の初回ポーリングで通知一覧を受け取る
- **THEN** 新着は 0 件として扱い、OS 通知を送らない
<!-- test: diff_test.go TestDiffer（初回は新着なし） -->

#### Scenario: 2 回目以降は新規 ID のみ
- **WHEN** 既知 ID に加えて新規 ID が現れる
- **THEN** 新規 ID のみを新着として返し、記憶を更新する
<!-- test: diff_test.go TestDiffer（新規 c のみ・再度 c は既知） -->

#### Scenario: 未読のみが対象
- **WHEN** 取得結果に既読スレッドが含まれる
- **THEN** 既読スレッドは新着判定の対象に含めない

### Requirement: OS デスクトップ通知
システムは新着を OS のデスクトップ通知として送らなければならない（MUST）。通知は fire-and-forget でありクリックアクションには対応しない。送信失敗はアプリの動作を妨げてはならない（MUST NOT）。

#### Scenario: 新着があれば通知を送る
- **WHEN** 新着が 1 件以上ある
- **THEN** デスクトップ通知を送信する（macOS は osascript / terminal-notifier、Linux は D-Bus / notify-send）

### Requirement: 通知文面の整形
システムは新着件数に応じて通知の件名・本文を整形しなければならない（MUST）。

#### Scenario: 1 件は repo とタイトル
- **WHEN** 新着が 1 件
- **THEN** 件名は `GitHub: <リポジトリ full_name>`、本文は subject のタイトルになる

#### Scenario: 複数件は件数と先頭 3 件
- **WHEN** 新着が複数件
- **THEN** 件名は `GitHub: 新着通知 <N> 件`、本文は先頭 3 件のタイトルを列挙し、超過分は「…ほか <M> 件」と表示する

### Requirement: 通知の無効化
システムは通知が無効なとき、デスクトップ通知を一切送ってはならない（MUST NOT）。

#### Scenario: notify=false は何も送らない
- **WHEN** 設定 `notify` が false、または `--no-notify` 指定
- **THEN** Noop 実装が用いられ、通知は送られない

