## Why

通知一覧の副行には GitHub Notifications API の `reason` がそのまま出ているだけで、`state_change` の場合「状態が変わった」事実しか分からない。閉じられたのか・マージされたのか・再オープンされたのかを区別できず、ユーザーは結局ブラウザを開いて確認することになる。Notifications API は subject の状態を返さないため、subject の詳細（Issue/PR）を追加取得して実状態を導出すれば、この区別を一覧上で完結できる。

## What Changes

- `reason` が `state_change` の Issue / PullRequest 通知について、subject 詳細を追加取得し**実状態**を導出する。
  - Issue: `open` / `closed-completed` / `closed-not_planned`
  - PullRequest: `open` / `draft` / `merged` / `closed`（未マージ）
- 副行で、`reason=state_change` かつ実状態を取得できている場合は `state_change(merged)` のように **reason の後ろに `(実状態)` を付記**して表示する。`(実状態)` を括弧ごと GitHub 標準カラーで色分けし、`reason` 本体は通常色のまま残す。取得未完了・取得失敗・`state_change` 以外の reason・対象外種別（Commit / Release / Discussion / CheckSuite 等）は従来どおり `reason` のみを表示する。
- 実状態の取得対象は `reason=state_change` の Issue/PR に限定する（取得＝表示が一致し、API 消費を最小化する）。
- 実状態の取得は一覧描画を**ブロックしない**非同期処理とする。結果が届いた項目から順に副行を更新する。
- 取得結果は `(通知ID, updated_at)` をキーにキャッシュし、`updated_at` が変わらない通知は再取得しない。並行取得には上限を設け、エラーは当該項目のみ `reason` のみ表示にフォールバックする。
- エンリッチ機能の ON/OFF を設定で切り替え可能にする（既定 ON、部分設定可）。OFF 時は追加取得を一切行わず従来表示に戻る。

## Capabilities

### New Capabilities

（なし）

### Modified Capabilities

- `notification-tui`: 副行の表示内容を変更する。`reason=state_change` の Issue/PR で実状態が取得できている場合は `state_change(実状態)` と付記し、`(実状態)` を括弧ごと色分けする（取得前・取得不能・他 reason・対象外種別は従来どおり `reason` のみ）要件を追加する。
- `notification-polling`: 一覧取得後に `reason=state_change` の Issue/PR の subject 詳細を非同期で追加取得し実状態を導出する要件を追加する。`(通知ID, updated_at)` キャッシュ、並行取得上限、エラー時の局所フォールバックを含む。X-Poll-Interval 尊重・304 はレート制限を消費しない既存不変条件は維持する。
- `configuration`: エンリッチ機能の ON/OFF トグルを追加する（既定 ON、部分設定可）。

## Impact

- `internal/github`: subject 詳細（Issue/PR）の取得メソッドと、実状態を表す型・導出ロジックを追加。既存のクライアント使い分け方針（GET は素の `*http.Client`、2xx 完結の書き込みは `RESTClient`）を踏襲。
- `internal/tui`: `handleFetched` 後にエンリッチ Cmd を発行する非同期フロー、エンリッチ結果メッセージのハンドリング、`(通知ID, updated_at)` キャッシュ、副行描画（`item.Description`）の状態置換・色分けを追加。
- `internal/config`: エンリッチ ON/OFF フィールドを追加（既定値 ON）。
- `docs/manual-verification.md`: 実状態表示・色分け・OFF 設定の手動検証項目を追加。
- レート制限: 追加 GET が Core API を消費するため、キャッシュ・並行上限・OFF トグルで圧迫を抑える。
