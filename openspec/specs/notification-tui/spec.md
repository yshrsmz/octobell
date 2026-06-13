# notification-tui Specification

## Purpose
TBD - created by archiving change document-existing-behavior. Update Purpose after archive.
## Requirements
### Requirement: 通知一覧の表示
システムは取得した通知を一覧表示し、未読には印（`●`）を付け、各項目にリポジトリ名・種別・理由を副行として表示しなければならない（MUST）。空リストでもパニックせず描画しなければならない（MUST）。

#### Scenario: 一覧と未読マークを表示
- **WHEN** 通知一覧を受け取って描画する
- **THEN** 各項目が主行（未読マーク + タイトル）と副行（リポ名 · 種別 · 理由）で表示される
<!-- test: tui_test.go TestModelFlowHeadless（fetchedMsg 反映後 View） -->

#### Scenario: 空リストで描画
- **WHEN** 通知が 0 件の状態で描画する
- **THEN** パニックせずに描画できる
<!-- test: tui_test.go TestModelFlowHeadless（初期 View）/ TestEmptyListActions -->

### Requirement: インクリメンタルフィルタ
システムは `/` で一覧をインクリメンタルに絞り込めなければならない（MUST）。フィルタ対象はリポジトリ名・種別・理由・タイトルとする。

#### Scenario: フィルタ入力中はキーを list に委ねる
- **WHEN** フィルタ入力中である
- **THEN** アプリ独自キーは発火せず、入力は list の絞り込みに使われる

### Requirement: ブラウザ起動と URL 変換
システムは選択中の通知を既定ブラウザで開けなければならない（MUST）。API URL（`subject.url`）を `repository.html_url` を基点とした閲覧用 URL に変換し、Enterprise host にも追従する。`enter` は開いた上で（設定が有効かつ未読なら）既読化し、`o` は開くのみとする。

#### Scenario: PullRequest は単数形 /pull/
- **WHEN** subject が `.../pulls/42` の PullRequest
- **THEN** `…/pull/42` を開く
<!-- test: notifications_test.go TestBrowserURL（PR は単数形 pull） -->

#### Scenario: Issue は /issues/
- **WHEN** subject が `.../issues/7` の Issue
- **THEN** `…/issues/7` を開く
<!-- test: notifications_test.go TestBrowserURL（Issue は issues のまま） -->

#### Scenario: Commit は /commit/
- **WHEN** subject が `.../commits/<sha>` の Commit
- **THEN** `…/commit/<sha>` を開く
<!-- test: notifications_test.go TestBrowserURL（Commit は commit に変換） -->

#### Scenario: 不明種別・空はリポにフォールバック
- **WHEN** 解決できない種別（Release / Discussion 等）、または `subject.url` が空
- **THEN** `repository.html_url` を開く
<!-- test: notifications_test.go TestBrowserURL（フォールバック 2 ケース） -->

#### Scenario: enter は開く + 既読
- **WHEN** 未読通知で `enter` を押し、`mark_read_on_open` が有効
- **THEN** ブラウザで開き、楽観的にローカル既読化する

#### Scenario: o は開くのみ
- **WHEN** 通知で `o` を押す
- **THEN** ブラウザで開き、既読化しない
<!-- test: tui_test.go TestModelFlowHeadless（openSelected(false)） -->

### Requirement: 定期ポーリングと手動更新
システムは設定された間隔で自動再取得しなければならない（MUST）。取得中（loading）の tick では多重取得を避けて再取得を行わない。`ctrl+r` で手動更新できる。

#### Scenario: 取得中の tick は再取得しない
- **WHEN** 取得処理中に tick が発生する
- **THEN** 新たな取得を開始しない（多重取得を避ける）

#### Scenario: 手動更新
- **WHEN** 取得中でないときに `ctrl+r` を押す
- **THEN** ただちに再取得を開始する

### Requirement: ステータス行
システムは画面最下部にステータス行を表示しなければならない（MUST）。取得中はスピナー、通常時は最終更新時刻と現在の間隔、操作直後は一時メッセージ（約 4 秒で自動消去）を表示する。

#### Scenario: 通常時は最終更新と間隔
- **WHEN** 取得中でなく一時メッセージもない
- **THEN** 最終更新時刻と現在のポーリング間隔を表示する

### Requirement: 終了
システムは `q` または `ctrl+c` で終了できなければならない（MUST）。

#### Scenario: q で終了
- **WHEN** フィルタ入力中でないときに `q` を押す
- **THEN** アプリが終了する
<!-- test: tui_test.go TestQuitKey（q で QuitMsg） -->

