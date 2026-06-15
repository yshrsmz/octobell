# notification-tui Specification

## Purpose
TBD - created by archiving change document-existing-behavior. Update Purpose after archive.
## Requirements
### Requirement: 通知一覧の表示
システムは取得した通知を一覧表示し、未読には印（`●`）を付け、各項目にリポジトリ名・種別・理由を副行として表示しなければならない（MUST）。ただし `reason` が `state_change` の Issue または PullRequest で**実状態**（後述のエンリッチ）が取得できている場合は、副行の理由を `state_change(<実状態>)`（例 `state_change(merged)`）の形式で表示しなければならない（MUST）。このとき `(<実状態>)` を括弧ごと GitHub 標準に倣って色分けし、`state_change` 本体は通常色のままとしなければならない（MUST）: Issue は `open`=緑 / `closed-completed`=紫 / `closed-not_planned`=灰、PullRequest は `open`=緑 / `draft`=灰 / `merged`=紫 / `closed`（未マージ）=赤。実状態が未取得・取得失敗・`reason` が `state_change` 以外・対象外種別（Commit / Release / Discussion / CheckSuite など）の場合は、従来どおり `reason` のみを表示しなければならない（MUST）。主行には、対象が Issue または PullRequest の場合、未読マークの直後に `#<番号>` を表示しなければならない（MUST）。番号を持たない種別および `subject.url` が空の場合は、番号を付さず従来どおり表示しなければならない（MUST）。空リストでもパニックせず描画しなければならない（MUST）。空リストの案内表示（「No items.」）およびヘルプ（キーバインド一覧）は、ダークターミナルで背景に沈まず視認できる前景色で描画しなければならない（MUST）。

#### Scenario: 一覧と未読マークを表示
- **WHEN** 通知一覧を受け取って描画する
- **THEN** 各項目が主行（未読マーク + タイトル）と副行（リポ名 · 種別 · 理由）で表示される
<!-- test: tui_test.go TestModelFlowHeadless（fetchedMsg 反映後 View） -->

#### Scenario: state_change の Issue/PR は実状態を付記する
- **WHEN** `reason=state_change` の Issue/PR 通知について実状態のエンリッチが完了している
- **THEN** 副行の理由が `state_change(<実状態>)`（例 PR の `state_change(merged)`、Issue の `state_change(closed-completed)`）で表示され、`(<実状態>)` が括弧ごと状態に応じた前景色で色分けされる
<!-- test: tui_test.go（item.Description が state_change(<状態>) を含む） -->

#### Scenario: 実状態が未取得・他 reason・対象外は理由のみ表示
- **WHEN** 実状態がまだ取得できていない／取得に失敗した／`reason` が `state_change` 以外である／番号を持たない対象外種別である
- **THEN** 副行は従来どおり `reason` のみを表示する
<!-- test: tui_test.go（エンリッチ前後で表示が切り替わる / 他 reason は付記されない） -->

#### Scenario: Issue/PR は主行に番号を表示
- **WHEN** `subject.url` が `.../issues/7` または `.../pulls/42` の通知を描画する
- **THEN** 主行が `<未読マーク> #<番号> <タイトル>`（例 `● #42 …`）で表示される
<!-- test: tui_test.go（item.Title に #N が含まれる） -->

#### Scenario: 番号を持たない種別は番号を付さない
- **WHEN** Commit / Release / Discussion などの番号を持たない種別、または `subject.url` が空の通知を描画する
- **THEN** 主行に `#<番号>` を付さず、従来どおり `<未読マーク> <タイトル>` で表示される
<!-- test: tui_test.go（番号なし種別の item.Title に # が含まれない） -->

#### Scenario: 空リストで描画
- **WHEN** 通知が 0 件の状態で描画する
- **THEN** パニックせずに描画できる
<!-- test: tui_test.go TestModelFlowHeadless（初期 View）/ TestEmptyListActions -->

#### Scenario: 空表示とヘルプが視認可能な前景色で描画される
- **WHEN** TUI を初期化して空表示・ヘルプを描画する
- **THEN** 「No items.」とヘルプの key / desc / 区切りには bubbles デフォルトの低コントラストなグレーではなく、上書きした明るめの前景色が適用される
<!-- 実機 render 検証は docs/manual-verification.md のチェックリストに従う（headless では色を自動検証できない） -->

### Requirement: インクリメンタルフィルタ
システムは `/` で一覧をインクリメンタルに絞り込めなければならない（MUST）。フィルタ対象はリポジトリ名・種別・理由・タイトル、および Issue/PR の番号とする。Issue/PR の実状態が取得できている場合は、その実状態（例 `merged` / `closed`）もフィルタ対象に含めなければならない（MUST）。番号は `42` でも `#42` でも該当通知にマッチしなければならない（MUST）。

#### Scenario: フィルタ入力中はキーを list に委ねる
- **WHEN** フィルタ入力中である
- **THEN** アプリ独自キーは発火せず、入力は list の絞り込みに使われる

#### Scenario: 番号でフィルタできる
- **WHEN** `subject.url` が `.../pulls/42` の通知があり、フィルタに `42` または `#42` を入力する
- **THEN** その通知が絞り込み結果に含まれる
<!-- test: tui_test.go（FilterValue に番号が含まれる） -->

#### Scenario: 実状態でフィルタできる
- **WHEN** 実状態が `merged` の PR 通知があり、フィルタに `merged` を入力する
- **THEN** その通知が絞り込み結果に含まれる
<!-- test: tui_test.go（FilterValue にエンリッチ済み実状態が含まれる） -->

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

