# 用語辞典(Compose)

← [document 一覧](./README.md)

docker/compose のドキュメントとコードを読むときに出てくる、compose 固有の言葉。
1 語は 1〜3 行の短い定義だけにして、詳しい説明はリンク先に任せる。
TTY、moby、dockerd、namespace など Docker 全般の言葉は [docker-doc の用語辞典](../../docker-doc/glossary.md) にある。

分類の中は、[roadmap.md](./roadmap.md) の読む順に近い並びにしている。

---

## 名前と仕様

### Docker Compose v1 / v2

v1 は Python で書かれた旧版の `docker-compose`(ハイフン付き。サポート終了)。v2 は Go で書き直した今の版で、`docker compose` として動く。このリポジトリは v2。

→ [intro.md の「名前まわりの整理」](../../docker-doc/concepts/intro.md)

### プラグイン版 / スタンドアロン版

プラグイン版は、docker CLI から `docker compose` として呼ばれる形。スタンドアロン版は、`docker-compose` という単体のバイナリとして動く形。中身は同じバイナリで、スタンドアロンのときは引数をプラグインの形に並べ替えてから同じ処理を通す。

→ [codebase/cli.md](./codebase/cli.md)、[codebase/cli-qa.md](./codebase/cli-qa.md)

### compose-spec

`compose.yaml` に何を書けて、それぞれどういう意味かを決めた仕様。Compose はその実装の 1 つ。

