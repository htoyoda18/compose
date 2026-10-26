# compose-concepts.md を読んで出た疑問 Q&A

[compose-concepts.md](./compose-concepts.md) を読んでいて出た疑問への短い回答。手元の compose と、go.mod が指す compose-go(v2.16.2-0.20261015044941-5515644a24af)のコードで確認した。
compose.yaml を読み込む話(YAML → プロジェクト名)→ 読み込んだサービスがコンテナになる話 → サービス同士の順序 → コンテナにならないサービス(provider)の順に並べている。

← [document 一覧](../README.md)

## 1. 「YAML 自体の挙動」とは何か

- compose-spec の意味(services、depends_on など)が乗る前の、**YAML という書式そのものの決まり**のこと。compose-go は `go.yaml.in/yaml/v4` で読む(compose-go `loader/loader.go:47`)。compose で特に効くのは次の 4 つ。
  - **アンカーとエイリアス**: `&name` で印を付けた値を `*name` で使い回す。`<<: *name` で map を混ぜ込める(compose-go は `<<` を扱う処理を持つ。`loader/reset.go:90-92`)
  - **`x-` で始まるキー**: compose は無視するので、使い回す設定の置き場(アンカーの定義場所)にする(`loader/loader.go:826`)
  - **1 ファイルに複数の文書**: `---` で区切った文書を順に読む(`loader/loader.go:574-590`)
  - **compose 独自のタグ**: `!reset`(前のファイルの値を消す)と `!override`(混ぜずに置き換える)。`-f` で複数のファイルを重ねるときに使う(`loader/reset.go:95-99`)
- 落とし穴の「`yes` / `no` が真偽値になる」は YAML 1.1 の決まりで、YAML 1.2 では文字列。compose-go がどちらの扱いになるかは、コードでは確かめていない。迷う値(`"yes"`、ポートの `"80:80"` など)は、引用符で囲めば必ず文字列になる。

## 2. プロジェクト名は、ディレクトリのパスではなくディレクトリ名か

- **ディレクトリ名(パスの最後の要素)**。パス全体ではない。名前は次の順で決まる(compose-go `cli/options.go:637-655`)。
  1. `-p` / `--project-name`
  2. 環境変数 `COMPOSE_PROJECT_NAME`
  3. compose.yaml のトップレベルの `name:`
  4. どれも無ければ、作業ディレクトリの `filepath.Base`(例: `/home/me/work/myapp` → `myapp`)
- 4 のときは、小文字にして `a-z 0-9 _ -` 以外の文字を消す(`loader/loader.go:799-804`)。例: `My App` → `myapp`。
- パスは別のラベル(`com.docker.compose.project.working_dir`)に記録されるだけで、名前には使われない(`pkg/api/labels.go`)。だから、**別の場所にある同じ名前のディレクトリは、同じプロジェクトとして扱われる**。`down -v` が他のプロジェクトのボリュームを消すバグ(issue #14373、[next-tasks.md の N2](../tracking/next-tasks.md))は、これが原因だった。

## 3. 1 サービス = 1 コンテナという認識で合っているか

- **既定ではそうだが、いつもではない**。サービスは「同じ設定で動くコンテナの型」で、そこから作られるコンテナの数は変わる。
  - **0 個**: `profiles` が有効でないとき。`provider:` のサービス(コンテナではなく外部のリソースになる。5 を参照)
  - **複数個**: `--scale` や `deploy.replicas` で台数を増やしたとき。名前は `<プロジェクト>-<サービス>-<連番>` になる(`pkg/compose/service_containers.go:127-137`)。`container_name` を指定したサービスは名前がぶつかるので、2 台以上にできない(同 `:60-63`)
  - **ほかに**: `docker compose run` の 1 回だけのコンテナ(`com.docker.compose.oneoff` ラベルで区別。`pkg/api/labels.go`)や、`pre_start` などのフックのコンテナも、同じサービスに属する
- どのコンテナがどのサービスの何台目かは、ラベル(`com.docker.compose.service`、`com.docker.compose.container-number`)で見分ける。

## 4. トポロジカルソートとは。ほかで説明しているか

- **依存関係のある物を、「依存先が必ず先に来る」ように 1 列に並べること**。db ← web(web が db に依存)なら、起動は db → web、停止はその逆の web → db。
- 説明は [codebase/algorithms.md](../codebase/algorithms.md) の「3. 依存順に処理する(トポロジカル順序)+並列化」 にある。Compose は 1 列に並べる代わりに「今すぐ処理してよいものを全部同時に処理する」ことで並列化している、という点まで書いてある。
- plan engine での使われ方(計画のノードがトポロジカル順に並ぶ)は [codebase/plan-engine.md](../codebase/plan-engine.md) を参照。

## 5. Provider 拡張機構の具体的な使い方

- **コンテナの代わりに、外部のコマンドにサービスを用意させる仕組み**。書き方の例(`docs/extension.md:16-24`、`:189-203`):
  ```yaml
  services:
    app:
      image: myapp
      depends_on: [database]
    database:
      provider:
        type: awesomecloud        # 実行するコマンド名(docker CLI のプラグインか、PATH 上の実行ファイル)
        options:                  # コマンドのフラグに変換される
          type: mysql
          size: 256
  ```
- **`up` のとき**、Compose は `awesomecloud compose --project-name <名前> up --type=mysql --size=256 database` を実行する(`docs/extension.md:38-45`)。コマンドは標準出力に 1 行 1 つの JSON を書いて、Compose とやり取りする。
  - `{"type": "info", ...}` は進捗として表示される
  - `{"type": "setenv", "message": "URL=https://..."}` を返すと、依存するサービス(上の `app`)に `DATABASE_URL` が環境変数として渡される。サービス名が頭に付く(`docs/extension.md:205-211`)。`rawsetenv` なら名前はそのまま
- **`down` のとき**は `... compose --project-name <名前> down database` を実行し、確保したリソースの後始末をコマンドに任せる(`docs/extension.md:357-360`)。代表的な実例は Docker Model Runner(`type: model` で `docker-model` プラグインを呼ぶ。`docs/extension.md:27`)。
- 実装例は `docs/examples/provider.go`、Compose 側の実装は `internal/provider/`。
