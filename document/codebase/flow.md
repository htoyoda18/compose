# 処理の流れ

コマンドを打ってから Docker Engine API が呼ばれるまでの流れ。前半は全コマンド共通、後半は主要コマンドごとの `pkg/compose` 内の流れ。
各層の役割は [architecture.md](./architecture.md)、plan engine の詳細は [plan-engine.md](./plan-engine.md)、
コードを読む順番は [reading-guide.md](./reading-guide.md) を参照。

← [document 一覧](../README.md) / [コードノート入口](../../note.md)

## 共通の流れ

`cmd/main.go` → `cmd/compose/<コマンド>.go` → `pkg/api.Compose` のメソッド → `pkg/compose/<コマンド>.go` と降りていく。

1. 起動([cmd/main.go](../../cmd/main.go.md))
   - `main` はまず起動形態を判定する。単体バイナリ `docker-compose` として起動された(スタンドアロン)なら、
     `cmd/compatibility` が引数列を `docker compose ...` の並びに組み替える
   - その後 `pluginMain` が docker CLI プラグインとして `plugin.Run` を呼び、Docker デーモンへの接続設定などの初期化を docker/cli に任せる
2. コマンドの組み立て([cmd/compose](../../cmd/compose/note.md))
   - `RootCommand` が cobra のルートコマンドを作り、`up`/`down`/`ps` などのサブコマンドを登録する
   - コマンド実行前(`PersistentPreRunE`)に `cmd/cmdtrace` が OpenTelemetry のトレースを初期化する
3. サブコマンドの実行(`cmd/compose/<コマンド>.go`)
   - フラグを `upOptions` などの構造体に詰め、必要なら `ToProject` → `LoadProject` で compose.yaml を読み込んで `types.Project`(compose-go のモデル)にする
     - `down`/`logs` のようにプロジェクト名だけで動くコマンドは、compose.yaml がなくてもラベルから対象コンテナを探せる
   - `compose.NewComposeService` でバックエンド(`pkg/compose`)を作り、`pkg/api` のインターフェース越しに `backend.Up(...)` などを呼ぶ
4. Compose の本体([pkg/compose](../../pkg/compose/note.md))
   - コマンドごとの処理(下記)。必要に応じて `internal` の部品(`internal/projectloader`, `internal/oci`, `internal/sync`, `internal/provider` など)を使う
5. 表示([cmd/display](../../cmd/display/note.md), [cmd/formatter](../../cmd/formatter/note.md))
   - 実行中の進捗はイベントとして流れ、`cmd/display` が TTY/Plain/JSON などのモードに応じて描画する
   - `ps`/`images` などの一覧の出力は `cmd/formatter` が整形する

## コマンド別の流れ(`pkg/compose` 内)

### up([up.go](../../pkg/compose/up.go.md))

- `Up` は plan engine(観測 → 計画 → 実行)を回す
  - 観測: `collectObservedState` が、ラベルを手がかりにプロジェクトの既存コンテナ・ネットワーク・ボリュームを集める
  - 計画: `reconcile` が、望ましい状態(`Project`)との差分を埋める操作の列(`Plan`)を作る
  - 実行: `executePlan` が依存順に並列で Docker Engine API を呼ぶ
- `up -d` は `upDetached` が作成と起動を 1 つのプランでまとめて実行する
- `-d` なしは `create` でコンテナを作ったあと、`runInteractiveUp` が attach の準備をしてから `start` で起動し、ログを流し続ける
- `up --watch` の場合は、起動後に下記の watch も開始する

### down([down.go](../../pkg/compose/down.go.md))

plan engine は通らず、消す対象を直接集めて削除する。

1. `getContainers` でプロジェクトのコンテナをラベルから集める。compose.yaml がなければ `getProjectWithResources` でコンテナからプロジェクトを復元する
2. `InReverseDependencyOrder` で依存の逆順(依存される側を最後)に、サービスごとに停止・削除する(`downService`)
3. `--remove-orphans` なら compose.yaml から消えたサービスのコンテナ(孤児)も削除する
4. pre_start フック用のコンテナを片付ける
5. ネットワーク、`--rmi` 指定時はイメージ、`-v` 指定時はボリュームの削除を集め、`errgroup` で並列に実行する

### build([build.go](../../pkg/compose/build.go.md))

1. `build:` セクションを持つサービスをビルド対象として選ぶ
2. `buildWithBake` がビルド方式を決める
   - BuildKit が有効で buildx プラグインがあれば `doBuildBake`(`docker buildx bake` に全サービスをまとめて渡す)
   - なければ `doBuildClassic`(サービスごとに Engine API の旧ビルダーで 1 つずつビルド)
3. ビルドするイメージにはプロジェクト名・サービス名のラベル(`com.docker.compose.project`/`service`)が付く。`up --build` などからは `ensureImagesExists` 経由で内部の `build` が呼ばれる

### run(run.go)

`docker compose run <service>` は、通常のサービスコンテナとは別に one-off コンテナを 1 つ作って動かす。

1. `startDependencies` が対象サービス以外(依存先)を `create` → `start` で先に立ち上げる
2. `ensureImagesExists` でイメージを用意し、`depends_on` の条件(healthy など)を満たすまで待つ
3. `createContainer` で one-off コンテナを作り、secrets/configs を注入する
4. シグナルをコンテナへ転送する設定をしてから、docker/cli の `RunStart` で起動と attach を一度に行う。終了コードがそのままコマンドの終了コードになる

### logs(logs.go)

1. `selectLogsContainers` で対象コンテナを選ぶ
2. コンテナごとに `logContainer` が Engine API からログを取り、サービス名のプレフィックス付きで出力する
3. `-f` の場合は `monitor` がコンテナの起動イベントを監視し、あとから起動・再起動したコンテナのログも追いかける

### watch([watch.go](../../pkg/compose/watch.go.md))

`docker compose watch` は、まず `up` でサービスを起動し(`--no-up` で省略可)、そのあと `backend.Watch` を呼ぶ。

1. 準備: 各サービスの `develop.watch` 設定を集めて `watchRule`(監視パス・include/ignore・action)にし、`pkg/watch` の `NewWatcher` で OS のファイル監視を始める
2. イベントの束ね: `BatchDebounceEvents` が短時間に連続したファイル変更を 1 つのバッチにまとめる
3. 振り分け: `handleWatchBatch` が、変更されたファイルにマッチしたルールの action ごとに処理を分ける
   - `sync`: `internal/sync` がファイルを tar にまとめてコンテナへ転送する
   - `restart`: コンテナを再起動する(同期はしない)
   - `sync+restart`: 同期したあとコンテナを再起動する
   - `sync+exec`: 同期したあとコンテナ内でコマンドを実行する
   - `rebuild`: `rebuild` が `build` でイメージを作り直し、`createAndStart` でコンテナを作り直す
