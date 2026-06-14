## Why

現状の OS 通知は beeep（macOS は osascript / terminal-notifier、Linux は D-Bus / notify-send）の単一経路で、外部プロセス（osascript）の起動を伴う。Ghostty のように OSC エスケープシーケンスで端末自身がデスクトップ通知を出せる端末では、サブプロセスを起こさずより軽量・ネイティブに通知できる。この経路を追加し、対応端末ではそちらを使えるようにする。

## What Changes

- `internal/notify` に **OSC エスケープシーケンスでデスクトップ通知を出す `Notifier` 実装（`TerminalOSC`）** を追加する。
  - 主経路は **OSC 777**（`ESC ] 777 ; notify ; <title> ; <body> ST`、title/body 分離）。
  - フォールバックとして **OSC 9**（`ESC ] 9 ; <message> ST`、単一文字列）も選べる。OSC 9 の message は `数字 + ;` で始めない（ConEmu / `OSC 9;4` 進捗表示との衝突回避）。
- **端末検出**を追加する。`TERM=xterm-ghostty` / `TERM_PROGRAM=ghostty` 等を起点に、OSC 通知に対応すると判定できる端末でのみ OSC を撃つ。
- **通知バックエンドのセレクタ**を導入する。検出結果と設定に基づき `TerminalOSC` か `Beeep` の**どちらか一方のみ**を選ぶ。両者は最終的に同一の OS 通知センターへ届くため、同時に走らせると二重通知になる。両方をチェーンしてはならない。
- 設定に **`terminal_notify: auto | osc777 | osc9 | off`**（既定 `auto`）を追加する。`auto` は検出に従い、明示指定は検出を上書きし、`off` は常に beeep を使う。
- 未対応端末・検出不能時は従来どおり `Beeep` にフォールバックする（挙動の後方互換を維持）。

## Capabilities

### New Capabilities
<!-- 新規 capability はなし。既存の alerting / configuration の要件変更で表現する。 -->

### Modified Capabilities
- `notification-alerting`: 「OS デスクトップ通知」要件に、OSC エスケープシーケンス経由の通知バックエンドを追加する。検出と設定で beeep か OSC のいずれか一方を選ぶ排他選択を要件化し、二重通知を禁止する。
- `configuration`: 設定キー `terminal_notify`（既定 `auto`）と、その部分上書き・既定値挙動を追加する。

## Impact

- **コード**: `internal/notify`（`TerminalOSC` 実装・端末検出・セレクタの追加）、`internal/config`（`terminal_notify` キーと既定値）、`internal/tui` 配線（選んだ単一 `Notifier` を受け取る）、`cmd/octobell/main.go`（セレクタの組み立て）。
- **依存**: 追加なし（OSC は標準出力へのバイト列出力のみ。新規パッケージ不要）。
- **設計上の留意点（design.md で詳述）**:
  - Bubble Tea v2 の alt-screen 描画中に生エスケープシーケンスを安全に出力する方法（描画と競合しない出力経路）。
  - OSC は exit code を持たず**配信成功を検知できない**（beeep の osascript も exit 0 ≠ 配信だが、OSC はさらに戻り値自体が無い）。検証は手動（`docs/manual-verification.md`）に依存。
  - 通知の**名義は端末アプリ（Ghostty）**になり、octobell 名義にはならない（octobell 名義には署名済み `.app` バンドル化が必要で、本変更の対象外）。
- **スコープ外**: SSH 越し運用の最適化（OSC は本来そこで効くが、本変更の主眼はローカル Ghostty での軽量・ネイティブ通知経路の追加）、`.app` バンドル化による名義変更。
