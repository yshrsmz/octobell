# `opsx:verify` の直前に `/code-review` を走らせる

`opsx:verify` / `openspec-verify-change` スキルを実行する前に、当該 change の実装に対して `/code-review --fix` を 1 回回す。コードが整った状態に対して verify をかけることで、verify の PASS 結果がそのまま authoritative になり、二度手間（verify → review で修正 → 再 verify）を避ける。

## やること

- `opsx:verify` / `openspec-verify-change` を実行する直前に、`git diff` で実装変更があるか確認する。
- 変更がある場合は **先に `/code-review --fix` を実行**する（自動連続実行で構わない、ユーザー確認は挟まない）。
  - code-review が修正を入れた場合、通常通り「コミットしますか？」を `AskUserQuestion` で聞く（[[use-askuserquestion]]、[[commit-after-opsx-archive]] の同様の運用に合わせる）。
  - コミット可否のユーザー応答を得た後に `opsx:verify` を実行する。
- `git diff` が空（実装変更なし、artifacts のみ更新等）の場合は `/code-review --fix` を skip し、即座に `opsx:verify` を実行する。
- code-review の対象スコープは当該 change の実装ファイルを優先する。`openspec/changes/<name>/tasks.md` から関連ファイルを推測できる場合はそれを使い、難しければ `git diff` 全体を対象にしてよい。

## やらないこと

- `opsx:verify` の**後**に `/code-review --fix` を回す（after 順序）。verify 後に修正が入ると verify 結果が古くなり再 verify が必要になる。
- `/code-review --fix` の中で発生した修正をユーザー確認なしに自動コミットする（コミット提案までは自動、commit 実行は承認必須）。
- `opsx:verify` を実行しない時に予防的に `/code-review --fix` を回す（このルールは `opsx:verify` トリガに限定）。
- change と無関係なリポジトリ全体の大規模リファクタを code-review に任せる。スコープは当該 change の実装に絞る。

## 理由

- `opsx:verify` の役目は「実装が change artifacts（proposal / design / specs / tasks）と整合しているか」のチェック。`/code-review --fix` の役目は「reuse / quality / efficiency の観点で実装を整える」。両者が補完関係になる場面（change を archive する直前など）は多い。
- after 順序だと「verify PASS → code-review --fix が修正 → コミット → verify 状態が古くなり再 verify」と二度手間が発生する。before 順序なら「review で整える → verify で 1 回確定」で済む。
- `/code-review --fix` の修正が verify を落とすケース（例: 必須エクスポートを削除した、テスト名を変えた）も before のほうが原因切り分けが容易（直前の review が原因と即座に特定できる）。
