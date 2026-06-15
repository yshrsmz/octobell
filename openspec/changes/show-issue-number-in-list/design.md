## Context

通知一覧の 1 件は `internal/tui/item.go` の `item` が `list.Item` として描画する。主行 `Title()` は未読マーク + `subject.title`、副行 `Description()` はリポ名 · 種別 · 理由。GitHub Notifications API のレスポンスには Issue/PR 番号フィールドが無く、番号は `subject.url`（API URL `.../repos/owner/repo/(issues|pulls)/42`）に埋まっている。`internal/github/notifications.go` には既に `BrowserURL()` 用の正規表現 `reIssueOrPull`（group2 が番号）が存在する。

## Goals / Non-Goals

**Goals:**
- Issue / PullRequest の主行に `#<番号>` を表示する。
- 番号を持たない種別（Commit / Release / Discussion / CheckSuite など）は従来表示のまま。
- `/` フィルタで `42` / `#42` のどちらでも該当通知を絞れる。

**Non-Goals:**
- Issue と PR の視覚的な区別（種別は副行で既出のため番号表記は `#N` 共通とする）。
- 番号の色付け・ゼロ埋め・桁揃えなどの装飾。
- API レスポンスへの番号フィールド追加要求（API 仕様に番号は無い）。

## Decisions

**番号抽出は `internal/github` に置く（`Notification.SubjectNumber() string`）。**
理由: 番号は `subject.url` という GitHub API の構造に依存する知識であり、URL 解析を担う github パッケージが持つのが自然。既存 `reIssueOrPull` を `BrowserURL()` と共有でき追加コストがほぼゼロ。tui 側は整形だけに専念できる。
- 戻り値は番号文字列（例 `"42"`）。issues/pulls 以外、または `subject.url` が空のときは `""` を返す。
- 種別判定（Issue か PR か）は番号表記が `#N` で共通のため不要。メソッドは番号だけ返す。
- 代替案: tui 側で正規表現を持つ → github の URL 解析知識が二重化するため不採用。

**主行プレフィックスは未読マークの直後に挿入する（`● #42 Title`）。**
理由: 番号を左端付近に固定するとタイトルが長くても視認しやすい。未読マーク（`●` / 空白 2 文字）の整列は維持する。番号が無いときは現状の `● Title` と完全一致させ、既存テスト・体験を壊さない。

**フィルタは `FilterValue()` 末尾に `#<番号>` を 1 つ足すだけにする。**
理由: bubbles の `/` フィルタは部分一致なので、`#42` を含めておけば `42` でも `#42` でもマッチする。`42` を別途足す必要はない。番号が無ければ何も足さない。

## Risks / Trade-offs

- [タイトル先頭に番号が入ることで横幅が圧迫される] → 番号は通常 1〜6 桁程度で影響は軽微。装飾せず最小限の `#N ` のみ付与。
- [`subject.url` の形式が将来変わる] → 既存 `BrowserURL()` も同じ前提に依存しており、リスクは新規に増えない。正規表現を共有することで変更時の修正箇所も 1 箇所に集約される。

## Open Questions

なし（explore で表示位置・フィルタ方針は確定済み）。
