## Context

GitHub Notifications API（`GET /notifications`）は subject の状態（open/closed/merged 等）を返さない。`reason` フィールドの `state_change` は「状態が変わった」事実のみで、内容を区別できない。実状態を一覧で見せるには subject 詳細（Issue/PR）の追加取得が必須となる。

既存アーキテクチャの制約:
- ポーリングは `X-Poll-Interval` を下限に強制し、自動取得は `If-Modified-Since` の条件付きリクエストで `304` 時にレート制限を消費しない。
- 取得（GET）は go-gh の `DefaultHTTPClient()`（素の `*http.Client`）、書き込み（2xx 完結）は `RESTClient` という使い分けがある。
- TUI は Bubble Tea v2 の Model-Update-View。副作用は `tea.Cmd` で非同期に行い、結果を `tea.Msg` で受け取る。
- 既読化は楽観的更新、`Differ` の初回は必ず「新着なし」。

## Goals / Non-Goals

**Goals:**
- `reason=state_change` の Issue/PR で実状態（open / draft / merged / closed-completed / closed-not_planned）を `state_change(<実状態>)` 形式で副行に付記し、`(<実状態>)` を括弧ごと色分けする。
- 一覧描画をブロックせず、取得できた項目から順に副行を更新する。
- 取得対象を `reason=state_change` の Issue/PR に限定し、取得＝表示を一致させる。
- `(通知ID, updated_at)` キャッシュ・並行上限・OFF トグルで追加 API 消費を抑える。
- 既存の不変条件（X-Poll-Interval 尊重、304 でレート消費なし、Differ 初回抑制、楽観的既読）を壊さない。

**Non-Goals:**
- `state_change` 以外の reason での実状態表示（取得もしない。将来 reason に依らず出したくなったら拡張）。
- state_change 以外の reason 値の整形・国際化。
- subject 詳細の ETag/If-None-Match による条件付きキャッシュ（updated_at キャッシュで十分。将来拡張）。
- レビュー状態・CI 状態・ラベルなど state 以外のメタ情報表示。
- Commit / Release / Discussion / CheckSuite への状態表示（state を持たないため対象外）。

## Decisions

### D1: エンリッチは非同期 Cmd、結果は per-item メッセージで反映
`handleFetched` で一覧を即描画した後、対象（`reason=state_change` の Issue/PR かつキャッシュミス）ごとにエンリッチ Cmd を発行する。各取得完了で `enrichedMsg{id, updatedAt, state, err}` を返し、`Update` でキャッシュとアイテム副行を更新→再描画する。
- 代替案: 取得を待って一括描画 → 一覧表示が subject 取得の遅延に引きずられ UX 劣化。却下。
- 代替案: 1 つの Cmd 内で全件を順次取得し最後にまとめて返す → 部分反映できず、失敗の局所化も難しい。却下。

### D2: キャッシュキーは `(通知ID, updated_at)`、Model に保持
状態変化は `state_change` 通知を生み `updated_at` を更新するため、`updated_at` 不変なら状態も不変とみなせる。`map[string]enrichEntry`（key=通知ID, value={updatedAt, state}）を Model に持ち、ポーリングをまたいで再利用する。`updated_at` 不一致のエントリのみ再取得。
- これによりポーリングごとの追加リクエストを「実際に変化した Issue/PR だけ」に限定できる。
- 一覧から消えた通知のエントリは任意で掃除（メモリは小さいので必須ではない）。

### D3: 並行取得に上限を設ける
初回起動や「全部動いた」ピーク時に最大 `per_page`（50）件のバーストが起きうる。固定上限（例: 同時 4〜8）で取得を絞り、Core API（5000/hr）を圧迫しない。実装は Cmd 発行数の制御、または取得側の semaphore で行う。

### D4: 実状態の型と導出は internal/github に置く
`SubjectState`（enum 相当の型 + String()）と、Issue/PR 詳細レスポンスのデコード・導出関数を `internal/github` に追加する。取得は既存の GET 方針（素の `*http.Client`）に合わせる。TUI は導出済みの状態値のみを扱い、API 詳細に依存しない。
- 導出規則: Issue は `state==closed` のとき `state_reason`（`completed`/`not_planned`/`reopened`）で分岐、open は `open`。PR は `merged==true`→`merged`、`state==closed`→`closed`、`draft==true`→`draft`、それ以外 `open`。

### D5: 副行の付記と色分けは item.Description で行う
`item.Description` は現在 `repo · type · reason`。`reason=state_change` かつエンリッチ済みなら reason を `state_change(<実状態>)` に整形し、`(<実状態>)` を括弧ごと状態に応じた lipgloss 前景色で着色する（`state_change` 本体は通常色）。`item` は状態を参照できる必要があるため、`item` 自体に状態フィールドを持たせて `refreshItems` で詰める（list との相性がよい）。
- 色は GitHub 標準に倣う: open=緑 / merged・closed-completed=紫 / draft・closed-not_planned=灰 / closed(PR 未マージ)=赤。`applyContrastStyles` と同様、ダーク端末で沈まない明るめの色を選ぶ。
- 着色は文字列末尾に置く。中間に着色を挟むと lipgloss の reset（`\x1b[0m`）で list delegate が被せる副行色が打ち切られ、reset より後ろの文字（閉じ括弧）が端末既定色に化けるため。`state_change` + 着色した `(<実状態>)` の順に連結する。

### D6: フィルタ対象に実状態を含める
`FilterValue` は reason を含む。エンリッチ済み（=state_change の項目）なら実状態文字列も連結し、`merged` 等で絞り込めるようにする。

### D7: config に `enrich_state`（bool, 既定 true）を追加
`Default()` に追加し、部分上書き規則に従う。OFF 時は `handleFetched` 後のエンリッチ Cmd を発行しない。

## Risks / Trade-offs

- [追加 GET による Core API 消費] → `(ID, updated_at)` キャッシュで変化分のみ取得、並行上限でバースト抑制、`enrich_state=false` で完全停止。通常運用では数件/poll に収まる見込み。
- [起動直後のバースト（最大 50 件）] → 並行上限 + 非同期反映で UX への影響を局所化。必要なら初回のみ取得を間引く余地を残す。
- [headless テストの限界] → 色分け・実機 render は自動検証できないため、状態導出ロジック・キャッシュ判定・置換文字列は Model/github の headless 単体テストで担保し、色・render は `docs/manual-verification.md` のチェックリストに追加する。
- [updated_at 以外で状態が変わるケース] → GitHub 仕様上 PR/Issue の状態遷移は通知と updated_at 更新を伴うため実害は小さい。手動更新（`ctrl+r`）はキャッシュを無効化せず使えるが、必要なら force 時にエンリッチキャッシュも破棄する選択肢を残す。
- [Enterprise host] → subject.url は API URL なのでそのまま GET でき、host 追従は既存どおり。

## Open Questions

- 並行取得上限の具体値（4 / 8 など）— 実装時にレート影響を見て確定。
- 手動更新（`ctrl+r`）時にエンリッチキャッシュを破棄して再取得するか、updated_at 判定に委ねるか。
- 一覧から消えた通知のキャッシュ掃除を行うか（メモリは小さいので任意）。
