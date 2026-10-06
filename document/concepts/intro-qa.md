# intro.md を読んで出た疑問 Q&A

[intro.md](./intro.md) を読んでいて出た疑問への短い回答。OCI の詳細は [oci.md](./oci.md) を参照。

← [document 一覧](../README.md)

## 1. BE はコンテナをよく使うのに、FE はあまり使わないのはなぜか

- BE は DB・キャッシュ・ミドルウェアなど**OS レベルの依存**が多く、本番(Linux サーバ)と同じ環境を再現する価値が大きい。
- FE は最終的に**ブラウザで動く静的ファイル**を作るだけで、依存は Node と npm パッケージ程度。`nvm` + `package-lock.json` で再現性は十分取れる。
- さらに Mac 上のコンテナはファイル監視(HMR)やボリュームの I/O が遅く、開発体験が落ちるのでわざわざ使う理由が薄い。

## 2. Apple の `container`(apple/container)とは何か

- Apple 製の、Mac 上で Linux コンテナを動かすための Swift 製 CLI。macOS 26 の `Containerization` フレームワークを使う(Apple Silicon 専用)。
- Docker Desktop が「1 台の大きな Linux VM の中で全コンテナを動かす」のに対し、**コンテナ 1 つごとに軽量 VM を 1 つ起動**する方式。隔離が強く、起動も速く、使わない分のメモリを抱えない。
- OCI 準拠のイメージをそのまま pull/run できるが、Compose 相当の機能やエコシステムはまだ Docker に及ばない。「VM コストがなくなる」わけではなく、VM の持ち方を変えたもの。

## 3. Apple Silicon(M1)登場時に Docker が動かなかったのは何だったのか

- コンテナは**ホストのカーネル・CPU 上で直接動く**ので、CPU アーキテクチャ(x86_64 / arm64)が一致したイメージが必要。当時は Docker Hub のイメージの多くが x86_64 版しかなかった。
- Docker Desktop 自体も Intel Mac 向けの VM(HyperKit)に依存しており、Apple の Virtualization.framework 対応版が出るまで数か月かかった。
- 今はマルチアーキテクチャイメージ(manifest list)が普及し、どうしても x86 版しかない場合は QEMU / Rosetta でエミュレーションして動かせる(ただし遅い)。

## 4. サーバ用途で Windows が衰退し Linux 一強になった背景

- **ライセンス費用がゼロ**で、何千台とスケールするクラウド/Web 企業にとってコスト差が決定的だった(LAMP スタックの普及)。
- ソースが公開されていて自由にチューニング・軽量化でき、GUI 不要の CLI 前提の運用(SSH・スクリプト自動化)と相性が良い。
- AWS などのクラウド、そして Docker/Kubernetes が Linux カーネル機能(namespaces, cgroups)の上に作られたことで、エコシステムが Linux に一極集中した。

## 5. OCI 標準の具体的な中身

- OCI(Open Container Initiative)は Docker の形式を元に策定された、コンテナの**ベンダー非依存の標準**。主に 3 つの仕様がある。
- **Image Spec**: イメージの形式(レイヤーの tar、各レイヤーの digest を並べた manifest、環境変数・エントリポイント等を書いた config JSON)。**Runtime Spec**: 展開済みの rootfs と `config.json` をどう起動するか(参照実装が `runc`)。**Distribution Spec**: レジストリとの push/pull の HTTP API。
- これがあるので、Docker でビルドしたイメージを Podman・containerd・Kubernetes・apple/container などどこでも動かせる。

## 6. DB はローカルではコンテナなのに、本番では専用サービス(RDS 等)を使う理由

- ローカルでは「すぐ立ち上げて、壊れたら捨てる」ことが重要で、データが消えても困らないのでコンテナが最適。
- 本番では**データを絶対に失わない**ことが最優先。バックアップ、ポイントインタイムリカバリ、レプリケーション、フェイルオーバー、パッチ適用など運用の手間が大きく、マネージドサービスに任せた方が安全で安い。
- コンテナは「ステートレスで使い捨て」が前提の設計なので、永続データ・ストレージ性能・安定した配置が求められる DB とは相性が悪い(Kubernetes で動かすこともできるが運用難度が高い)。
