# 前提知識

← [roadmap 一覧](./README.md) ／ [document 一覧](../README.md)

## Go 言語

- **goroutine / channel**
  - Go の軽量スレッドである goroutine と、goroutine 間でデータをやり取りする channel の基礎。
  - buffered/unbuffered の違いや `select` による多重待受を押さえる。
  - Compose 本体は `up`/`down` 時の複数サービス並列処理などで多用している。
- **context**
  - キャンセル伝播・タイムアウト・リクエストスコープの値渡しを行う標準パターン。
  - `context.Context` を関数の第一引数として引き回すのが Go の慣習。
  - Compose では Ctrl+C 等の中断シグナルを、起動した子 goroutine やコマンド実行まで伝搬させるのに使われる。
- **interface 設計**
  - 実装ではなく振る舞いに対して小さなインターフェースを定義し、組み合わせて使う Go 流の設計思想。
  - `pkg/api` の `Service` インターフェースのように、呼び出し側と実装を疎結合にする境界の切り方。
  - モックによるテスト容易性にも直結する。
- **error wrapping**
  - `fmt.Errorf("...: %w", err)` でエラーに文脈を追加し、元のエラーを保持したまま連鎖させる書き方。
  - `errors.Is` / `errors.As` で特定のエラー種別を判定する。
  - スタックトレースを持たない Go では、呼び出し階層をエラー文脈で辿ることが多い。
- **go modules / vendoring**
  - `go.mod` / `go.sum` によるバージョン付き依存管理の仕組み。
  - `vendor/` ディレクトリに依存コードそのものを同梱する vendoring。
  - 本リポジトリは vendoring しており、依存差分は `vendor/` にも反映される点に注意。

## Linux カーネルの基礎

- **namespace**
  - プロセスから見える PID・ネットワーク・マウント・ホスト名などのリソースを分離する仕組み。
  - コンテナ間で「見えるもの」を隔離する、コンテナ分離の中核技術。
  - `pid` / `net` / `mnt` / `uts` / `ipc` / `user` など種類ごとに分離対象が異なる。
- **cgroup**
  - CPU・メモリ・IO などの計算リソースの使用量を制限・計測する Linux カーネル機能。
  - コンテナごとにリソース上限を割り当て、他コンテナへの影響を防ぐ。
  - v1/v2 でインターフェースが異なり、コンテナランタイムはこれを抽象化して利用する。
- **capability**
  - root 権限を細分化し、必要な権限だけをプロセスに付与する仕組み。
  - コンテナはデフォルトで多くの capability を落として起動し、攻撃面を減らす。
  - `--cap-add` / `--cap-drop` のような CLI オプションの背景にある概念。
- **chroot / pivot_root**
  - プロセスから見えるファイルシステムのルートを変更する仕組み。
  - `chroot` は簡易的なルート変更、`pivot_root` はより安全にルートを入れ替える手段。
  - コンテナのファイルシステム分離（イメージの root fs をコンテナの `/` にする）の基礎。
- **overlayfs**
  - 複数のディレクトリ層を重ね合わせて 1 つのファイルシステムに見せる union filesystem。
  - Docker イメージの read-only レイヤーと、コンテナの read-write レイヤーを重ねる仕組みに使われる。
  - レイヤーの再利用によりイメージのサイズや起動速度を効率化する。

## ネットワーキングの基礎

- **bridge / veth**
  - bridge はホスト内の仮想スイッチ、veth はその両端をつなぐ仮想ケーブルのようなペアインターフェース。
  - コンテナのネットワーク namespace とホスト側 bridge を veth で接続し、外部と通信できるようにする。
  - Docker のデフォルトネットワークドライバである `bridge` の実体でもある。
- **iptables / nftables**
  - Linux のパケットフィルタリング・NAT の仕組みとその設定ツール。
  - Docker はコンテナのポート公開（`-p`）やコンテナ間通信の制御にこれらのルールを自動生成する。
  - nftables は iptables の後継で、内部的に置き換えが進んでいる。
- **DNS 解決**
  - コンテナ名・サービス名から IP アドレスを引く仕組み。
  - Compose はプロジェクト内のサービス同士が名前で通信できるよう、内蔵 DNS を提供する。
  - `/etc/resolv.conf` の扱いや埋め込み DNS サーバーの挙動を理解しておくとネットワーク周りのデバッグに役立つ。

## コンテナ関連の標準規格

- **OCI Image Spec**
  - コンテナイメージのフォーマット（レイヤー構成・manifest・config）を定義する仕様。
  - Docker イメージもこの仕様に準拠しており、他のツール（Podman 等）とも互換性がある。
  - manifest がレイヤーの digest リストと実行時の設定情報を持つ。
- **OCI Runtime Spec**
  - コンテナをどう起動するか（namespace/cgroup/mount 設定等）を定義する仕様。
  - `runc` はこの仕様のリファレンス実装であり、`config.json` を受け取ってコンテナを起動する。
  - dockerd/containerd から見ると、この仕様に従うランタイムを差し替え可能。
- **OCI Distribution Spec（レジストリ API）**
  - イメージの pull/push を行うレジストリ HTTP API の仕様。
  - Docker Hub や GHCR など各種レジストリはこの仕様に準拠している。
  - `pkg/remote` のような OCI レジストリ操作コードを読む際の前提知識になる。

## その他

- **YAML**
  - インデントによる構造表現、アンカー/エイリアス、マルチドキュメントなどの基本文法。
  - `compose.yaml` はこの上に compose-spec 独自の意味論が乗る形なので、まず YAML 自体の挙動を理解しておく。
  - タブ不可・スカラー型の曖昧な解釈（`yes`/`no` 等）といった落とし穴も把握しておくと良い。
- **Git / GitHub ワークフロー（OSS コントリビューションの基本作法）**
  - フォーク・ブランチ運用、Conventional Commits 的な粒度でのコミット分割。
  - DCO（Signed-off-by）や CLA など、OSS プロジェクトごとのコントリビューション規約。
  - PR テンプレート、レビュー対応、CI（golangci-lint・テスト）通過までの一連の流れ。

