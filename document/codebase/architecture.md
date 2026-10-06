# リポジトリのアーキテクチャ

ディレクトリ構成・レイヤー構成と、コード全体に共通する設計パターン。
CLI 層の詳細（cobra・トレーシング・プラグイン版/スタンドアロン版）は [cli.md](./cli.md)、
`internal/` 配下は [internal.md](./internal.md) を参照。

← [document 一覧](../README.md)

## ディレクトリ構造

```
compose/
├── cmd/                   # コマンドライン実装
│   ├── compose/           # Composeコマンド実装（45以上のコマンド）
│   ├── compatibility/     # v1互換性レイヤー
│   ├── display/           # UI表示ロジック
│   └── main.go            # エントリーポイント
│
├── pkg/                   # コアパッケージ
│   ├── api/               # APIインターフェース定義
│   ├── compose/           # メインCompose実装
│   ├── watch/             # ファイル監視機能
│   ├── remote/            # リモートリソース（Git、OCI）
│   └── e2e/               # E2Eテスト（約50ファイル）
│
├── internal/              # 内部パッケージ
│   ├── desktop/           # Docker Desktop統合
│   ├── tracing/           # 分散トレーシング
│   └── oci/               # OCIレジストリ操作
│
└── docs/                  # ドキュメント
    ├── reference/         # コマンドリファレンス（95+ファイル）
    └── examples/          # サンプルコード
```

| ディレクトリ | 役割 |
| --- | --- |
| `cmd/compose/` | CLI のコマンド定義（cobra）。フラグ解析とユーザー入力の受け取り |
| `pkg/api/` | Compose の操作を表すインターフェース定義 |
| `pkg/compose/` | 実際のロジック本体（up/down/build/収束処理など） |
| `pkg/e2e/` | E2E テスト |

## レイヤー構成（cmd / pkg/api / pkg/compose / internal）

```
CLI Layer (cmd/)
    ↓ コマンド解析・バリデーション
API Layer (pkg/api/)
    ↓ インターフェース定義
Service Layer (pkg/compose/)
    ↓ リソース収束・ライフサイクル管理
Docker Engine API
```

- `cmd` は CLI のコマンド定義・引数パース・表示ロジックを担当するエントリーポイント層。
- `pkg/api` は Compose の機能をインターフェースとして定義する層で、外部の Go プログラムからも SDK として利用可能。
- `pkg/compose` が実際のロジック（収束処理・Docker Engine API 呼び出し等）を持つ実装層。
- `internal` はこのリポジトリ専用の非公開パッケージ（Desktop 連携・トレーシング・OCI 操作等）を置く。

### 主要コンポーネント

**pkg/api/api.go**: Compose インターフェース

- Build, Push, Pull, Create, Start, Up, Down など多数のメソッド定義
- サードパーティアプリケーションが Compose をプログラムから利用可能

**pkg/compose/compose.go**: composeService 実装

- Docker Engine API とのやり取り
- リソースの収束処理
- 進捗トラッキング

**pkg/watch/**: ファイル監視機能（OS ごとの実装は [concepts/compose-concepts.md #watch 機能](../concepts/compose-concepts.md#watch-機能ファイル監視同期)）

## インターフェース駆動設計と SDK としての利用

- `pkg/api/api.go` の `Service` インターフェースが、CLI 層と実装層（`pkg/compose`）の境界になっている。
- テストではこのインターフェースをモック化し、Docker Engine に接続せずにロジックを検証できる。
- SDK としての利用時も、この抽象化のおかげで実装詳細を意識せずに Compose を組み込める。

```go
dockerCLI, _ := command.NewDockerCli()
service, _ := compose.NewComposeService(dockerCLI)
project, _ := service.LoadProject(ctx, api.ProjectLoadOptions{
    ConfigPaths: []string{"compose.yaml"},
})
service.Up(ctx, project, api.UpOptions{})
```

## 並行処理パターン（errgroup 等）

- `golang.org/x/sync/errgroup` を使い、複数 goroutine のエラーをまとめて収集しつつ並列実行する（errgroup 自体は [concepts/go.md #errgroup](../concepts/go.md#errgroup)）。
- サービスの並列起動・並列ビルド等、依存グラフ上で並列実行可能な単位ごとにこのパターンが使われる。
- 1 つでもエラーが出れば context をキャンセルし、他の goroutine に中断を伝える設計。

## プログレス表示・イベントシステム

- `up` / `build` 実行中の進捗を、TTY / plain / JSON 形式でリアルタイムに描画する仕組み。
- 内部的にはイベント（開始・進捗・完了等）を発行し、writer 実装がそれを購読して描画する構成。
- この仕組みにより、ターミナル出力だけでなく外部 UI（Docker Desktop 等）にも同じイベントを流用できる（リアルタイム進捗通知、カスタム UI 統合可能）。

## パフォーマンス最適化

- **並列処理**: 最大並列度設定可能（`COMPOSE_PARALLEL_LIMIT`）
- **増分更新**: 変更されたサービスのみ再作成
- **BuildKit キャッシュ**: ビルドキャッシュの効率的な活用
- **プログレス表示**: TTY/Plain/JSON 形式の出力

## OpenTelemetry によるトレーシング

→ [cli.md #トレーシング（OpenTelemetry）](./cli.md#トレーシングopentelemetry)
