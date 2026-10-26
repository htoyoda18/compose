# 問題集の計画

← [document 一覧](../README.md)

compose 側のドキュメントは、量はそろってきた。これからは書き足すよりも、**読んだことを自分でテストして思い出す**ための問題集を作る。
問題を見てからドキュメントを読むと、何を探しながら読めばよいかがはっきりする。読み返すだけより定着する。

- 作成日: 2026-10-27
- 形式の手本: [english-study の問題集](https://github.com/htoyoda18/english-study/blob/main/note/quiz/01-basics.md)(答えを `<details>` で折りたたみ、答えの末尾に参照の節へのリンク)

## 置き場所

問題集は、出題範囲のドキュメントの隣に置く([docker-doc/README.md](../../../docker-doc/README.md) の方針と同じ)。

| 置き場所 | 出題範囲 | 例 |
| -------- | -------- | -- |
| `docker-doc/quiz/` | Docker の一般的な概念(`docker-doc/concepts/`) | コンテナと VM の違いは何か |
| `compose/document/quiz/` | compose の概念とコード(`compose/document/`、コードノート) | `up` は既存のコンテナを作り直すかどうかを何で決めるか |

- docker-doc の問題集は、あとで moby 側からもそのまま使える
- compose の問題集の README は、roadmap の段階の順に、docker-doc の問題集も含めて並べる。読む順に解けるようにする

## 問題の種類

english-study の 4 種類(説明 / 穴埋め / 誤り訂正 / 和文英訳)を、コードを読む勉強向けに置き換える。

| 種類 | 何を問うか | 例 |
| ---- | ---------- | -- |
| 説明 | 紛らわしい概念の違いや、設計の理由を自分の言葉で説明できるか | コンテナと VM の違いは何か。Compose はなぜ状態の DB を持たないのか |
| 予想 | compose.yaml と操作を見て、結果を当てられるか | この構成で `down -v` を打つと、どのボリュームが消えるか |
| 場所 | その処理がコードのどこにあるか言えるか | `-f` の引数をプロジェクトに変換するのはどのパッケージか |
| 穴埋め | 用語・名前・順序を正確に言えるか | プロジェクト名の優先順位は `-p` > ( ) > `name:` > ディレクトリ名 |

- **説明は、紛らわしいものだけに絞る**。単純な事実は穴埋めにする
- 予想と場所は、コードを読んだ人にしか解けない問題。compose の問題集の中心にする

## 1 問の形式

`docker-doc/quiz/01-intro.md` に置く問題の例(参照のリンクは、そのファイルからの相対パス)。

````markdown
### Q1 [説明]

コンテナと VM の違いは何か。

<details><summary>答え</summary>

**VM はコンピュータそのものを仮想化してゲスト OS をカーネルごと動かす。コンテナはホストのカーネルを共有した、隔離されたプロセスにすぎない。**

隔離は Linux カーネルの namespace(見える範囲)と cgroup(使える量)で実現する。だからコンテナは起動が速く、オーバーヘッドが小さい。

**確かめ方**

- `docker run --rm alpine uname -r` と、ホスト(Docker Desktop なら VM)の `uname -r` が同じになる(カーネルを共有している)

参照：[intro.md > VM との違い](../concepts/intro.md#vm-との違い)

</details>
````

- 答えは、**太字の 1 文**と、**1〜2 行の理由**
- **確かめ方**: 手元で試せるコマンドか、読むべきコードの場所(`path:line`)。english-study の「他の例」の代わりで、答えを自分の手で確かめられるようにする。無理に付けない
- **参照**: 出題元のドキュメントの節へのリンク。間違えたらそこへ戻って読み直す
- 答えはドキュメントとコードに基づいて書き、推測で書かない(Q&A の回答ルールと同じ)

## 問題の材料

1. **Q&A(`*-qa.md`)を最優先にする**。自分が実際に疑問に思ったことなので、問題にする価値が一番高い。1 つの Q&A から 1〜2 問作る
2. 各ドキュメントの「混乱しやすい点」「注意」「落とし穴」の節
3. next-tasks.md で見つけたバグ(N1、N2 など)。「予想」の問題にすると、仕様の細かいところを覚えられる

## 章立て

roadmap の段階に合わせる。目安は 1 章 10〜15 問。

### docker-doc/quiz/

| 章 | 出題範囲 | 主な問題 | 目安 |
| -- | -------- | -------- | ---- |
| 1. 入門 | intro.md(+ qa) | コンテナと VM、イメージとコンテナ、Docker が Linux でしか動かない理由、名前の整理(v1 / v2 / compose-spec) | 12 |
| 2. Docker Engine | docker-engine.md(+ qa)、docker-runtime-deepdive.md | dockerd / containerd / shim / runc の役割、Docker コンテキストで変わるもの | 12 |
| 3. コンテナの基盤 | container-tech.md(+ qa) | namespace と cgroup の違い、overlayfs、capability | 10 |
| 4. ネットワーク | networking.md(+ qa)、docker-networking.md | bridge / veth / DNS、ポート公開の仕組み | 10 |
| 5. イメージと OCI | oci.md(+ qa)、docker-build.md | レイヤー、manifest、digest とタグ | 10 |
| 6. Go | go.md(+ qa) | goroutine と channel、context のキャンセル、error wrapping | 10 |

### compose/document/quiz/

| 章 | roadmap | 出題範囲 | 主な問題 | 目安 |
| -- | ------- | -------- | -------- | ---- |
| 1. 全体像 | Stage 1 | how-compose-works.md、overview.md、architecture.md、ルートの note.md | Compose が状態を持たずにラベルで動く理由、cmd / pkg/api / pkg/compose の役割 | 10 |
| 2. コマンドが本体に届くまで | Stage 2 | flow.md の共通の流れ(+ qa)、cli.md(+ qa)、cmd / pkg/api のノート(+ qa) | プラグイン版とスタンドアロン版、`PersistentPreRunE`、`api.Compose` の 6 つの小さなインターフェース、Options の `Services` は絞り込みではない | 15 |
| 3. 概念と意味論 | Stage 3 前半 | compose-concepts.md(+ qa)、compose-lifecycle.md | プロジェクト名の決まり方、1 サービスとコンテナの数、depends_on と healthcheck、作り直しの判定(config hash) | 15 |
| 4. plan engine | Stage 3 後半 | algorithms.md、plan-engine.md、pkg/compose のノート | 観測 → 計画 → 実行、トポロジカル順序と並列化、DAG の実行 | 12 |
| 5. 他のコマンド | Stage 5 | flow.md のコマンド別の流れ、down / build / watch / run のノート(+ qa) | `down -v` で消えるもの、`run` が複雑な理由、job と `up` | 12 |
| 6. 開発と PR | Stage 6 | development.md、testing.md、contributing/ | DCO、AI_POLICY の「No contribution farming」、Scenario DSL の決まり、PR が閉じられる理由 | 10 |

合計の目安: docker-doc 64 問、compose 74 問。

## タスク

1 タスク 15〜30 分で、上から順に進める。読む順(roadmap)に合わせて作るので、作った章からすぐ使える。

| # | タスク | 成果物 | 状態 |
| - | ------ | ------ | ---- |
| Q0 | 2 つの問題集の README(使い方、問題の種類、一覧、収録の方針)を作る | `docker-doc/quiz/README.md`、`compose/document/quiz/README.md` | ✅ 2026-10-27 |
| Q1 | docker-doc 1. 入門 | `docker-doc/quiz/01-intro.md` | ✅ 2026-10-27 |
| Q2 | compose 1. 全体像 | `compose/document/quiz/01-overview.md` | |
| Q3 | compose 2. コマンドが本体に届くまで | `compose/document/quiz/02-cli-to-backend.md` | |
| Q4 | compose 3. 概念と意味論 | `compose/document/quiz/03-concepts.md` | |
| Q5 | compose 4. plan engine | `compose/document/quiz/04-plan-engine.md` | |
| Q6 | docker-doc 2. Docker Engine | `docker-doc/quiz/02-engine.md` | |
| Q7 | docker-doc 3〜5(コンテナの基盤、ネットワーク、イメージと OCI)。1 章ずつ | `docker-doc/quiz/03〜05` | |
| Q8 | compose 5. 他のコマンド | `compose/document/quiz/05-commands.md` | |
| Q9 | compose 6. 開発と PR | `compose/document/quiz/06-contributing.md` | |
| Q10 | docker-doc 6. Go | `docker-doc/quiz/06-go.md` | |
| Q11 | roadmap.md の各段階の終わりに、対応する問題集へのリンクを足す | `roadmap.md` | |

## 続け方

- **Q&A を書いたら、同じ時に問題を 1〜2 問足す**。Q&A は「読んでいて出た疑問への回答」、問題集は「あとで思い出すためのテスト」と役割を分ける
- 間違えた問題には、問題の見出しに印を付けておき(例: `### Q3 [予想] ✗`)、次に解くときはそこから始める
- ドキュメントやコードが変わって答えが古くなったら、問題も直す。答えに `path:line` を書いたものは、コードの変更で古くなりやすい
