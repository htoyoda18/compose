# Docker エンジンの理解

← [roadmap 一覧](./README.md) ／ [document 一覧](../README.md)

- **dockerd / containerd / runc / shim の関係**
  - dockerd は API サーバーとして高レベルの操作を受け付け、コンテナのライフサイクル管理は containerd に委譲する。
  - containerd はイメージ管理やコンテナの実行制御を担当するデーモンで、containerd-shim を介して各コンテナプロセスを監視する。
  - runc は OCI Runtime Spec に従って namespace/cgroup を設定し、実際にプロセスを起動する低レベルランタイム。
  - shim は dockerd/containerd が再起動してもコンテナプロセスが道連れにならないよう、親プロセスとして常駐する。
- **Docker Engine API（REST）**
  - dockerd が Unix ソケット（または TCP）越しに公開する REST API。
  - `docker` / `docker compose` CLI はこの API を叩いてコンテナ・イメージ・ネットワーク・ボリュームを操作する。
  - Compose 本体もこの API 経由でリソースを作成・削除しており、`pkg/api` の実装の多くは API クライアント呼び出しになる。
- **イメージの layer 構造と manifest**
  - Docker イメージは複数の読み取り専用レイヤー（差分ファイルシステム）の重ね合わせで構成される。
  - 各レイヤーは digest で一意に識別される content-addressable なデータで、複数イメージ間で共有・再利用される。
  - manifest はレイヤーの並び順や実行時設定（entrypoint, env 等）をまとめたメタデータで、イメージの pull/push 単位となる。
- **レジストリと pull/push**
  - レジストリはイメージ（manifest とレイヤー）を保存・配布するサーバー（Docker Hub, GHCR 等）。
  - pull は必要なレイヤーだけをダウンロードし、ローカルにキャッシュ済みのレイヤーは再利用する。
  - push はローカルにあってレジストリ側に無いレイヤーのみアップロードする、差分転送が基本。
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
- **ネットワークドライバ（bridge / overlay / macvlan）**
  - bridge は単一ホスト内でのコンテナ間通信用のデフォルトドライバ。
  - overlay は複数ホストにまたがるコンテナ間通信を実現する（Swarm 等で使用）。
  - macvlan はコンテナに物理 NIC 相当の MAC アドレスを割り当て、外部ネットワークに直接参加させる。
- **ボリュームドライバ**
  - コンテナのライフサイクルとは独立して永続化されるデータ領域（ボリューム）の実体を管理する仕組み。
  - ローカルドライバはホストのファイルシステム上にデータを保持するが、NFS やクラウドストレージ向けのドライバも存在する。
  - Compose の `volumes` 定義は、最終的にこのボリュームドライバへの操作に変換される。
- **Docker CLI プラグイン機構**
  - `docker <subcommand>` として動作する外部バイナリを `~/.docker/cli-plugins/` に配置して追加できる仕組み。
  - `docker compose` 自体がこのプラグインとして実装されており、単体で動く `docker-compose` とは別に、`docker` からプラグインとして解決される。
  - プラグインはメタデータ（`docker-cli-plugin-metadata`）を返すことで、`docker` コマンドに自身を登録する。

