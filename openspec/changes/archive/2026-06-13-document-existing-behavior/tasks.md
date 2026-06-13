<!-- これは既存実装の spec 化（後付け文書化）であり、コード変更を伴わない。
     各タスクは「spec の scenario が実装・既存テストと一致することの確認」を指す。 -->

## 1. spec と実装の突き合わせ

- [x] 1.1 `notification-polling` を `internal/github/client.go`（304 / X-Poll-Interval / クエリ / restPrefix）と突き合わせ
- [x] 1.2 `notification-alerting` を `internal/notify/diff.go` `tui.go notifyCmd`（初回抑制・未読のみ・文面整形）と突き合わせ（test: diff_test.go）
- [x] 1.3 `read-management` を `tui.go`（二度押し確認・楽観的更新）と突き合わせ（test: tui_test.go TestMarkAllReadConfirm / TestModelFlowHeadless）
- [x] 1.4 `notification-tui` を `tui.go` `item.go` `keys.go` と `notifications.go BrowserURL` に突き合わせ（test: tui_test.go / notifications_test.go TestBrowserURL）
- [x] 1.5 `configuration` を `internal/config/config.go` と `cmd/octobell/main.go`（フラグ・モード）と突き合わせ

## 2. 検証

- [x] 2.1 `openspec validate document-existing-behavior --strict` で構造を検証
- [x] 2.2 既存テストが全 scenario の固定済み振る舞いをカバーしているか確認（未カバー scenario は spec 上の注記で明示）
