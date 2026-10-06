# Docker Compose 特有の概念

compose.yaml の仕様からリソース管理・収束処理・各種機能まで、Compose 固有の概念。
「なぜ Compose が必要か」「宣言的とは何か」という入門的な話は [intro.md](./intro.md) を参照。

← [document 一覧](../README.md)

## compose-spec と compose-go

- **compose-spec（compose.yaml 仕様）**
  - `compose.yaml` のスキーマやフィールドの意味論を定義する、Docker/OCI から独立したオープン仕様。
  - services / networks / volumes / configs / secrets など、書けるキーとその挙動の「正本」。
  - Compose はこの仕様の実装の一つで、他ツール（Kubernetes 変換ツール等）も同じ仕様を参照する。
- **YAML**
  - インデントによる構造表現、アンカー/エイリアス、マルチドキュメントなどの基本文法。
  - `compose.yaml` はこの上に compose-spec 独自の意味論が乗る形なので、まず YAML 自体の挙動を理解しておく。
  - タブ不可・スカラー型の曖昧な解釈（`yes`/`no` 等）といった落とし穴も把握しておくと良い。
- **compose-go（Loader / Merge / Interpolation）**
  - compose-spec を Go で実装したパーサー/ローダーライブラリ（`compose-spec/compose-go`）で、Compose 本体が依存として利用する。
  - Loader は複数の compose.yaml ファイルをパースし、内部モデルに変換する。
  - Merge は `-f` で複数ファイルを指定した際の上書き・マージルールを実装する。
  - Interpolation は `${VAR}` のような環境変数展開を担当する。

## Compose の「3 レイヤ」とモデル

- Spec: ユーザーが書く宣言（compose.yaml）。services / networks / volumes / configs / secrets…
- Model: YAML をパース・正規化した“内部表現”
- Runtime: 最終的に Container / Network / Volume として作られる世界

**Project / Service / Network / Volume / Config / Secret モデル**（Model 層）

- compose-go が YAML をパースした後の内部データモデル。Project がトップレベルで、その配下に Service やリソース定義が並ぶ。
- このモデルが Compose 本体（`pkg/compose`）の入力となり、Docker Engine API への操作に変換される。

## Project という単位

- 「プロジェクト」= サービス集合 + 付随リソースの束
- プロジェクト名はデフォルトでディレクトリ名（→ [intro.md #プロジェクトという単位](./intro.md#プロジェクトという単位)）

## ラベルによるリソース管理（状態 DB を持たない設計）

- Compose は Docker Engine 上に“状態 DB”を持てない。自前の状態ストアを持たず、Docker Engine 上のリソースに付与したラベル（`com.docker.compose.project` 等）だけを頼りに管理対象を判定する。
- 内部で compose 管理下のリソースを探す処理は、ほぼ filters + labels。`up` / `down` / `ps` 等はいずれもラベルによる filter 検索が起点になる。
- この設計により、Compose CLI を再起動・別マシンから実行しても、ラベルさえ読めれば既存の状態を復元・把握できる。

## 収束処理（convergence）と「望ましい状態」への到達

- Compose のコアはコントローラ
- モデルに書かれた「望ましい状態」と、Docker Engine 上の「実際の状態」を比較し、差分を埋める処理。
- `up` 実行時、既存リソースを再利用するか作り直す（recreate）かを判定する中心ロジック。
  - 既存があれば差分で recreate
  - 依存関係順に起動順も制御
- config-hash と recreate の条件
- Kubernetes のコントローラパターンに近い設計思想。

## 依存関係グラフと起動順序制御

- 内部ではサービスをグラフとして扱う。`depends_on` 等のサービス間依存関係を有向グラフとして構築する。
- トポロジカルソートにより、依存先から順に起動・依存元から順に停止する順序を決定する。
- 循環依存の検出もこのグラフ構築時に行われる。

## スケーリング・レプリカ管理

- 1 つのサービス定義から複数のコンテナインスタンス（レプリカ）を起動・管理する仕組み。
- `docker compose up --scale` や compose.yaml の `deploy.replicas` に対応する。
- レプリカごとにコンテナ名へ連番のサフィックスを振り、ラベルで同一サービスのグループとして管理する。

## watch 機能（ファイル監視・同期）

- ソースコードの変更を検知して、自動でビルド・再起動・ファイル同期を行う開発者向け機能。
- compose.yaml の `develop.watch` で sync/rebuild などのアクションをファイルパターンごとに指定する。
- OS ごとに異なる監視実装を抽象化している（`pkg/watch/`）
  - macOS: FSEvents 使用
  - Windows: 専用実装
  - Linux: 汎用実装
- Mac で不安定になりやすい背景は [intro.md #Docker は Linux 上でしか動かない](./intro.md#docker-は-linux-上でしか動かない)

## Provider 拡張機構

- サービスの `provider` 属性を使い、Compose 本体が知らない外部リソース（クラウドサービス等）をプラグイン的に扱う仕組み。
- CLI プラグインまたは実行ファイルとして実装され、JSON 形式で Compose とやり取りする。
- AWS/GCP/Azure のマネージドサービスなどを compose.yaml から宣言的に扱えるようにする。

## ビルドとプラットフォーム解決

- マルチプラットフォームビルドとplatform解決（docker/compose内部）
  - `build.platforms`（サービスがビルドをサポートするプラットフォーム一覧）と
    `service.platform`（実行時に使うプラットフォーム）は別概念
  - `docker compose build` は複数プラットフォームでのビルドを許容するが、
    `up`/`create`/`run`/`watch`/`config` は単一プラットフォームでの実行が前提
    （`buildForSinglePlatform`フラグで制御）
  - `service.platform`が未指定かつ`build.platforms`が複数ある場合、最終的に
    「ビルダーに選択を委ねる」ためリストを空にする分岐があるが、ビルダーが実際に
    選ぶプラットフォームが宣言済みリストに含まれているかは検証されない、という
    既知の検証漏れがある（`cmd/compose/options.go`のTODO、未対応。詳細な調査は [tracking/todo.md](../tracking/todo.md)）
- ビルドエンジン（BuildKit / Buildx）は [docker-engine.md #BuildKit / Buildx](./docker-engine.md#buildkit--buildx)

## docker compose コマンド

| コマンド  | 説明                                             |
| --------- | ------------------------------------------------ |
| `up`      | コンテナを起動                                    |
| `down`    | コンテナ・ネットワークを停止＆削除                |
| `start`   | 停止中のコンテナを再起動                          |
| `stop`    | コンテナを停止                                    |
| `ps`      | Compose 配下のコンテナ一覧                        |
| `logs`    | ログを見る                                        |
| `top`     | コンテナ内プロセス一覧                            |
| `build`   | イメージをビルド                                  |
| `pull`    | イメージを取得                                    |
| `exec`    | 起動中コンテナに入る                              |
| `run`     | 一時コンテナを起動してコマンド実行                |
| `restart` | 再起動                                            |
| `config`  | compose.yaml を展開・検証                         |
| `ls`      | Compose プロジェクト一覧                          |
| `rm`      | 停止中コンテナを削除                              |
| `events`  | イベント監視                                      |
| `attach`  | 起動中コンテナの標準入力・出力に接続する（挙動をそのまま見る） |
| `watch`   | ファイル監視と自動再ビルド                        |

- oneOff
  - `docker compose run` で起動される、一時的・使い捨てのコンテナを指す概念
- IO ストリームと TTY
- docker compose コマンドの解説チャット: https://chatgpt.com/c/695374b5-06e8-8324-97b3-f423942ba35a
