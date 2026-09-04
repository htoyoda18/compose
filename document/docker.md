# Docker / Docker Compose 学習ノート

Docker および Docker Compose の一般的な概念に関する学習ノート。このリポジトリ（docker/compose）固有の構成・アーキテクチャは [note.md](./note.md)、既知の FIXME/TODO 調査は [fixme.md](./fixme.md) / [todo.md](./todo.md) を参照。

## 標準規格

- OCI (Open Container Initiative)
  - コンテナ技術の標準仕様
  - この形式で作れば、どのコンテナ環境でも動く
- OCI Image Spec 1.1 と 1.0 の違い
  - 1.1 でマニフェストに `artifactType` フィールドが追加され、コンテナイメージ以外の
    任意アーティファクト（Compose定義ファイル一式など）を、実体のないconfigの代わりに
    `artifactType` で明示的に表現できるようになった
  - 1.0（＝古いレジストリ/Distribution仕様）は `artifactType` を認識しない。代わりに
    config media typeを見て種別を判別する慣習で後方互換を取る
  - docker/composeの`publish`では、まず1.1形式でpushを試み、レジストリが理解できず
    4xx（authエラー以外）を返した場合のみ1.0形式にフォールバックする、という設計になっている
    （`internal/oci/push.go`）

## docker compose コマンド

| コマンド  | 説明                                             |
| --------- | ------------------------------------------------ |
| `up`      | コンテナを起動                                    |
| `down`    | コンテナ・ネットワークを停止＆削除                |
| `start`   | 停止中のコンテナを再起動                          |
| `stop`    | コンテナを停止                                    |
| `ps`      | Compose 配下のコンテナ一覧                        |
| `logs`    | ログを見る                                        |
| `top`     | コンテナ内プロセス一覧                            |
| `build`   | イメージをビルド                                  |
| `pull`    | イメージを取得                                    |
| `exec`    | 起動中コンテナに入る                              |
| `run`     | 一時コンテナを起動してコマンド実行                |
| `restart` | 再起動                                            |
| `config`  | compose.yaml を展開・検証                         |
| `ls`      | Compose プロジェクト一覧                          |
| `rm`      | 停止中コンテナを削除                              |
| `events`  | イベント監視                                      |
| `attach`  | 起動中コンテナの標準入力・出力に接続する（挙動をそのまま見る） |
| `watch`   | ファイル監視と自動再ビルド                        |

- oneOff
  - `docker compose run` で起動される、一時的・使い捨てのコンテナを指す概念

## Docker Compose の内部アーキテクチャ

- Compose の「3 レイヤ」
  - Spec: ユーザーが書く宣言（compose.yaml）。services / networks / volumes / configs / secrets…
  - Model: YAML をパース・正規化した“内部表現”
  - Runtime: 最終的に Container / Network / Volume として作られる世界
- Project という単位
  - 「プロジェクト」= サービス集合 + 付随リソースの束
- ラベル設計
  - Compose は Docker Engine 上に“状態 DB”を持てない
  - 内部で compose 管理下のリソースを探す処理は、ほぼ filters + labels
- 望ましい状態と実際の状態
  - Compose のコアはコントローラ
  - 望ましい状態 / 実際の状態 の差分を埋める
  - up の動作
    - 既存があれば差分で recreate
    - 依存関係順に起動順も制御
- config-hash と recreate の条件
- 依存関係のグラフ
  - 内部ではサービスをグラフとして扱う
- IO ストリームと TTY

## ビルド

- Docker Buildx
  - Docker の次世代ビルド機能を CLI から使いやすくした拡張
  - 複数アーキテクチャ向けのマルチプラットフォームビルドを 1 コマンドで実行
- マルチプラットフォームビルドとplatform解決（docker/compose内部）
  - `build.platforms`（サービスがビルドをサポートするプラットフォーム一覧）と
    `service.platform`（実行時に使うプラットフォーム）は別概念
  - `docker compose build` は複数プラットフォームでのビルドを許容するが、
    `up`/`create`/`run`/`watch`/`config` は単一プラットフォームでの実行が前提
    （`buildForSinglePlatform`フラグで制御）
  - `service.platform`が未指定かつ`build.platforms`が複数ある場合、最終的に
    「ビルダーに選択を委ねる」ためリストを空にする分岐があるが、ビルダーが実際に
    選ぶプラットフォームが宣言済みリストに含まれているかは検証されない、という
    既知の検証漏れがある（`cmd/compose/options.go`のTODO、未対応）

