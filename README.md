# octobell 🔔

GitHub の通知（Notifications）を定期的に取得して未読管理する TUI アプリ。
[gitify](https://github.com/gitify-app/gitify) の TUI 版のようなもので、`gh`（GitHub CLI）の認証を再利用し、新着を OS のデスクトップ通知で知らせる。

## 特徴

- **gh 認証を再利用**: `gh auth login` 済みのトークンをそのまま使う（`go-gh` 経由）。追加のトークン設定は不要
- **効率的なポーリング**: `X-Poll-Interval` ヘッダを尊重し、`If-Modified-Since` 条件付きリクエストで変更がなければ `304 Not Modified`（API レート制限を消費しない）
- **OS デスクトップ通知**: 新着（前回ポーリングからの差分）のみを通知（macOS / Linux）
- **TUI で未読管理**: 一覧・既読化・ブラウザ起動・インクリメンタルフィルタ

## 必要なもの

- [GitHub CLI (`gh`)](https://cli.github.com/) がインストール済みで `gh auth login` 済みであること
  - 通知 API には `notifications` または `repo` スコープが必要。`gh auth login` の既定スコープには `repo` が含まれるため通常は追加設定不要
  - private リポジトリの通知には `repo` スコープが必要
- OS デスクトップ通知:
  - **macOS**: 追加不要（`osascript` にフォールバック。`terminal-notifier` があればより高機能）
  - **Linux**: D-Bus（デスクトップ環境）または `notify-send`（`libnotify`）

## インストール

```sh
go install github.com/yshrsmz/octobell/cmd/octobell@latest
```

または、このリポジトリをクローンしてビルド（Go バージョンは `mise` でリポジトリ単位に固定）:

```sh
mise install        # go.mod の Go バージョンを導入
go build -o octobell ./cmd/octobell
```

## 使い方

```sh
octobell                 # TUI を起動（定期ポーリング + OS 通知）
octobell --once          # 通知を一度だけ取得して一覧表示し終了（TUI を起動しない）
octobell --once --all    # 既読も含めて一覧表示
octobell --no-notify     # OS 通知を無効化して TUI 起動
octobell --version
```

### キーバインド

| キー | 動作 |
|---|---|
| `j` / `k`, `↓` / `↑` | カーソル移動 |
| `enter` | ブラウザで開く + 既読化 |
| `o` | ブラウザで開く（既読化しない） |
| `r` / `.` | 選択中を既読化 |
| `ctrl+a` | すべて既読化（二度押しで確認） |
| `ctrl+r` | 手動更新 |
| `/` | インクリメンタルフィルタ（リポ名・種別・理由・タイトル） |
| `?` | ヘルプの開閉 |
| `q` / `ctrl+c` | 終了 |

## 設定

設定ファイルは XDG Base Directory 準拠で `~/.config/octobell/config.json`（`$XDG_CONFIG_HOME` があればそちら）。
存在しない場合は既定値で動作する。

```json
{
  "poll_seconds": 60,
  "all": false,
  "participating": false,
  "per_page": 50,
  "mark_read_on_open": true,
  "notify": true,
  "terminal_notify": "auto"
}
```

| キー | 既定 | 説明 |
|---|---|---|
| `poll_seconds` | `60` | 希望ポーリング間隔（秒）。実際の間隔は GitHub の `X-Poll-Interval` を下限に強制される |
| `all` | `false` | 既読も含めて取得する |
| `participating` | `false` | 参加（mention / review 依頼など）通知のみに絞る |
| `per_page` | `50` | 1 ページあたりの取得件数（最大 50） |
| `mark_read_on_open` | `true` | `enter` で開いたときに既読化する |
| `notify` | `true` | OS デスクトップ通知を有効にする |
| `terminal_notify` | `"auto"` | 通知バックエンドの選択。`auto`（対応端末を検出したら OSC、未検出なら beeep）／ `osc777` ／ `osc9` ／ `off`（常に beeep）。不正値は `auto` 扱い |

### 通知バックエンド（`terminal_notify`）

通常は OS 通知（beeep: macOS は osascript / terminal-notifier、Linux は D-Bus / notify-send）を使う。
対応端末（現状 **Ghostty**）では、OSC エスケープシーケンスで端末自身が通知を出す経路も選べる。外部プロセスを起動せず軽量。

- `auto`（既定）: `TERM=xterm-ghostty` / `TERM_PROGRAM=ghostty` を検出したら **OSC 777**、未検出なら beeep。
- `osc777` / `osc9`: 端末検出によらず OSC を強制（SSH 越し等、検出が効かない場合に明示指定する）。
- `off`: 常に beeep。
- beeep と OSC は同時に使わない（同じ通知が二重に出るのを防ぐため、いずれか一方のみ）。`--no-notify` / `notify=false` のときは何も送らない。
- OSC を選んでも controlling terminal（`/dev/tty`）を開けない場合（パイプ経由・`--once` 等）は beeep にフォールバックする。
- 注意: OSC は端末アプリ（Ghostty）名義の通知になる（octobell 名義にはならない）。また配信成否は検知できず、Ghostty 側で `desktop-notifications` が無効だと黙って出ない。

## 設計メモ

- **取得（読み取り）** は `go-gh` の `RESTClient` ではなく `DefaultHTTPClient()` の素の `*http.Client` を使う。`RESTClient` は `304` を含む非 2xx をエラー化してレスポンスヘッダ（`X-Poll-Interval` / `Last-Modified`）を失うため、条件付きリクエストに使えない。認証ヘッダは `go-gh` の RoundTripper が自動付与する
- **既読化（書き込み）** は 2xx で完結するため `RESTClient` を使う
- **OS 通知** はクリックアクション非対応（fire-and-forget）。Issue/PR を開く操作は TUI 側のキー（`enter` / `o`）で行う
- **新着判定** は前回ポーリングで見たスレッド ID との差分。初回起動時は既存バックログ全件を通知しない

## ライセンス

MIT
