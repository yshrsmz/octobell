## 1. fetchCmd の条件付き / 強制切り替え

- [x] 1.1 `fetchCmd` に `force bool` 引数を追加し、`force` のとき `lastModified` を空文字で `client.List` に渡す
- [x] 1.2 `tickMsg` ハンドラの `fetchCmd()` 呼び出しを `fetchCmd(false)` にする
- [x] 1.3 `Refresh`（`ctrl+r`）ハンドラの `fetchCmd()` 呼び出しを `fetchCmd(true)` にする
- [x] 1.4 `fetchCmd` の他の呼び出し箇所がないか grep で確認し、引数を網羅する

## 2. テスト

- [x] 2.1 手動更新（force=true）で `If-Modified-Since` を送らない／無条件取得になることの headless テストを追加
- [x] 2.2 自動ポーリング（force=false）では従来どおり `lastModified` を送ることの headless テストを追加
- [x] 2.3 強制取得の 200 応答後に `m.lastModified` が更新され（`handleFetched` は `NotModified` チェック前に無条件更新）、次回 tick が条件付きになることを確認するテストを追加
- [x] 2.4 強制取得で同じ未読セットが返っても `Differ` が空を返し OS 通知が発火しないことの headless テストを追加（304 スキップが無くなっても通知スパムしない不変条件の固定）

## 3. 検証とドキュメント

- [x] 3.1 `go vet ./...` / `go build ./cmd/octobell` / `go test ./...` を実行して通す
- [x] 3.2 `docs/manual-verification.md` に「`ctrl+r` で一覧が最新化される（既読化済み項目が消える）」のチェック項目を追加
