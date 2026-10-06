# 直近の PR / コミットの傾向（受け入れられる PR とは）

docker/compose の直近のクローズ済み PR 150 件と、upstream `main` のコミット履歴を調べ、
「どんな PR がマージされ、どんな PR が閉じられているか」「どんなコミットが積まれているか」
をまとめたもの。自分の PR に付いたレビュー指摘の分析は [pr-review-patterns.md](./pr-review-patterns.md) にある。

- 調査日: 2026-10-03
- 対象: https://github.com/docker/compose/pulls?q=is%3Apr+state%3Aclosed の直近 150 件
  （作成日 2026-08-17 〜 クローズ日 2026-10-02）
- コミット: `upstream/main` の 2026-08-15 以降、マージコミット以外の 260 件
- 根拠の一次資料: upstream の `CONTRIBUTING.md`、`AI_POLICY.md`、`AGENTS.md`(=`CLAUDE.md`)、
  `.github/PULL_REQUEST_TEMPLATE.md`

> 注意: `origin` は自分の fork (`htoyoda18/compose`) で、`origin/main` は 2025-12 で止まっている。
> 最新の状態は `git fetch upstream main` してから `upstream/main` を見ること。

---

## 1. 数字で見る全体像

| 作成者の区分                                         | マージ | マージされずクローズ | 作成からマージまで(中央値) |
| ---------------------------------------------------- | -----: | -------------------: | -------------------------: |
| メンテナ (ndeloof / glours / thaJeztah)              |     77 |                    9 |                    0.7 日 |
| bot (dependabot / docker-agent)                      |     21 |                   15 |                    0.4 日 |
| 外部コントリビューター                               |      4 |                   18 |                    0.8 日 |
| 自分 (htoyoda18)                                     |      5 |                    1 |                    3.8 日 |

- **マージの大半はメンテナ自身の PR**。ndeloof 44 件、glours 21 件、thaJeztah 12 件。
- **外部の PR は 22 件中 4 件しかマージされていない（約 18%）**。マージされた 4 件は
  ricardobranco777 ×2（openQA でのテスト失敗修正）、denisolnce ×1（hook 出力）、Endika ×1（watch のシンボリックリンク）。
- メンテナの PR でも閉じられたものは「同じ作業を別 PR で出し直した」ケースがほとんど
  （例: #14083/#14082 は試作として出したあと「#14081 で設計し直した計画に置き換える」として閉じ、
  中身を小さな PR に分けて出し直した。#14156 → #14200、#14201 → #14221 も同じ作業の出し直し）。
- dependabot の PR は、新しいバージョンが出ると古い bump PR が自動で閉じられるので、クローズ数が多い。
- bot 以外でマージされた PR の差分は**中央値 84 行**。ただし大きい PR もある
  （#14087 Scenario DSL 導入 +3952/-3907、#14093 compose-go jobs 対応 +2048/-572）。こうした大きい PR は
  エピック（#14074 / #14081）の一部として、メンテナが出したものだけ。

## 2. いま main で動いているテーマ（メンテナの作業領域）

直近 1.5 か月のメンテナ PR は、次のテーマに集中している。
外部から触ると衝突しやすい領域でもある。

| テーマ                         | 代表的な PR                                                                                             | 内容                                                                                     |
| ------------------------------ | ------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| **plan engine（エピック #14081）** | #14200, #14221, #14077, #14078, #14104, #14105, #14106                                                  | 命令的だった start の処理を、reconciler が組む「計画（plan）」に置き換えていく            |
| **provider / relay**           | #14193, #14175, #14236, #14238, #14252, #14258, #14264, #14275                                          | provider サービス用のネットワーク relay、イメージ配布、移行時のクリーンアップ           |
| **e2e テスト基盤の刷新**       | #14087（Scenario DSL）, #14132（`e2e` ビルドタグ）, #14226, #14228, #14234, #14242                       | 宣言的な `Scenario` DSL を導入。ログ行ではなく状態を待つ。不具合はまず e2e で再現する   |
| **「コードが正しく自分を説明する」（エピック #14074）** | #14129, #14130, #14131, #14133, #14135, #14136                                        | コメントには「経緯」ではなく「振る舞い」を書く。存在しない API を説明するドキュメントを直す |
| **依存の整理**                 | #14123（buildx を Go 依存から外す）, #14161（survey 削除）, #14185（clockwork → synctest）, #14207（perfsprint） | 依存ライブラリを減らし、lint を強化する                                                  |
| **compose-go 追従**            | #14093, #14178, #14215, #14253, #14281                                                                  | compose-go の jobs 対応、構造体のリファクタに引きずられないハッシュの固定                 |
| **バグ修正（glours 中心）**    | #14094, #14095, #14096, #14112, #14119, #14165, #14177, #14202, #14213, #14266, #14270                  | watch / build / display / network / ps / `--parallel` などの個別の不具合                  |

