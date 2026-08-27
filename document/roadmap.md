# Docker / Docker Compose 内部理解ロードマップ

Docker および Docker Compose の OSS 内部実装まで理解するために必要な概念の一覧。詳細は各項目を学習しながら追記していく。

## 前提知識

- Go 言語
  - goroutine / channel
  - context
  - interface 設計
  - error wrapping（`fmt.Errorf %w`, errors.Is/As）
  - go modules / vendoring
- Linux カーネルの基礎
  - namespace
  - cgroup
  - capability
  - chroot / pivot_root
  - overlayfs
- ネットワーキングの基礎
  - bridge / veth
  - iptables / nftables
  - DNS 解決
- コンテナ関連の標準規格
  - OCI Image Spec
  - OCI Runtime Spec
  - OCI Distribution Spec（レジストリ API）
- YAML
- Git / GitHub ワークフロー（OSS コントリビューションの基本作法）

## Docker エンジンの理解

- dockerd / containerd / runc / shim の関係
- Docker Engine API（REST）
- イメージの layer 構造と manifest
- レジストリと pull/push
- BuildKit
  - LLB（Low-Level Build definition）
  - Dockerfile frontend
  - ビルドキャッシュ
- ネットワークドライバ（bridge / overlay / macvlan）
- ボリュームドライバ
- Docker CLI プラグイン機構

## Docker Compose 特有の概念

- compose-spec（compose.yaml 仕様）
- compose-go（Loader / Merge / Interpolation）
- Project / Service / Network / Volume / Config / Secret モデル
- 収束処理（convergence）と「望ましい状態」への到達
- 依存関係グラフと起動順序制御
- watch 機能（ファイル監視・同期）
- Provider 拡張機構
- スケーリング・レプリカ管理
- ラベルによるリソース管理（状態 DB を持たない設計）

## リポジトリのアーキテクチャ

- レイヤー構成（cmd / pkg/api / pkg/compose / internal）
- インターフェース駆動設計
- 並行処理パターン（errgroup 等）
- プログレス表示・イベントシステム
- OpenTelemetry によるトレーシング

## 開発・運用

- E2E テスト（Scenario DSL）
- Lint / CI（golangci-lint, GitHub Actions）
- リリース・バージョニング
- コントリビューションフロー（DCO, PR テンプレート）
