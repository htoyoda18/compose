# Docker Compose v2 プロジェクト概要

## プロジェクトについて

Docker Compose v2は、Docker上でマルチコンテナアプリケーションを実行するための公式ツールです。Python版のCompose v1からGoで完全に書き直されたバージョンで、Docker CLIのプラグインとして動作します。

- **ライセンス**: Apache License 2.0
- **言語**: Go 1.24.11
- **総Goファイル数**: 279ファイル
- **テストファイル数**: 89ファイル
- **コア実装**: 約16,558行

## 主な特徴

### 動作モード
1. **CLIプラグインモード**: `docker compose` コマンドとして動作
2. **スタンドアロンモード**: `docker-compose` 単体で動作

### コア機能
- **ライフサイクル管理**: up, down, start, stop, restart, pause/unpause
- **イメージ管理**: Buildx統合、マルチプラットフォームビルド、pull/push
- **開発者向け機能**: watch（ファイル監視）、exec、logs、cp
- **高度な機能**: publish、viz、scale、config

### 提供コマンド
45以上のコマンドを提供し、主要なものは以下の通り:
- `up` - サービスの作成と起動
- `down` - サービスの停止と削除
- `build` - イメージのビルド
- `ps` - コンテナのリスト表示
- `logs` - ログの表示
- `exec` - コンテナ内でコマンド実行
- `watch` - ファイル監視と自動再ビルド

## ディレクトリ構造

```
compose/
├── cmd/                    # コマンドライン実装
│   ├── compose/           # Composeコマンド実装（45以上のコマンド）
│   ├── compatibility/      # v1互換性レイヤー
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

**pkg/api/api.go**: Composeインターフェース
- Build, Push, Pull, Create, Start, Up, Down など多数のメソッド定義
- サードパーティアプリケーションがComposeをプログラムから利用可能

**pkg/compose/compose.go**: composeService実装
- Docker Engine APIとのやり取り
- リソースの収束処理
- 進捗トラッキング

**pkg/watch/**: ファイル監視機能
- macOS: FSEvents使用
- Windows: 専用実装
- Linux: 汎用実装

## 技術スタック

### 主要な依存ライブラリ

**Dockerエコシステム**:
- `github.com/docker/cli` v28.5.2
- `github.com/docker/docker` v28.5.2
- `github.com/docker/buildx` v0.30.1
- `github.com/compose-spec/compose-go/v2` v2.10.0

**CLI/UI**:
- `github.com/spf13/cobra` v1.10.2 - CLIフレームワーク
- `github.com/AlecAivazis/survey/v2` v2.3.7 - インタラクティブプロンプト

**監視/テレメトリ**:
- `go.opentelemetry.io/otel` v1.38.0 - OpenTelemetry対応

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

計11プラットフォーム

## 拡張性

### Provider拡張機能
- サービスの`provider`属性で外部プロバイダーを指定可能
- CLIプラグインまたは実行ファイルとして実装
- JSON形式でComposeと通信
- AWS/GCP/Azureなどのクラウドサービスと統合可能

### SDKとしての利用
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
- **BuildKitキャッシュ**: ビルドキャッシュの効率的な活用
- **プログレス表示**: TTY/Plain/JSON形式の出力

## 可観測性

- **OpenTelemetry統合**: 分散トレーシング対応
- **OTLP exporter**: gRPC/HTTP対応
- **イベントシステム**: リアルタイム進捗通知、カスタムUI統合可能

## CI/CD

GitHub Actionsで以下を実行:
- **validate**: lint、vendor、headers、docsの検証
- **binary**: マルチプラットフォームバイナリビルド
- **test**: ユニットテストとカバレッジ
- **e2e**: E2Eテスト

## まとめ

Docker Compose v2は、Goで実装された高性能なマルチコンテナオーケストレーションツールです。Python版から完全に書き直され、Docker CLIプラグインとして動作しながらも、SDKとして他のアプリケーションから利用することも可能です。

包括的なテストスイート、11プラットフォーム対応、拡張可能なアーキテクチャ、OpenTelemetry対応など、エンタープライズグレードの品質を備えており、開発から本番環境まで幅広く使用されています。
