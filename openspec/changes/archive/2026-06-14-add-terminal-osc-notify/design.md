## Context

現状の OS 通知は `internal/notify` の `Notifier` interface（`Beeep` / `Noop`）で抽象化され、`Beeep` は gen2brain/beeep を通じて macOS では外部プロセス `osascript`（または `terminal-notifier`）を起動する。Ghostty は `desktop-notifications` が既定 ON で、OSC 9 / OSC 777 エスケープシーケンスを受け取ると端末自身がデスクトップ通知を出す。これを使えばサブプロセス起動なしに通知できる。

octobell は Bubble Tea v2（`charm.land/bubbletea/v2`）の alt-screen で描画しており、通知送信はポーリング結果を処理する `Update`／`tea.Cmd` の流れの中で `Notifier.Notify` を呼ぶ形になっている。`Beeep` は stdout と独立に動くが、OSC 経路は端末の出力ストリームへバイト列を流す必要があり、描画との競合を考慮しなければならない。

制約: 開発環境は headless（TTY・OS 通知・macOS 無し）のため、OSC の実配信は自動検証できず手動検証（`docs/manual-verification.md`）に依存する。

## Goals / Non-Goals

**Goals:**
- 対応端末（まず Ghostty）で OSC エスケープシーケンスによるネイティブなデスクトップ通知経路を `Notifier` 実装として追加する。
- 検出と設定に基づき `TerminalOSC` か `Beeep` の**いずれか一方のみ**を選び、二重通知を防ぐ。
- 未対応端末・検出不能時は従来どおり `Beeep` にフォールバックし、後方互換を保つ。
- 既存の `Notifier` 抽象・通知文面整形・`notify=false`／`--no-notify` の挙動を変えない。

**Non-Goals:**
- 通知の名義を octobell にすること（OSC では端末アプリ名義になる。octobell 名義は署名済み `.app` バンドルが必要で対象外）。
- SSH 越し運用の最適化（OSC が本来効く領域だが、検出変数が SSH を越えにくく、本変更の主眼はローカル Ghostty）。
- iTerm2 / kitty など他端末の網羅対応（抽象は一般化するが、初期実装は Ghostty を主対象とする）。
- OSC 配信成否の検知（OSC は戻り値を持たない）。

## Decisions

