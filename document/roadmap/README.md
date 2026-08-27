# Docker / Docker Compose 内部理解ロードマップ

Docker および Docker Compose の OSS 内部実装まで理解するために必要な概念の一覧。詳細は各項目を学習しながら追記していく。

内容が長くなったため、セクションごとにファイルを分割している。

| ファイル                                          | 内容                                       |
| -------------------------------------------------- | ------------------------------------------ |
| [prerequisites.md](./prerequisites.md)             | 前提知識（Go / Linux カーネル / ネットワーキング / OCI 標準規格 / YAML / Git） |
| [docker-engine.md](./docker-engine.md)             | Docker エンジンの理解（dockerd/containerd/runc, BuildKit 等） |
| [compose-concepts.md](./compose-concepts.md)       | Docker Compose 特有の概念（compose-spec, 収束処理, watch 等） |
| [architecture.md](./architecture.md)               | リポジトリのアーキテクチャ（レイヤー構成, 並行処理パターン等）   |
| [development.md](./development.md)                 | 開発・運用（E2E テスト, Lint/CI, リリース, コントリビューションフロー） |

← [document 一覧](../README.md)
