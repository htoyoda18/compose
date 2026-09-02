# Docker Compose v2 プロジェクト概要

## プロジェクトについて

Docker Compose v2 は、Docker 上でマルチコンテナアプリケーションを実行するための公式ツールです。Python 版の Compose v1 から Go で完全に書き直されたバージョンで、Docker CLI のプラグインとして動作します。

- **ライセンス**: Apache License 2.0
- **言語**: Go 1.24.11
- **総 Go ファイル数**: 279 ファイル
- **テストファイル数**: 89 ファイル
- **コア実装**: 約 16,558 行

## ルートファイル/ルートディレクト

- .github
  - GitHub 設定
- cmd
  - エントリーポイント
- docs
  - ドキュメント
- internal
  - 内部パッケージ
- pkg
  - 公開パッケージ
- .dockerignore
  - Docker ビルド時に無視するファイル
- .gitattributes
  - Git 属性の設定
- .gitignore
  - Git で追跡しないファイル/ディレクトリ
- .go-version
  - 使用する Go のバージョン
- .golangci.yml
  - コード品質チェックツール
- AGENTS.md
  - AI コーディングエージェント用のドキュメント
- BUILDING.md
  - ビルド方法の詳細説明
- codecov.yml
  - コードカバレッジ測定
- CONTRIBUTING.md
  - 貢献者向けのガイドライン
- docker-bake.hcl
  - Docker Buildx の設定ファイル
- Dockerfile
  - Docker Compose のバイナリをビルドする
- go.mod
  - Go モジュールの依存関係管理
- go.sum
  - Go モジュールの依存関係管理
- LICENSE
  - Apache License 2.0 のライセンス
- logo.png
  - Docker Compose のロゴ画像
- Makefile
  - ビルド、テスト、リリースなどの主要なコマンドを定義
- NOTICE
  - 著作権表示
- README.md
  - プロジェクトの概要説明

## 主な特徴

### 動作モード

1. **CLI プラグインモード**: `docker compose` コマンドとして動作
2. **スタンドアロンモード**: `docker-compose` 単体で動作

### コア機能

- **ライフサイクル管理**: up, down, start, stop, restart, pause/unpause
- **イメージ管理**: Buildx 統合、マルチプラットフォームビルド、pull/push
- **開発者向け機能**: watch（ファイル監視）、exec、logs、cp
- **高度な機能**: publish、viz、scale、config

### 提供コマンド

