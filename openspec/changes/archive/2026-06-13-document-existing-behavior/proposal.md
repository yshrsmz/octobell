## Why

octobell は既に動作する TUI アプリだが、振る舞いを定義した spec が存在しない。今後の変更が「何を壊してはいけないか」の基準を持たないため、リグレッションを検知できない。既存実装の振る舞い（特に 304 処理・新着抑制・URL 変換・二度押し確認といった壊れやすいエッジケース）を spec として後付けで記録し、以降の変更が delta する土台を作る。

## What Changes

- 既存実装の振る舞いを 5 つの capability spec として記録する（新規の振る舞い追加・変更は行わない。**コードは変更しない**）。
- 各 scenario は可能な限り既存テスト（`diff_test.go` / `notifications_test.go` / `tui_test.go`）が固定している振る舞いに対応させ、spec をリグレッション基準として機能させる。

## Capabilities

### New Capabilities
- `notification-polling`: gh 認証の再利用、条件付きリクエスト（If-Modified-Since / 304）、X-Poll-Interval を下限とする間隔強制、クエリオプション、GitHub Enterprise host 追従、取得エラー。
- `notification-alerting`: 新着差分判定（初回バックログ抑制・未読のみ）、OS デスクトップ通知、1 件/N 件の文面整形、通知無効化。
- `read-management`: 単一スレッド既読化、全件既読化と二度押し確認、楽観的ローカル更新。
- `notification-tui`: 通知一覧表示、インクリメンタルフィルタ、ブラウザ起動と URL 変換、定期ポーリングループと手動更新、ステータス行、終了。
- `configuration`: XDG 準拠 config.json の読み込み（部分上書き）、既定値、CLI フラグ上書き、実行モード（--once / --version）。

### Modified Capabilities
<!-- 既存 spec は無し（初回 spec 化）。 -->

## Impact

- 影響コード: なし（ドキュメント化のみ。`internal/github`, `internal/notify`, `internal/tui`, `internal/config`, `cmd/octobell` の現状を spec に写す）。
- archive 後に `openspec/specs/` 配下へ 5 つの spec が新設される。
