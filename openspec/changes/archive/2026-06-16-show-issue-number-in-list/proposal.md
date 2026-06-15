## Why

通知一覧では主行にタイトル、副行にリポジトリ名・種別・理由しか出ておらず、Issue / PullRequest の番号が分からない。番号が見えれば「同じタイトルの別 PR」の判別や、CLI・ブラウザ・チャットでのやり取りで頻出する `#42` という参照との突合が一目でできる。番号は既に `subject.url` に含まれており、抽出ロジック（`reIssueOrPull`）も `BrowserURL()` 用に存在するため、低コストで実現できる。

## What Changes

- 一覧の主行（Title）に、Issue / PullRequest の場合は `#<番号>` を未読マークの直後に表示する（例: `● #42 Fix the login bug`）。
- 番号を持たない種別（Commit / Release / Discussion / CheckSuite など）は従来どおり番号を付さず素通しする。
- `/` インクリメンタルフィルタの対象に番号を加え、`42` でも `#42` でも該当通知を絞り込めるようにする。

## Capabilities

### New Capabilities
<!-- なし -->

### Modified Capabilities
- `notification-tui`: 通知一覧の主行表示に Issue/PR 番号を含める要件を追加し、フィルタ対象に番号を加える。

## Impact

- `internal/github/notifications.go`: `subject.url` から番号を取り出す `SubjectNumber()` を新設（既存 `reIssueOrPull` を `BrowserURL()` と共有）。
- `internal/tui/item.go`: `Title()` に番号プレフィックスを追加、`FilterValue()` に番号を追加。
- テスト: `internal/github/notifications_test.go`（`SubjectNumber`）、`internal/tui/tui_test.go`（番号表示・フィルタ）。
- API 依存・通信・既存のポーリング/既読化挙動には影響なし。
