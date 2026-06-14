## 1. 設定（internal/config）

- [x] 1.1 `Config` に `TerminalNotify string`（JSON キー `terminal_notify`）を追加する
- [x] 1.2 `Default()` の既定値を `auto` にする
- [x] 1.3 部分上書き・欠損キー維持が `terminal_notify` でも効くことを確認・テストする
- [x] 1.4 未知値は `auto` にフォールバックする正規化を加え、テストする

## 2. 端末検出（internal/notify）

- [x] 2.1 `TERM=xterm-ghostty` / `TERM_PROGRAM=ghostty` を判定する検出関数を追加する（env をパラメータ化してテスト可能にする）
- [x] 2.2 検出関数の headless 単体テストを追加する（Ghostty 検出 / 未検出ケース）

## 3. OSC 通知バックエンド（internal/notify）

- [x] 3.1 `TerminalOSC` 実装を追加する（出力先 `io.Writer` を注入。既定は `/dev/tty`、開けなければ Beeep にフォールバック）
- [x] 3.2 OSC 777（`ESC ] 777 ; notify ; <件名> ; <本文> ST`）を組み立て、シーケンス全体を単一 `Write` で出す
- [x] 3.3 OSC 9 経路（`ESC ] 9 ; <メッセージ> ST`、件名+本文を結合、数字+`;` 始まりを回避）を追加する
- [x] 3.4 `bytes.Buffer` を出力先に差し替え、生成シーケンスのバイト列を検証する単体テストを追加する

## 4. バックエンドセレクタ

- [x] 4.1 「`notify=false`→`Noop` 最優先 / `terminal_notify` と端末検出で `TerminalOSC` か `Beeep` を排他選択。OSC を選ぶ条件でも `/dev/tty` を開けなければ `Beeep` に落とす」セレクタを実装する（`auto` は OSC 777 のみ自動、OSC 9 は明示 `osc9` 時のみ）
- [x] 4.2 セレクタの分岐（auto+検出, auto+未検出, osc777, osc9, off, notify=false, tty なし）を網羅する単体テストを追加する
- [x] 4.3 `cmd/octobell/main.go` でセレクタを使い、単一 `Notifier` を組み立てて TUI に渡す配線にする

## 5. ドキュメント・検証

- [x] 5.1 `README.md` に `terminal_notify` 設定（値と既定 `auto`）を追記する
- [x] 5.2 `docs/manual-verification.md` に「Ghostty で OSC 通知が出るか」「Ghostty の desktop-notifications off 時に静かに失敗するか」「alt-screen 描画が崩れないか」のチェック項目を追加する
- [x] 5.3 `go vet ./...` / `go build` / `go test ./...` が通ることを確認する
