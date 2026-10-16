# Compose のライフサイクルと意味論

← [document 一覧](../README.md)

compose.yaml の各機能が、実際にどう振る舞うか。依存、ヘルスチェック、作り直しの判定、ラベル、読み込みの順番、ボリューム、停止の仕方を、コードの場所つきでまとめる。
概念の入口は [compose-concepts.md](./compose-concepts.md)、差分を埋める仕組みは [codebase/plan-engine.md](../codebase/plan-engine.md) を参照。

- 基準コード: `upstream/main` 289d8a037（2026-10-16）。`file:line` は upstream の行番号（GitHub のリンクの基点は `https://github.com/docker/compose/blob/289d8a037874a5197c54cfa055d28cc5d980593f/`）
- compose-go の部分は、少し古いバージョン（v2.16.2-0.20261012）で読んだので、行番号は少しずれているかもしれない
- **（推定）** と書いたものは、コードから推論しただけで、実機では確かめていない

## 1. depends_on

| 条件 | 何を待つか | plan での形 |
| ---- | ---------- | ----------- |
| `service_started`（既定） | 依存先が起動したこと | 待ちのノードは作らない。依存先の起動の鎖への辺だけ |
| `service_healthy` | 依存先が healthy になること | `OpWaitCondition` ノード |
| `service_completed_successfully` | 依存先が exit 0 で終わること | `OpWaitCondition` ノード |

- **待つかどうか**は `shouldWaitForDependency`（`service_containers.go:362-383`）が決める。次の場合は待たない:
  - `service_started`
  - `--no-deps` で無効にしたサービス
  - scale 0
  - provider サービス
- **待ち方**: 500ms ごとに `ContainerInspect` を呼んで状態を見る（`waitDependency`、`service_containers.go:215-273`）。イベントは使わない。待っている間は `--parallel` の枠を使わない。
- **`required: false`（任意の依存）**:
  - 依存先がモデルに無い場合（profiles で無効など）: 読み込みの段階で依存ごと消える（`WithoutUnresolvedOptionalDependencies`）。
  - 依存先のコンテナが 0 個、unhealthy、exit ≠ 0、自分の待ち時間切れの場合: 警告を出して先に進む。
  - ただし、**依存先のコンテナの作成や起動そのものが失敗すると、任意の依存でも plan 全体が失敗する**（推定）。
- **`restart: true`**: 依存先を作り直すとき、依存元を先に止め（`planStopDependents`）、Start フェーズで起動し直す。`docker compose restart X` でも、`restart: true` の依存元は一緒に再起動される。
- **止める順番**: `stop` と `down` は依存の逆順（`InReverseDependencyOrder`）。`down web` のようにサービスを指定すると、**その依存元も対象になる**（`WithRootNodesAndDown`）。
- **暗黙の depends_on**: compose-go が、次の参照から自動で `service_started` の依存を足す。
  - `links`、`network_mode: service:x`、`ipc` / `pid` / `uts` / `cgroup` の `service:`（`restart: true` 付き）
  - `volumes_from`（`restart: false`）

## 2. healthcheck と `--wait`

- healthcheck は Engine にそのまま渡す（`ToMobyHealthCheck`、`convert.go:43-85`）。`start_period` や `retries` を数えるのは **Engine**。
  - `start_interval` を書くなら、`start_period` も必須（無いとエラー）。
- Compose は inspect で状態を見るだけ（`isServiceHealthy`、`service_containers.go:578-612`）:
  - `starting`（`start_period` の間を含む）→ 待ち続ける
  - `unhealthy` → **すぐにエラー**。その後 healthy に戻るのを待たない
  - exited → エラー
  - healthcheck が無い → 場合により、running なら OK か、「no healthcheck configured」エラー
- **depends_on の待ちには、既定のタイムアウトが無い**。`--wait-timeout` を付けると、待ちのノードごとに独立した時間枠で制限する（`--wait` 無しでも効く）。
- **`--wait`**: 自動的に detached になる。plan が終わってから全サービスを見て、running か healthy になるまで待つ。
  - 落とし穴: exit 0 で終わる使い捨てのサービスでも、他のサービスから `service_completed_successfully` で参照されていなければ、`--wait` は「exited」で失敗する。

## 3. restart ポリシー

