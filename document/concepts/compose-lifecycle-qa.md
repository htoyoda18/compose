# compose-lifecycle.md を読んで出た疑問 Q&A

[compose-lifecycle.md](./compose-lifecycle.md) を読んでいて出た疑問への短い回答。手元の compose と、go.mod が指す compose-go（v2.16.2-0.20261015044941-5515644a24af）、moby の `toyo/study` ブランチのコードで確認した。
登場人物（compose-go と docker/compose、Engine）→ 作ったものをどう見分けるか（config hash、slug）→ 作り直すかの方針 → ネットワークの中身、の順に並べている。

← [document 一覧](../README.md)

## 1. compose-go と docker/compose の関係

- **compose-go は「compose.yaml を読んで、Go のデータにする」ライブラリ。docker/compose は「そのデータを見て、Docker を操作する」ツール**。docker/compose は compose-go を依存として import している（`go.mod:10`）。
  - compose-go の担当: YAML の読み込み、`${VAR}` の展開、複数の `-f` のマージ、`include` / `extends`、検証。結果が `types.Project`
  - docker/compose の担当: `types.Project` と Docker の今の状態を比べて、コンテナなどを作る・消す（plan engine）
- 別のリポジトリ（`compose-spec/compose-go`）なので、**読み込みのバグは compose-go に PR を出す必要がある**（[compose-lifecycle.md の 7](./compose-lifecycle.md)）。go.mod のバージョンが `v2.16.2-0.2026...` という形なのは、リリース前の commit を直接指しているから。
- 同じ compose-spec の実装なので、ほかのツール（Kubernetes への変換ツールなど）も compose-go を使える。両方のリポジトリのメンテナが同じ人たちかどうかは、コードでは確かめていない。

## 2. 「healthcheck は Engine にそのまま渡す」とは。plan engine のことか、dockerd のことか

- **Engine は dockerd（Docker Engine）のこと**。plan engine ではない。compose は healthcheck を自分では実行しない。
  - compose は compose.yaml の `healthcheck` を Engine API の形（`container.HealthConfig`）に変換するだけ（`pkg/compose/convert.go:42-85`）。それをコンテナを作るリクエストに入れて dockerd に送る（`pkg/compose/create.go:351`、`:380`）
  - その後、決まった間隔でチェックのコマンドを実行し、失敗の回数を数えて healthy / unhealthy を決めるのは dockerd（moby `daemon/health.go`。`monitor` が `:256`、失敗の連続回数 `FailingStreak` が `:229-231`、`retries` の既定 3 が `:41`）
- compose は、`depends_on: condition: service_healthy` や `--wait` のときに、コンテナを inspect して dockerd が決めた状態を**読むだけ**（`pkg/compose/service_containers.go:578-608`）。

## 3. config hash は、なぜわざわざハッシュにするのか。設定ファイルを文字列で比べてはだめか

- **ファイルの文字列で比べると、意味が同じでも違うと判定してしまう**。コメントや並び順を変えただけ、`${VAR}` の値が同じ、別のファイルに分けた（`-f` や `include`）などでも文字列は変わる。だから compose-go が読み込んで正規化した**サービスごとのデータ**を JSON にして、SHA-256 を取る（`pkg/compose/hash.go:169-181`）。
  - サービスごとに取るので、1 つのサービスを変えても、ほかのサービスは作り直されない
  - 作り直しに関係ない項目（`build`、`scale`、`depends_on` など）は、計算の前に外す（`trimServiceHashFields`）
- **前の設定はどこにも保存されていない**（compose は状態の DB を持たない）。比べる相手は、コンテナに付けたラベル `com.docker.compose.config-hash` だけ（`pkg/compose/create.go:312`）。設定全体をラベルに入れる代わりに、短い固定長のハッシュにしている。
  - 「ラベルの大きさを抑えるため」というのは私の解釈で、理由をはっきり書いたコメントは見つからなかった。
- ハッシュならではの注意点もある。compose-go の中の項目の並びが変わるだけでハッシュが変わり、全コンテナが作り直されてしまう。それを防ぐために、JSON のキーの順番を固定している（`hash.go:27-39` のコメント）。

## 4. diverged と never

- **作り直しの方針**で、`up` が既存のコンテナを見つけたときにどうするかを決める（`pkg/compose/reconcile.go:1025-1048` の `mustRecreate`）。
  - **diverged**（既定）: 設定がずれていたら作り直す。config hash、イメージの digest、つながっているネットワーク、ボリュームのマウントのどれかが違えば作り直す
  - **never**（`--no-recreate`）: **何が変わっていても作り直さない**。最初の行で `false` を返すので、ずれの確認すらしない。設定を変えても、古い設定のコンテナが動き続ける
  - **force**（`--force-recreate`）: 何も変わっていなくても作り直す
- 指定したサービスと、その依存先とで、別々の方針が付く（[compose-lifecycle.md の「フラグと方針」の表](./compose-lifecycle.md)）。

## 5. slug はどう割り当てられるのか。連番か

- **連番ではなく、ランダムな ID**。`docker compose run` のたびに、32 バイトの乱数を 16 進数にした 64 文字を作る（`pkg/compose/run.go:159`、moby の `client/pkg/stringid` の `GenerateRandomID`）。
  - 全体をラベル `com.docker.compose.slug` に入れ（`run.go:170`）、先頭 12 文字をコンテナ名に使う: `<project>-<service>-run-<12文字>`（`run.go:161`）
- 連番（ラベル `container-number`）は、`scale` で増やす普通のコンテナに使う。`run` のコンテナは、同時にいくつ動いても名前がぶつからないことだけが必要なので、ランダムな ID にしている（理由はコードのコメントには無く、私の解釈）。

## 6. Docker のネットワークを作るとは、内部で何をしているのか

Docker 全般の話なので docker-doc に置いた → [networking-qa.md の 6](../../../docker-doc/concepts/networking-qa.md)

compose の側でやっていることは、ネットワークの設定にラベル（プロジェクト名と `config-hash`）を付けて、dockerd に作成を頼むことだけ（`pkg/compose/create.go:1485-1492`）。
