# 問題集 compose 1. 全体像

出題範囲：[how-compose-works.md](../concepts/how-compose-works.md)、[overview.md](../codebase/overview.md)、[architecture.md](../codebase/architecture.md)、[ルートの note.md](../../note.md) ｜ [問題集トップ](README.md) ｜ 前：[docker-doc 1. 入門](../../../docker-doc/quiz/01-intro.md) ｜ 次：docker-doc 6. Go(未作成)

答えは折りたたんでいる。**先に自分の言葉で答えてから**開く。

Q3〜Q7 は、how-compose-works.md の例(ディレクトリ `myapp/` の compose.yaml。`web` が `db` に依存し、`db` が名前付きボリューム `db-data` を使う)を前提にする。

---

## Compose は自分で状態を持たない

### Q1 [説明]

Compose は「前回何を作ったか」を記録するファイルやデータベースを持たない。それでも、2 回目の `up` で既存のコンテナを見つけられるのはなぜか。

<details><summary>答え</summary>

**作ったリソースすべてにラベルを貼り、Docker Engine 自体を状態の保管場所にしているから。**

次に実行したときは、ラベル `com.docker.compose.project=<プロジェクト名>` で Docker Engine を検索して、自分が管理しているものを見つける。だから CLI を何度起動し直しても、別の端末から実行しても、ラベルさえ読めれば現状が分かる。

**確かめ方**

- `docker ps --filter label=com.docker.compose.project=myapp --format '{{.Names}}'` で、そのプロジェクトのコンテナだけが出る
- `pkg/compose/filters.go:33-35`(`projectFilter`): コンテナなどを探すときの絞り込みが、このラベルでできている

