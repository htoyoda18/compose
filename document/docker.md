# Docker / Docker Compose 学習ノート

Docker および Docker Compose の一般的な概念に関する学習ノート。このリポジトリ（docker/compose）固有の構成・アーキテクチャは [note.md](./note.md)、既知の FIXME/TODO 調査は [fixme.md](./fixme.md) / [todo.md](./todo.md) を参照。

## 標準規格

- OCI (Open Container Initiative)
  - コンテナ技術の標準仕様
  - この形式で作れば、どのコンテナ環境でも動く

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

## Docker Compose の内部アーキテクチャ

- Compose の「3 レイヤ」
  - Spec: ユーザーが書く宣言（compose.yaml）。services / networks / volumes / configs / secrets…
  - Model: YAML をパース・正規化した“内部表現”
  - Runtime: 最終的に Container / Network / Volume として作られる世界
- Project という単位
  - 「プロジェクト」= サービス集合 + 付随リソースの束
- ラベル設計
  - Compose は Docker Engine 上に“状態 DB”を持てない
  - 内部で compose 管理下のリソースを探す処理は、ほぼ filters + labels
- 望ましい状態と実際の状態
  - Compose のコアはコントローラ
  - 望ましい状態 / 実際の状態 の差分を埋める
  - up の動作
    - 既存があれば差分で recreate
    - 依存関係順に起動順も制御
- config-hash と recreate の条件
- 依存関係のグラフ
  - 内部ではサービスをグラフとして扱う
- IO ストリームと TTY

## ビルド

- Docker Buildx
  - Docker の次世代ビルド機能を CLI から使いやすくした拡張
  - 複数アーキテクチャ向けのマルチプラットフォームビルドを 1 コマンドで実行

## コンテナ基盤技術

- VM と Docker
  - VM はコンピュータ自体を抽象化する
  - Docker はプロセス自体の抽象化をする
- cgroup
  - メモリ・CPU のような計算リソースを隔離するための機能
  - コンテナはそれぞれ専用の計算リソースを割り当て、他のコンテナにはお互いにアクセスできないようにする
- namespace
  - プロセスやネットワーク、ファイルアクセスなど複数の種類がある
  - それぞれが異なるリソースの隔離を行う
- Capability
  - スーパーユーザーとしての権限を制限するための機能
- Docker Image
  - 任意のタイミングのスナップショット
- Immutable Infrastructure
  - サーバー内への変更を行わないアプローチ
  - 変更や追加が発生する場合
    - 新しく構築してスナップショットを保存
    - スナップショットを元にサーバーを新しく立ち上げる

## その他の用語

- パストラバーサル攻撃
  - 本来アクセスできないファイルを、URL や入力値を細工して読み取る攻撃
- MAC アドレス
  - ネットワーク機器ごとに割り当てられる固有の識別番号

## 参考リンク

- https://y-ohgi.com/introduction-docker/2_component/image/
- docker compose コマンドの解説チャット: https://chatgpt.com/c/695374b5-06e8-8324-97b3-f423942ba35a
