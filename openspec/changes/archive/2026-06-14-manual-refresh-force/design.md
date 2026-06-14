## Context

`internal/tui/tui.go` の `fetchCmd()` は常に `m.lastModified` を条件付きリクエスト値として `client.List` に渡す。これは `tickMsg`（自動ポーリング）と `Refresh`（`ctrl+r` 手動更新）の両方から呼ばれる。

`handleFetched` は `res.NotModified`（304）のとき早期 return して `m.notifs` を更新しない。自動ポーリングではこれが正しい（レート節約）が、手動更新では「ユーザーが最新化を求めたのに何も起きない」結果になり、既読化済みなどで変化した一覧が古いまま残る。

## Goals / Non-Goals

**Goals:**
- 手動更新（`ctrl+r`）を無条件取得にし、常に最新の一覧を反映する。
- 自動ポーリングの条件付きリクエスト（レート節約）は維持する。

**Non-Goals:**
- `internal/github` の `List` シグネチャ変更。`lastModified` を空文字で渡せば既存実装が無条件 GET になるため不要。
- ポーリング間隔・`X-Poll-Interval` 関連の挙動変更。
- 既読化の楽観的更新（別変更 `clear-read-from-view` の範囲）。

## Decisions

### 決定1: `fetchCmd` に強制フラグを追加する

`fetchCmd(force bool)` とし、`force` のとき `lastModified` に空文字を渡す。

- `tickMsg` ハンドラ: `m.fetchCmd(false)`（従来どおり条件付き）。
- `Refresh` ハンドラ: `m.fetchCmd(true)`（無条件）。

`client.List` は `lastModified` が空なら `If-Modified-Since` を付けないため、`github` パッケージの変更は不要。

- **代替案A**: `handleFetched` で「手動更新由来なら 304 を無視」する。→ 304 を返すために既にリクエストを 1 回投げており、無条件取得の方が確実かつ単純。`fetchedMsg` に由来フラグを持たせる必要も生じるため不採用。
- **代替案B**: 手動更新時に `m.lastModified` をクリアする。→ 副作用で次回の自動ポーリングも無条件になりレート節約が効かなくなるため不採用。`fetchCmd` 引数で局所化する。

### 決定2: 強制取得後の Last-Modified 更新は従来ロジックに委ねる

無条件 GET は `200` を返し `handleFetched` が `m.lastModified = res.LastModified` を更新する。よって手動更新後も次回以降の自動ポーリングの条件付きリクエストは正しく機能する。

`handleFetched` は `m.lastModified = res.LastModified` を `NotModified` チェックの**前**に無条件で実行する（`internal/tui/tui.go`）。したがって強制取得（必ず `200`）でも `m.lastModified` が確実に更新され、次回 tick が条件付きに戻る挙動は既存ロジックのままで成立する。タスク 2.3 はこの不変条件をテストで固定する。

### 決定3: 手動更新の `loading` ガードは据え置く

`Refresh`（`ctrl+r`）ハンドラは `if !m.loading` の内側にあり、背景の条件付き tick が in-flight（`m.loading==true`）のときは強制取得が発火せず `return m, nil` となる。「明示的に最新化を求める操作」という思想からはわずかにズレるが、二重 fetch 防止の既存挙動であり、本変更で悪化はしない（force 化は発火した場合の取得方法を変えるだけ）。force を loading 中でも割り込ませる改修は本変更のスコープ外とし、必要なら別変更で扱う。

## Risks / Trace-offs

- [手動更新が毎回レート制限を消費する] → 手動更新はユーザー操作で頻度が低く、最新化の確実性を優先する。自動ポーリングは条件付きのまま据え置くので通常運用のレート消費は変わらない。
- [`fetchCmd` の呼び出し箇所すべての引数追加漏れ] → 呼び出しは tick と Refresh の 2 箇所のみ。grep で網羅を確認し、テストで両者の挙動を固定する。
- [強制取得で OS 通知が増えるのではないか] → 増えない。従来 `304` のときは `handleFetched` が早期 return し `Differ` を通らなかったが、force 化で必ず `200` → `m.differ.New()` を通る。ただし `Differ` は前回見たスレッド ID を記憶し新着のみ返す不変条件のため、同じ未読セットでは空を返し OS 通知は発火しない。レビュー時に疑われやすい点なので明記する。