## 3. マージされずに閉じられた外部 PR と、その理由

閉じられた理由は、次の 5 パターンに分けられる。

### ① 実在する問題ではない・再現しない

| PR                                                     | 理由（メンテナのコメント要旨）                                                                                                |
| ------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------- |
| [#14153](https://github.com/docker/compose/pull/14153) formatter の nil 対応 | 「再現できる問題が無い。issue へのリンクも無い。`formatter.Print` の呼び出し元 4 か所はすべて具体的な型を渡しており、nil になり得ない」 |
| [#14126](https://github.com/docker/compose/pull/14126) watch 後の依存再起動   | 「本当にバグを再現したのか？ #13717 と同じ手順を main で試したが再現しなかった」                                       |
| [#14243](https://github.com/docker/compose/pull/14243) 引数名 `dependant`→`dependent` | 「どちらの綴りも正しく、直すべき誤りが無い。しかもこの PR はコンパイルが通らない（呼び出し側の名前を変え忘れている）」 |
| [#14248](https://github.com/docker/compose/pull/14248) README の言い回し     | 「改善になっていない。この文は compose-file の話で、"A Compose" というものは存在しない」                                |

### ② 修正する場所が違う・根本原因を外している（メンテナが別 PR で直す）

| PR                                                     | 理由                                                                                                                                         | 代わりにマージされた PR |
| ------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------- |
| [#14184](https://github.com/docker/compose/pull/14184) bake の進捗表示 | 「分析は正しいが、直す場所が違う。根本原因は `display.Mode` が stderr を基準に `auto` を解決するのに、bake は stdout に書くという食い違い」 | #14194                  |
| [#14181](https://github.com/docker/compose/pull/14181) exec の stdin    | 「`-i` はローカルの stdin をつなぐかどうかだけを決める。stdout が TTY かどうか（`-T` の判断材料）とは別の話」                          | —                       |
| [#14169](https://github.com/docker/compose/pull/14169) 読めない `.env` を無視する | 「`.env` が存在するのに読めないのは本物のエラーなので、無視してはいけない」。直すなら compose-go 側                                | compose-go#925          |
| [#14115](https://github.com/docker/compose/pull/14115) ttyWriter の Done    | 「バグ自体は本物だが、この直し方は安全ではない」                                                                                       | #14119                  |
| [#14168](https://github.com/docker/compose/pull/14168) 非 amd64 でテストをスキップ | 「スキップではなく、古いイメージ（v0.0.3）を v0.0.7 に上げるのが正しい直し方」→ 作者自身が出し直した                              | #14170                  |
| [#14237](https://github.com/docker/compose/pull/14237) down の進捗表示     | コード自体は承認されたが、より広い修正に吸収された                                                                                     | #14266                  |

### ③ 過去に議論して却下済み・メンテナが望んでいない機能

- [#14241](https://github.com/docker/compose/pull/14241) `run` で ports が無視されるときに警告を出す:
  元の issue #14229 が「#10138 と重複。すでに議論して却下済み。`compose run` のたびに警告が出ると、
  ports に頼っていない大多数のユーザーにとってうるさい」として閉じられた。
  → **`status/approved` が付いていない issue に対して PR を出すと、こうなる。**

### ④ コントリビューションファーミング（実績稼ぎ）・AI に丸投げした PR

- [#14186〜#14189](https://github.com/docker/compose/pull/14189)（NAVEENKUMARKR777）、[#14191/#14192](https://github.com/docker/compose/pull/14192)（veerareddyvishal144）、
  locker95（#14169 ほか）: 短時間に大量の、よく似た AI 生成 PR を無関係なリポジトリへ出していたため、
  **アカウントごとブロック**されたか、PR を閉じられた。
- メンテナは「AI を使うこと自体を否定しているのではない。開示されていて、本人がレビュー済みなら歓迎する」と毎回明言している。

### ⑤ AI 開示ファイルの扱い

- [#14237](https://github.com/docker/compose/pull/14237): コードは承認されたが、`AI_AGENT_DISCLOSURE.md` が
  PR に含まれていたため **AI disclosure gate の CI が失敗し、マージできなかった**。
  ndeloof から「差分を自分でレビューしてから、このファイルを消してほしい」と依頼されている。

## 4. マージされた外部 PR に共通すること

| PR                                                     | 何が良かったか                                                                                                                         |
| ------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------- |
| [#14245](https://github.com/docker/compose/pull/14245) | 実際の CI（openQA）のログへのリンク付き。原因（エラー文言がデーモンのバックエンドで違う／stderr に出る）を具体的に説明し、修正後の検証結果のリンクも貼っている。1 行の修正 |
| [#14170](https://github.com/docker/compose/pull/14170) | メンテナの提案（イメージを上げる）を素直に反映して出し直した。上流のどのコミットで直ったかを明記                                     |
| [#14088](https://github.com/docker/compose/pull/14088) | 「現状こう出る → 出力が無いのは `AttachStdout: false` が原因 → 影響する呼び出し経路は restart / run / ライブラリ利用」と、症状・原因・影響範囲を順に説明 |
| [#14084](https://github.com/docker/compose/pull/14084) | 2 回 CHANGES_REQUESTED（所有者の uid/gid がおかしくなる退行、どのエラーでもリトライしてしまう）を受けたが、粘り強く直してマージ。「Stick around」の実例 |

共通点:

1. **実際に起きている問題**で、ログや再現手順がある（CI ログ、エラーメッセージの実物）
2. **根本原因を説明している**（症状を抑えるだけの直し方ではない）
3. 差分が小さく、目的が 1 つに絞られている
4. レビューの指摘に応じて最後まで直している

## 5. コミットの積み方

### 5-1. マージ方式: rebase merge（マージコミットは作らない）

- 2026-08-15 以降の `upstream/main` には**マージコミットが 0 件**。PR のコミットがそのまま main に並ぶ。
- そのため **PR 内の各コミットメッセージが、そのまま main の履歴に残る**。
  「fix typo」「address review」のような雑なコミットを残すと、main に残ってしまう。
- CONTRIBUTING の方針: 「論理的な作業単位ごとに squash する。ほとんどの PR は 1 コミットになるはず。迷ったら 1 つに squash」
  「main に追従するときは `merge main` ではなく `rebase main`」「コミットごとにテストが通ること」
- 実際、外部 PR と小さな修正はほぼ 1 コミット。メンテナの大きい PR は複数コミットになることもある
  （例: #14221 は 5 コミット。うち 2 つは「copilot / docker-agent の指摘への対応」。#14202 は 3 コミット）。

### 5-2. コミットメッセージ: Conventional Commits がほぼ標準

260 コミット中 **249 件（96%）が `type(scope): summary` 形式**。

| type       | 件数 | 備考                                                    |
| ---------- | ---: | ------------------------------------------------------- |
| `fix`      |   95 | 圧倒的に多い                                            |
| `test`     |   45 | ほぼ `test(e2e): ...`                                   |
| `build`    |   32 | ほぼ `build(deps): bump ...`（dependabot）              |
| `refactor` |   16 |                                                         |
| `feat`     |   15 |                                                         |
| `docs`     |   12 |                                                         |
| `chore`    |   11 |                                                         |
| `e2e` / `ci` / `api` / `lint` など | 少数 | type ではなく領域名を先頭に置く書き方も一部ある |

よく使われる scope: `e2e`(45), `deps`(34), `provider`(18), `compose`(10), `watch`(8), `display`(6),
`hooks`(5), `up`(4), `build`(4), `hash`(3), `run`(3), `relay`(3), `publish`(3)。
**scope は「コマンド名」か「機能領域名」**（`watch`, `down`, `ps`, `port`, `logs`, `dry-run` など）。

### 5-3. サマリー行の書き方

- 平均 61 文字、最大 93 文字。
- **修正の作業内容ではなく、修正後の「振る舞い」を書く**のがメンテナの書き方:
  - `fix(watch): sync into a symlinked directory instead of failing`
  - `fix(up): drop benign context.Canceled noise from the final report`
  - `fix(down): spare dangling images of orphaned services, like their tagged image`
  - `cli: display.Mode always resolves to the mode actually rendered`
  - `api: option structs document what the implementation actually honors`
- e2e で不具合を再現するテストは `test(e2e): reproduce #14224 -- ...` の形で、**修正の前に別の PR として出す**ことがある（#14234, #14242）。

### 5-4. 本文・トレーラー

- **92% のコミットに本文がある**（平均 10 行強）。本文には「なぜこの変更が必要か」を書く。
- **DCO の `Signed-off-by:` はほぼ全コミットに付いている**（260 件中 258 件）。CI で必須。
- `Co-Authored-By: Claude ...` が 50 件ある。**メンテナ自身も AI を使って書いている**
  （AI_POLICY: 「この規則は外部からのコントリビューションに適用する。メンテナは実績に基づいて自分で判断してよい」）。

## 6. PR の説明文の型

`.github/PULL_REQUEST_TEMPLATE.md` の構成:

```
## What this PR does
## Related issue        ← "Fixes #1234"。AI を使った PR は approved な issue が必須
## Changes made
## Testing done         ← make test / e2e / make lint / make fmt / 手動テスト
## AI tool used (if applicable)
## Additional context
(not mandatory) A picture of a cute animal
```

メンテナ（特に ndeloof）の大きい PR は、次の型に寄せていることが多い:

```
## What this PR does, in one sentence
## Context            ← 今どう壊れているか。エラーの実物を貼る
## What the PR brings ← 何をどう変えたか
（テスト・残課題）
```

glours はテンプレートの旧版に近い `**What I did**` / `**Related issue**` を使っている。
どちらの型でも共通しているのは、**「何を」と「なぜ」が最初の数行でわかる**こと。

## 7. レビューの流れ

- **承認は最低 1 人のメンテナ**（GitHub の Approve）。実際は ndeloof・glours・thaJeztah が互いにレビューしている。
- **AI レビュー bot が先にコメントする**: `docker-agent`（pr-review ワークフロー）と
  `copilot-pull-request-reviewer` が自動でレビューし、作者はその指摘にも対応する
  （#14221 のコミット「address copilot-pull-request-reviewer's 3 findings」）。
- codecov がカバレッジをコメントする。
- メンテナ同士の PR は早い（中央値 0.7 日）。外部 PR は、指摘のやり取りが続くと数日かかる。

## 8. CONTRIBUTING / AI_POLICY のルール（外部コントリビューター向けの要点）

1. **大きな変更・新機能は、先に issue を立てて合意を取る**
2. **AI を使った PR は、`status/approved` ラベルの付いた issue に紐づける**。紐づいていない思いつきの PR は閉じられる
3. **使った AI ツールを開示する**（PR テンプレートの「AI tool used」欄）
4. **`make test` / `make lint` / `make fmt` / 関係する e2e を実際に走らせる**。自分でテストできない環境（OS・アーキテクチャ）向けのコードは書かない
5. **近くのコードのパターンに合わせる**（AGENTS.md を読む）
6. **全行を自分で説明できること**
7. **`AI_AGENT_DISCLOSURE.md` は PR を開く前に削除する**。エージェントは「人間が未レビュー」という印としてこのファイルを作る（AGENTS.md の指示）。人間がレビューを終えたら消す。残っていると CI でマージがブロックされる
8. **Stick around**: マージで終わりではない。レビュー指摘や後から出た退行にも対応する
9. **コントリビューションファーミング禁止**: 検証していない PR、issue を立てずに出す AI 生成 PR、1 つの変更をわざと細かく分けること、具体的な利益の無い見た目だけの変更
10. 変更はタスクの範囲に絞る（バグ修正のついでに無関係なリファクタ・typo 修正・整形をしない）

### 補足: next-tasks の調査中に分かったプロセス上の注意

- **`status/approved` ラベルはリポジトリに存在しない**。CONTRIBUTING.md は「AI を使った PR は `status/approved` の issue に紐づけること」と書いているが、
  実際には付けようがない。だから「メンテナがレビューや issue で明示的に後押ししたもの」を選ぶのが現実的な代わりになる（[tracking/next-tasks.md](../tracking/next-tasks.md) の 1 と 3）。
- **AI_POLICY.md:「AI の `Co-Authored-By` トレーラーは付けない」**。開示はコミットではなく PR の説明に書く。
  メンテナのコミットに Claude のトレーラーがあるのは「この規則は外部からのコントリビューションに適用する」という例外のため（§5-4）。
- PR を開く前に `AI_AGENT_DISCLOSURE.md` を、自分でレビューした上で削除する（上の 7）。

> 自分の作業ブランチ (`toyo/study`) の `CLAUDE.md` は upstream より古い。upstream の `AGENTS.md` には
> 「All agents must conform to AI_POLICY.md」と「Keep changes scoped to the task」が追加されている。

## 9. 次に PR を出す前のチェックリスト

上の調査から引き出したチェック項目は、[pr-review-patterns.md](./pr-review-patterns.md) のものと合わせて
[checklist.md](./checklist.md) にまとめた。