→ [concepts/compose-concepts.md の「compose-spec と compose-go」](./concepts/compose-concepts.md#compose-spec-と-compose-go)

### compose-go

compose.yaml を読み込んで Go のデータ(`types.Project`)にするライブラリ。変数の展開、複数ファイルのマージ、検証もここが受け持つ。別のリポジトリなので、読み込みのバグは compose-go 側で直す。

→ [concepts/compose-lifecycle-qa.md の 1](./concepts/compose-lifecycle-qa.md)

---

## compose.yaml のモデル

### プロジェクト

1 つの compose.yaml から作られる、サービスとネットワーク・ボリュームなどのまとまり。`up` や `down` はプロジェクトの単位で動く。

→ [intro.md の「プロジェクトという単位」](../../docker-doc/concepts/intro.md)

### プロジェクト名

プロジェクトを見分ける名前。`-p` > `COMPOSE_PROJECT_NAME` > `name:` > ディレクトリ名(パスの最後の部分)の順で決まる。別の場所にある同じ名前のディレクトリは、同じプロジェクトとして扱われる。

→ [concepts/compose-concepts-qa.md の 2](./concepts/compose-concepts-qa.md)

### サービス

同じ設定で動くコンテナの型。既定では 1 サービスにつき 1 コンテナだが、scale で増やしたり、profile や provider で 0 個になったりする。

→ [concepts/compose-concepts-qa.md の 3](./concepts/compose-concepts-qa.md)

### レプリカ / scale

1 つのサービスから作る、同じ設定のコンテナの台数。`--scale` や `deploy.replicas` で決める。コンテナ名は `<プロジェクト>-<サービス>-<連番>`。

### profiles

指定したときだけ有効になるサービスのまとまり。有効でないサービスは、作られない。

### depends_on

サービス同士の依存関係。依存先を先に起動する。`condition: service_healthy` などで、依存先がどの状態になるまで待つかを決められる。

→ [concepts/compose-lifecycle.md の 1](./concepts/compose-lifecycle.md)

### healthcheck

コンテナが正常かどうかを確かめるコマンドの設定。compose は設定を dockerd に渡すだけで、実際に実行して healthy かどうかを決めるのは dockerd。

→ [concepts/compose-lifecycle-qa.md の 2](./concepts/compose-lifecycle-qa.md)

### job

1 回実行して終わる処理を、`services` とは別に `jobs:` として書く仕組み。`up` は起動せず、`docker compose run <job>` で起動する。

→ [cmd/compose/up.go-qa.md の 2](../cmd/compose/up.go-qa.md)

### provider

コンテナの代わりに、外部のコマンド(クラウドのサービスや Docker Model Runner など)にサービスを用意させる仕組み。

→ [concepts/compose-concepts-qa.md の 5](./concepts/compose-concepts-qa.md)

### リモートの compose ファイル

git リポジトリや `oci://` に置いた compose ファイルを、`-f` で直接指定して動かす機能。他人が書いた設定を実行するので、`up` の前に確認を挟む。

→ [cmd/compose/up.go-qa.md の 1](../cmd/compose/up.go-qa.md)

---

## 状態の管理(ラベル)

### ラベル(`com.docker.compose.*`)

compose がコンテナ・ネットワーク・ボリュームに付ける目印。compose は状態の DB を持たず、このラベルで「どれが自分のプロジェクトのものか」を見分ける。

→ [concepts/how-compose-works.md](./concepts/how-compose-works.md)、[pkg/api/labels.go.md](../pkg/api/labels.go.md)

### config hash

サービスの設定から計算したハッシュ。コンテナにラベルとして付け、次の `up` で今の設定のハッシュと比べて、作り直すかを決める。

→ [concepts/compose-lifecycle-qa.md の 3](./concepts/compose-lifecycle-qa.md)

### 作り直しの方針(diverged / never / force)

既存のコンテナを作り直すかの方針。diverged は設定がずれていたら作り直す(既定)、never は何が変わっても作り直さない(`--no-recreate`)、force は必ず作り直す(`--force-recreate`)。

→ [concepts/compose-lifecycle-qa.md の 4](./concepts/compose-lifecycle-qa.md)

### 匿名ボリュームの引き継ぎ(Inherit)

コンテナを作り直すとき、古いコンテナの匿名ボリュームを新しいコンテナに付け直すこと。`-V` を付けると引き継がない。

→ [concepts/compose-lifecycle.md の 5](./concepts/compose-lifecycle.md)

### orphan

compose.yaml から消したのに、ラベルが残っているために見つかるコンテナ。`--remove-orphans` で消す。

### oneoff

`docker compose run` で作る、1 回だけのコンテナ。ラベル `com.docker.compose.oneoff` で普通のコンテナと見分ける。

→ [concepts/compose-lifecycle.md の 10](./concepts/compose-lifecycle.md)

### slug

oneoff のコンテナを見分けるためのランダムな ID(64 文字)。先頭の 12 文字をコンテナ名に使う。連番ではない。

→ [concepts/compose-lifecycle-qa.md の 5](./concepts/compose-lifecycle-qa.md)

---

## 中の仕組み

### 望ましい状態 / 実際の状態

compose.yaml に書かれた「こうあってほしい」姿と、Docker の上に今ある姿。compose の中心の仕事は、この 2 つの差を埋めること。

→ [concepts/compose-concepts.md](./concepts/compose-concepts.md)

### plan engine

今の `up` などの中身。観測(ObservedState)→ 計画(reconcile が操作の DAG を作る)→ 実行(executor)の 3 段で、差を埋める。2026 年に convergence engine から置き換わった。

→ [codebase/plan-engine.md](./codebase/plan-engine.md)

### ObservedState

ラベルで集めた、Docker の上の今の状態のスナップショット。plan engine の観測の結果。

→ [codebase/plan-engine.md の 1](./codebase/plan-engine.md)

### reconcile

望ましい状態と ObservedState を比べて、必要な操作の一覧(Plan)を作る計算。Docker の API は一切呼ばない。

→ [codebase/plan-engine.md の 2](./codebase/plan-engine.md)

### Plan / DAG

Plan は、Docker の API の呼び出し 1 回を 1 つのノードにした、操作のグラフ。DAG は「向きがあって、輪の無いグラフ」のことで、依存の順番を表せる。

→ [codebase/plan-engine.md の 3](./codebase/plan-engine.md)、[codebase/algorithms.md](./codebase/algorithms.md)

### トポロジカル順序

依存関係のあるものを、依存先が必ず先に来るように並べた順番。起動は依存先から、停止はその逆。

→ [codebase/algorithms.md の 3](./codebase/algorithms.md)

### convergence engine

plan engine の前の、差を埋める仕組み(`convergence.go`)。今はもう無い。古いドキュメントや issue に出てくる。

---

## コード

### cobra

Go の CLI のフレームワーク。`up` や `down` などのサブコマンドは、すべて `cobra.Command` として定義されている。

→ [codebase/cli.md](./codebase/cli.md)

### PersistentPreRunE

cobra の、どのサブコマンドの前にも必ず走るフック。compose はここで docker CLI の初期化とトレースの準備をする。

→ [cmd/main.go.md](../cmd/main.go.md)

### api.Compose

compose の公開 API のインターフェース(`pkg/api`)。CLI も SDK として使う人も、これを通して compose を操作する。用途ごとの 6 つの小さなインターフェースを束ねたもの。

→ [pkg/api/note.md](../pkg/api/note.md)、[pkg/api/api.go.md](../pkg/api/api.go.md)

### composeService

`api.Compose` の実装の本体(`pkg/compose`)。`compose.NewComposeService` で作る。

→ [codebase/flow-qa.md の 1](./codebase/flow-qa.md)

### XxxOptions

`api.Compose` の各メソッドに渡すオプションの構造体(`UpOptions`、`DownOptions` など)。フラグとほぼ 1 対 1。中の `Services` フィールドは対象を絞り込むためのものではないので注意。

→ [pkg/api/api.go.md の「読むときに大事な約束」](../pkg/api/api.go.md)

### EventProcessor

進捗のイベントを受け取るインターフェース。`pkg/compose` がイベントを出し、`cmd/display` が画面に描く。

→ [pkg/api/note.md](../pkg/api/note.md)

### 進捗表示のモード(TTY / Plain / JSON / Quiet)

`--progress` か、出力先が端末かどうかで決まる表示の仕方。TTY は同じ行を書き換えて描き、Plain は 1 行ずつ追記し、JSON は機械が読む形で出し、Quiet は何も出さない。

→ [cmd/note.md の用語](../cmd/note.md)

### dry-run

`--dry-run` を付けると、Docker に何も送らずに、何をするかだけを表示する。Docker の API を真似る `DryRunClient` で実現している。

→ [pkg/note.md](../pkg/note.md)

---

## 開発と PR

### Scenario DSL

新しい e2e テストを書くときの決まった書き方(`NewScenario`)。compose のモデルと、「コマンド → 期待する結果」の手順を並べて書く。

→ [pkg/e2e/SCENARIO.md](../pkg/e2e/SCENARIO.md)、[codebase/testing.md](./codebase/testing.md)

### DCO

Developer Certificate of Origin。コミットに `Signed-off-by` を付けて、「自分が書いたものを出す権利がある」と宣言すること。`git commit -s` で付く。compose では必須。

→ [contributing/README.md](./contributing/README.md)

### AI_POLICY

AI を使ったコントリビューションのルール。まだ報告されていない問題については、先に issue でメンテナに必要性を確かめてもらう(No contribution farming)。全行を自分で説明できること。

→ [contributing/pr-trends.md の 8](./contributing/pr-trends.md)
