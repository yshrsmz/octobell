# opsx:archive 完了後はコミットを提案する

`opsx:archive` / `openspec-archive-change` スキルでアーカイブ作業が完了したら、ユーザーに明示的に頼まれなくても、その場で「コミットしますか？」と提案すること。

グローバル CLAUDE.md の「コミットは明示的に頼まれるまでしない」既定をこのケースでは上書きする。**承認は依然としてユーザーから取る**（提案するだけで、勝手にコミットはしない）。

## やること

- `opsx:archive` または `openspec-archive-change` の最終ステップ完了直後に `git status` を実行する。
- アーカイブ作業に由来する変更（`openspec/changes/archive/...` 配下のファイル追加・移動、`openspec/specs/<capability>/spec.md` の更新、関連実装ファイルの変更など）が staged / unstaged に残っている場合、`AskUserQuestion` を使ってコミット提案を行う（地の文で「コミットしますか？」と書かない。詳細は [[use-askuserquestion]]）。
  - `header`: `コミット` 程度
  - 選択肢例: `今すぐコミット (推奨)` / `内容を確認してから` / `コミットしない`
  - 提案するコミットメッセージは `description` に含めてユーザーが判断できるようにする
- 提案するコミットメッセージは Conventional Commits 準拠で、変更内容を 1-2 行で要約する。例:
  - `feat: archive <change-name>` （実装＋アーカイブが同一ブランチで完結している場合）
  - `chore(openspec): archive <change-name>` （実装は別コミット済みでアーカイブ移動のみの場合）
- アーカイブと無関係な変更が混ざっている場合も `AskUserQuestion` で「アーカイブ分だけ抽出 / 全部まとめる / コミットしない」を選ばせる。
- ユーザーが承認した場合は通常の commit 手順（HEREDOC + Co-Authored-By trailer）に従う。

## やらないこと

- 承認なしに `git commit` を実行する。
- 地の文で「コミットしますか？」と聞く（必ず `AskUserQuestion`）。
- 複数の change を一度に archive した場合に、それぞれ別コミットを強制する（基本は 1 コミットで OK、ユーザーが分けたい場合のみ分割）。

## 理由

- OpenSpec の運用上、アーカイブはレビュー済み・実装完了済みの change を確定させる節目であり、コミット粒度として自然な区切り。
- 毎回ユーザーが「コミットして」と打つのは手間。提案までは自動化することで作業フローを短縮する。
- ただし自動 commit は意図しないファイルを巻き込むリスクがあるため、承認は必ず取る（hook ではなくこのソフトルールに留める理由）。
