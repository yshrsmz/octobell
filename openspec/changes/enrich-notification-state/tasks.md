## 1. config: enrich_state トグル

- [x] 1.1 `internal/config` に `EnrichState bool`（JSON `enrich_state`）を追加し、`Default()` を `true` にする
- [x] 1.2 部分上書き規則（存在キーのみ上書き・欠損は既定維持）に沿うことをテストで確認（既定 true / 明示 false / 部分設定）

## 2. github: subject 実状態の型と取得

- [x] 2.1 `internal/github` に `SubjectState` 型（open / draft / merged / closed-completed / closed-not_planned）と `String()` を追加
- [x] 2.2 Issue/PR 詳細レスポンスのデコード構造体を追加（Issue: state, state_reason / PR: state, merged, draft）
- [x] 2.3 実状態の導出関数を実装（PR: merged→merged / closed→closed / draft→draft / else open、Issue: closed は state_reason で分岐、else open）し、headless 単体テストを書く
- [x] 2.4 `subject.url`（API URL）を GET して実状態を返すメソッドを追加（既存方針どおり素の `*http.Client` を使用）。state_change 以外・対象外種別・空 URL は取得しない

## 3. tui: 非同期エンリッチとキャッシュ

- [x] 3.1 `enrichedMsg{id, updatedAt, state, err}` メッセージ型と Model のキャッシュ（`map[string]enrichEntry{updatedAt,state}`）を追加
- [x] 3.2 `handleFetched` 後、`enrich_state` 有効時に `reason=state_change` の Issue/PR かつキャッシュミス（updated_at 不一致）の通知へエンリッチ Cmd を発行する
- [x] 3.3 並行取得に上限を設ける（同時実行数を制限）
- [x] 3.4 `enrichedMsg` を `Update` で処理：成功はキャッシュ更新＋該当アイテム副行を更新して再描画、失敗は当該項目のみ reason のみ表示にフォールバック
- [x] 3.5 一覧描画がエンリッチ完了を待たないこと（即描画→順次反映）を headless テストで確認

## 4. tui: 副行の置換・色分け・フィルタ

- [x] 4.1 `item` に実状態を持たせ、`refreshItems` でキャッシュから詰める
- [x] 4.2 `item.Description` を `reason=state_change` かつエンリッチ済みなら `state_change(<実状態>)` に整形し、`(<実状態>)` を括弧ごと GitHub 標準カラーで色分け（open=緑 / merged・closed-completed=紫 / draft・closed-not_planned=灰 / closed(PR)=赤）。`state_change` 本体は通常色。着色は末尾に置き lipgloss の reset で delegate 色が打ち切られないようにする
- [x] 4.3 `FilterValue` にエンリッチ済み実状態を含め、`merged` 等で絞り込めるようにする（テスト追加）
- [x] 4.4 エンリッチ前・他 reason・対象外種別・OFF 時は従来どおり reason のみ表示になることをテストで確認

## 5. 検証・ドキュメント

- [x] 5.1 `go vet ./...` / `go test ./...` / `go build ./cmd/octobell` を通す
- [x] 5.2 `docs/manual-verification.md` に `state_change(<状態>)` 表示・`(状態)` 括弧ごと色分け・`enrich_state=false` の手動検証項目を追加
- [x] 5.3 `README.md` の設定一覧に `enrich_state` を追記し、`state_change` 通知の副行に実状態を付記する旨を記載
