# skill が指示するサブエージェント fan-out を省略しない

`code-review` / `simplify` のように、skill 本文がサブエージェント（Agent ツール）への fan-out を明示的に指示している場合は、その手順を踏むこと。差分が小さい・自明に見えるという理由で、サブエージェントを起動せず自分の頭の中だけで簡易的なインラインレビューに置き換えてはいけない。

## やること

- review/simplify 系 skill を起動したら、本文の「Agent ツールで N 個の finder/verifier を起動」という指示をそのまま実行する。
- 差分が小さくても省略しない。skill が fan-out を設計に組み込んでいるのは、独立した視点で recall を上げるためで、インライン代替はその設計目的を損なう。

## やらないこと

- 「22 行の styling-only 差分だから overkill」といった自己判断でサブエージェント起動をスキップする。
- どうしても fan-out が過剰だと判断する場合に、勝手に省略する。その場合は省略せず先にユーザーへ確認する（[[ask-user-question]]）。

## 理由

- このリポジトリの `code-review-before-opsx-verify.md` 運用では、`opsx:verify` の前に `/code-review --fix` を 1 回回して結果を authoritative にする。fan-out を省略したインラインレビューは、その「1 回で確定」の前提（十分な網羅性）を満たさない。
