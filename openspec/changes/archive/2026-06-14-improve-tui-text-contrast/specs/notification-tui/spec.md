## MODIFIED Requirements

### Requirement: 通知一覧の表示
システムは取得した通知を一覧表示し、未読には印（`●`）を付け、各項目にリポジトリ名・種別・理由を副行として表示しなければならない（MUST）。空リストでもパニックせず描画しなければならない（MUST）。空リストの案内表示（「No items.」）およびヘルプ（キーバインド一覧）は、ダークターミナルで背景に沈まず視認できる前景色で描画しなければならない（MUST）。

#### Scenario: 一覧と未読マークを表示
- **WHEN** 通知一覧を受け取って描画する
- **THEN** 各項目が主行（未読マーク + タイトル）と副行（リポ名 · 種別 · 理由）で表示される
<!-- test: tui_test.go TestModelFlowHeadless（fetchedMsg 反映後 View） -->

#### Scenario: 空リストで描画
- **WHEN** 通知が 0 件の状態で描画する
- **THEN** パニックせずに描画できる
<!-- test: tui_test.go TestModelFlowHeadless（初期 View）/ TestEmptyListActions -->

#### Scenario: 空表示とヘルプが視認可能な前景色で描画される
- **WHEN** TUI を初期化して空表示・ヘルプを描画する
- **THEN** 「No items.」とヘルプの key / desc / 区切りには bubbles デフォルトの低コントラストなグレーではなく、上書きした明るめの前景色が適用される
<!-- 実機 render 検証は docs/manual-verification.md のチェックリストに従う（headless では色を自動検証できない） -->
