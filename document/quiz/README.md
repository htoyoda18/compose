# 問題集

← [document 一覧](../README.md)

compose/document の概念・コード解説とコードノートを、**自分でテストして思い出す**形にしたもの。読み返すだけより定着する。
問題を見てから出題範囲を読むと、何を探しながら読めばよいかもはっきりする。

Docker の一般的な概念の問題集は [docker-doc/quiz/](../../../docker-doc/quiz/README.md) に置いている。下の一覧は、それも含めて [roadmap.md](../roadmap.md) の段階の順に並べた。上から解けば、読む順と同じになる。

## 使い方

1. 問題を読んで、**答えを開く前に**自分の言葉で答える
2. `答え` を開いて確かめる。**確かめ方** があれば、手元で打つか、書かれたコードの場所を開く
3. 間違えた問題は、答えの末尾の `参照：` のリンクから出題元の節へ飛んで読み直す。問題の見出しに印を付けておき(例: `### Q3 [予想] ✗`)、次に解くときはそこから始める

答えは `<details>` で折りたたんでいる。VSCode のプレビューか GitHub の上でクリックすると開く。

## 問題の種類

| 種類 | 何を問うか | 例 |
| ---- | ---------- | -- |
| 説明 | 紛らわしい概念の違いや、設計の理由を自分の言葉で説明できるか | Compose はなぜ状態の DB を持たないのか |
| 予想 | compose.yaml と操作を見て、結果を当てられるか | この構成で `down -v` を打つと、どのボリュームが消えるか |
| 場所 | その処理がコードのどこにあるか言えるか | `-f` の引数をプロジェクトに変換するのはどのパッケージか |
| 穴埋め | 用語・名前・順序を正確に言えるか | プロジェクト名の優先順位は `-p` > ( ) > `name:` > ディレクトリ名 |

予想と場所は、コードを読んだ人にしか解けない問題。compose の問題集はこの 2 つを中心にする。

## 一覧

| 段階 | 問題集 | 出題範囲 | 問題数 |
| ---- | ------ | -------- | ------ |
| Stage 1 | [docker-doc 1. 入門](../../../docker-doc/quiz/01-intro.md) | intro.md、intro-qa.md | 15 |
| Stage 1 | [compose 1. 全体像](01-overview.md) | how-compose-works.md、overview.md、architecture.md、ルートの note.md | 11 |
| Stage 2 | docker-doc 6. Go(未作成) | go.md、go-qa.md | |
| Stage 2 | [compose 2. コマンドが本体に届くまで](02-cli-to-backend.md) | flow.md の共通の流れ、flow-qa.md、cli.md、cli-qa.md、cmd / pkg/api のノート | 15 |
| Stage 3 | compose 3. 概念と意味論(未作成) | compose-concepts.md、compose-concepts-qa.md、compose-lifecycle.md | |
| Stage 3 | compose 4. plan engine(未作成) | algorithms.md、plan-engine.md、pkg/compose のノート | |
| Stage 4 | docker-doc 2. Docker Engine(未作成) | docker-engine.md、docker-engine-qa.md、docker-runtime-deepdive.md | |
| Stage 4 | docker-doc 3. コンテナの基盤(未作成) | container-tech.md、container-tech-qa.md | |
| Stage 4 | docker-doc 4. ネットワーク(未作成) | networking.md、networking-qa.md、docker-networking.md | |
| Stage 4 | docker-doc 5. イメージと OCI(未作成) | oci.md、oci-qa.md、docker-build.md | |
| Stage 5 | compose 5. 他のコマンド(未作成) | flow.md のコマンド別の流れ、down / build / watch / run のノート | |
| Stage 6 | compose 6. 開発と PR(未作成) | development.md、testing.md、contributing/ | |
| **合計** | | | **41** |

## 収録の方針

- **問題の材料は Q&A(`*-qa.md`)を最優先にする**。自分が実際に疑問に思ったことだから。そのあとに、ドキュメントの「注意」「落とし穴」「混乱しやすい点」の節と、[next-tasks.md](../tracking/next-tasks.md) で見つけたバグ
- **説明の問題は、紛らわしいものだけ**に絞る。単純な事実は穴埋めにする
- 答えは、**太字の 1 文**と 1〜2 行の理由。答えは出題範囲のドキュメントとコードに書かれていることだけで作り、推測で書かない
- **確かめ方** のコマンドは、実際に打って結果を確かめたものだけ載せる。打てないものは、読むべきコードの場所(`path:line`)にする。行番号はコードの変更で古くなりやすいので、ずれていたら直す
- Q&A を書いたら、同じ時に問題を 1〜2 問足す。ドキュメントやコードが変わって答えが古くなったら、問題も直す
- 計画と進み具合は [tracking/quiz-plan.md](../tracking/quiz-plan.md) にある
