# OCI 標準とイメージ・レジストリ

コンテナイメージの形式・起動方法・配布方法を定める標準規格と、それに沿ったイメージ／レジストリの仕組み。
OCI の 3 仕様の短い要約は [intro-qa.md #5](./intro-qa.md#5-oci-標準の具体的な中身) にもある。
このリポジトリでの実装（`internal/oci`, `internal/registry`）は [codebase/internal.md](../codebase/internal.md) を参照。

← [document 一覧](../README.md)

## OCI (Open Container Initiative)

- コンテナ技術の標準仕様
- この形式で作れば、どのコンテナ環境でも動く

### OCI Image Spec

- コンテナイメージのフォーマット（レイヤー構成・manifest・config）を定義する仕様。
- Docker イメージもこの仕様に準拠しており、他のツール（Podman 等）とも互換性がある。
- manifest がレイヤーの digest リストと実行時の設定情報を持つ。

### OCI Runtime Spec

- コンテナをどう起動するか（namespace/cgroup/mount 設定等）を定義する仕様。
- `runc` はこの仕様のリファレンス実装であり、`config.json` を受け取ってコンテナを起動する。
- dockerd/containerd から見ると、この仕様に従うランタイムを差し替え可能。

### OCI Distribution Spec（レジストリ API）

- イメージの pull/push を行うレジストリ HTTP API の仕様。
- Docker Hub や GHCR など各種レジストリはこの仕様に準拠している。
- `pkg/remote` のような OCI レジストリ操作コードを読む際の前提知識になる。

### OCI Image Spec 1.1 と 1.0 の違い

- 1.1 でマニフェストに `artifactType` フィールドが追加され、コンテナイメージ以外の
  任意アーティファクト（Compose定義ファイル一式など）を、実体のないconfigの代わりに
  `artifactType` で明示的に表現できるようになった
- 1.0（＝古いレジストリ/Distribution仕様）は `artifactType` を認識しない。代わりに
  config media typeを見て種別を判別する慣習で後方互換を取る
- docker/composeの`publish`では、まず1.1形式でpushを試み、レジストリが理解できず
  4xx（authエラー以外）を返した場合のみ1.0形式にフォールバックする、という設計になっている
  （`internal/oci/push.go`。フォールバック時の警告の TODO 対応は [tracking/todo.md](../tracking/todo.md) を参照）

## イメージの layer 構造と manifest

- Docker イメージは複数の読み取り専用レイヤー（差分ファイルシステム）の重ね合わせで構成される。
- 各レイヤーは digest で一意に識別される content-addressable なデータで、複数イメージ間で共有・再利用される。
- manifest はレイヤーの並び順や実行時設定（entrypoint, env 等）をまとめたメタデータで、イメージの pull/push 単位となる。
- レイヤーの重ね合わせを実現する overlayfs は [container-tech.md](./container-tech.md) を参照。

## レジストリと pull/push

- レジストリはイメージ（manifest とレイヤー）を保存・配布するサーバー（Docker Hub, GHCR 等）。
- pull は必要なレイヤーだけをダウンロードし、ローカルにキャッシュ済みのレイヤーは再利用する。
- push はローカルにあってレジストリ側に無いレイヤーのみアップロードする、差分転送が基本。
