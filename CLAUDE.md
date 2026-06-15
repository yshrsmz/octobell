# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

octobell は GitHub の通知（Notifications）を定期取得して未読管理する TUI アプリ（Go）。`gh`（GitHub CLI）の認証を再利用し、新着を OS デスクトップ通知で知らせる。詳細な仕様・キーバインド・設定は `README.md` を参照。

## コマンド

```sh
go build -o octobell ./cmd/octobell   # ビルド
go test ./...                         # 全テスト
go test ./internal/tui/ -run TestName # 単一テスト
go vet ./...                          # CI と同等の静的検査
go mod verify                         # CI と同等のモジュール検証

# 認証・ネットワークを要するライブテスト（CI ではスキップ。実 GitHub に接続する）
OCTOBELL_LIVE=1 go test ./internal/github/ -run TestLiveConditionalRequest -v
```

Go バージョンはリポジトリ単位で `mise` により固定（`go.mod` の `go 1.26.4`）。CI（`.github/workflows/ci.yml`）は `go mod verify` → `go vet` → `go build` → `go test` を実行する。

## アーキテクチャ

依存方向は `cmd/octobell/main.go` → `internal/{config,github,notify,tui}`。TUI が他 3 パッケージをオーケストレーションし、各 internal パッケージは互いに独立（tui のみ全てに依存）。

- **`internal/github`** — Notifications API クライアント。**読み取り（GET）と書き込み（PATCH/PUT）で意図的にクライアントを使い分ける**:
  - 取得は go-gh の `DefaultHTTPClient()`（素の `*http.Client`）を使う。`RESTClient` は `304` 等の非 2xx をエラー化してレスポンスヘッダ（`X-Poll-Interval` / `Last-Modified`）を失い、条件付きリクエストに使えないため。認証ヘッダは go-gh の RoundTripper が自動付与。
  - 既読化は 2xx で完結するため `RESTClient` を使う。
  - `BrowserURL()` は API URL（`subject.url`）を `repository.html_url` ベースのブラウザ URL に変換（PR は `/pulls/N`→`/pull/N`、未対応種別は repo URL にフォールバック）。Enterprise host にも追従。
- **`internal/notify`** — OS 通知の抽象化。`Notifier` interface（`Beeep` 実機 / `Noop` 無効・headless）で実機差を吸収。`Differ` は前回ポーリングで見たスレッド ID を記憶し**新着のみ**を返す（初回起動時はバックログ全件を抑制）。
- **`internal/tui`** — Bubble Tea v2（`charm.land/bubbletea/v2`）の Model-Update-View。ポーリングは `X-Poll-Interval` を下限に強制（`pollSeconds()` = max(設定, サーバ値)）。既読化は**楽観的更新**（再ポーリングを待たずローカル状態を即更新）。`ctrl+a`（全既読）は取り消し不能なので二度押し確認。
  - **lipgloss v2 gotcha**: `lipgloss.Color` は型ではなく `color.Color`（`image/color`）を返す**関数**。色を返すヘルパーの戻り型は `image/color.Color` にする。
  - **list の部分着色は文字列末尾に置く**: `item.Description` / `Title` を lipgloss で部分着色するとき、着色セグメントは末尾に連結する。中間に置くと着色末尾の reset（`\x1b[0m`）で list delegate が行全体に被せる前景色が打ち切られ、reset より後ろの文字が端末既定色に化ける（例 `state_change(merged)` で着色を `merged` だけにすると閉じ括弧が化けるため `(merged)` を括弧ごと末尾着色する）。
- **`internal/config`** — XDG 準拠の JSON 設定（`~/.config/octobell/config.json`）。ファイルが無ければ `Default()`、存在するキーのみ上書き（部分設定可）。

### 重要な不変条件

- **新着通知のスパム防止**: `Differ` の初回呼び出しは必ず「新着なし」を返す。起動直後に既存通知全件が一斉通知される不具合（gitify #1437 系）を避けるための仕様。
- **ポーリング間隔**: ユーザー設定値より GitHub の `X-Poll-Interval` が大きければそちらを優先（レート制限尊重）。条件付きリクエストで `304` のときは API レート制限を消費しない。

## テスト方針

開発環境は headless（TTY・OS 通知・macOS 無し）のため、自動検証できる範囲が限られる。TUI は render/入力を介さない **Model ロジックの headless 単体テスト**（`internal/tui/tui_test.go`）で検証する。実機検証が必要な範囲（render・キー入力・OS 通知・macOS）は `docs/manual-verification.md` のチェックリストに従う。新機能を加えたら同ドキュメントの更新も検討する。

## OpenSpec ワークフロー

このリポジトリは spec-driven な変更管理に OpenSpec を用いる（`openspec/` 配下、`.claude/skills/openspec-*`）。仕様変更は `openspec/specs/` の各 capability spec（configuration / notification-polling / notification-alerting / notification-tui / read-management）が基準。変更提案・実装・アーカイブは `openspec-propose` / `openspec-apply-change` / `openspec-archive-change` スキルを使う。OpenSpec CLI は必ず `pnpm openspec <subcommand>` で実行する（`.claude/rules/openspec-command.md`）。

## 規約

- ユーザーへの確認・選択は `AskUserQuestion` ツールを使う（`.claude/rules/ask-user-question.md`）。
- 依存追加・更新はバージョン固定・lock ファイル優先・postinstall 系スクリプト回避（グローバル CLAUDE.md のサプライチェーン方針に従う）。
- GitHub Actions は組織方針によりコミット SHA で固定する。
- コメント・コミットメッセージ・ドキュメントは日本語、技術用語は英語のまま。
