# 問題集 compose 3. 概念と意味論

出題範囲：[compose-concepts.md](../concepts/compose-concepts.md)、[compose-concepts-qa.md](../concepts/compose-concepts-qa.md)、[compose-lifecycle.md](../concepts/compose-lifecycle.md)、[compose-lifecycle-qa.md](../concepts/compose-lifecycle-qa.md) ｜ [問題集トップ](README.md) ｜ 前：[compose 2. コマンドが本体に届くまで](02-cli-to-backend.md) ｜ 次：compose 4. plan engine(未作成)

答えは折りたたんでいる。**先に自分の言葉で答えてから**開く。

確かめ方のコマンドは、手元の `docker compose`(5.0.1)で打った。`image: alpine` と `command: sleep infinity` の短い compose.yaml で試せる。

---

## compose.yaml の読み込み

### Q1 [場所]

次の 3 つのバグを直すには、docker/compose と compose-go のどちらに PR を出すか。

1. `${VAR}` の展開がおかしい
2. `-f` で重ねたときの `!reset` が効かない
3. 設定を変えたのに、`up` でコンテナが作り直されない

<details><summary>答え</summary>

**1 と 2 は compose-go、3 は docker/compose。**

compose-go は「compose.yaml を読んで、Go のデータ(`types.Project`)にする」ライブラリで、YAML の読み込み、`${VAR}` の展開、複数の `-f` のマージ、`include` / `extends`、検証を受け持つ。docker/compose は「そのデータと Docker の今の状態を比べて、コンテナなどを作る・消す」ツール。別のリポジトリなので、読み込みのバグは compose-go に PR を出す。

**確かめ方**

- `go.mod:10`: docker/compose は compose-go を依存として import している
- 3 の判定は `pkg/compose/reconcile.go:1025-1048`(`mustRecreate`)

