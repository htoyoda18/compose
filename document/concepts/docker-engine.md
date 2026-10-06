# Docker エンジンの理解

dockerd / containerd / runc の関係と、Engine API・BuildKit・ボリュームなど Docker エンジンが提供する機能。
全体像（`docker CLI → dockerd → containerd → runc`）は [intro.md #動かすときの構成要素](./intro.md#動かすときの構成要素) を参照。

← [document 一覧](../README.md)

## dockerd とは

**Docker Engine の本体であるデーモンプロセス**。`docker` コマンドの実体はここにある。

### 役割

- **REST API サーバー**: `docker` CLI からのリクエストを受け付ける HTTP API を公開する。
  待ち受け先はデフォルトで Unix ドメインソケット `/var/run/docker.sock`
  （TCP で公開する設定も可能だが、実質ホストの root 権限を渡すことになるので危険）
- **上位オブジェクトの管理**: イメージ、コンテナ、ボリューム、ネットワークといった
  「Docker の概念」を管理するのは dockerd。コンテナの状態や設定を保持している
- **ネットワーク構築**: bridge ネットワークの作成、iptables/nftables 操作による
  ポートフォワード（`-p 8080:80`）、コンテナ名の DNS 解決（→ [networking.md](./networking.md)）
- **ボリューム管理**: ボリュームの作成・マウント・ドライバの扱い
- **イメージの取得と保管**: レジストリ認証、pull/push、レイヤーの展開と保管（→ [oci.md](./oci.md)）
- **ログ収集**: コンテナの stdout/stderr を logging driver 経由で集める（`docker logs`）
- **下位ランタイムへの委譲**: 実際のコンテナ起動は自分でやらず containerd に任せる

### containerd / runc との分担

もともと dockerd が全部やっていたが、OCI 標準化にあわせて層が分離された。

| コンポーネント | 担当 |
| --- | --- |
| `dockerd` | Docker としての高レベル機能（API, イメージ, ネットワーク, ボリューム, ログ） |
| `containerd` | コンテナのライフサイクル管理（作成/開始/停止/監視）、イメージの実体管理。常駐デーモン。containerd-shim を介して各コンテナプロセスを監視する |
| `containerd-shim` | コンテナ 1 つごとに常駐し、親デーモンとコンテナを切り離す。これにより dockerd/containerd を再起動しても稼働中コンテナは死なない |
| `runc` | 最終的に namespace / cgroup を設定してプロセスを起動する。OCI Runtime Spec の実装（低レベルランタイム）。起動したら終了する短命プロセス |

つまり `docker run` は
`docker CLI → (API) → dockerd → (gRPC) → containerd → shim → runc → コンテナ`
という流れ。

### 押さえておくと役立つ性質

- **クライアント / サーバー構成である**: CLI とデーモンは別プロセスなので、
  リモートの dockerd を操作することもできる（`DOCKER_HOST`, `docker context`。
  → [codebase/cli.md #Docker コンテキスト](../codebase/cli.md#docker-コンテキスト)）。
  Mac / Windows で動くのはまさにこれで、CLI はホスト側、dockerd は Linux VM 側にいる
- **root で動いている**: namespace/cgroup 操作に特権が必要なため。
  `docker.sock` へのアクセス権 ≒ root 権限、という点はセキュリティ上の定番の注意点
  （権限を落として動かす rootless モードもある）
- **設定ファイルは `/etc/docker/daemon.json`**: ログドライバ、レジストリミラー、
  デフォルトアドレスプールなどを指定する
- **状態の持ち主は dockerd**: Compose が自前の状態 DB を持たず、ラベル検索で
  自分の管理対象を見つけられるのは、dockerd が状態を持っていてくれるから
  （→ [compose-concepts.md #ラベルによるリソース管理](./compose-concepts.md#ラベルによるリソース管理状態-db-を持たない設計)）

## Docker Engine API（REST）

- dockerd が Unix ソケット（または TCP）越しに公開する REST API。
- `docker` / `docker compose` CLI はこの API を叩いてコンテナ・イメージ・ネットワーク・ボリュームを操作する。
- Compose 本体もこの API 経由でリソースを作成・削除しており、`pkg/api` の実装の多くは API クライアント呼び出しになる。

### Docker Compose から見た dockerd

このリポジトリのコードも、結局は **dockerd の API を叩くクライアント**にすぎない。

- `compose.yaml` を読んで「望ましい状態」を作る
- dockerd の API に問い合わせて「実際の状態」を得る（ラベルで絞り込み）
- 差分を埋めるように dockerd の API を呼ぶ（コンテナ作成/削除/起動…）

## BuildKit / Buildx

- **BuildKit**
  - 従来の Docker ビルドを置き換える、並列実行・キャッシュ効率に優れた次世代ビルドエンジン。
  - ビルド処理をグラフ（DAG）として表現し、依存関係のないステップを並列実行できる。
  - Docker CLI / Compose からは buildx を経由して呼び出される。
  - **LLB（Low-Level Build definition）**
    - BuildKit が内部で扱う、ビルド処理を表す低レベルの中間表現（DAG）。
    - Dockerfile 等のフロントエンドはユーザー入力を LLB に変換し、BuildKit は LLB を最適化・実行する。
  - **Dockerfile frontend**
    - LLB への変換を担うプラガブルな「フロントエンド」の一つで、Dockerfile の構文を解釈する。
    - フロントエンドを差し替えることで、Dockerfile 以外の記法でもビルドグラフを生成できる設計になっている。
  - **ビルドキャッシュ**
    - 各ビルドステップの入力（コマンドやファイル内容）からキャッシュキーを計算し、変化がなければ再実行をスキップする仕組み。
    - `--cache-to` / `--cache-from` によるキャッシュのエクスポート/インポートで、CI など別環境間でも共有できる。
- **Docker Buildx**
  - Docker の次世代ビルド機能を CLI から使いやすくした拡張
  - 複数アーキテクチャ向けのマルチプラットフォームビルドを 1 コマンドで実行
  - Compose 側のプラットフォーム解決は [compose-concepts.md #ビルドとプラットフォーム解決](./compose-concepts.md#ビルドとプラットフォーム解決) を参照

## ボリュームドライバ

- コンテナのライフサイクルとは独立して永続化されるデータ領域（ボリューム）の実体を管理する仕組み。
- ローカルドライバはホストのファイルシステム上にデータを保持するが、NFS やクラウドストレージ向けのドライバも存在する。
- Compose の `volumes` 定義は、最終的にこのボリュームドライバへの操作に変換される。

## Docker CLI プラグイン機構

`docker compose` 自体がこの仕組みで動いているので、[codebase/cli.md #プラグイン版とスタンドアロン版](../codebase/cli.md#プラグイン版とスタンドアロン版) にまとめた。
