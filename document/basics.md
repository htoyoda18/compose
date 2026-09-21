# 入門: そもそも Docker / Docker Compose とは

前提知識ゼロから読み始めるための入門メモ。個々の用語の詳細や実装との対応は
[docker.md](./docker.md)、このリポジトリの構成は [note.md](./note.md)、
内部理解のための学習順序は [roadmap/](./roadmap/README.md) を参照。

## 1. Docker とは何か

**アプリを「動く環境ごと」パッケージ化して、どのマシンでも同じように動かすための仕組み。**

### 解決したい問題

- 「自分の PC では動くのに、サーバーでは動かない」
  - OS のバージョン、ライブラリの有無、環境変数、ミドルウェアの設定などが違うため
- 環境構築手順書が長大になり、誰がやっても同じ結果にならない
- 1 台のサーバーで複数アプリを動かすと、依存ライブラリのバージョンが衝突する

### Docker のアプローチ

アプリ本体 + 依存ライブラリ + OS のユーザーランド一式を 1 つの**イメージ**に固め、
それを**コンテナ**として起動する。イメージは「どこに持っていっても同じ中身」なので、
開発 PC / CI / 本番で同じものが動く。

ただし「どこでも動く」には条件がある（後述の
[Docker は Linux 上でしか動かない](#docker-は-linux-上でしか動かない)）。

### VM との違い

| | 仮想マシン (VM) | コンテナ (Docker) |
| --- | --- | --- |
| 抽象化の対象 | コンピュータそのもの（CPU/メモリ/ディスクを仮想化） | プロセス（ホスト OS のカーネルを共有） |
| ゲスト OS | 必要（カーネルごと起動） | 不要（カーネルはホストと共有） |
| 起動時間 | 数十秒〜分 | 数百ミリ秒〜秒 |
| オーバーヘッド | 大きい | 小さい |

※ この比較は**ホストが Linux の場合**の話。Mac / Windows では Docker 自身が
Linux VM の上で動くので、VM のコストは別途かかる。

コンテナは「隔離されたプロセス」にすぎない。隔離は Linux カーネルの機能で実現している。

- **namespace**: プロセス・ネットワーク・ファイルシステムなどの「見える範囲」を隔離する
- **cgroup**: CPU・メモリなどの「使える量」を制限する
- **capability**: root 権限を細分化して、必要な分だけ与える

→ 詳細は [docker.md #コンテナ基盤技術](./docker.md)

### Docker は Linux 上でしか動かない

上のとおり、コンテナの隔離は **Linux カーネルの機能そのもの**（namespace / cgroup /
capability / overlayfs など）で実現している。つまりコンテナは「Linux カーネルの上で動く
Linux プロセス」であり、**Linux カーネルがなければ動かない**。

にもかかわらず Mac や Windows で `docker run` が使えるのは、裏側で**Linux VM を起動して
いる**から。「コンテナは VM より軽い」という説明は Linux ホスト上での話で、
Mac / Windows では VM 1 台分のコストが土台として常にかかっている。

| ホスト OS | 実際の動き |
| --- | --- |
| Linux | ホストのカーネルをそのまま使う。VM なし |
| macOS | Docker Desktop が軽量な Linux VM（Apple Virtualization.framework 等）を立て、その中の `dockerd` が動く。`docker` CLI だけが Mac 側 |
| Windows | 同様に Linux VM を使う（WSL2 バックエンド、または Hyper-V） |

Mac / Windows で困りやすい点は、ほぼこの「VM 境界」が原因。

- **ファイル共有が遅い**: ホストのディレクトリを bind mount すると、VM 境界を越える
  ファイルシステム共有（virtiofs / gRPC-FUSE など）になるため、ネイティブより遅い
- **`localhost` の意味がずれる**: コンテナ内の `localhost` は VM 内のそれ。ホストを
  指したいときは `host.docker.internal` を使う
- **`--network host` が期待どおりに動かない**: 「ホスト」は Mac ではなく Linux VM
- **inode 監視やファイル変更検知が取りこぼす**: `docker compose watch` の
  ファイル監視が Mac で不安定になりやすいのはこれが背景

### Windows コンテナという例外

Windows には「Windows カーネルの上で動く Windows コンテナ」も存在し、Docker Desktop の
コンテナモードを切り替えて使う。ただし、

- Linux イメージと Windows イメージは**互換性がない**（カーネルが違うので相互に動かない）
- Windows コンテナはホストの Windows バージョンとの対応がかなり厳しい

ため、一般に「Docker」と言った場合はほぼ Linux コンテナを指す。このリポジトリが
主に扱うのも Linux コンテナ。

### アーキテクチャ (CPU) も揃っている必要がある

カーネルだけでなく CPU アーキテクチャも一致が必要。`linux/amd64` のイメージは
そのままでは `linux/arm64`（Apple Silicon など）で動かない。

- イメージは `linux/amd64`, `linux/arm64` のように **OS + アーキテクチャ**単位で存在する
- 1 つのタグで複数アーキテクチャを指す仕組みがマルチプラットフォームイメージ
  （マニフェストリスト）。`docker pull` はホストに合うものを自動で選ぶ
- 合わないものは QEMU によるエミュレーションで一応動くが、大幅に遅い

→ Compose 側のプラットフォーム解決は
[docker.md #ビルド](./docker.md) を参照

### 基本用語

- **イメージ (image)**: アプリと依存一式のスナップショット。読み取り専用のテンプレート
- **コンテナ (container)**: イメージを起動した実行中のインスタンス
- **Dockerfile**: イメージのビルド手順を書いたテキストファイル
- **レジストリ (registry)**: イメージを保管・配布する場所（Docker Hub, GHCR など）
- **ボリューム (volume)**: コンテナを消しても残したいデータの保存先
- **ネットワーク (network)**: コンテナ同士を通信させるための仮想ネットワーク

### 動かすときの構成要素

```
docker CLI  ──(REST API / Unix socket)──>  dockerd  ──>  containerd  ──>  runc  ──> コンテナ
（コマンド）                              （デーモン）  （ランタイム管理）（実際に隔離して起動）
```

- ユーザーが叩く `docker` は CLI にすぎず、実際の処理は常駐デーモン `dockerd` が行う
- ビルドは `BuildKit`（`docker buildx`）が担当する

### dockerd とは

**Docker Engine の本体であるデーモンプロセス**。`docker` コマンドの実体はここにある。

#### 役割

- **REST API サーバー**: `docker` CLI からのリクエストを受け付ける HTTP API を公開する。
  待ち受け先はデフォルトで Unix ドメインソケット `/var/run/docker.sock`
  （TCP で公開する設定も可能だが、実質ホストの root 権限を渡すことになるので危険）
- **上位オブジェクトの管理**: イメージ、コンテナ、ボリューム、ネットワークといった
  「Docker の概念」を管理するのは dockerd。コンテナの状態や設定を保持している
- **ネットワーク構築**: bridge ネットワークの作成、iptables/nftables 操作による
  ポートフォワード（`-p 8080:80`）、コンテナ名の DNS 解決
- **ボリューム管理**: ボリュームの作成・マウント・ドライバの扱い
- **イメージの取得と保管**: レジストリ認証、pull/push、レイヤーの展開と保管
- **ログ収集**: コンテナの stdout/stderr を logging driver 経由で集める（`docker logs`）
- **下位ランタイムへの委譲**: 実際のコンテナ起動は自分でやらず containerd に任せる

#### containerd / runc との分担

もともと dockerd が全部やっていたが、OCI 標準化にあわせて層が分離された。

| コンポーネント | 担当 |
| --- | --- |
| `dockerd` | Docker としての高レベル機能（API, イメージ, ネットワーク, ボリューム, ログ） |
| `containerd` | コンテナのライフサイクル管理（作成/開始/停止/監視）、イメージの実体管理。常駐デーモン |
| `containerd-shim` | コンテナ 1 つごとに常駐し、親デーモンとコンテナを切り離す。これにより dockerd を再起動しても稼働中コンテナは死なない |
| `runc` | 最終的に namespace / cgroup を設定してプロセスを起動する。OCI Runtime Spec の実装。起動したら終了する短命プロセス |

つまり `docker run` は
`docker CLI → (API) → dockerd → (gRPC) → containerd → shim → runc → コンテナ`
という流れ。

#### 押さえておくと役立つ性質

- **クライアント / サーバー構成である**: CLI とデーモンは別プロセスなので、
  リモートの dockerd を操作することもできる（`DOCKER_HOST`, `docker context`）。
  Mac / Windows で動くのはまさにこれで、CLI はホスト側、dockerd は Linux VM 側にいる
- **root で動いている**: namespace/cgroup 操作に特権が必要なため。
  `docker.sock` へのアクセス権 ≒ root 権限、という点はセキュリティ上の定番の注意点
  （権限を落として動かす rootless モードもある）
- **設定ファイルは `/etc/docker/daemon.json`**: ログドライバ、レジストリミラー、
  デフォルトアドレスプールなどを指定する
- **状態の持ち主は dockerd**: Compose が自前の状態 DB を持たず、ラベル検索で
  自分の管理対象を見つけられるのは、dockerd が状態を持っていてくれるから
  （→ [プロジェクトという単位](#プロジェクトという単位)）

#### Docker Compose から見た dockerd

このリポジトリのコードも、結局は **dockerd の API を叩くクライアント**にすぎない。

- `compose.yaml` を読んで「望ましい状態」を作る
- dockerd の API に問い合わせて「実際の状態」を得る（ラベルで絞り込み）
- 差分を埋めるように dockerd の API を呼ぶ（コンテナ作成/削除/起動…）

→ 詳細は [roadmap/docker-engine.md](./roadmap/docker-engine.md)

### 最小の流れ

```bash
docker build -t myapp .   # Dockerfile からイメージを作る
docker run -p 8080:80 myapp   # イメージからコンテナを起動する
docker ps                 # 動いているコンテナを確認する
docker stop <id>          # 止める
```

## 2. なぜ Docker だけでは足りないのか

現実のアプリは 1 コンテナでは終わらない。たとえば Web アプリなら:

- Web アプリ本体
- データベース (PostgreSQL など)
- キャッシュ (Redis など)
- リバースプロキシ (nginx など)

これを `docker run` だけでやると、こうなる。

```bash
docker network create myapp-net
docker volume create db-data
docker run -d --name db --network myapp-net -v db-data:/var/lib/postgresql/data \
  -e POSTGRES_PASSWORD=secret postgres:16
docker run -d --name cache --network myapp-net redis:7
docker run -d --name web --network myapp-net -p 8080:80 \
  -e DATABASE_URL=postgres://db:5432 myapp
```

問題点:

- コマンドが長く、手順として人に渡せない（履歴や README に埋もれる）
- 起動順序（DB が先、Web が後）を人間が覚えておく必要がある
- 止めるとき・消すときも同じ数のコマンドが必要で、消し忘れが起きる
- 設定を変えたいとき、どのコンテナを作り直せばいいのか分からない

## 3. Docker Compose とは何か

**複数コンテナの構成を 1 つの YAML ファイルに宣言して、まとめて起動・停止・管理するためのツール。**

上の例は `compose.yaml` 1 枚になる。

```yaml
services:
  web:
    build: .
    ports:
      - "8080:80"
    environment:
      DATABASE_URL: postgres://db:5432
    depends_on:
      - db
      - cache
  db:
    image: postgres:16
    environment:
      POSTGRES_PASSWORD: secret
    volumes:
      - db-data:/var/lib/postgresql/data
  cache:
    image: redis:7

volumes:
  db-data:
```

操作はこれだけになる。

```bash
docker compose up -d   # 全部まとめて起動（ネットワークもボリュームも自動作成）
docker compose ps      # 状態確認
docker compose logs -f # ログをまとめて追う
docker compose down    # まとめて停止・削除
```

### 命令的 (imperative) から宣言的 (declarative) へ

これが Compose の本質。

- `docker run` は**命令的**: 「何をするか」を手順として並べる
- `compose.yaml` は**宣言的**: 「どうあってほしいか（望ましい状態）」を書く

Compose は「YAML に書かれた望ましい状態」と「Docker Engine の実際の状態」を比較し、
**差分だけを埋める**ように動く。だから `docker compose up` は何度打っても安全で、
すでに正しく動いているコンテナはそのまま残り、定義が変わったものだけ作り直される
（recreate）。この差分埋め＝**収束処理 (convergence)** が Compose の中核ロジック。

### プロジェクトという単位

Compose は「サービス集合 + 付随するネットワーク/ボリューム」の束を
**プロジェクト (project)** として扱う。プロジェクト名はデフォルトでディレクトリ名。

Compose は自前の状態データベースを持たない。代わりに、作成したリソースすべてに
`com.docker.compose.project` などの**ラベル**を付け、次回はラベルで絞り込んで
「自分が管理しているリソース」を見つけ出す。Docker Engine 自体が状態の保管場所になっている。

→ 詳細は [docker.md #Docker Compose の内部アーキテクチャ](./docker.md)

### Docker Compose が提供する主なもの

- サービス定義（イメージ指定 or ビルド、環境変数、ポート、ボリューム）
- サービス間の依存関係と起動順序 (`depends_on`)
- サービス名での名前解決（`web` から `db` というホスト名で DB に繋がる）
- 複数ファイルの合成・環境ごとの上書き (`-f`, `override`, profiles)
- 開発向けの機能（`watch` によるファイル変更検知と自動再ビルド、`run` による使い捨てコンテナ）

## 4. 名前まわりの整理（混乱しやすい点）

- **`docker-compose` (v1)**: Python 実装の旧バージョン。ハイフン付き。EOL
- **`docker compose` (v2)**: Go 実装の現行バージョン。サブコマンド形式。**このリポジトリがそれ**
  - `docker` CLI の**プラグイン**として動くのが基本形
  - `docker-compose` という名前の**スタンドアロンバイナリ**としても配布されており、
    v1 からの移行のため同じ実装が両方の形で動く
- **Compose Specification (compose-spec)**: `compose.yaml` の書式そのものの仕様。
  実装（このリポジトリ）とは別リポジトリで管理されている
- **ファイル名**: `compose.yaml` が現在の推奨。`docker-compose.yml` は後方互換で読まれる

→ 詳細は [cli.md](./cli.md)（プラグイン版/スタンドアロン版の差異）

## 5. このリポジトリは何なのか

`docker/compose` は **`docker compose` コマンドの実装**（Go）。

ざっくり言うと、やっていることは次の 3 段。

1. `compose.yaml` を読んでパース・検証し、内部表現（Project モデル）にする
   → 主に別リポジトリの `compose-go` が担当
2. 内部表現と Docker Engine の実際の状態を比較し、差分を計算する
3. Docker Engine API を叩いて、コンテナ・ネットワーク・ボリュームを作る/消す/作り直す

| ディレクトリ | 役割 |
| --- | --- |
| `cmd/compose/` | CLI のコマンド定義（cobra）。フラグ解析とユーザー入力の受け取り |
| `pkg/api/` | Compose の操作を表すインターフェース定義 |
| `pkg/compose/` | 実際のロジック本体（up/down/build/収束処理など） |
| `pkg/e2e/` | E2E テスト |

→ 詳細は [note.md](./note.md) と [roadmap/architecture.md](./roadmap/architecture.md)

## 次に読むもの

1. [docker.md](./docker.md) — 用語・コマンド・内部アーキテクチャの詳細メモ
2. [note.md](./note.md) — このリポジトリの構成と読み方
3. [roadmap/README.md](./roadmap/README.md) — 内部実装まで理解するための学習ロードマップ

← [document 一覧](./README.md)
