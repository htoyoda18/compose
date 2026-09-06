# cmd/compose メモ

`docker compose` の各サブコマンド(up, down, ps, logs, ...)を定義するcobraコマンド層。
基本的には薄いCLIラッパーで、フラグを`*Options`構造体に詰め替え、
`Backend`インターフェース(実体は`pkg/compose`)を呼ぶだけの構成。

## ファイル一覧・責務

### コマンド共通基盤

- [compose.go](compose.go.md)
  - ルートコマンド`RootCommand`の構築
  - `ProjectOptions`(`--file`/`--profile`/`--project-name`等の共通フラグとプロジェクトロード)
  - `Adapt`/`AdaptCmd`(SIGINT/SIGTERMで`context.WithCancel`をキャンセルしつつcobraのRunE形式に変換)
  - `selectEventProcessor`/`resolveMaxConcurrency`など全コマンド共通の設定・起動処理
- backend.go
  - `withBackend`(バックエンド生成ヘルパー。kill/start/stop/pause/restartから利用)
  - `optionalTimeout`(int秒+changedフラグから`*time.Duration`に変換するヘルパー)
- options.go
  - `DOCKER_DEFAULT_PLATFORM`のservice.Platformへの反映(`resolvePlatforms`/`applyPlatforms`)
  - リモート(git://, oci://)compose設定の検出と警告(`confirmRemoteIncludes`)
  - 変数展開前の確認プロンプト(`promptForInterpolatedVariables`)。`up`/`run`から呼ばれる
- alpha.go
  - `viz`/`publish`/`generate`をまとめる実験的コマンド群の親(`Hidden: true`)
- completion.go
  - シェル補完用関数群(サービス名/プロジェクト名/プロファイル名/`--scale`引数)

### ライフサイクル系コマンド

- [up.go](up.go.md)
  - `docker compose up`。`upOptions`/`createOptions`/`buildOptions`を束ね、`validateFlags`で
    フラグの非互換組み合わせを検証、`runUp`でcreate→attach対象決定→`backend.Up`を実行
- [down.go](down.go.md)
  - サービス停止・削除。`--remove-orphans`/`--timeout`/`--volumes`/`--rmi`を定義
- create.go
  - コンテナ作成のみ。`recreateStrategy`で再作成方針を決定し`backend.Create`を呼ぶ
- start.go / stop.go / restart.go / pause.go(pause/unpause) / kill.go / remove.go(rm) / scale.go / wait.go
  - いずれもオプション構造体を組み立てて対応する`backend.*`を呼ぶ薄いラッパー

### 実行・アタッチ系

- run.go
  - 使い捨てコンテナ実行。本ディレクトリで最も複雑なコマンドの一つ
- exec.go
  - 実行中コンテナでコマンド実行。非ゼロ終了コードを`cli.StatusError`に変換して`os.Exit`
- [attach.go](attach.go.md)
  - 起動中コンテナの標準入出力にアタッチ
- commit.go
  - コンテナをイメージ化

### 情報表示系

- ps.go / list.go(ls) / logs.go / events.go / images.go / top.go / stats.go / port.go / volumes.go / version.go
  - `backend.*`の結果をtable/json/quiet等のフォーマットで出力するだけの薄いコマンド群
  - stats.goのみComposeバックエンドを経由せず`docker/cli`の`container.RunStats`に直接委譲

### ビルド・イメージ配布系

- build.go
  - イメージビルド。`--ssh`/`--build-arg`/`--memory`等を`api.BuildOptions`に変換
- pull.go / push.go / publish.go(OCI公開) / export.go / cp.go
  - イメージのpull/push/publish、コンテナのファイルシステムexport、ファイルコピー

### 実験的・その他

- viz.go
  - Graphvizのdot形式でプロジェクト構造を可視化(EXPERIMENTAL)
- generate.go
  - 起動中コンテナからcompose定義を逆生成(EXPERIMENTAL)
- bridge.go
  - `convert`/`transformations list`/`transformations create`
- hooks.go
  - Docker Desktop連携用のpost-commandフック(Logsタブへのディープリンクヒント)
- config.go
  - `docker compose config`。設定の表示・検証。本ディレクトリ最大級のファイル
- watch.go
  - ファイル変更監視と自動リビルド。`locker.NewPidfile`でプロジェクト単位の多重起動防止

---

## 確認できた実害のあるバグ(コードを直接読んで確認済み)

### 1. run.go — `--env-from-file`のエラーが握りつぶされ、環境変数がまるごと消える

`getEnvironment`(L122-143):
```go
f, err := os.Open(file)          // Closeされていない(fdリーク)
...
vars, err := dotenv.ParseWithLookup(f, ...)
if err != nil {
    return nil, nil              // ← エラーなのにnilを返している
}
```
`broken.env`のように構文が壊れた`--env-from-file`を指定すると、パースエラーが握りつぶされ
`environment`が`nil`のまま返る。呼び出し元`runRun`はエラーが無いので処理を続行し、
`-e FOO=bar`のように明示指定した環境変数まで含めて**コンテナが環境変数なしで起動**する。
ユーザーには何もエラーが出ない。加えて`os.Open`した`f`も`Close`されておらず、
`--env-from-file`を都度指定するたびfdがリークする。
→ `return nil, err`と`defer f.Close()`で直せる単純な修正。

### 2. build.go — コピペミスで`--progress`警告が出ない条件になっている

L113-118:
```go
if cmd.Flags().Changed("ssh") && opts.ssh == "" {
    opts.ssh = "default"
}
if cmd.Flags().Changed("progress") && opts.ssh == "" {   // ← 前の行からのコピペ
    fmt.Fprint(os.Stderr, "--progress is a global compose flag, ...")
}
```
`opts.ssh == ""`という条件が意味的に無関係。`docker compose build --progress=plain --ssh=default web`
のように`--ssh`も同時指定すると`opts.ssh`が非空になり、本来出るべき警告が出なくなる。
実害は警告抑制のみで低リスクだが、意図とコードが不一致。

### 3. viz.go — `backend.Viz`のエラーを握りつぶし

L82:
```go
graphStr, _ := backend.Viz(ctx, project, api.VizOptions{...})
```
現状`pkg/compose`側の実装は常に`nil`エラーしか返さないため今は実害ゼロだが、他の全コマンドは
一貫して`backend.*`のエラーを`return err`しており、ここだけ握りつぶしている。将来`Viz`にエラー
パスが追加されると気づかれずに空の出力で正常終了してしまう。

## 確度は中程度だが把握しておく価値のある点

- **hooks.go**: `docker compose logs -f`のように`-f`(followの短縮形)を使うと、
  `resolveAppIDIn`が`--file`の短縮形`-f`と誤認してappId解決をスキップする可能性(要検証)。
- **up.go**: `--attach`は存在しないサービス名を検証してエラーにするが、`--no-attach`には
  同じ検証が無く、タイポしても黙って無視される。
- **start.go**: `up.go`は`--wait-timeout`の負数を`validateFlags`で弾くが、`start.go`の
  同名フラグには対応するチェックが無い。
- **create.go**: `up.go`/`run.go`にある`--always-recreate-deps`相当や`COMPOSE_IGNORE_ORPHANS`が
  `create`単体からは効かない(意図的な仕様の可能性もあり)。
- **config.go**: `lockModel`だけ他の姉妹関数と違い型アサーションに`ok`チェックが無く、
  想定外のモデル形状でpanicする可能性(要検証)。
- **logs.go**: `logsOptions`が`*ProjectOptions`を直接埋め込みつつ`composeOptions`
  (内部にも`*ProjectOptions`を持つ)を値型で埋め込んでおり、後者は常にnilのまま
  (現状未使用なので実害なし、同パターンがpush.go/pull.goにも既存)。
- **port.go**: `Args: cobra.MinimumNArgs(2)`だが実際は`args[0]`/`args[1]`しか使わず、
  3個目以降の引数は黙って無視される(`ExactArgs(2)`の方が意図に合致しそう)。

## 軽微な一貫性の問題

- 出力に`dockerCli.Out()`/`dockerCli.Err()`を経由せず生の`os.Stdout`/`os.Stderr`を直接使う
  ファイルがある(viz.go, generate.go, config.goの一部関数)。テストでの出力キャプチャが
  できず、他コマンドとスタイルが非対称。
- `withBackend`ヘルパーを使わず`compose.NewComposeService`を直接呼ぶファイルが多数
  (down.go/events.go/exec.go/export.go/images.go/list.go/logs.goなど)、パッケージ内で
  バックエンド生成のパターンが統一されていない。
- `down.go`/`restart.go`/`stop.go`でタイムアウト変換ロジック(`timeChanged`/`timeout`→
  `*time.Duration`)が`optionalTimeout()`(backend.go)と重複して再実装されている箇所がある。

## テストカバレッジ

`ps.go`/`ps_test.go`のようにテストがあるファイルもあるが、`alpha.go`/`attach.go`/`build.go`/
`commit.go`/`create.go`/`scale.go`/`start.go`/`stop.go`/`wait.go`/`watch.go`などcmd/compose単体の
ユニットテストが存在しないファイルも多い(e2eでのカバーの可能性はあるが未確認)。