参照：[compose-lifecycle-qa.md > 1. compose-go と docker/compose の関係](../concepts/compose-lifecycle-qa.md#1-compose-go-と-dockercompose-の関係)

</details>

### Q2 [予想]

次の compose.yaml で `docker compose config` を打つ。サービス `a` と `b` の `environment` は、それぞれどうなるか。`x-common` はサービスになるか。

```yaml
x-common: &common
  image: alpine
  environment:
    TZ: Asia/Tokyo
services:
  a:
    <<: *common
    command: echo a
  b:
    <<: *common
    environment:
      LANG: C
```

<details><summary>答え</summary>

**`a` は `TZ: Asia/Tokyo`、`b` は `LANG: C` だけ(`TZ` は消える)。`x-common` はサービスにならない。**

`&common` で印を付けた map を、`<<: *common` で混ぜ込む。混ぜ込みはキーの単位なので、`b` が自分で書いた `environment` が丸ごと勝ち、中身は足し合わされない。`x-` で始まるキーは compose が無視するので、使い回す設定の置き場になる。

**確かめ方**

- 上の compose.yaml で `docker compose config` を打つと、`b` の `environment` は `LANG: C` だけ。`x-common` は出力の最後にそのまま残るが、`services` には入らない

参照：[compose-concepts-qa.md](../concepts/compose-concepts-qa.md) の「1. 「YAML 自体の挙動」とは何か」

</details>

### Q3 [穴埋め]

プロジェクト名は、`-p` > ( ) > ( ) > 作業ディレクトリ名、の順に決まる。

<details><summary>答え</summary>

**環境変数 `COMPOSE_PROJECT_NAME`(`.env` に書いたものを含む)/ compose.yaml のトップレベルの `name:`**

作業ディレクトリは、`--project-directory` > 最初の `-f` のディレクトリ > カレントディレクトリ の順に決まる。

**確かめ方**

- `name: fromyaml` を書いた compose.yaml で `docker compose config | head -1` → `name: fromyaml`。`COMPOSE_PROJECT_NAME=fromenv` を付けると `fromenv`、さらに `-p fromflag` を付けると `fromflag`
- compose-go の `cli/options.go:637-655`(`withNamePrecedenceLoad`)

参照：[compose-lifecycle.md > プロジェクト名の決まり方](../concepts/compose-lifecycle.md#プロジェクト名の決まり方)

</details>

### Q4 [予想]

`name:` を書いていない compose.yaml を、`/home/me/work/My App/` に置いて `up` した。プロジェクト名は何になるか。別の場所の `/home/me/other/My App/` で同じように `up` すると、同じプロジェクトになるか。

<details><summary>答え</summary>

**どちらも `myapp` になり、同じプロジェクトとして扱われる。**

使うのはパスの最後の要素(ディレクトリ名)だけで、小文字にして `a-z 0-9 _ -` 以外の文字を消す。パスは別のラベル `com.docker.compose.project.working_dir` に記録されるだけで、名前には使われない。`down -v` が他のプロジェクトのボリュームを消すバグ(next-tasks.md の N2)は、これが原因だった。

**確かめ方**

- `My App` というディレクトリで `docker compose config | head -1` → `name: myapp`
- compose-go の `loader/loader.go:799-804`(`NormalizeProjectName`)

参照：[compose-concepts-qa.md](../concepts/compose-concepts-qa.md) の「2. プロジェクト名は、ディレクトリのパスではなくディレクトリ名か」

</details>

---

## サービスとコンテナの数

### Q5 [予想]

プロジェクト `scale` のサービス `a` を、`docker compose up -d --scale a=3` で起動する。コンテナの名前は何になるか。`a` に `container_name: fixed` を書いて、`--scale a=2` で起動するとどうなるか。

<details><summary>答え</summary>

**`scale-a-1`、`scale-a-2`、`scale-a-3` の 3 つになる。`container_name` があると、2 台以上にできずにエラーになる。**

サービスは「同じ設定で動くコンテナの型」で、コンテナの数は `--scale` や `deploy.replicas` で変わる。名前は `<プロジェクト>-<サービス>-<連番>`。`container_name` を指定すると全部が同じ名前になってぶつかるので、台数が 2 以上なら作る前に断る。どのコンテナが何台目かは、ラベル `container-number` で見分ける。

**確かめ方**

- `docker compose up -d --scale a=3` のあと、`docker compose ps --format '{{.Name}}'` に 3 つ並ぶ
- `container_name: fixed` を足して `--scale a=2` → `WARNING: The "a" service is using the custom container name "fixed". Docker requires each container to have a unique name. Remove the custom name to scale the service`(終了コード 1)
- `pkg/compose/service_containers.go:55-66`(`getScale`)

参照：[compose-concepts-qa.md > 3. 1 サービス = 1 コンテナという認識で合っているか](../concepts/compose-concepts-qa.md#3-1-サービス--1-コンテナという認識で合っているか)

</details>

### Q6 [穴埋め]

`docker compose run` で作るコンテナの名前は `<project>-<service>-run-<( )>`。ラベル `oneoff` は ( ) で、識別子のラベル `slug` には、連番ではなく ( ) が入る。レプリカの番号のラベル `container-number` は ( )。

<details><summary>答え</summary>

**slug の先頭 12 文字 / `True` / 32 バイトの乱数を 16 進数にした 64 文字のランダムな ID / 付かない**

`run` のコンテナは、同時にいくつ動いても名前がぶつからないことだけが要る。連番(`container-number`)は、`scale` で増やす普通のコンテナに使う。`run` のコンテナは、普通の `docker compose ps` には出ない。

**確かめ方**

- `docker compose run -d a sleep 60` のあと、そのコンテナの名前は `scale-a-run-cc64b2178261`、ラベル `slug` は `cc64b2178261...`(64 文字)、`container-number` は空。最後に `docker rm -f` で消す
- `pkg/compose/run.go:159-170`

参照：[compose-lifecycle-qa.md](../concepts/compose-lifecycle-qa.md) の「5. slug はどう割り当てられるのか。連番か」

</details>

---

## depends_on と healthcheck

### Q7 [穴埋め]

`depends_on` の `condition` は 3 つ。既定の ( ) は依存先が起動したことを、( ) は依存先が healthy になることを、( ) は依存先が exit 0 で終わることを待つ。待つときは、( ) ミリ秒ごとに `ContainerInspect` を呼んで状態を見る。

<details><summary>答え</summary>

**`service_started` / `service_healthy` / `service_completed_successfully` / 500**

`service_started` は待ちのノードを作らず、依存先の起動への辺だけになる。待ち方はイベントではなく inspect の繰り返しで、待っている間は `--parallel` の枠を使わない。

**確かめ方**

- `pkg/compose/service_containers.go:215-216`(`waitDependency` の ticker)

参照：[compose-lifecycle.md > 1. depends_on](../concepts/compose-lifecycle.md#1-depends_on)

</details>

### Q8 [説明]

compose.yaml に `healthcheck` を書いた。決まった間隔でチェックのコマンドを実行し、失敗の回数を数えて healthy / unhealthy を決めているのは、Compose(plan engine)か、Docker Engine(dockerd)か。

<details><summary>答え</summary>

**Docker Engine(dockerd)。Compose は healthcheck を Engine API の形に変えて渡し、決まった状態を inspect で読むだけ。**

`start_period` や `retries` を数えるのも Engine。Compose が状態を読むのは、`depends_on` の `service_healthy` と `--wait` のとき。

**確かめ方**

- `pkg/compose/convert.go:43-85`(`ToMobyHealthCheck`): `container.HealthConfig` に変えるだけ
- `pkg/compose/service_containers.go:578-612`(`isServiceHealthy`): inspect した `State.Health.Status` を見るだけ

参照：[compose-lifecycle-qa.md](../concepts/compose-lifecycle-qa.md) の「2. 「healthcheck は Engine にそのまま渡す」とは。plan engine のことか、dockerd のことか」

</details>

### Q9 [予想]

`db` の healthcheck は必ず失敗する(`test: ["CMD", "false"]`、`interval: 1s`、`retries: 1`)。`web` は `depends_on: db: condition: service_healthy`。`docker compose up -d` はどうなるか。`web` はどうなるか。

<details><summary>答え</summary>

**`db` が unhealthy になった時点ですぐにエラーで終わる(終了コード 1)。`web` は作られるが、起動されない。**

Compose は unhealthy を見たら、healthy に戻るのを待たずにエラーにする。`starting` の間(`start_period` を含む)は待ち続ける。`depends_on` の待ちには、既定のタイムアウトが無い。

**確かめ方**

- 打つと、2 秒ほどで `dependency failed to start: container health-db-1 is unhealthy`。そのあと `docker compose ps -a` で、`db` は running / unhealthy、`web` は created

参照：[compose-lifecycle.md > 2. healthcheck と --wait](../concepts/compose-lifecycle.md#2-healthcheck-と---wait)

</details>

### Q10 [予想]

`migrate`(`command: "true"` ですぐ exit 0 で終わる)と `app`(`sleep infinity`)の 2 つのサービスがあり、どちらも `depends_on` を書いていない。`docker compose up --wait` は成功するか。成功させるには、compose.yaml をどう変えるか。

<details><summary>答え</summary>

**失敗する(`exited (0)`)。`app` に `depends_on: migrate: condition: service_completed_successfully` を書けば成功する。**

`--wait` は自動的に detached になり、plan が終わってから全サービスを見て、running か healthy になるまで待つ。exit 0 で終わる使い捨てのサービスでも、ほかのサービスから `service_completed_successfully` で参照されていなければ、「exited」で失敗になる。

**確かめ方**

- そのままだと `container wait-migrate-1 exited (0)` で終了コード 1。`depends_on` を足すと `Container wait-migrate-1 Exited` → `Container wait-app-1 Healthy` で終了コード 0

参照：[compose-lifecycle.md > 2. healthcheck と --wait](../concepts/compose-lifecycle.md#2-healthcheck-と---wait) の「落とし穴」

</details>

---

## 作り直しの判定

### Q11 [説明]

`up` が「設定が変わったか」を見るとき、compose.yaml の文字列を比べずに、サービスごとの config hash を比べるのはなぜか。

<details><summary>答え</summary>

**文字列で比べると、意味が同じでも違うと判定してしまうから。それに、前の設定はどこにも保存されていない。**

コメントや並び順を変えただけ、`${VAR}` の値が同じ、別のファイルに分けた、などでも文字列は変わる。だから compose-go が読み込んで正規化したサービスごとのデータを JSON にして SHA-256 を取る。比べる相手は、コンテナに付けたラベル `config-hash` だけ。サービスごとに取るので、1 つのサービスを変えても、ほかのサービスは作り直されない。

**確かめ方**

- `pkg/compose/hash.go:169-181`(`ServiceHash`)、`pkg/compose/create.go:312`(ラベルに付ける)

参照：[compose-lifecycle-qa.md](../concepts/compose-lifecycle-qa.md) の「3. config hash は、なぜわざわざハッシュにするのか。設定ファイルを文字列で比べてはだめか」

</details>

### Q12 [予想]

`up -d` で動いているサービス `a` に、(1) `depends_on: [b]` を足して `up -d` を打つ。次に (2) `labels: [team=x]` を足して `up -d` を打つ。それぞれ `a` は作り直されるか。

<details><summary>答え</summary>

**(1) は作り直されない。(2) は作り直される。**

config hash を計算する前に、作り直しに関係ない項目(`build`、`pull_policy`、`scale`、`deploy.replicas`、`depends_on`、`profiles`)を外す。ユーザーが書いた `labels` は入る。そのため (1) のあとも、ラベル `depends_on` は古いまま残る。

**確かめ方**

- (1) のあとは `Container recreate-a-1 Running` で、ラベル `com.docker.compose.depends_on` は空のまま。(2) で `Recreate` → `Recreated` が出て、ラベルが `b:service_started:false` になる
- `pkg/compose/hash.go:154-166`(`trimServiceHashFields`)

参照：[compose-lifecycle.md](../concepts/compose-lifecycle.md) の「5. 作り直し(recreate)の判定」>「config hash に何が入るか」

</details>

### Q13 [穴埋め]

作り直しの方針は、指定したサービスと依存先のサービスで別々に決まる。フラグなしなら両方 `diverged`(違えば作り直す)。`--no-recreate` なら指定したサービスは ( )、依存先は ( )。`--force-recreate` なら ( ) と ( )。`--always-recreate-deps` なら ( ) と ( )。`-V`(`--renew-anon-volumes`)なら、指定したサービスは ( )。

<details><summary>答え</summary>

**never / never / force / diverged / diverged / force / force**

`-V` は匿名ボリュームを新しくするので、必ず作り直す。`never` は最初に `false` を返すので、ずれの確認すらしない(設定を変えても、古い設定のコンテナが動き続ける)。

**確かめ方**

- `cmd/compose/create.go:150-171`(`recreateStrategy` と `dependenciesRecreateStrategy`)

参照：[compose-lifecycle-qa.md > 4. diverged と never](../concepts/compose-lifecycle-qa.md#4-diverged-と-never)、[compose-lifecycle.md](../concepts/compose-lifecycle.md) の「フラグと方針」

</details>

### Q14 [予想]

動いているサービス `a` の `environment` を `MODE: one` から `MODE: two` に変えた。次の順に打つ。それぞれのあとで、コンテナの中の `MODE` は何か。

1. `docker compose restart a`
2. `docker compose up -d --no-recreate`
3. `docker compose up -d`

<details><summary>答え</summary>

**1 は `one`、2 も `one`、3 で初めて `two` になる。**

`docker compose restart` は `ContainerRestart` を呼ぶだけで、作り直さない。設定の変更は反映されない。`--no-recreate` は方針が `never` なので、何が変わっていても作り直さない。方針が既定の `diverged` の `up` で、config hash の違いを見て作り直す。

**確かめ方**

- それぞれのあとで `docker compose exec a printenv MODE` を打つ。3 では `Container recreate-a-1 Recreate` が出る
- `pkg/compose/restart.go:125`(`ContainerRestart`)

参照：[compose-lifecycle.md > 3. restart ポリシー](../concepts/compose-lifecycle.md#3-restart-ポリシー)、[compose-lifecycle-qa.md > 4. diverged と never](../concepts/compose-lifecycle-qa.md#4-diverged-と-never)

</details>

---

## 環境変数

### Q15 [予想]

compose.yaml に `X: ${VAR:-unset}` と `Y: ${OTHER:-unset}` を書いた。`.env` には `VAR=from-dotenv` と `OTHER=dotenv-other`、`other.env` には `VAR=from-envfile` だけがある。次の 3 つで、`X` と `Y` はどうなるか。

1. `docker compose config`
2. `VAR=from-shell docker compose config`
3. `docker compose --env-file other.env config`

<details><summary>答え</summary>

**1 は `from-dotenv` / `dotenv-other`。2 は `from-shell` / `dotenv-other`。3 は `from-envfile` / `unset`。**

`${VAR}` の補間は、シェルの環境変数 > `--env-file` / `.env` の順。`--env-file` を指定すると `.env` は読まないので、3 では `OTHER` がどこにも無く、既定値の `unset` になる。

**確かめ方**

- 上の 3 つを打って、`environment` の `X` と `Y` を見る

参照：[compose-lifecycle.md > 環境変数の優先順位](../concepts/compose-lifecycle.md#環境変数の優先順位)

</details>
