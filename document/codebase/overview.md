# Docker Compose v2 プロジェクト概要

このリポジトリ（docker/compose）自体の概要。構成・アーキテクチャは [architecture.md](./architecture.md)、
ビルド・テストは [development.md](./development.md)、コードの読み方は [reading-guide.md](./reading-guide.md) を参照。

← [document 一覧](../README.md)

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
- [cmd](../../cmd/note.md)
  - エントリーポイント
- docs
  - ドキュメント
- [internal](../../internal/note.md)
  - 内部パッケージ
- [pkg](../../pkg/note.md)
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

CLI プラグインモード（`docker compose`）とスタンドアロンモード（`docker-compose`）の 2 つ。
→ [cli.md #プラグイン版とスタンドアロン版](./cli.md#プラグイン版とスタンドアロン版)

### コア機能

- **ライフサイクル管理**: up, down, start, stop, restart, pause/unpause
- **イメージ管理**: Buildx 統合、マルチプラットフォームビルド、pull/push
- **開発者向け機能**: watch（ファイル監視）、exec、logs、cp
- **高度な機能**: publish、viz、scale、config

### 提供コマンド

45 以上のコマンドを提供。各コマンドの説明は [concepts/compose-concepts.md](../concepts/compose-concepts.md#docker-compose-コマンド) を参照。

### 拡張性

- Provider 拡張機能 → [concepts/compose-concepts.md #Provider 拡張機構](../concepts/compose-concepts.md#provider-拡張機構)
- SDK としての利用 → [architecture.md #インターフェース駆動設計と SDK としての利用](./architecture.md#インターフェース駆動設計と-sdk-としての利用)

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

## まとめ

Docker Compose v2 は、Go で実装された高性能なマルチコンテナオーケストレーションツールです。Python 版から完全に書き直され、Docker CLI プラグインとして動作しながらも、SDK として他のアプリケーションから利用することも可能です。

包括的なテストスイート、11 プラットフォーム対応、拡張可能なアーキテクチャ、OpenTelemetry 対応など、エンタープライズグレードの品質を備えており、開発から本番環境まで幅広く使用されています。
