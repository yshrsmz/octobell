## 1. 楽観的更新の挙動を `All` で分岐させる

- [x] 1.1 `setReadLocal`（単一既読）を、`m.cfg.All` が false なら `m.notifs` から該当 ID を除去、true なら従来どおり `Unread = false` にするよう変更する
- [x] 1.2 `markedAllMsg` ハンドラ（全件既読成功）を、`m.cfg.All` が false なら `m.notifs` を空にする、true なら従来どおり全 `Unread = false` にするよう変更する
- [x] 1.3 単一・全件で重複が出るなら「ID 集合を除去 or 未読落とし」する共通ヘルパーに寄せる
- [x] 1.4 `refreshItems()` の選択インデックスのクランプを確認し、末尾除去であふれる場合は `min(idx, len-1)` 相当で補強する

## 2. テスト

- [x] 2.1 `All=false` で単一既読すると `m.notifs` から該当項目が消えることの headless テストを追加
- [x] 2.2 `All=false` で全件既読すると `m.notifs` が空になることの headless テストを追加
- [x] 2.3 `All=true` で既読化しても項目が残り `Unread=false` になることの headless テストを追加
- [x] 2.4 末尾項目を除去しても選択インデックスが破綻しないことの headless テストを追加
- [x] 2.5 既存テスト（`TestModelFlowHeadless` など）が新挙動と整合するよう更新する

## 3. 検証とドキュメント

- [x] 3.1 `go vet ./...` / `go build ./cmd/octobell` / `go test ./...` を実行して通す
- [x] 3.2 `docs/manual-verification.md` に「既読化後に一覧から項目が消える（`All=false`）／残る（`All=true`）」のチェック項目を追加
