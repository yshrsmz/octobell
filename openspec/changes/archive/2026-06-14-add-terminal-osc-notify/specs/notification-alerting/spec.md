## MODIFIED Requirements

### Requirement: OS デスクトップ通知
システムは新着を OS のデスクトップ通知として送らなければならない（MUST）。通知は fire-and-forget でありクリックアクションには対応しない。送信失敗はアプリの動作を妨げてはならない（MUST NOT）。通知は選択された単一の通知バックエンドを通じて配信されなければならず（MUST）、同一の新着を複数バックエンドから重複して通知してはならない（MUST NOT）。

#### Scenario: 新着があれば通知を送る
- **WHEN** 新着が 1 件以上ある
- **THEN** 選択された単一バックエンドでデスクトップ通知を送信する（既定の `Beeep` は macOS で osascript / terminal-notifier、Linux で D-Bus / notify-send）

#### Scenario: 重複通知を出さない
- **WHEN** 端末ネイティブ通知（OSC）が選択されている
- **THEN** 同じ新着に対して `Beeep` 経路の通知は送らない

## ADDED Requirements

### Requirement: 端末ネイティブ通知バックエンド（OSC）
システムは、対応端末で OSC エスケープシーケンスによりデスクトップ通知を出す通知バックエンドを提供しなければならない（MUST）。主経路は OSC 777（件名と本文を分離）とし、OSC 9（単一文字列）をフォールバックとして選べなければならない（MUST）。OSC バックエンドは戻り値を持たず、配信成否を呼び出し側に伝えない。

#### Scenario: OSC 777 で件名と本文を送る
- **WHEN** OSC 777 バックエンドが選択され、新着通知を送る
- **THEN** `ESC ] 777 ; notify ; <件名> ; <本文> ST` 形式のシーケンスを controlling terminal へ単一の write で出力する

#### Scenario: OSC 9 は単一文字列に整形する
- **WHEN** OSC 9 バックエンドが選択され、新着通知を送る
- **THEN** 件名と本文を結合した単一メッセージを `ESC ] 9 ; <メッセージ> ST` で出力し、メッセージは数字と `;` で始めない

#### Scenario: 描画を壊さない
- **WHEN** TUI が alt-screen で描画中に OSC 通知を出力する
- **THEN** シーケンス全体を 1 回の write で出力し、画面表示を破壊しない

### Requirement: 通知バックエンドの選択と排他
システムは、端末検出と設定 `terminal_notify` に基づき、通知バックエンドを起動時に単一へ解決しなければならない（MUST）。`Beeep` と OSC バックエンドを同時に使ってはならない（MUST NOT）。通知が無効（`notify=false` / `--no-notify`）の場合は他に優先して `Noop` を選ばなければならない（MUST）。

#### Scenario: 対応端末を検出したら OSC を選ぶ
- **WHEN** `terminal_notify=auto` かつ `TERM=xterm-ghostty` または `TERM_PROGRAM=ghostty` を検出する
- **THEN** OSC 777 バックエンドを選択する

#### Scenario: 未対応・未検出は Beeep にフォールバック
- **WHEN** `terminal_notify=auto` で対応端末を検出できない
- **THEN** 既存の `Beeep` バックエンドを選択する

#### Scenario: 明示指定は検出を上書きする
- **WHEN** `terminal_notify` が `osc777` または `osc9` に明示設定されている
- **THEN** 端末検出の結果によらず指定された OSC バックエンドを選択する

#### Scenario: off は常に Beeep
- **WHEN** `terminal_notify=off`
- **THEN** 端末検出によらず `Beeep` バックエンドを選択する

#### Scenario: tty を開けない場合は Beeep
- **WHEN** OSC バックエンドが選ばれる条件でも controlling terminal（`/dev/tty`）を開けない（パイプ経由・`--once` 等）
- **THEN** `TerminalOSC` を選ばず `Beeep` バックエンドにフォールバックする

#### Scenario: 通知無効は Noop が最優先
- **WHEN** `notify=false` または `--no-notify` が指定されている
- **THEN** `terminal_notify` の値によらず `Noop` を選択し、通知を一切送らない
