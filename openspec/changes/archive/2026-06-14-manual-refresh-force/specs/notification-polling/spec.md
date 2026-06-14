## MODIFIED Requirements

### Requirement: 条件付きリクエストによるレート節約
システムは**自動ポーリング**時に直近の `Last-Modified` を `If-Modified-Since` として送り、サーバが変更なしを返した場合はレート制限を消費せずに「変化なし」と扱わなければならない（MUST）。このため取得には go-gh の RESTClient ではなく素の `*http.Client` を用いる（RESTClient は 304 をエラー化しヘッダを失うため）。

#### Scenario: 変化がなければ 304 で一覧は空
- **WHEN** 自動ポーリングで前回取得の `Last-Modified` を付けて取得し、サーバが `304 Not Modified` を返す
- **THEN** 結果は `NotModified=true`・通知一覧は空となり、レート制限を消費しない

#### Scenario: 200 ではデコードして Last-Modified を保持
- **WHEN** サーバが `200 OK` を返す
- **THEN** ボディを通知一覧へデコードし、レスポンスの `Last-Modified` を次回の条件付きリクエスト用に保持する

## ADDED Requirements

### Requirement: 手動更新は強制取得
システムはユーザーによる手動更新（`ctrl+r`）時、条件付きリクエスト（`If-Modified-Since`）を行わず、常に最新の一覧を無条件で取得して反映しなければならない（MUST）。手動更新はユーザーが明示的に最新化を求める操作であり、`304 Not Modified` による「変化なし」スキップを適用してはならない（MUST NOT）。

#### Scenario: 手動更新は If-Modified-Since を送らない
- **WHEN** ユーザーが `ctrl+r` で手動更新する
- **THEN** `If-Modified-Since` を付けずに取得し、サーバは `200` で最新の一覧を返す

#### Scenario: 手動更新で一覧が最新化される
- **WHEN** 既読化などで実際の一覧が変化した状態で `ctrl+r` を押す
- **THEN** 取得結果で `m.notifs` を置き換え、古い項目が残らない

#### Scenario: 自動ポーリングは従来どおり条件付き
- **WHEN** tick による自動ポーリングが発火する
- **THEN** 直近の `Last-Modified` を `If-Modified-Since` として送り、レート制限を節約する
