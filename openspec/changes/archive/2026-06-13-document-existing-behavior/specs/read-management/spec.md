## ADDED Requirements

### Requirement: 単一スレッドの既読化
システムは選択中の未読スレッドを既読にできなければならない（MUST）。既読化は `PATCH /notifications/threads/{id}` で行う。

#### Scenario: 未読スレッドを既読化
- **WHEN** 未読スレッドを選択して既読操作（`r` / `.`）を行う
- **THEN** 対象スレッドの既読化リクエストを送る

#### Scenario: 既読スレッドは対象外
- **WHEN** 既読スレッドに対して既読操作を行う
- **THEN** リクエストを送らない（何もしない）

### Requirement: 全件既読化と二度押し確認
システムは全通知を既読にできなければならない（MUST）。取り消し不能な一括操作のため、`ctrl+a` の一度押しでは実行せず確認状態に入り、再度の `ctrl+a` で初めて `PUT /notifications` を実行しなければならない（MUST）。確認状態で別キーが押されたら実行をキャンセルしなければならない（MUST）。

#### Scenario: 一度押しは確認のみ
- **WHEN** 確認状態でないときに `ctrl+a` を押す
- **THEN** 確認状態に入り、全件既読は実行しない
<!-- test: tui_test.go TestMarkAllReadConfirm（1 回目は確認待ち・cmd nil） -->

#### Scenario: 二度押しで実行
- **WHEN** 確認状態で再度 `ctrl+a` を押す
- **THEN** 確認状態を解除し、全件既読を実行する
<!-- test: tui_test.go TestMarkAllReadConfirm（2 回目で全既読 cmd） -->

#### Scenario: 別キーでキャンセル
- **WHEN** 確認状態で `ctrl+a` 以外のキーを押す
- **THEN** 確認状態を解除し、全件既読を実行しない
<!-- test: tui_test.go TestMarkAllReadConfirm（別キーでキャンセル） -->

### Requirement: 楽観的ローカル更新
システムは既読操作時、サーバ応答を待たずに対象スレッドの未読状態をローカルで既読に反映しなければならない（MUST）。全件既読の成功時はローカルの全スレッドを既読に反映する。

#### Scenario: 既読操作で即座に未読が落ちる
- **WHEN** 未読スレッドを既読化する
- **THEN** サーバ応答前にローカルの該当スレッドが既読表示になる
<!-- test: tui_test.go TestModelFlowHeadless（markSelectedRead 後 Unread=false） -->

#### Scenario: 全件既読の成功で全件が既読
- **WHEN** 全件既読が成功する
- **THEN** ローカルの全スレッドが既読表示になる
<!-- test: tui_test.go TestModelFlowHeadless（markedAllMsg 後 全 Unread=false） -->
