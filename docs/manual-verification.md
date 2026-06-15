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
- [ ] **Issue/PR の主行に番号が出る**: 主行が `● #42 タイトル` のように未読マーク直後に `#<番号>` 付きで表示される。Commit / Release / Discussion など番号を持たない種別では番号が付かず従来どおり表示される
- [ ] `j`/`k` でカーソル移動できる
- [ ] `/` でフィルタが効く（リポ名・タイトル等で絞り込み。`42` や `#42` でも該当 Issue/PR が絞り込める）
- [ ] `?` でヘルプ（octobell 独自キー含む）が開閉する
- [ ] **ダークターミナルで文字色が沈まない**: 通知 0 件時の「No items.」、および short help（最下部のキー一覧）と `?` の full help の key / desc / 区切りが、背景に埋もれず読める明るさで表示される（bubbles デフォルトの薄いグレーに戻っていない）
- [ ] **state_change の実状態付記**: `reason=state_change` の Issue/PR の副行が `state_change(merged)` のように `(実状態)` 付きで表示される。括弧内の状態のみ色が付く（`state_change` 本体は通常色）。merged・closed-completed=紫 / open=緑 / draft・closed-not_planned=灰 / closed(PR 未マージ)=赤。取得が届くまで（一瞬）は `state_change` のみ表示で、一覧描画はブロックされない
- [ ] **他 reason・対象外種別は従来表示**: `mention` / `review_requested` など `state_change` 以外の通知、および Commit / Release / Discussion 等は副行に `(実状態)` が付かず reason のみが出る
- [ ] **実状態でフィルタできる**: `/` フィルタに `merged` 等を入力すると、その実状態の Issue/PR が絞り込まれる
- [ ] **`enrich_state: false` で無効化**: 設定で `enrich_state` を `false` にすると `(実状態)` の付記が消え、追加取得も行われない（従来どおり reason のみ表示）
- [ ] 最下部のステータス行に「最終更新 / 間隔」が出る
- [ ] `ctrl+r` で手動更新すると一覧が最新化される（別端末で既読化した項目が消える等、`304 Not Modified` でスキップされず必ず反映される）
- [ ] `q` で終了する

### 2. ブラウザ起動と既読化

> ⚠️ 既読化は実際の GitHub 通知の状態を変更する。テスト用の不要な通知で試すこと。

- [ ] `o` で選択中の Issue/PR がブラウザで開く（既読化されない）
- [ ] `enter` で開く + 既読化される（`●` が消える。`mark_read_on_open: true` の場合）
- [ ] `r` で選択中が既読化される
- [ ] `ctrl+a` で確認メッセージが出て、もう一度 `ctrl+a` を押すとすべて既読化される（別キーでキャンセルされる）
- [ ] 既読化直後に未読がゼロ件に戻って再表示される不具合（gitify #1437）が起きない
- [ ] **既読のみ表示（`all: false`）**: `r` で既読化した項目が一覧から即座に消える。`ctrl+a` の全既読で一覧が空になる
- [ ] **全件表示（`all: true`）**: `r` / `ctrl+a` で既読化しても項目は一覧に残り、`●`（未読マーク）だけが消える
- [ ] `all: false` で末尾の項目を既読化しても選択カーソルが破綻せず、残りの項目に移動する

### 3. ポーリングと OS 通知

- [ ] 起動後、`poll_seconds`（既定 60 秒、`X-Poll-Interval` 下限）ごとに自動更新される
- [ ] 新しい通知が来たとき OS のデスクトップ通知が出る
- [ ] 起動直後に既存通知全件の通知が一斉に出ない（差分のみ）
- [ ] 同じ未読が毎回再通知されない

#### OS 別の通知バックエンド確認

- **macOS**: 通知センターに表示される（`osascript` ベース。`terminal-notifier` 導入時はそちら）
- **Linux（デスクトップ）**: D-Bus または `notify-send`（`libnotify` 要）。
  headless/SSH 環境では通知が出ない場合がある。その場合は `--no-notify` で TUI のみ利用するか、`notify: false` を設定

#### 端末ネイティブ通知（OSC / `terminal_notify`）

> OSC 経路は自動検証できない（OSC は戻り値を持たず、配信成否を検知できない）。Ghostty 実機で目視確認する。

- [ ] **Ghostty で OSC 通知が出る**: Ghostty で `terminal_notify: "auto"`（または `"osc777"`）にして起動し、新着で通知センターにバナーが出る。通知名義は **Ghostty**（octobell 名義ではない）
- [ ] **desktop-notifications off 時に静かに失敗する**: Ghostty 設定で `desktop-notifications = false` にすると、新着でもバナーが出ず、かつアプリは異常終了・エラー表示しない（fire-and-forget）
- [ ] **alt-screen 描画が崩れない**: OSC 通知が出る瞬間に TUI 画面の表示崩れ・ちらつき・エスケープ列の漏れが無い
- [ ] **二重通知が出ない**: OSC 経路選択時、同じ新着で beeep 経路の通知が重複して出ない（バナーは 1 つ）
- [ ] **フォールバック**: 非対応端末（例: VS Code 統合端末）で `auto` なら beeep 経路になり通知が出る。`./octobell --once`（tty 無し相当）や `osc777` 指定でも beeep に落ちて動作する
- [ ] **`off`**: `terminal_notify: "off"` で常に beeep 経路になる

### 4. 設定ファイル

- [ ] `~/.config/octobell/config.json` を作成し、`poll_seconds` / `participating` / `all` 等が反映される
- [ ] 設定ファイルが無くても既定値で動作する

## 既知の制約・要確認事項

- 通知のクリックでアクション（クリックして Issue を開く）は OS 通知ライブラリ（beeep）が非対応。TUI 側のキー操作で開く
- `per_page` の最大値は GitHub のドキュメント上 50。100 を受け付けるかは要確認
- fine-grained PAT での通知 API 利用可否は未確認（classic スコープ `repo` / `notifications` を推奨）
