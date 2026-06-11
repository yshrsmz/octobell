# 手動検証手順

octobell は開発環境（headless Linux・TTY 無し）の制約上、自動で検証できる範囲が限られる。
本書は「自動検証済みの範囲」と「実機（macOS / Linux デスクトップ）で手動検証すべき範囲」を分けて示す。

## 自動検証済み（CI / 開発コンテナ）

これらは `go test ./...` および開発コンテナでの実 GitHub 検証で確認済み:

- ✅ ビルド・`go vet`（CI で実行）
- ✅ `BrowserURL` の API URL → html_url 変換（PR の `/pulls/N`→`/pull/N` 等、単体テスト）
- ✅ 新着差分判定 `Differ`（初回バックログ抑制・新規のみ抽出、単体テスト）
- ✅ **TUI Model ロジック**（`Update`/`handleFetched`/`refreshItems`/`View`、既読の楽観的更新・全既読・quit キー・空リスト安全性）を headless 単体テストで実行・検証（`internal/tui/tui_test.go`）。render/入力/OS 通知/macOS のみ実機手動検証に残る
- ✅ **実 GitHub に対する取得と条件付きリクエスト**: 初回 `200`（通知取得 + `Last-Modified`）→ 2 回目 `If-Modified-Since` で `304 Not Modified`。`X-Poll-Interval` ヘッダ取得も確認
  （`OCTOBELL_LIVE=1 go test ./internal/github/ -run TestLiveConditionalRequest -v` で再現可能。認証要・CI ではスキップ）

## 実機で手動検証すべき範囲

開発コンテナでは TTY・OS 通知・macOS が無いため未検証。各自のマシンで以下を確認する。

### 1. ビルドと起動（macOS / Linux 共通）

```sh
gh auth status                 # gh ログイン済みを確認
go build -o octobell ./cmd/octobell
./octobell --once              # 非 TUI で通知一覧が出ることを確認（スパインの再確認）
./octobell                     # TUI 起動
```

確認:
- [ ] 通知一覧が表示される（未読は `●` マーク付き）
- [ ] `j`/`k` でカーソル移動できる
- [ ] `/` でフィルタが効く（リポ名・タイトル等で絞り込み）
- [ ] `?` でヘルプ（octobell 独自キー含む）が開閉する
- [ ] 最下部のステータス行に「最終更新 / 間隔」が出る
- [ ] `q` で終了する

### 2. ブラウザ起動と既読化

> ⚠️ 既読化は実際の GitHub 通知の状態を変更する。テスト用の不要な通知で試すこと。

- [ ] `o` で選択中の Issue/PR がブラウザで開く（既読化されない）
- [ ] `enter` で開く + 既読化される（`●` が消える。`mark_read_on_open: true` の場合）
- [ ] `r` で選択中が既読化される
- [ ] `ctrl+a` で確認メッセージが出て、もう一度 `ctrl+a` を押すとすべて既読化される（別キーでキャンセルされる）
- [ ] 既読化直後に未読がゼロ件に戻って再表示される不具合（gitify #1437）が起きない

### 3. ポーリングと OS 通知

- [ ] 起動後、`poll_seconds`（既定 60 秒、`X-Poll-Interval` 下限）ごとに自動更新される
- [ ] 新しい通知が来たとき OS のデスクトップ通知が出る
- [ ] 起動直後に既存通知全件の通知が一斉に出ない（差分のみ）
- [ ] 同じ未読が毎回再通知されない

#### OS 別の通知バックエンド確認

- **macOS**: 通知センターに表示される（`osascript` ベース。`terminal-notifier` 導入時はそちら）
- **Linux（デスクトップ）**: D-Bus または `notify-send`（`libnotify` 要）。
  headless/SSH 環境では通知が出ない場合がある。その場合は `--no-notify` で TUI のみ利用するか、`notify: false` を設定

### 4. 設定ファイル

- [ ] `~/.config/octobell/config.json` を作成し、`poll_seconds` / `participating` / `all` 等が反映される
- [ ] 設定ファイルが無くても既定値で動作する

## 既知の制約・要確認事項

- 通知のクリックでアクション（クリックして Issue を開く）は OS 通知ライブラリ（beeep）が非対応。TUI 側のキー操作で開く
- `per_page` の最大値は GitHub のドキュメント上 50。100 を受け付けるかは要確認
- fine-grained PAT での通知 API 利用可否は未確認（classic スコープ `repo` / `notifications` を推奨）
