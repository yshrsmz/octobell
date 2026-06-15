## MODIFIED Requirements

### Requirement: 通知一覧の表示
システムは取得した通知を一覧表示し、未読には印（`●`）を付け、各項目にリポジトリ名・種別・理由を副行として表示しなければならない（MUST）。主行には、対象が Issue または PullRequest の場合、未読マークの直後に `#<番号>` を表示しなければならない（MUST）。番号を持たない種別（Commit / Release / Discussion / CheckSuite など）および `subject.url` が空の場合は、番号を付さず従来どおり表示しなければならない（MUST）。空リストでもパニックせず描画しなければならない（MUST）。空リストの案内表示（「No items.」）およびヘルプ（キーバインド一覧）は、ダークターミナルで背景に沈まず視認できる前景色で描画しなければならない（MUST）。

#### Scenario: 一覧と未読マークを表示
- **WHEN** 通知一覧を受け取って描画する
- **THEN** 各項目が主行（未読マーク + タイトル）と副行（リポ名 · 種別 · 理由）で表示される
<!-- test: tui_test.go TestModelFlowHeadless（fetchedMsg 反映後 View） -->

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
システムは `/` で一覧をインクリメンタルに絞り込めなければならない（MUST）。フィルタ対象はリポジトリ名・種別・理由・タイトル、および Issue/PR の番号とする。番号は `42` でも `#42` でも該当通知にマッチしなければならない（MUST）。

#### Scenario: フィルタ入力中はキーを list に委ねる
- **WHEN** フィルタ入力中である
- **THEN** アプリ独自キーは発火せず、入力は list の絞り込みに使われる

#### Scenario: 番号でフィルタできる
- **WHEN** `subject.url` が `.../pulls/42` の通知があり、フィルタに `42` または `#42` を入力する
- **THEN** その通知が絞り込み結果に含まれる
<!-- test: tui_test.go（FilterValue に番号が含まれる） -->
