## 1. 番号抽出（internal/github）

- [x] 1.1 `Notification.SubjectNumber() string` を新設し、既存 `reIssueOrPull` を共有して `subject.url` から番号（group2）を返す。issues/pulls 以外・空 URL は `""` を返す。
- [x] 1.2 `notifications_test.go` に `TestSubjectNumber` を追加（issue / pull / commit / release / 空 URL の各ケース）。

## 2. 一覧表示とフィルタ（internal/tui）

- [x] 2.1 `item.Title()` を更新し、`SubjectNumber()` が非空なら未読マーク直後に `#<番号> ` を挿入する。番号なしは現状の `● <title>` / `  <title>` を維持。
- [x] 2.2 `item.FilterValue()` を更新し、番号が非空なら末尾に `#<番号>` を追加する（`42` / `#42` の両方で部分一致）。
- [x] 2.3 `tui_test.go` に検証を追加（Issue/PR の Title に `#N` が出る・番号なし種別は出ない・FilterValue に番号が入る）。

## 3. 検証

- [x] 3.1 `go test ./...` / `go vet ./...` / `go build ./cmd/octobell` を通す。
- [x] 3.2 `docs/manual-verification.md` の一覧表示チェック項目に番号表示の確認を追記する。
