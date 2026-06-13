# openspec コマンドは `pnpm openspec` で使う

OpenSpec の CLI は、グローバルや `npx` で直接呼ばず、プロジェクトローカルの devDependency 経由で実行すること。`package.json` の `openspec` スクリプトを使う。

```sh
pnpm openspec <subcommand>
```

**理由**: `@fission-ai/openspec` は devDependency としてバージョン固定でインストールされており（`pnpm-lock.yaml` 準拠）、グローバル版や `npx` の最新版とバージョンがずれると挙動が変わるため。

**してはいけない例**
- `openspec ...`（グローバルインストール前提）
- `npx openspec ...`（バージョン未固定で取得される）
