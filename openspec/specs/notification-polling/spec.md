# notification-polling Specification

## Purpose
TBD - created by archiving change document-existing-behavior. Update Purpose after archive.
## Requirements
### Requirement: gh 認証の再利用
システムは GitHub Notifications API へのアクセスに、`gh`（GitHub CLI）が保存した認証情報を再利用し、追加のトークン設定を要求してはならない（MUST NOT）。

#### Scenario: gh ログイン済みで取得できる
- **WHEN** `gh auth login` 済みの環境でクライアントを生成する
- **THEN** go-gh の RoundTripper が認証ヘッダを自動付与し、追加設定なしで `GET /notifications` を実行できる

### Requirement: 条件付きリクエストによるレート節約
システムは取得時に直近の `Last-Modified` を `If-Modified-Since` として送り、サーバが変更なしを返した場合はレート制限を消費せずに「変化なし」と扱わなければならない（MUST）。このため取得には go-gh の RESTClient ではなく素の `*http.Client` を用いる（RESTClient は 304 をエラー化しヘッダを失うため）。

#### Scenario: 変化がなければ 304 で一覧は空
- **WHEN** 前回取得の `Last-Modified` を付けて取得し、サーバが `304 Not Modified` を返す
- **THEN** 結果は `NotModified=true`・通知一覧は空となり、レート制限を消費しない

#### Scenario: 200 ではデコードして Last-Modified を保持
- **WHEN** サーバが `200 OK` を返す
- **THEN** ボディを通知一覧へデコードし、レスポンスの `Last-Modified` を次回の条件付きリクエスト用に保持する

### Requirement: ポーリング間隔の下限強制
システムは実際のポーリング間隔を、ユーザー設定値とサーバの `X-Poll-Interval` のうち大きい方とし、GitHub が要求する下限を尊重しなければならない（MUST）。いずれも有効値でない場合は 60 秒にフォールバックする。

#### Scenario: サーバ値がユーザー設定より大きい
- **WHEN** `X-Poll-Interval` がユーザー設定の `poll_seconds` より大きい
- **THEN** サーバ値を採用する
<!-- test: tui_test.go TestModelFlowHeadless（PollInterval=60 反映） -->

#### Scenario: 有効値がなければ 60 秒
- **WHEN** ユーザー設定・サーバ値ともに 1 未満／未指定
- **THEN** 間隔は 60 秒となる

### Requirement: 取得クエリオプション
システムは `all`・`participating`・`per_page` のクエリで取得対象を制御できなければならない（MUST）。`per_page` は最大 50 とし、未指定または 0 以下のときは 50 を用いる。

#### Scenario: per_page 未指定は 50
- **WHEN** `PerPage` が 0 以下で取得する
- **THEN** クエリの `per_page` は 50 になる

#### Scenario: all / participating の反映
- **WHEN** `All` または `Participating` が true で取得する
- **THEN** 対応するクエリ（`all=true` / `participating=true`）が付与される

### Requirement: GitHub Enterprise host への追従
システムは API のエンドポイント prefix を認証ホストから組み立て、GitHub Enterprise Server に追従しなければならない（MUST）。

#### Scenario: github.com は公開 API
- **WHEN** ホストが空・`github.com`・`api.github.com` のいずれか
- **THEN** prefix は `https://api.github.com/` になる

#### Scenario: Enterprise host は /api/v3/
- **WHEN** ホストが上記以外（例: `ghe.example.com`）
- **THEN** prefix は `https://ghe.example.com/api/v3/` になる

### Requirement: 取得失敗時のエラー
システムは 200 でも 304 でもない応答を、応答本文を含むエラーとして返さなければならない（MUST）。

#### Scenario: 非 2xx はエラー
- **WHEN** サーバが 401 など 304 以外の非 2xx を返す
- **THEN** ステータスと応答本文を含むエラーを返す