- `restart:` は `HostConfig.RestartPolicy` になる（`getRestartPolicy`、`create.go:735-760`）。`deploy.restart_policy` があれば、そちらで**上書き**する。
- **再起動するのは Engine**。Compose 自身は再起動しない。monitor は、コンテナが死んだ後に inspect して「再起動中か」を見るだけ。
- `restart` は config hash に含まれるので、変えるとコンテナが作り直される。
- `run` の oneoff コンテナでは、restart ポリシーを消す。
- **`docker compose restart`** は `ContainerRestart` を呼ぶだけで、作り直さない（設定の変更は反映されない）。前に pre_stop、後に post_start を実行する。

## 4. ライフサイクルフック

| フック | 実行のしかた | いつ |
| ------ | ------------ | ---- |
| `pre_start` | **専用のコンテナ**（`<project>-<service>-pre_start-<i>`）を作って実行する。サービスのコンテナと `VolumesFrom` でボリュームを共有する | 観測した時点で、そのサービスの running なコンテナが 1 つも無いとき、サービスごとに 1 回 |
| `post_start` | 起動したコンテナの中で `docker exec` | レプリカごとに、起動の直後 |
| `pre_stop` | 止めるコンテナの中で `docker exec` | `stop`、`down`、前景の `up` の Ctrl-C、`restart` の前 |

- pre_start のコンテナ（`pre_start.go`、`hook.go`）:
  - Create フェーズで作り、Start フェーズで実行する。
  - 成功したら消す。**失敗したら調べられるように残す**（エラーメッセージに、出力の末尾とコンテナの short ID を含める）。
  - 残ったコンテナは `down` が消す（`removePreStartHookContainers`）。この部分のバグを #14310 で直した。
  - config-hash ラベルを付けないので、`ps` などの普通の一覧には出ない。
  - `per_replica: true` は未対応（エラー）。
- **気付いた点**（推定・未報告）:
  - **plan の `StopContainer` は pre_stop を実行しない**（`executor_ops.go:309-314`）。そのため、作り直し、scale down、orphan の削除、ネットワークの作り直しで止めるときは、pre_stop が動かない。意図した挙動かは不明。
  - ラベルから組み立て直したプロジェクト（`-p` だけ）には、フックの情報が無い。だからフックは動かない。

## 5. 作り直し（recreate）の判定

### 判定の順番（`mustRecreate`、`reconcile.go:1025-1048`）

1. 方針が `never` なら作り直さない。`force` なら作り直す
2. 名前空間やボリュームを共有する親（`network_mode: service:x`、`volumes_from` など）が作り直されるなら、作り直す
3. **config hash** が違う
4. イメージの digest（ラベル `image`）、image volume の digest が違う
5. 動いているコンテナが、あるべきネットワークにつながっていない
6. 名前付きボリュームのマウントが違う

### config hash に何が入るか（`hash.go`）

- サービスの設定を JSON にしてハッシュを取る（`ServiceHash`）。
- **除外するもの**（`trimServiceHashFields`）: `build`、`pull_policy`、`scale`、`deploy.replicas`、`depends_on`、`profiles`。
- Compose が内部で付けるラベル（`CustomLabels`）は入らない。ユーザーが書いた `labels` は入る。
- `network_mode: service:x` のような参照は、実際のコンテナ ID に置き換えてから計算する。
- 結果として、たとえば `depends_on` を変えても作り直されない。ラベル `depends_on` は古いまま残る（推定）。
- ネットワークとボリュームにも、それぞれのハッシュ（`NetworkHash` / `VolumeHash`）がラベルとして付く。

### フラグと方針（`cmd/compose/create.go:150-171`）

| フラグ | 指定したサービス | 依存先のサービス |
| ------ | ---------------- | ---------------- |
| （なし） | diverged（違えば作り直す） | diverged |
| `--no-recreate` | never | never |
| `--force-recreate` | force | diverged |
| `-V` / `--renew-anon-volumes` | **force**（匿名ボリュームを新しくするので、必ず作り直す） | diverged |
| `--always-recreate-deps` | diverged | force |

### 匿名ボリュームの引き継ぎ（`Inherit`）

- 作り直すとき、`Inherit` が true なら、古いコンテナの匿名ボリュームを新しいコンテナに付け直す（`reconcile.go:1629-1632` → `buildContainerMountOptions`、`create.go:1159-1203`）。
  - 対象は、イメージの `VOLUME` と、compose.yaml で source を書かなかったボリューム。