### D1. OSC 777 を主経路、OSC 9 をフォールバックとする
- OSC 777（`ESC ] 777 ; notify ; <title> ; <body> ST`）は title と body を分離できるため、octobell の「件名／本文」整形（既存の `notification-alerting` 要件）にそのまま載る。
- OSC 9（`ESC ] 9 ; <message> ST`）は単一文字列のみ。採用時は `件名: 本文` を連結する。message を `数字 + ;` で始めない（ConEmu 拡張・`OSC 9;4` 進捗表示との衝突回避）。
- 終端は ST（`ESC \`）を用いる。
- **代替案**: OSC 9 のみ採用 → title/body を活かせず整形要件と噛み合わないため却下。

### D2. 端末検出は環境変数のホワイトリスト方式
- `TERM=xterm-ghostty` または `TERM_PROGRAM=ghostty` を Ghostty と判定する。
- **確信が持てる端末のみ OSC を撃つ**。未知端末へ盲撃ちすると、OSC を解釈しない端末では `9;…` 等が alt-screen に漏れて描画を壊すため。
- SSH 越しでは `TERM_PROGRAM` 等の検出変数が消えるため、`auto` では未検出として `Beeep` に落ちる。必要なら設定の明示指定で強制する（D5）。
- **代替案**: 全端末へ盲撃ちして「対応端末だけ反応する」前提 → 行儀の悪い端末で画面汚染リスクがあるため却下。

### D3. 通知バックエンドは排他セレクタで単一の Notifier に解決する
- `Beeep` と `TerminalOSC` は最終的に同一の OS 通知センターへ届くため、両方走らせると同じ新着が2回通知される。**両者を同時に使ってはならない**。
- `cmd/octobell/main.go` で「検出 + 設定」から **1 個の `Notifier`** を組み立てて TUI に渡す（TUI 側は従来どおり単一 `Notifier` を受け取るだけ）。
- 優先順位: `notify=false` / `--no-notify` なら `Noop`（最優先）→ それ以外は `terminal_notify` と端末検出で `TerminalOSC` か `Beeep` を選ぶ。
- **代替案**: TUI 内で複数 Notifier を保持し条件分岐 → 二重通知の不変条件が配線に散らばり壊しやすいため、組み立てを main に集約する。

### D4. OSC は controlling terminal へシーケンス全体を単一 write で出力する
- `TerminalOSC` は出力先 `io.Writer`（既定は `/dev/tty`、開けなければ `os.Stdout`）を注入で受け取る。
- **エスケープシーケンス全体を 1 回の `Write` 呼び出しで出す**。小サイズの write はアトミックで、Bubble Tea の frame 書き込みとバイト単位で混ざらない。
- OSC 777 / 9 の通知シーケンスは画面セルを変化させない（端末が消費して何も描かない）ため、alt-screen 描画中でも理屈上は安全。`/dev/tty` を使うことで Bubble Tea の stdout バッファとは別 fd になり、frame 出力と干渉しにくい。
- **代替案**: Bubble Tea の出力ライタを共有しロックする → v2 が安全なフックを公開していない場合に複雑化するため、単一 write + `/dev/tty` を第一候補とし、手動検証で描画崩れがないか確認する（Risks 参照）。

### D5. 設定 `terminal_notify`（既定 `auto`）
- 値: `auto | osc777 | osc9 | off`。
  - `auto`: 端末検出に従う。Ghostty 検出時は OSC 777、未検出時は `Beeep`。
  - `osc777` / `osc9`: 検出を上書きして強制（SSH 越し等、検出が効かない場合の逃げ道）。
  - `off`: 常に `Beeep` を使う。
- 既存の `configuration` 要件にならい、ファイル未記載なら既定 `auto`、存在するキーのみ上書き。
- CLI フラグは今回追加しない（将来の余地として残す）。

## Risks / Trade-offs

- **配信成否を検知できない** → OSC は戻り値を持たず、Ghostty 側で `desktop-notifications` が off／OS 通知許可が無い場合でも黙って出ない。`Beeep` の `exit 0 ≠ 配信` と同根。緩和: 手動検証チェックリスト（`docs/manual-verification.md`）に「対応端末で実際にバナーが出るか」「Ghostty 設定 off 時に静かに失敗するか」を追加する。
- **未対応端末への盲撃ちで画面汚染** → 検出ホワイトリストで確信時のみ撃つ（D2）。`auto` 既定では未知端末は `Beeep` に落ちる。
- **alt-screen 描画との競合** → 単一 write + `/dev/tty`（D4）で緩和するが、実機検証が必要。手動検証で描画崩れ・ちらつきが無いか確認する。
- **通知名義が端末アプリ（Ghostty）になる** → 受容する。octobell 名義は `.app` 化が必要でスコープ外。
- **検出の取りこぼし** → Ghostty 以外や将来のバージョンで `TERM`/`TERM_PROGRAM` がずれた場合は `Beeep` に落ちるだけで、通知自体は失われない（安全側のフォールバック）。

## 確定事項（旧 Open Questions）

- **`auto` は OSC 777 のみ自動採用する**。`auto` で対応端末を検出したら常に OSC 777 を使い、OSC 9 は `terminal_notify=osc9` の明示指定時のみ用いる。検出ロジックを単純・予測可能に保つため。
- **`/dev/tty` を開けない実行時は `Beeep` にフォールバックする**。パイプ経由・`--once` 等で controlling terminal を開けない場合、セレクタは `TerminalOSC` を選ばず `Beeep` に落とす（tty の有無もセレクタの判定材料にする）。安全側に倒し、`--once` の一覧出力をエスケープ列で汚さない。
- **本変更は Ghostty に絞る**。検出は `TERM=xterm-ghostty` / `TERM_PROGRAM=ghostty` のみ。iTerm2（`LC_TERMINAL=iTerm2`）/ kitty（`TERM=xterm-kitty`, OSC 99）への検出拡張は後続 change に分離する（スコープと手動検証範囲を Ghostty 1 端末に限定するため）。
