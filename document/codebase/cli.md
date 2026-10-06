# CLI 実装まわり

`cmd/` 配下の CLI 実装にかかわる概念（cobra、トレーシング、Docker コンテキスト、プラグイン版/スタンドアロン版）。

← [document 一覧](../README.md)

## cobra

- Go の代表的な CLI フレームワークで、`docker compose` のサブコマンド群 (`up`, `down`, `build` 等) はすべて `cobra.Command` として定義される
- `compose.go` の `RootCommand` がルートコマンドを構築し、`AddCommand` で各サブコマンドをツリー状に登録する
- フラグ解析・ヘルプ生成・シェル補完・`PersistentPreRunE` などのライフサイクルフックを提供する

## トレーシング（OpenTelemetry）

### OTel とは

- アプリケーションの分散トレーシング・メトリクス・ログを計測するためのベンダー中立な
  標準規格＋SDK群（ログ・メトリクス・トレースを統一的に扱うためのオブザーバビリティの標準規格）。
  特定の監視ベンダー（Datadog, Honeycomb等）にロックインされずに計装できるのが利点
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

### compose におけるトレースとは

- compose では分散トレーシングに利用している。各コマンド実行の span を生成し、内部処理の所要時間や呼び出し関係を可視化できるようにする
- `cmd/cmdtrace` が担う機能で、コマンド実行のたびにルートスパンを作り、実行全体を OpenTelemetry のトレースとして記録する
- コマンド終了時にスパンを終了・エクスポートし、外部のオブザーバビリティ基盤で CLI の実行を可視化できるようにする
- `internal/tracing` にラッパーがあり、tracer の初期化と OTLP exporter（gRPC/HTTP）への送出を担う。コマンドや API 呼び出しの主要ポイントに計装されている

### docker/compose での実装（`internal/tracing`, `cmd/cmdtrace`）

- `cmd/cmdtrace/cmd_span.go`の`Setup`が各CLIコマンド実行のたびに（`PersistentPreRunE` から）呼ばれ、コマンド名を
  スパン名にしたルートスパンを1つ作る（`otel.Tracer("").Start(...)`）
- `internal/tracing.InitTracing`が実際のトレーサー・エクスポータを初期化。
  設定元は2種類:
  - 標準の`OTEL_*`環境変数（`traceClientFromEnv`）
  - Docker contextのメタデータ（`traceClientFromDockerContext`）による自動検出（→ [Docker コンテキスト](#docker-コンテキスト)）
- コマンド終了時（成功/失敗どちらでも）に`wrapRunE`がスパンへ結果（成功/エラー/exit code）を
  記録し、`tracingShutdown`でスパンをflushしてエクスポータを終了する
- デフォルトでは、OTel SDK内部のエラー（`otel.ErrorHandler`）やshutdown時のエラーは
  CLIの通常出力を汚さないよう完全に握りつぶされる設計だったが、デバッグ目的で見えないと
  診断できない問題があった。`--debug`/`-D`を付けると、これらのエラー（エクスポート失敗など OTel 自体の内部動作）が
  `logrus.WithError(err).Debug(...)`経由でstderrに出力されるようになる
  （自分たちで実装したTODO対応。[PR #14152](https://github.com/docker/compose/pull/14152)。
  当初は専用の`COMPOSE_OTEL_DEBUG`環境変数を追加する案だったが、レビューで既存の
  `--debug`機構に寄せる方針に変更。あわせて`docker/cli`の`plugin.Run()`が起動時に
  `otel.SetErrorHandler`を呼び直し、`internal/tracing`側の`init()`によるハンドラ登録を
  無効化していた不具合も判明・解消した）

## Docker コンテキスト

- 接続先の Docker エンジン（ローカル/リモート、TLS 設定など）をひとまとめにした設定情報
- `docker context use` で切り替え、CLI はこの情報をもとに Docker Engine API へ接続する
- compose のトレース初期化でも、コンテキストのメタデータから OTel の送出先エンドポイントを自動検出する

## プラグイン版とスタンドアロン版

- **Docker CLI プラグイン機構**
  - `docker <subcommand>` として動作する外部バイナリを `~/.docker/cli-plugins/` に配置して追加できる仕組み。
  - `docker compose` 自体がこのプラグインとして実装されており、単体で動く `docker-compose` とは別に、`docker` からプラグインとして解決される。
  - プラグインはメタデータ（`docker-cli-plugin-metadata`）を返すことで、`docker` コマンドに自身を登録する。
- スタンドアロン版 docker-compose と CLI プラグイン版 docker compose
  - スタンドアロン版は単独の実行ファイル `docker-compose` として動作し、独自の引数形式（`--tls` などのグローバルフラグ）を持つ
  - プラグイン版は `docker` 本体からサブコマンド `compose` として呼び出される、現在推奨されている形態
  - `cmd/main.go` の `compatibility.Convert` がスタンドアロン形式の引数をプラグイン形式に変換し、内部的には同じ `RootCommand` で処理される
  - 両モードで同じバイナリを配布している（→ [development.md #リリース・バージョニング](./development.md#リリースバージョニング)）
- 名前の整理（`docker-compose` v1 / `docker compose` v2 / compose-spec）は [concepts/intro.md #4](../concepts/intro.md#4-名前まわりの整理混乱しやすい点)
