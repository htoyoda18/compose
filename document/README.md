# document/

Docker Compose リポジトリを学習・調査した際の個人メモ置き場。
1 つのトピックは 1 つのファイルにだけ書き、関連する話はリンクでたどる。

## concepts/ — 一般的な概念（このリポジトリに依存しない）

| ファイル | 内容 |
| --- | --- |
| [intro.md](./concepts/intro.md) | 入門: そもそも Docker とは何か / Docker Compose とは何か（前提知識ゼロから読む用） |
| [intro-qa.md](./concepts/intro-qa.md) | intro.md を読んで出た疑問の Q&A（FE とコンテナ、apple/container、M1、Linux 一強、OCI、DB） |
| [container-tech.md](./concepts/container-tech.md) | コンテナ基盤技術（namespace / cgroup / capability / chroot / overlayfs、Immutable Infrastructure） |
| [container-tech-qa.md](./concepts/container-tech-qa.md) | container-tech.md を読んで出た疑問の Q&A（マウントの仕組み、mnt / uts / ipc、cgroup の上限超過、ファイルシステムとは、virtiofs、ホスト名） |
| [networking.md](./concepts/networking.md) | ネットワーキング（bridge / veth / iptables / DNS、ネットワークドライバ） |
| [networking-qa.md](./concepts/networking-qa.md) | networking.md を読んで出た疑問の Q&A（ネットワークドライバ、パケットフィルタリング、NAT、IP アドレスの割り振り、MAC と IP のレイヤー） |
| [oci.md](./concepts/oci.md) | OCI 標準（Image / Runtime / Distribution Spec、1.1 と 1.0）、イメージのレイヤーとレジストリ |
| [oci-qa.md](./concepts/oci-qa.md) | oci.md を読んで出た疑問の Q&A（manifest、digest、コンテナの起動の流れ、Dockerfile とレイヤー） |
| [docker-engine.md](./concepts/docker-engine.md) | Docker エンジン（dockerd / containerd / runc、Engine API、BuildKit / Buildx、ボリュームドライバ） |
| [docker-engine-qa.md](./concepts/docker-engine-qa.md) | docker-engine.md を読んで出た疑問の Q&A（dockerd の居場所、状態の保存先、shim、ボリュームの実体、ポートフォワード、BuildKit、rootless、セキュリティ） |
| [compose-concepts.md](./concepts/compose-concepts.md) | Compose 特有の概念（compose-spec、3 レイヤ、ラベル、収束処理、依存グラフ、watch、provider、コマンド一覧） |
| [compose-lifecycle.md](./concepts/compose-lifecycle.md) | Compose の意味論（depends_on、healthcheck、フック、作り直しの判定と config hash、ラベル、読み込み順と環境変数、ボリューム、ネットワーク、run、停止と終了コード） |
| [go.md](./concepts/go.md) | Go 言語（goroutine / channel / errgroup / context / interface / error wrapping / modules） |
| [go-qa.md](./concepts/go-qa.md) | go.md を読んで出た疑問の Q&A（goroutine leak、buffered channel、他言語のキャンセル、Go の設計思想、go.mod / go.sum） |

## codebase/ — このリポジトリ（docker/compose）の構成

| ファイル | 内容 |
| --- | --- |
| [overview.md](./codebase/overview.md) | プロジェクト概要（ルートファイル、主な特徴、技術スタック） |
| [architecture.md](./codebase/architecture.md) | ディレクトリ構造・レイヤー構成、インターフェース駆動設計と SDK、並行処理、イベントシステム |
| [plan-engine.md](./codebase/plan-engine.md) | plan engine（観測 → 計画 → 実行）。`up` などが差分を埋める仕組みと、どのコマンドがそれを通るか |
| [cli.md](./codebase/cli.md) | CLI 実装（cobra、トレーシング / OpenTelemetry、Docker コンテキスト、プラグイン版 / スタンドアロン版） |
| [internal.md](./codebase/internal.md) | internal/ 配下（インメモリソケット、PID ファイル、実行時ディレクトリ、レジストリ、マニフェスト / blob、シンボリックリンク） |
| [development.md](./codebase/development.md) | ビルド・テスト（E2E Scenario DSL）・Lint / CI・リリース |
| [reading-guide.md](./codebase/reading-guide.md) | コードを読む順番 |
| [testing.md](./codebase/testing.md) | テストの作法（gotest.tools、gomock、Scenario DSL）、make のターゲット、CI の実際、落とし穴 |

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
| [next-tasks.md](./tracking/next-tasks.md) | pr-trends.md の基準で選んだ、まだ誰も着手していないタスク候補（検証済みの上位 13 件と、保留・次点・見守るもの・やらないもの。2026-10-16 更新） |
| [issues.md](./tracking/issues.md) | 個別 issue の調査状況 |
| [todo.md](./tracking/todo.md) | リポジトリ内 TODO コメントの調査状況 |
| [fixme.md](./tracking/fixme.md) | リポジトリ内 FIXME コメントの調査状況（随時再スキャンして更新） |

## 学習ロードマップ（読む順）

上から順に読む。（Q&A）は直前のファイルを読んで疑問が出たときに開けばよい。

1. [concepts/intro.md](./concepts/intro.md)
   - （Q&A）[concepts/intro-qa.md](./concepts/intro-qa.md)
2. [concepts/container-tech.md](./concepts/container-tech.md)
   - （Q&A）[concepts/container-tech-qa.md](./concepts/container-tech-qa.md)
3. [concepts/networking.md](./concepts/networking.md)
   - （Q&A）[concepts/networking-qa.md](./concepts/networking-qa.md)
4. [concepts/oci.md](./concepts/oci.md)
   - （Q&A）[concepts/oci-qa.md](./concepts/oci-qa.md)
5. [concepts/docker-engine.md](./concepts/docker-engine.md)
   - （Q&A）[concepts/docker-engine-qa.md](./concepts/docker-engine-qa.md)
6. [concepts/compose-concepts.md](./concepts/compose-concepts.md)
7. [concepts/compose-lifecycle.md](./concepts/compose-lifecycle.md)
8. [concepts/go.md](./concepts/go.md)
   - （Q&A）[concepts/go-qa.md](./concepts/go-qa.md)
9. [codebase/overview.md](./codebase/overview.md)
10. [codebase/architecture.md](./codebase/architecture.md)
11. [codebase/cli.md](./codebase/cli.md)
12. [codebase/plan-engine.md](./codebase/plan-engine.md)
13. [codebase/internal.md](./codebase/internal.md)
14. [codebase/reading-guide.md](./codebase/reading-guide.md)
15. [codebase/development.md](./codebase/development.md)
16. [codebase/testing.md](./codebase/testing.md)
17. [pkg/e2e/SCENARIO.md](../pkg/e2e/SCENARIO.md)
18. [contributing/README.md](./contributing/README.md)
19. [CONTRIBUTING.md](../CONTRIBUTING.md) と [AI_POLICY.md](../AI_POLICY.md)
20. [contributing/pr-trends.md](./contributing/pr-trends.md)
21. [contributing/pr-review-patterns.md](./contributing/pr-review-patterns.md)
22. [contributing/checklist.md](./contributing/checklist.md)
23. [tracking/next-tasks.md](./tracking/next-tasks.md)
24. [tracking/issues.md](./tracking/issues.md)
25. [tracking/todo.md](./tracking/todo.md)
26. [tracking/fixme.md](./tracking/fixme.md)
