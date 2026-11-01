# 問題集 compose 2. コマンドが本体に届くまで

出題範囲：[flow.md の共通の流れ](../codebase/flow.md#共通の流れ)、[flow-qa.md](../codebase/flow-qa.md)、[cli.md](../codebase/cli.md)、[cli-qa.md](../codebase/cli-qa.md)、[cmd/note.md](../../cmd/note.md)、[cmd/note-qa.md](../../cmd/note-qa.md)、[cmd/main.go.md](../../cmd/main.go.md)、[cmd/compose/note.md](../../cmd/compose/note.md)、[cmd/compose/note-qa.md](../../cmd/compose/note-qa.md)、[cmd/compose/up.go.md](../../cmd/compose/up.go.md)、[pkg/note.md](../../pkg/note.md)、[pkg/api/note.md](../../pkg/api/note.md)、[pkg/api/api.go.md](../../pkg/api/api.go.md) ｜ [問題集トップ](README.md) ｜ 前：docker-doc 6. Go(未作成) ｜ 次：compose 3. 概念と意味論(未作成)

答えは折りたたんでいる。**先に自分の言葉で答えてから**開く。

---

## 共通の流れ

### Q1 [穴埋め]

`docker compose up` は、`cmd/main.go` → `cmd/compose/up.go` → ( ) のメソッド → ( ) と降りていく。実行中の進捗は ( ) が描き、`ps` などの一覧の出力は ( ) が整形する。

<details><summary>答え</summary>

**`pkg/api.Compose` / `pkg/compose/up.go` / `cmd/display` / `cmd/formatter`**

`cmd/compose/up.go` は、フラグを Options に詰めて、`pkg/api` のインターフェース越しに `backend.Up(...)` を呼ぶ。本体の `pkg/compose` は、必要に応じて `internal` の部品を使う。

参照：[flow.md > 共通の流れ](../codebase/flow.md#共通の流れ)

</details>

### Q2 [予想]

`compose.NewComposeService(dockerCli)` を呼んだ時点で、Docker Engine に接続したり、コンテナを調べたりするか。

<details><summary>答え</summary>

**しない。`composeService` という構造体を 1 つ作って、`api.Compose` として返すだけ。**

中身は「これからの処理で使う道具の詰め合わせ」。docker/cli が初期化済みの `dockerCli`、並列数の上限(既定は無制限の `-1`)、レジストリ認証の取り方を持たせる。進捗表示の送り先やプロンプトが渡されなければ、何もしない既定値を入れる。だから SDK として使うときも、`dockerCli` だけ渡せば動く。

**確かめ方**

- `pkg/compose/compose.go:69-102`: 構造体を作り、オプションを当て、既定値を埋めて返すだけ。Engine API の呼び出しは無い

参照：[flow-qa.md](../codebase/flow-qa.md) の「1. `compose.NewComposeService` でバックエンドを作るとは、具体的に何をしているのか」

</details>

### Q3 [予想]

compose の外に Go のモジュールを作り、`github.com/docker/compose/v5/pkg/api` と `github.com/docker/compose/v5/internal/oci` を import してビルドする。どうなるか。

<details><summary>答え</summary>

**`internal/oci` の import でコンパイルエラーになる。`pkg/api` だけなら通る。**

`internal/` の下のパッケージは、その親ディレクトリより下のコードからしか import できない。これは Go 言語そのものの決まりで、コンパイラが禁止する。公開すると互換性を守る約束が生まれるので、外に使わせる理由がないものは `internal` に置き、あとから自由に変えられるようにしている。

**確かめ方**

- 外のモジュールの go.mod で `replace github.com/docker/compose/v5 => <compose のパス>` を書いて `go vet .` を打つと、`use of internal package github.com/docker/compose/v5/internal/oci not allowed`。`internal/oci` の import を消すと通る

参照：[flow-qa.md > 2. Go では internal ディレクトリを使うことが多いのか](../codebase/flow-qa.md#2-go-では-internal-ディレクトリを使うことが多いのか)

</details>

---

## プラグイン版とスタンドアロン版

### Q4 [予想]

スタンドアロン版で、次の 2 つを打つ。それぞれ、内部ではどんな引数の並びに組み替えられるか。

1. `docker-compose --context foo -f compose.yaml up`
2. `docker-compose exec mongo mongo --host mongo`

<details><summary>答え</summary>

**1 は `--context foo compose -f compose.yaml up`、2 は `compose exec mongo mongo --host mongo`(`--host` は動かない)。**

docker CLI はプラグインを `docker <グローバルフラグ> compose <サブコマンド>` の並びで呼ぶので、`--context` / `--host` / `--tls` 系は `compose` より前に置く必要がある。`compatibility.Convert` はこれを組み替える。ただし最初のフラグでない引数(ここでは `exec`)から後ろはそのまま残すので、2 の `--host` は `exec` で動かすコマンドに渡る。組み替えたあと、先頭に `docker` を足して、プラグイン版と同じ処理を通す。

**確かめ方**

- `cmd/compatibility/convert_test.go:41-43`(with context)と `:81-83`(issues/8648)が、この 2 つの例。`go test ./cmd/compatibility/` が通る
- `cmd/compatibility/convert.go:67-70`: フラグでない引数が来たら、そこから後ろをそのまま付けて止まる

参照：[cmd/note.md > 用語](../../cmd/note.md#用語) の「プラグイン呼び出し」、[cli.md > プラグイン版とスタンドアロン版](../codebase/cli.md#プラグイン版とスタンドアロン版)

</details>

### Q5 [場所]

同じバイナリが、プラグイン版(`docker compose`)として呼ばれたのか、スタンドアロン版(`docker-compose`)として呼ばれたのかを判定しているのは、どこの何という関数か。

<details><summary>答え</summary>

**`cmd/main.go` の `main` が呼ぶ、docker/cli の `plugin.RunningStandalone()`。**

docker CLI から呼ばれたことを表す環境変数 `DOCKER_CLI_PLUGIN_ORIGINAL_CLI_COMMAND` が無く、第 1 引数がメタデータの問い合わせ(`docker-cli-plugin-metadata`)でもなければ、スタンドアロンとみなす。スタンドアロンなら `compatibility.Convert` で引数を組み替えてから、どちらも `pluginMain` に進む。

**確かめ方**

- `cmd/main.go:86-89`、docker/cli の `cli-plugins/plugin/plugin.go:229-234`(go.mod が指す v29.9.1)
- Docker Desktop のプラグインのバイナリを直接 `~/.docker/cli-plugins/docker-compose version` と打つと、スタンドアロンとして動いて `Docker Compose version v5.0.1` が出る。`~/.docker/cli-plugins/docker-compose docker-cli-plugin-metadata` だと、メタデータの JSON が出る

参照：[cmd/main.go.md > main](../../cmd/main.go.md#main)

</details>

### Q6 [説明]

Compose v2 が、単体のコマンドではなく `docker` CLI のプラグインになったことで、実装の上で何が楽になったか。

<details><summary>答え</summary>

**`docker` CLI の初期化と設定を、そのまま使い回せるようになった。**

`plugin.Run` と `plugin.PersistentPreRunE` が `docker` 本体と同じ手順で CLI を初期化するので、コンテキスト、`-H`、TLS、`~/.docker/config.json`(認証情報など)を compose 側で作り直さずに済む。docker/cli の実装も直接呼べる。スタンドアロン版も同じバイナリで動くので、v1 からの移行の互換を保ったまま実装は 1 つで済む。なお、プラグインにすると決めた理由そのものは、コードや文書には書かれていない。

**確かめ方**

- `cmd/main.go:39`(`plugin.Run`)、`:55`(`plugin.PersistentPreRunE`)

参照：[cli-qa.md](../codebase/cli-qa.md) の「1. なぜプラグイン版になったのか。どんな利点があったのか」

</details>

### Q7 [場所]

`cmd/compose` のコマンドのほとんどは、Options を組み立てて Compose のバックエンド(`pkg/compose`)を呼ぶ。バックエンドを通らず、docker/cli の実装に直接任せているコマンドはどれか。どこで任せているか。

<details><summary>答え</summary>

**`stats`。`cmd/compose/stats.go:73` で docker CLI の `container.RunStats` を呼ぶ。**

プラグインなので、docker/cli の実装をそのまま呼べる(Q6)。

参照：[cmd/compose/note.md](../../cmd/compose/note.md) の「情報表示系」、[cli-qa.md](../codebase/cli-qa.md) の「1. なぜプラグイン版になったのか。どんな利点があったのか」

</details>

---

## PersistentPreRunE

### Q8 [穴埋め]

cobra の `PersistentPreRunE` は、そのコマンドと下のすべてのサブコマンドの本体の直前に呼ばれるフック。`Persistent` は ( ) こと、`E` は ( ) ことを表す。Compose では、`up` でも `ps` でも、(1) docker CLI の初期化(`plugin.PersistentPreRunE`)→ (2) ( ) → (3) Compose 側の準備(`.env` の読み込み、`--ansi` の解決、( ) の選択、`--parallel` や `--dry-run` をバックエンドのオプションに追加)の順に走る。

<details><summary>答え</summary>

**子コマンドにも引き継がれる / エラーを返せる版である / OTel のトレースの開始(`cmdtrace.Setup`)/ 進捗表示のモード(`selectEventProcessor`)**

Compose 側の準備は `RootCommand` の `PersistentPreRunE` にある。cobra は親の `PersistentPreRunE` を自動では呼ばないので、`runParentPreRun` で docker CLI 側のものを明示的に呼ぶ。さらに `cmd/main.go` がこの関数を包み直して、(1) と (2) を前に足している。トレースの開始に失敗しても、debug のログを出すだけで止まらない。

**確かめ方**

- `cmd/main.go:52-66`: 包み直し。`cmd/compose/compose.go:488-534`: Compose 側の準備

参照：[cmd/note-qa.md > 2. PersistentPreRunE とは何か](../../cmd/note-qa.md#2-persistentprerune-とは何か)

</details>

---

## 進捗表示と確認プロンプト

### Q9 [穴埋め]

`pkg/compose` は画面に直接書かず、( ) を送るだけ。受け取る側は `pkg/api` の ( ) インターフェースで、メソッドは ( ) / ( ) / ( ) の 3 つ。その実装は ( ) にある。

<details><summary>答え</summary>

**イベント / `EventProcessor` / `Start` / `On` / `Done` / `cmd/display`**

どの実装(TTY / Plain / JSON / Quiet)を使うかは、`PersistentPreRunE` の中の `selectEventProcessor` が決める(Q10)。本体が画面を知らないので、同じイベントを別の UI にも流せる。

**確かめ方**

- `pkg/api/event.go:122-129`: `EventProcessor` の定義

参照：[cmd/note-qa.md](../../cmd/note-qa.md) の「3. display は、コマンド実行中のイベントをどう画面に描画しているのか」

</details>

### Q10 [予想]

`--progress` を付けずに、端末から次の 2 つを打つ。進捗はそれぞれ TTY(同じ行を書き直す)と Plain(1 行ずつ追記)のどちらで出るか。

1. `docker compose up -d | tee log`(stdout だけをパイプにする)
2. `docker compose up -d 2>&1 | cat`(stderr もパイプにする)

<details><summary>答え</summary>

**1 は TTY、2 は Plain。自動のときは、stdout ではなく stderr が端末かどうかで決める。**

進捗を書く先が stderr だから。stdout を見ると、`up | tee log` のように stdout をリダイレクトしただけで、stderr が端末なのに Plain になってしまう。`--ansi never` なら、端末でも Plain になる。

**確かめ方**

- main から作ったバイナリ(`make build` の `bin/build/docker-compose`)で、`script -q /dev/null sh -c '<コマンド>'` として端末から打ったのと同じ状況にすると、1 は `[+] up 2/2` と ANSI の書き直し、2 は ` Container ... Created` の 1 行ずつになる
- 手元の 5.0.1(Docker Desktop の `docker compose`)では 1 も Plain になる。5.0.1 は stdout を見ていたため
- `cmd/compose/compose.go:715-728`(`selectEventProcessor` の自動判定)

参照：[cmd/note-qa.md](../../cmd/note-qa.md) の「3. display は、コマンド実行中のイベントをどう画面に描画しているのか」、[cmd/note.md > 用語](../../cmd/note.md#用語) の「TTY/Plain/JSON/Quiet」

</details>

### Q11 [穴埋め]

破壊的な操作の前の「本当に実行しますか? [y/N]」は、( ) が端末かどうかで実装を切り替える。端末なら、端末を raw モードにして 1 文字ずつ読む `User`。そうでなければ、`echo y | docker compose ...` のように ( ) で渡された入力から 1 行読む `Pipe`。

<details><summary>答え</summary>

**stdin / パイプ**

cmd/note.md の「survey を使った確認」は古い。survey は 2026-07-06 に取り除かれ、今は自前の実装。CI などで確認を飛ばしたいときは、パイプで入力を渡すか、`-y` などのフラグがあるコマンドではそれを使う。

**確かめ方**

- `cmd/prompt/prompt.go:41-46`(`NewPrompt`)

参照：[cmd/note-qa.md](../../cmd/note-qa.md) の「4. 対話端末の survey と、非対話環境でのパイプ経由とは何か」

</details>

---

## pkg/api: CLI と本体をつなぐインターフェース

### Q12 [穴埋め]

`api.Compose` は、用途ごとの 6 つの小さなインターフェースを束ねたもの。リソースを作る・動かす・消す ( )、状態を見る ( )、イメージを扱う ( )、コンテナの中で何かをする ( )、ファイルの同期・再ビルドの ( )、compose ファイルや既存のコンテナからプロジェクトを作る ( )。

<details><summary>答え</summary>

**`Lifecycle` / `Inspector` / `ImageManager` / `Runner` / `Watcher` / `ProjectLoader`**

例えば `Up` と `Down` は `Lifecycle`、`Ps` と `Logs` は `Inspector`、`Build` は `ImageManager`、`Exec` は `Runner`、`LoadProject` は `ProjectLoader`。

**確かめ方**

- `pkg/api/api.go:102-109`

参照：[pkg/api/note.md > Compose インターフェースの中身](../../pkg/api/note.md#compose-インターフェースの中身)

</details>

### Q13 [予想]

SDK として、compose.yaml(サービス `web` と `db`)を読み込んだプロジェクト全体を渡し、`CreateOptions{Services: []string{"web"}}` で `Create` を呼ぶ。作られるのは `web` だけか。

<details><summary>答え</summary>

**`web` だけではない。渡したプロジェクト全体(`web` と `db`)が対象になる。**

Options の `Services` は絞り込みではなく、利用者が名前を挙げたサービスを示すだけ。`CreateOptions.Services` は、挙げたサービスには `Recreate`、それ以外には `RecreateDependencies` の作り直しの方針を使い分けるためにある。一部だけを対象にしたいなら、呼ぶ前にプロジェクト自体を絞る(読み込みのときの `ProjectLoadOptions.Services` か、`WithSelectedServices`)。

**確かめ方**

- `pkg/api/api.go:95-101`(`Compose` のコメント)、`:341-346`(`CreateOptions.Services` のコメント)

参照：[pkg/api/api.go.md > 読むときに大事な約束](../../pkg/api/api.go.md#読むときに大事な約束)

</details>

### Q14 [穴埋め]

`api.Compose` のメソッドの第 2 引数は 2 種類ある。compose ファイルを読み込んだ結果が要る `Up` などは ( ) を取り、ラベルで探せば済む `Down` などは ( ) を取る。また `UpOptions` は、( ) と ( ) を並べただけの型。

<details><summary>答え</summary>

**`project *types.Project` / `projectName string` / `CreateOptions` / `StartOptions`**

`Down` がプロジェクト名だけで動けるので、`docker compose -p NAME down` は compose ファイル無しでも片付けられる。`UpOptions` の形のとおり、`up` は create + start。

**確かめ方**

- `pkg/api/api.go:431-434`(`UpOptions`)

参照：[pkg/api/api.go.md > Compose インターフェース](../../pkg/api/api.go.md#compose-インターフェース)、同じファイルの[読むときに大事な約束](../../pkg/api/api.go.md#読むときに大事な約束)

</details>

---

## up のフラグ

### Q15 [予想]

`docker compose up -d --abort-on-container-exit` と `docker compose up --wait --attach web` は、それぞれどうなるか。

<details><summary>答え</summary>

**どちらもエラーで止まり、何も起動しない。**

`-d` は起動したらすぐ戻るので、コンテナの終了を見張る `--abort-on-*` や、ログを流す `--attach` と組み合わせられない。`--wait` は `-d` を含むので、同じ理由で `--attach` と組み合わせられない。この検証は、プロジェクトを読み込む前に `validateFlags` がする。

**確かめ方**

- `docker compose up -d --abort-on-container-exit` → `--detach cannot be combined with --abort-on-container-exit, ...`(終了コード 1)
- `docker compose up --wait --attach a` → `--wait cannot be combined with --abort-on-container-exit, --attach or --attach-dependencies`(終了コード 1)

参照：[cmd/compose/up.go.md > validateFlags](../../cmd/compose/up.go.md#validateflags)、[cmd/compose/note-qa.md > 1. ライフサイクル系コマンドでよく使われるフラグ](../../cmd/compose/note-qa.md#1-ライフサイクル系コマンドでよく使われるフラグ)

</details>
