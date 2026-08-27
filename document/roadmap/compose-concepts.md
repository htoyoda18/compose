# Docker Compose 特有の概念

← [roadmap 一覧](./README.md) ／ [document 一覧](../README.md)

- **compose-spec（compose.yaml 仕様）**
  - `compose.yaml` のスキーマやフィールドの意味論を定義する、Docker/OCI から独立したオープン仕様。
  - services / networks / volumes / configs / secrets など、書けるキーとその挙動の「正本」。
  - Compose はこの仕様の実装の一つで、他ツール（Kubernetes 変換ツール等）も同じ仕様を参照する。
- **compose-go（Loader / Merge / Interpolation）**
  - compose-spec を Go で実装したパーサー/ローダーライブラリ（`compose-spec/compose-go`）で、Compose 本体が依存として利用する。
  - Loader は複数の compose.yaml ファイルをパースし、内部モデルに変換する。
  - Merge は `-f` で複数ファイルを指定した際の上書き・マージルールを実装する。
  - Interpolation は `${VAR}` のような環境変数展開を担当する。
- **Project / Service / Network / Volume / Config / Secret モデル**
  - compose-go が YAML をパースした後の内部データモデル。Project がトップレベルで、その配下に Service やリソース定義が並ぶ。
  - このモデルが Compose 本体（`pkg/compose`）の入力となり、Docker Engine API への操作に変換される。
  - [docker.md](./docker.md) の「3 レイヤ」でいう Model 層に相当する。
- **収束処理（convergence）と「望ましい状態」への到達**
  - モデルに書かれた「望ましい状態」と、Docker Engine 上の「実際の状態」を比較し、差分を埋める処理。
  - `up` 実行時、既存リソースを再利用するか作り直す（recreate）かを判定する中心ロジック。
  - Kubernetes のコントローラパターンに近い設計思想。
- **依存関係グラフと起動順序制御**
  - `depends_on` 等のサービス間依存関係を有向グラフとして構築する。
  - トポロジカルソートにより、依存先から順に起動・依存元から順に停止する順序を決定する。
  - 循環依存の検出もこのグラフ構築時に行われる。
- **watch 機能（ファイル監視・同期）**
  - ソースコードの変更を検知して、自動でビルド・再起動・ファイル同期を行う開発者向け機能。
  - OS ごとに異なる監視実装（macOS は FSEvents 等）を抽象化している。
  - compose.yaml の `develop.watch` で sync/rebuild などのアクションをファイルパターンごとに指定する。
- **Provider 拡張機構**
  - `provider` 属性を使い、Compose 本体が知らない外部リソース（クラウドサービス等）をプラグイン的に扱う仕組み。
  - CLI プラグインまたは実行ファイルとして実装され、JSON 形式で Compose とやり取りする。
  - AWS/GCP/Azure のマネージドサービスなどを compose.yaml から宣言的に扱えるようにする。
- **スケーリング・レプリカ管理**
  - 1 つのサービス定義から複数のコンテナインスタンス（レプリカ）を起動・管理する仕組み。
  - `docker compose up --scale` や compose.yaml の `deploy.replicas` に対応する。
  - レプリカごとにコンテナ名へ連番のサフィックスを振り、ラベルで同一サービスのグループとして管理する。
- **ラベルによるリソース管理（状態 DB を持たない設計）**
  - Compose は自前の状態ストアを持たず、Docker Engine 上のリソースに付与したラベル（`com.docker.compose.project` 等）だけを頼りに管理対象を判定する。
  - `up` / `down` / `ps` 等はいずれもラベルによる filter 検索が起点になる。
  - この設計により、Compose CLI を再起動・別マシンから実行しても、ラベルさえ読めれば既存の状態を復元・把握できる。