参照：[how-compose-works.md > 2. Compose は自分で状態を持たない](../concepts/how-compose-works.md#2-compose-は自分で状態を持たない)

</details>

### Q2 [穴埋め]

Compose がコンテナに貼るラベル。どのプロジェクトのものか → `com.docker.compose.( )`。どのサービスのものか → `com.docker.compose.( )`。そのサービスの何台目か → `com.docker.compose.( )`。どの設定で作ったか(設定から計算したハッシュ)→ `com.docker.compose.( )`。

<details><summary>答え</summary>

**project / service / container-number / config-hash**

ネットワークとボリュームにも `project` のラベルが付き、加えて `network` / `volume` のラベルが付く。

**確かめ方**

- `docker inspect myapp-web-1 --format '{{range $k,$v := .Config.Labels}}{{$k}}={{$v}}{{"\n"}}{{end}}'` で、`com.docker.compose.` で始まるラベルが並ぶ
- `pkg/api/labels.go:26-37`: ラベルの定数

参照：[how-compose-works.md > 3. 例で追う](../concepts/how-compose-works.md#3-例で追う) の場面 1

</details>

### Q3 [予想]

`up -d` のあと、手で `docker rename myapp-web-1 renamed-by-hand` を打った。続けて `docker compose ps` と `docker compose up -d` を打つ。`web` のコンテナは見つかるか。作り直されるか。名前は `myapp-web-1` に戻るか。

<details><summary>答え</summary>

**見つかる。作り直されず、名前も `renamed-by-hand` のまま。**

Compose が頼りにするのは名前ではなくラベル。コンテナ名 `myapp-web-1` はラベルから組み立てたものにすぎない。名前を変えてもラベルは残るので、`web` の 1 台目として見つかり、config hash も変わっていないので何もしない。

**確かめ方**

- 上の順に打つと、`docker compose ps` に `renamed-by-hand` が `web` として出て、`up -d` は `Container renamed-by-hand Running` を出すだけ

参照：[how-compose-works.md > 3. 例で追う](../concepts/how-compose-works.md#3-例で追う) の場面 1(最後の段落)

</details>

---

## 例で追う: up と down

### Q4 [予想]

初めて `docker compose up -d` を打つ。作られるネットワーク・ボリューム・コンテナの名前と、作る順を答える。

<details><summary>答え</summary>

**ネットワーク `myapp_default` とボリューム `myapp_db-data` → コンテナ `myapp-db-1` → `myapp-web-1` の順に作る。**

プロジェクト名はディレクトリ名から `myapp`。ネットワークとボリュームは `<プロジェクト>_<名前>`、コンテナは `<プロジェクト>-<サービス>-<連番>` で、区切りの記号が違う。ラベルで探しても何も見つからないので全部が作る対象になり、依存関係の順(`db` が先、`db` に依存する `web` が後)に作る。

**確かめ方**

- 例の `image` を `alpine` に、`command` を `sleep infinity` にして `up -d` を打つと、`Network myapp_default Created` → `Volume myapp_db-data Created` → `Container myapp-db-1 Created` → `Container myapp-web-1 Created` の順に出る

参照：[how-compose-works.md > 3. 例で追う](../concepts/how-compose-works.md#3-例で追う) の場面 1

</details>

### Q5 [予想]

`web` の `environment` の `LOG_LEVEL` を `info` から `debug` に変えて、もう一度 `up -d` を打つ。`web` と `db` はそれぞれどうなるか。何を比べてそう決めるか。

<details><summary>答え</summary>

**`web` だけが作り直され、`db` には触れない。サービスごとに、今の compose.yaml から計算した config hash とコンテナのラベル `config-hash` を比べて決める。**

`db` は設定が変わっていないのでハッシュが一致する。`web` は `environment` が変わったのでハッシュが違う。正しく動いているものに触れないので、`up` は何度打っても安全。

**確かめ方**

- 打つと `Container myapp-db-1 Running` と `Container myapp-web-1 Recreate` → `Recreated` が出る。`db` のコンテナ ID は変わらない
- `pkg/compose/reconcile.go:1035`(`mustRecreate`): `oc.ConfigHash != expectedHash` なら作り直す。ほかにイメージ、ネットワーク、ボリュームのずれも見る

参照：[how-compose-works.md > 3. 例で追う](../concepts/how-compose-works.md#3-例で追う) の場面 2

</details>

### Q6 [穴埋め]

Q5 で `web` を作り直す手順。新しいコンテナを ( ) で作る → 古いコンテナを ( ) → 古いコンテナを ( ) → 新しいコンテナを ( )。新しいものを作れなかったときは、( ) がそのまま残る。

<details><summary>答え</summary>

**一時的な名前 / 止める / 消す / 正式な名前に変える / 古いコンテナ**

先に作ってから古いものを消すので、作り直しに失敗しても動いていたものは失われない。一時的な名前は `<古いコンテナの ID の先頭 12 文字>_<正式な名前>`。

**確かめ方**

- Q5 の `up -d` の間 `docker events --filter label=com.docker.compose.project=myapp --filter type=container` を流すと、`create 05a2930a8575_myapp-web-1` → `stop myapp-web-1` → `destroy myapp-web-1` → `rename myapp-web-1` → `start myapp-web-1` の順に出る(`kill` などは省いた)
- `pkg/compose/reconcile.go:1615-1620`(`planRecreateContainer`): この 4 つの操作に分けて計画する

参照：[how-compose-works.md > 3. 例で追う](../concepts/how-compose-works.md#3-例で追う) の場面 2

</details>

### Q7 [予想]

compose.yaml の無い、別のディレクトリで `docker compose -p myapp down` を打つ。片付けられるか。片付けられるなら、何がどの順で消え、何が残るか。

<details><summary>答え</summary>

**片付けられる。`myapp-web-1` → `myapp-db-1` → ネットワーク `myapp_default` の順に消え、ボリューム `myapp_db-data` は残る。**

`down` も compose.yaml の中身ではなく、ラベル `project=myapp` で消す対象を探す。止めて消す順は依存の逆順。ボリュームはデータを守るため、`-v` を付けたときだけ消える。

**確かめ方**

- 空のディレクトリで `docker compose -p myapp down` を打つと、`web` → `db` → ネットワークの順に `Removed` が出る。そのあと `docker volume ls --filter label=com.docker.compose.project=myapp` に `myapp_db-data` が残る
- `pkg/compose/down.go:74-80`: `-f` などでプロジェクトを渡されていなければ、ラベルで見つけたコンテナからプロジェクトを組み立て直す

参照：[how-compose-works.md > 3. 例で追う](../concepts/how-compose-works.md#3-例で追う) の場面 3

</details>

---

## リポジトリの構成

### Q8 [場所]

次の 3 つは、それぞれどのディレクトリにあるか。(a) cobra のコマンド定義と、フラグを Options に詰め替える処理。(b) Compose の操作を表すインターフェースの定義。(c) 収束処理(差分埋め)と Docker Engine API の呼び出し。

<details><summary>答え</summary>

**(a) `cmd/compose/`、(b) `pkg/api/`、(c) `pkg/compose/`。**

呼び出しは `cmd/compose` → `pkg/api` のインターフェース → 実体の `pkg/compose` → Docker Engine API と流れる。`cmd` はフラグを詰め替えて `pkg/compose` を呼ぶ薄いラッパー。

**確かめ方**

- `cmd/compose/up.go:342`: `up` のフラグを `api.UpOptions` に詰めて `backend.Up` を呼ぶ
- `pkg/api/api.go:102`: インターフェース `Compose` の定義

参照：[architecture.md](../codebase/architecture.md) の「レイヤー構成(cmd / pkg/api / pkg/compose / internal)」

</details>

### Q9 [説明]

`cmd` から `pkg/compose` を直接呼ばずに、間に `pkg/api` のインターフェースを挟んでいるのはなぜか。

<details><summary>答え</summary>

**インターフェースを CLI 層と実装層の境界にすると、テストでモックに差し替えられ、外部の Go プログラムからも SDK として同じ形で使えるから。**

テストではこのインターフェースをモックにして、Docker Engine につながずにロジックを確かめる。SDK として使う側も、実装の細部を意識せずに Compose を組み込める。

**確かめ方**

- `pkg/mocks/mock_docker_compose_api.go`: `pkg/api/api.go` から mockgen で生成したモック(`Makefile:111`)

参照：[architecture.md > インターフェース駆動設計と SDK としての利用](../codebase/architecture.md#インターフェース駆動設計と-sdk-としての利用)

</details>

### Q10 [場所]

外部の Go プログラムから Compose を SDK として使う。最初に呼ぶ関数はどのパッケージの何か。返ってくるのは、どのパッケージの何というインターフェースか。

<details><summary>答え</summary>

**`pkg/compose` の `NewComposeService` を呼び、`pkg/api` の `Compose` インターフェースが返る。**

返ってきた値の `LoadProject` で compose.yaml を読み、`Up` などを呼ぶ。architecture.md は境界のインターフェースを「`Service`」と書いているが、コードでの名前は `Compose`。

**確かめ方**

- `pkg/compose/compose.go:69`: `func NewComposeService(dockerCli command.Cli, options ...Option) (api.Compose, error)`
- `pkg/api/api.go:102`: `type Compose interface`

参照：[architecture.md > インターフェース駆動設計と SDK としての利用](../codebase/architecture.md#インターフェース駆動設計と-sdk-としての利用)

</details>

### Q11 [穴埋め]

外部から SDK として使える公開のパッケージは ( ) の下に、このリポジトリ専用の非公開のパッケージ(Docker Desktop 連携、トレーシング、OCI レジストリ操作など)は ( ) の下に置く。

<details><summary>答え</summary>

**`pkg/` / `internal/`**

`pkg` には公開 API(`pkg/api`)とその実装(`pkg/compose`)のほか、watch・bridge・remote・dry-run・e2e テストなどがある。

参照：[architecture.md > ディレクトリ構造](../codebase/architecture.md#ディレクトリ構造)、[note.md > ディレクトリ](../../note.md#ディレクトリ)

</details>