45 以上のコマンドを提供。各コマンドの説明は [docker.md](./docker.md#docker-compose-コマンド) を参照。

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

## アーキテクチャ

### レイヤー構造

```
CLI Layer (cmd/)
    ↓ コマンド解析・バリデーション
API Layer (pkg/api/)
    ↓ インターフェース定義
Service Layer (pkg/compose/)
    ↓ リソース収束・ライフサイクル管理
Docker Engine API
```

### 主要コンポーネント

**pkg/api/api.go**: Compose インターフェース

- Build, Push, Pull, Create, Start, Up, Down など多数のメソッド定義
- サードパーティアプリケーションが Compose をプログラムから利用可能

**pkg/compose/compose.go**: composeService 実装

- Docker Engine API とのやり取り
- リソースの収束処理
- 進捗トラッキング

**pkg/watch/**: ファイル監視機能

- macOS: FSEvents 使用
- Windows: 専用実装
- Linux: 汎用実装

## 技術スタック

### 主要な依存ライブラリ

**Docker エコシステム**:

- `github.com/docker/cli` v28.5.2
- `github.com/docker/docker` v28.5.2
- `github.com/docker/buildx` v0.30.1
- `github.com/compose-spec/compose-go/v2` v2.10.0

**CLI/UI**:

- `github.com/spf13/cobra` v1.10.2 - CLI フレームワーク
- `github.com/AlecAivazis/survey/v2` v2.3.7 - インタラクティブプロンプト

**監視/テレメトリ**:

- `go.opentelemetry.io/otel` v1.38.0 - OpenTelemetry 対応

**テスト**:

- `github.com/stretchr/testify` v1.11.1
- `go.uber.org/mock` v0.6.0

## ビルド・テスト・実行

### ビルド

```bash
# 基本ビルド
make build

# クロスコンパイル（11プラットフォーム対応）
make cross

# Dockerビルド
docker buildx bake binary
```

### テスト

```bash
# ユニットテスト
make test

# E2Eテスト（プラグインモード）
make e2e-compose

# E2Eテスト（スタンドアロンモード）
make e2e-compose-standalone

# 全テスト
make e2e
```

### 開発ワークフロー

```bash
# コード検証
make validate  # lint + vendor-validate + headers + docs

# プリコミットチェック
make pre-commit  # validate + lint + build + test + e2e
```

### インストール

```bash
# ローカルインストール（~/.docker/cli-plugins/にインストール）
make install
```

## 対応プラットフォーム

- darwin/amd64, darwin/arm64
- linux/amd64, linux/arm/v6, linux/arm/v7, linux/arm64
- linux/ppc64le, linux/riscv64, linux/s390x
- windows/amd64, windows/arm64

計 11 プラットフォーム

## 拡張性

### Provider 拡張機能

- サービスの`provider`属性で外部プロバイダーを指定可能
- CLI プラグインまたは実行ファイルとして実装
- JSON 形式で Compose と通信
- AWS/GCP/Azure などのクラウドサービスと統合可能

### SDK としての利用

```go
dockerCLI, _ := command.NewDockerCli()
service, _ := compose.NewComposeService(dockerCLI)
project, _ := service.LoadProject(ctx, api.ProjectLoadOptions{
    ConfigPaths: []string{"compose.yaml"},
})
service.Up(ctx, project, api.UpOptions{})
```

## パフォーマンス最適化

- **並列処理**: 最大並列度設定可能（`COMPOSE_PARALLEL_LIMIT`）
- **増分更新**: 変更されたサービスのみ再作成
- **BuildKit キャッシュ**: ビルドキャッシュの効率的な活用
- **プログレス表示**: TTY/Plain/JSON 形式の出力

## 可観測性

- **OpenTelemetry 統合**: 分散トレーシング対応
- **OTLP exporter**: gRPC/HTTP 対応
- **イベントシステム**: リアルタイム進捗通知、カスタム UI 統合可能

## CI/CD

GitHub Actions で以下を実行:

- **validate**: lint、vendor、headers、docs の検証
- **binary**: マルチプラットフォームバイナリビルド
- **test**: ユニットテストとカバレッジ
- **e2e**: E2E テスト

## まとめ

Docker Compose v2 は、Go で実装された高性能なマルチコンテナオーケストレーションツールです。Python 版から完全に書き直され、Docker CLI プラグインとして動作しながらも、SDK として他のアプリケーションから利用することも可能です。

包括的なテストスイート、11 プラットフォーム対応、拡張可能なアーキテクチャ、OpenTelemetry 対応など、エンタープライズグレードの品質を備えており、開発から本番環境まで幅広く使用されています。

## コードの読み方

### Phase 1: エントリーポイントの理解

1. cmd/main.go (86 行)
   ↓ main() → pluginMain() → RootCommand()
2. cmd/compose/compose.go:424 (RootCommand)
   ↓ 全サブコマンドを登録
3. cmd/compose/up.go (upCommand)
   ↓ 1 つのコマンドを深堀り
4. pkg/api/api.go
   ↓ Service インターフェース定義を確認
5. pkg/compose/compose.go
   ↓ Service インターフェースの実装

### Phase 2: 1 つのコマンドを完全に理解

docker compose up から始める

### Phase 3: 横展開 - 他のコマンド

優先度 高:
├─ down.go # up の逆、削除ロジック
├─ build.go # Buildkit 連携
├─ logs.go # ストリーミング処理
└─ run.go # 一時コンテナ

優先度 中:
├─ exec.go # 実行中コンテナ操作
├─ ps.go # 状態取得
└─ scale.go # レプリカ制御

### Phase 4: 深い部分を理解

Docker API 連携
Compose ファイルパース
並行制御・同期
