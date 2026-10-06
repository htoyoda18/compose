# document/

Docker Compose リポジトリを学習・調査した際の個人メモ置き場。
1 つのトピックは 1 つのファイルにだけ書き、関連する話はリンクでたどる。

## concepts/ — 一般的な概念（このリポジトリに依存しない）

| ファイル | 内容 |
| --- | --- |
| [intro.md](./concepts/intro.md) | 入門: そもそも Docker とは何か / Docker Compose とは何か（前提知識ゼロから読む用） |
| [intro-qa.md](./concepts/intro-qa.md) | intro.md を読んで出た疑問の Q&A（FE とコンテナ、apple/container、M1、Linux 一強、OCI、DB） |
| [container-tech.md](./concepts/container-tech.md) | コンテナ基盤技術（namespace / cgroup / capability / chroot / overlayfs、Immutable Infrastructure） |
| [networking.md](./concepts/networking.md) | ネットワーキング（bridge / veth / iptables / DNS、ネットワークドライバ） |
| [oci.md](./concepts/oci.md) | OCI 標準（Image / Runtime / Distribution Spec、1.1 と 1.0）、イメージのレイヤーとレジストリ |
| [docker-engine.md](./concepts/docker-engine.md) | Docker エンジン（dockerd / containerd / runc、Engine API、BuildKit / Buildx、ボリュームドライバ） |
| [compose-concepts.md](./concepts/compose-concepts.md) | Compose 特有の概念（compose-spec、3 レイヤ、ラベル、収束処理、依存グラフ、watch、provider、コマンド一覧） |
| [go.md](./concepts/go.md) | Go 言語（goroutine / channel / errgroup / context / interface / error wrapping / modules） |
| [go-qa.md](./concepts/go-qa.md) | go.md を読んで出た疑問の Q&A（goroutine leak、buffered channel、他言語のキャンセル、Go の設計思想、go.mod / go.sum） |

## codebase/ — このリポジトリ（docker/compose）の構成

| ファイル | 内容 |
| --- | --- |
| [overview.md](./codebase/overview.md) | プロジェクト概要（ルートファイル、主な特徴、技術スタック） |
| [architecture.md](./codebase/architecture.md) | ディレクトリ構造・レイヤー構成、インターフェース駆動設計と SDK、並行処理、イベントシステム |
| [cli.md](./codebase/cli.md) | CLI 実装（cobra、トレーシング / OpenTelemetry、Docker コンテキスト、プラグイン版 / スタンドアロン版） |
| [internal.md](./codebase/internal.md) | internal/ 配下（インメモリソケット、PID ファイル、実行時ディレクトリ、レジストリ、マニフェスト / blob、シンボリックリンク） |
| [development.md](./codebase/development.md) | ビルド・テスト（E2E Scenario DSL）・Lint / CI・リリース |
| [reading-guide.md](./codebase/reading-guide.md) | コードを読む順番 |

## contributing/ — PR を出すための情報

| ファイル | 内容 |
| --- | --- |
| [README.md](./contributing/README.md) | コントリビューションの基本作法（Git / DCO / フロー） |
| [checklist.md](./contributing/checklist.md) | PR を出す前のチェックリスト（下の 2 つから統合） |
| [pr-trends.md](./contributing/pr-trends.md) | 直近のクローズ済み PR 150 件と main のコミット履歴の調査。どんな PR がマージされ／閉じられているか、コミットの積み方 |
| [pr-review-patterns.md](./contributing/pr-review-patterns.md) | 自分が出した PR のレビュー指摘の傾向分析と対策 |

## tracking/ — 調査状況・タスク候補（随時更新）

| ファイル | 内容 |
| --- | --- |
| [next-tasks.md](./tracking/next-tasks.md) | pr-trends.md の基準で選んだ、まだ誰も着手していないタスク候補（検証済みの上位 6 件と、次点・やらないもの） |
| [issues.md](./tracking/issues.md) | 個別 issue の調査状況 |
| [todo.md](./tracking/todo.md) | リポジトリ内 TODO コメントの調査状況 |
| [fixme.md](./tracking/fixme.md) | リポジトリ内 FIXME コメントの調査状況（随時再スキャンして更新） |

## 学習ロードマップ（読む順）

OSS の内部実装まで理解するための順序。詳細は各項目を学習しながら追記していく。

1. [concepts/intro.md](./concepts/intro.md) — 全体像
2. 前提知識: [go.md](./concepts/go.md) → [container-tech.md](./concepts/container-tech.md) → [networking.md](./concepts/networking.md) → [oci.md](./concepts/oci.md)
3. [concepts/docker-engine.md](./concepts/docker-engine.md) — Docker エンジンの理解
4. [concepts/compose-concepts.md](./concepts/compose-concepts.md) — Docker Compose 特有の概念
5. [codebase/overview.md](./codebase/overview.md) → [architecture.md](./codebase/architecture.md) → [reading-guide.md](./codebase/reading-guide.md) — リポジトリのアーキテクチャ
6. [codebase/development.md](./codebase/development.md) と [contributing/](./contributing/README.md) — 開発・運用とコントリビューション