- `up` / `create` では `Inherit = !-V`。
- **`run` の依存サービスでは渡していない** → postgres などのデータが消えたように見える（[next-tasks.md の 2](../tracking/next-tasks.md#2-run-が依存サービスを作り直すと匿名ボリュームを引き継がない)）。
- 作り直しは「一時的な名前で作成 → 古いものを止める → 消す → 名前を変える」の 4 段階（[plan-engine.md](../codebase/plan-engine.md#作り直しは-4-つの操作に分かれる)）。

## 6. ラベル（`com.docker.compose.*`）

定義は [pkg/api/labels.go](https://github.com/docker/compose/blob/289d8a037874a5197c54cfa055d28cc5d980593f/pkg/api/labels.go)。

| ラベル | 意味 |
| ------ | ---- |
| `project` / `service` | どのプロジェクト・サービスのものか |
| `config-hash` | 作り直しを判定するための設定のハッシュ |
| `container-number` | レプリカの番号（oneoff には無い） |
| `oneoff` | `run` で作った使い捨てのコンテナか（`True` / `False`） |
| `slug` | oneoff の識別子 |
| `depends_on` | `サービス:条件:restart` を `,` でつないだもの。**`required` は保存されない** |
| `image` / `image-volume-digest` | イメージの digest |
| `project.working_dir` / `project.config_files` / `project.environment_file` | どのディレクトリ、どのファイルから作ったか |
| `network` / `volume` | ネットワーク・ボリュームの compose でのキー |
| `version` | 作った Compose のバージョン |
| `replace` | 作り直しで置き換えた古いコンテナ |
| `hook` / `hook-index` | pre_start のコンテナ |
| `relay` / `relay-network` | provider の relay |

- **自分のコンテナを探すときの既定のフィルタ**（`containers.go:90-104`）は、`project` ＋ `config-hash` が**あること** ＋ `oneoff`。そのため、config-hash の無いフックのコンテナは普通の一覧に出ない。
- **ラベルからプロジェクトを組み立て直す**（`projectFromName`、`compose.go:471-539`）:
  - いつ使うか: `-f` を付けずに `-p` を付けたとき。カレントディレクトリに compose.yaml があっても、ファイルは読まない（`projectOrName`）。
  - 戻せるもの: サービス名、イメージ、ラベル、コンテナの数（scale）、`depends_on`。
  - 戻せないもの: healthcheck、フック、stop_signal、environment など。
  - ここから生まれたバグが [next-tasks.md の保留](../tracking/next-tasks.md#start--p-で任意の依存サービスが必須扱いになる)。メンテナは「ラベルに設定を増やしたくない」という立場。

## 7. モデルの読み込み

### 読み込みの順番（`internal/projectloader/loader.go:81-121`）

1. 作業ディレクトリ
2. OS の環境変数
3. remote（git / OCI）の loader
4. `--env-file` / `.env`
5. `COMPOSE_FILE`、無ければ既定のファイル（`compose.yaml` などを親ディレクトリまでさかのぼって探し、override ファイルも自動で足す）
6. もう一度 `.env`（プロジェクトのディレクトリが決まった後で読み直す）
7. profiles
8. プロジェクト名

### 環境変数の優先順位

- **`${VAR}` の補間**: **シェルの環境変数 > `--env-file` / `.env`**。
  - `--env-file` を指定すると、`.env` は読まない。
  - `COMPOSE_DISABLE_ENV_FILE` で `.env` を無効にできる。
- **コンテナに渡る環境変数**: `env_file` を読んだ後、`environment` で上書きする。
- `.env` に書いた `COMPOSE_*` は、まだ設定されていなければプロセスの環境変数として export される（`setEnvWithDotEnv`）。
  - ただし、フラグの定義時に `os.Getenv` を読むもの（`COMPOSE_PROGRESS` など）には効かない（next-tasks の次点を参照）。

### profiles

- `--profile`、無ければ `COMPOSE_PROFILES` で有効にする。`profiles` を書いていないサービスは常に有効。
- **サービス名を明示すると、そのサービスの profile も有効になる**。
- 必須の依存先が無効だと、エラー「is disabled. Can be enabled by profiles」になる。

### プロジェクト名の決まり方

1. `-p`
2. `COMPOSE_PROJECT_NAME`（`.env` に書いたものを含む）
3. compose.yaml の `name:`
4. 作業ディレクトリ名（`--project-directory` > 最初の `-f` のディレクトリ > カレントディレクトリ）を正規化したもの

### include / extends / 複数の `-f`

- すべて compose-go が処理する。docker/compose 側は数を数えてメトリクスに送るだけ。
- **このあたりのバグは、compose-go に PR を出す必要があることが多い**（例: `!reset` の #11816 は compose-go#913 で直す）。

## 8. ボリューム

| 種類 | 書き方 | 実体 |
| ---- | ------ | ---- |
| 名前付き | `volumes: [data:/var/lib/x]` ＋ トップレベルの `volumes: {data: {}}` | `<project>_data`（external ならそのままの名前） |
| 匿名 | `volumes: [/var/lib/x]`、またはイメージの `VOLUME` | Engine がランダムな名前で作る |
| bind | `volumes: [./src:/app]` | ホストのパス（絶対パスに直す） |
| tmpfs | `tmpfs:` | メモリ上 |

- `volumes_from` は `HostConfig.VolumesFrom` になる。サービス名は、実行時に実際のコンテナ ID に置き換える。
- **設定が変わった名前付きボリューム**（`reconcileVolumes`、`reconcile.go:429-483`）:
  - 同じ名前のまま設定だけ変わった場合: **データが消えるので、確認のプロンプトを出す（既定は No）**。`-y` なら作り直す。
  - 名前が変わった場合: 新しいボリュームを作り、古いボリュームは残す。
  - 別のプロジェクトのボリュームや、ハッシュの無いボリュームには触らない。
- **`down -v`**: 匿名ボリュームはコンテナと一緒に消す。名前付きボリューム（external 以外）も消す。使用中なら警告だけ出す。

## 9. ネットワーク

- 何も書かなければ、`<project>_default` が使われる（使うサービスがあるときだけ作る）。
- エイリアス: コンテナ名は常に付く。サービス名と `aliases` も付く（サービス名を外すオプションが、10-16 に入った）。
- **設定が変わったネットワークは、確認せずに作り直す**: 止める → 切断する → 消す → 作る → つなぎ直す。**コンテナ自体は作り直さない**。
- `down` は external 以外のネットワークを消す。`run` のコンテナが動いていると「still in use」で残る（推定）。

## 10. oneoff（`docker compose run`）

- 名前は `<project>-<service>-run-<12文字>`（`--name` で変えられる）。ラベルは `oneoff=True` と `slug`。
- `--rm` → `AutoRemove`。`--service-ports` を付けないと ports を外す。restart ポリシーは消す。
- **依存サービスは、`create()` → `start()` で起動する**（`startDependencies`、`run.go:291-318`）。ここで `Inherit` と `Timeout` を渡していない。
- orphan として扱うのは、**終わった** oneoff だけ。動いている `run` は `--remove-orphans` でも消さない。

## 11. 止め方と終了コード

- `stop_signal` → `Config.StopSignal`、`stop_grace_period` → `Config.StopTimeout`、`init: true` → `HostConfig.Init`。
- シグナル（既定は SIGTERM）を送り、猶予を過ぎたら SIGKILL するのは **Engine**。`-t` を指定しなければ、コンテナの `StopTimeout` が使われる。
- `kill` の既定は SIGKILL。
- **前景の `up` の Ctrl-C**: 1 回目は普通に止める。2 回目は kill。
- **`up` の終了を決めるフラグ**:
  - `--abort-on-container-exit`: どれかのコンテナが終わったら全体を止め、そのコンテナの終了コードを返す。
  - `--abort-on-container-failure`: exit 0 以外で終わったときだけ止める。
  - `--exit-code-from svc`: そのサービスの終了コードを返す（`--abort-on-container-failure` が無ければ、`--abort-on-container-exit` も付いたことになる）。
  - `-d` / `--wait` と、これらのフラグは同時に使えない。
- **気付いた点**（推定・未報告）:
  - Ctrl-C で止めるとき（`stopApplication`）は timeout を渡していない。そのため、**`up -t` は Ctrl-C のときには効かない**かもしれない。
  - `--abort-on-container-exit` は、restart ポリシーで再起動中のコンテナが死んだときにも止めてしまうかもしれない（`Restarting` を見ていない）。
  - `compose wait` の終了コードを、複数の goroutine がロック無しで書いている（データ競合）。
