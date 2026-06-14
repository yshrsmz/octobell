# opsx:archive 完了後は session-review を実行する

`opsx:archive` / `openspec-archive-change` スキルでアーカイブ作業が完了したら、続けて `meta:session-review`（session-review）スキルを実行すること。アーカイブはひとつの change がレビュー済み・実装完了済みで確定する節目であり、そのセッションで得た学び（プロジェクト固有の gotcha・ユーザーの好み・再利用可能なワークフローなど）を振り返って保存するのに自然なタイミングである。

## やること

- `opsx:archive` または `openspec-archive-change` の最終ステップ完了後に `meta:session-review` スキルを起動する。
  - [[commit-after-opsx-archive]] のコミット提案とセットで運用する。順序は「アーカイブ → コミット提案・対応 → session-review」とし、コミット可否のユーザー応答を得た後に session-review を実行する。
- session-review の手順（学びの抽出・分類・`AskUserQuestion` での提案・承認分の適用）はスキル本文に従う。地の文で選択肢を並べない（[[ask-user-question]]）。

## やらないこと

- アーカイブ完了後に session-review を省略する（差分が小さい・学びが無さそうという自己判断でスキップしない。学びが無ければ session-review 自身が「保存対象なし」と結論づける）。
- session-review の提案を承認なしに勝手に適用する（適用はスキル本文どおりユーザー承認を取る）。
- `opsx:archive` 以外のタイミングで予防的に session-review を回す（このルールは `opsx:archive` 完了トリガに限定）。

## 理由

- アーカイブの節目はセッションの一区切りであり、文脈が薄れる前に学びを定着させる好機。毎回ユーザーが「振り返って」と打たなくても自動でレビュー機会を作ることで、知見の取りこぼしを減らす。
- session-review は保存先を checked-in な `.claude/rules/` / `CLAUDE.md` / skill 優先で分類するため、チームで共有すべき知見がローカルの auto-memory に埋もれにくくなる。
