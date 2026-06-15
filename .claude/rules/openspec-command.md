# openspec コマンドは `pnpm openspec` で使う

OpenSpec の CLI は、グローバルや `npx` で直接呼ばず、プロジェクトローカルの devDependency 経由で実行すること。`package.json` の `openspec` スクリプトを使う。

```sh
pnpm openspec <subcommand>
```

**理由**: `@fission-ai/openspec` は devDependency としてバージョン固定でインストールされており（`pnpm-lock.yaml` 準拠）、グローバル版や `npx` の最新版とバージョンがずれると挙動が変わるため。

**してはいけない例**
- `openspec ...`（グローバルインストール前提）
- `npx openspec ...`（バージョン未固定で取得される）

## `--json` を parse するときは `pnpm -s` を使う

`--json` 出力をそのままパイプして `python3 -c 'json.load(...)'` 等に渡す場合は `pnpm -s openspec ...`（`-s` = `--silent`）を使うこと。`-s` が無いと pnpm が先頭に `$ openspec <subcommand>` というスクリプト実行行を出力し、JSON パースが `Expecting value: line 1 column 1` で失敗する。

```sh
pnpm -s openspec status --change "<name>" --json | python3 -c '...'
```

