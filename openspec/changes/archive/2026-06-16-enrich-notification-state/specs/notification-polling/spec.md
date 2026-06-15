## ADDED Requirements

### Requirement: subject 実状態のエンリッチ
システムはエンリッチ機能が有効（既定 ON）の場合、一覧取得後に `reason` が `state_change` の Issue / PullRequest 通知について subject 詳細（`subject.url`）を追加取得し、実状態を導出しなければならない（MUST）。Issue は `state`（open/closed）と `state_reason` から `open` / `closed-completed` / `closed-not_planned` を、PullRequest は `state` / `merged` / `draft` から `open` / `draft` / `merged` / `closed`（未マージ）を導出しなければならない（MUST）。`reason` が `state_change` 以外の通知、number を持たない種別（Commit / Release / Discussion / CheckSuite など）、および `subject.url` が空の通知は対象外とし、追加取得を行ってはならない（MUST NOT）。エンリッチ機能が無効の場合は追加取得を一切行ってはならない（MUST NOT）。

追加取得は一覧の描画をブロックしてはならず（MUST NOT）、結果が得られた通知から順に副行へ反映できなければならない（MUST）。追加取得の結果は `(通知ID, updated_at)` をキーにキャッシュし、`updated_at` が変化していない通知は再取得してはならない（MUST NOT）。並行取得には上限を設けなければならない（MUST）。個々の追加取得が失敗しても UI 全体を壊さず、当該通知のみ `reason` のみ表示にフォールバックしなければならない（MUST）。追加取得は読み取り（GET）であり、既存のクライアント使い分け方針（GET は素の `*http.Client`、2xx で完結する書き込みは RESTClient）に従わなければならない（MUST）。

#### Scenario: state_change の Issue/PR の実状態を導出する
- **WHEN** エンリッチ有効で、`reason=state_change` の Issue/PR 通知について subject 詳細の取得に成功する
- **THEN** Issue は state/state_reason から、PR は state/merged/draft から実状態を導出し、当該通知に紐づけてキャッシュする

#### Scenario: state_change 以外・対象外種別は追加取得しない
- **WHEN** `reason` が `state_change` 以外の通知、Commit / Release / Discussion など番号を持たない種別、または `subject.url` が空の通知である
- **THEN** subject 詳細の追加取得を行わず、実状態を持たない

#### Scenario: updated_at が同じならキャッシュを使う
- **WHEN** 直近のポーリングと `updated_at` が同一の Issue/PR 通知である
- **THEN** subject 詳細を再取得せず、キャッシュ済みの実状態を再利用する

#### Scenario: updated_at が変われば再取得する
- **WHEN** Issue/PR 通知の `updated_at` が前回から変化している
- **THEN** subject 詳細を改めて取得し、実状態を更新する

#### Scenario: 追加取得の失敗は局所フォールバック
- **WHEN** ある通知の subject 詳細取得が失敗する
- **THEN** その通知のみ `reason` のみ表示にフォールバックし、他の通知の表示・取得は継続する

#### Scenario: エンリッチ無効時は追加取得しない
- **WHEN** エンリッチ機能が設定で無効になっている
- **THEN** subject 詳細の追加取得を一切行わず、副行は従来どおり `reason` を表示する

#### Scenario: 一覧描画をブロックしない
- **WHEN** 一覧取得が完了する
- **THEN** 実状態の取得完了を待たずに一覧を描画し、取得できた通知から順に副行を更新する