## OpenTelemetry (OTel)

- OTel とは
  - アプリケーションの分散トレーシング・メトリクス・ログを計測するためのベンダー中立な
    標準規格＋SDK群。特定の監視ベンダー（Datadog, Honeycomb等）にロックインされずに
    計装できるのが利点
- 基本概念
  - **Trace**: 1つのリクエスト/コマンド実行全体を表す一連の処理のまとまり
  - **Span**: Trace を構成する個々の作業単位（開始・終了時刻、属性、ステータスを持つ）。
    親子関係を持ちツリー状に連なる
  - **Context propagation**: 呼び出しを跨いでTrace/Spanの文脈情報を伝搬する仕組み
    （HTTPヘッダ等に埋め込む）
  - **Exporter**: 収集したTrace/Spanをバックエンド（Collectorやベンダー）に送信する部品
  - **OTLP (OpenTelemetry Protocol)**: SpanやMetricsをExporterからCollector/バックエンドへ
    送る際の標準プロトコル（gRPC/HTTP）
  - SDKとAPIの分離: アプリコードは薄い"API"に対して計装し、実際の収集・エクスポート方式は
    "SDK"側の設定で差し替えられる、という設計思想
- docker/composeでの実装（`internal/tracing`, `cmd/cmdtrace`）
  - `cmd/cmdtrace/cmd_span.go`の`Setup`が各CLIコマンド実行のたびに呼ばれ、コマンド名を
    スパン名にしたルートスパンを1つ作る（`otel.Tracer("").Start(...)`）
  - `internal/tracing.InitTracing`が実際のトレーサー・エクスポータを初期化。
    設定元は2種類:
    - 標準の`OTEL_*`環境変数（`traceClientFromEnv`）
    - Docker contextのメタデータ（`traceClientFromDockerContext`）による自動検出
  - コマンド終了時（成功/失敗どちらでも）に`wrapRunE`がスパンへ結果（成功/エラー/exit code）を
    記録し、`tracingShutdown`でスパンをflushしてエクスポータを終了する
  - デフォルトでは、OTel SDK内部のエラー（`otel.ErrorHandler`）やshutdown時のエラーは
    CLIの通常出力を汚さないよう完全に握りつぶされる設計だったが、デバッグ目的で見えないと
    診断できない問題があった。`--debug`/`-D`を付けると、これらのエラーが
    `logrus.WithError(err).Debug(...)`経由でstderrに出力されるようになる
    （自分たちで実装したTODO対応。[PR #14152](https://github.com/docker/compose/pull/14152)。
    当初は専用の`COMPOSE_OTEL_DEBUG`環境変数を追加する案だったが、レビューで既存の
    `--debug`機構に寄せる方針に変更。あわせて`docker/cli`の`plugin.Run()`が起動時に
    `otel.SetErrorHandler`を呼び直し、`internal/tracing`側の`init()`によるハンドラ登録を
    無効化していた不具合も判明・解消した）

## コンテナ基盤技術

- VM と Docker
  - VM はコンピュータ自体を抽象化する
  - Docker はプロセス自体の抽象化をする
- cgroup
  - メモリ・CPU のような計算リソースを隔離するための機能
  - コンテナはそれぞれ専用の計算リソースを割り当て、他のコンテナにはお互いにアクセスできないようにする
- namespace
  - プロセスやネットワーク、ファイルアクセスなど複数の種類がある
  - それぞれが異なるリソースの隔離を行う
- Capability
  - スーパーユーザーとしての権限を制限するための機能
- Docker Image
  - 任意のタイミングのスナップショット
- Immutable Infrastructure
  - サーバー内への変更を行わないアプローチ
  - 変更や追加が発生する場合
    - 新しく構築してスナップショットを保存
    - スナップショットを元にサーバーを新しく立ち上げる

## その他の用語

- パストラバーサル攻撃
  - 本来アクセスできないファイルを、URL や入力値を細工して読み取る攻撃
- MAC アドレス
  - ネットワーク機器ごとに割り当てられる固有の識別番号

## 参考リンク

- https://y-ohgi.com/introduction-docker/2_component/image/
- docker compose コマンドの解説チャット: https://chatgpt.com/c/695374b5-06e8-8324-97b3-f423942ba35a
