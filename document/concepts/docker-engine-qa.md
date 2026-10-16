# docker-engine.md を読んで出た疑問 Q&A

[docker-engine.md](./docker-engine.md) を読んでいて出た疑問への短い回答。理解しやすい順に並べてある。

← [document 一覧](../README.md)

## 1. dockerd はどこで動くのか

- dockerd は、namespace や cgroup といった Linux カーネルの機能を直接使うので、**Linux の上でしか動かない**。ふつうの Linux のプロセスとして、root 権限で常駐している。
- Mac や Windows の Docker Desktop は、裏で小さな Linux VM を動かしていて、dockerd はその VM の中にいる。`docker` CLI だけがホスト（macOS）側にいて、ソケット越しに VM の中の dockerd を呼んでいる。
- 詳しくは [intro.md の「Docker は Linux 上でしか動かない」](./intro.md#docker-は-linux-上でしか動かない) を参照。

## 2. 「Docker の概念」を管理するとは

- イメージ、コンテナ、ボリューム、ネットワークは、Linux カーネルには存在しない**Docker が作った概念**。カーネルが知っているのは、プロセス、namespace、マウント、iptables のルールといった部品だけ。
- dockerd は、これらに名前と ID を付け、部品との対応を覚えておく。たとえば「コンテナ `web` = このプロセス + この rootfs + このネットワーク namespace + この設定」「ボリューム `data` = このディレクトリ」といった対応。
- `docker ps` や `docker volume ls` に答えたり、`docker rm` で関係する部品をまとめて片付けたりできるのは、この対応表を dockerd が持っているから。

## 3. 「dockerd が状態を持つ」とは、ファイルに保存しているのか

- そのとおり。`/var/lib/docker` 以下にファイルとして保存している。たとえば次のとおり。
  - コンテナごとの設定（`containers/<id>/config.v2.json`、`hostconfig.json`。ラベルもここ）
  - イメージの情報（`image/`）
  - ボリュームの一覧（`volumes/metadata.db`）
  - ネットワーク（`network/files/local-kv.db`）
- dockerd を再起動すると、これらを読み直し、まだ動いているコンテナには containerd 経由でつなぎ直す。「いま動いているか」は、containerd に問い合わせて確かめる。
- Compose が自前のデータベースを持たずに、ラベルで検索するだけで済むのは、この保存を dockerd がやってくれるから。

## 4. containerd-shim は、どうやってデーモンとコンテナを切り離しているのか

- コンテナごとに shim という小さなプロセスを起動し、runc にコンテナを作らせる。runc はすぐ終わるので、その後は **shim がコンテナのプロセスの親**（正確には subreaper）になる。containerd や dockerd は親ではない。
- shim は、コンテナの標準入出力（パイプや pty）を握り、終了コードを受け取って保持する。containerd とは、ソケット経由で話すだけ。だから containerd が再起動しても、shim とコンテナは動き続け、containerd は起動し直した後に shim へつなぎ直す。
- 注意: dockerd を**普通に止める**と、既定ではコンテナも止める。dockerd を再起動してもコンテナを残したいときは、`daemon.json` に `"live-restore": true` を設定する。

## 5. ボリュームの実体は、ホストにある「ファイル」がマウントされたものか

- ほぼ合っている。正確には**ファイルではなくディレクトリ**。名前付きボリュームの実体は、ホストの `/var/lib/docker/volumes/<名前>/_data` というディレクトリで、それがコンテナの中のパスに bind mount される。
- マウントはコピーや同期ではない。同じディレクトリを別の場所からも見えるようにしているだけなので、コンテナを消してもデータは残る（[container-tech-qa.md の 1](./container-tech-qa.md#1-マウントはどういう仕組みかファイルの同期なのか)）。
- Mac では、この `/var/lib/docker` は Linux VM の中にあるので、Finder からは見えない。ホストのフォルダを直接見せたいときは、bind mount（`./src:/app`）を使う。

## 6. ポートフォワード（`-p 8080:80`）の仕組み

- dockerd が iptables（または nftables）に、「ホストの 8080 番に来たパケットの宛先を、コンテナの IP アドレスの 80 番に書き換える」という NAT のルールを書き込む。一部の経路（ホスト自身の localhost からなど）では、`docker-proxy` というプロセスが中継する。
- Mac の Docker Desktop では、さらに「macOS の 8080 番 → VM の 8080 番」の中継を Docker Desktop が行っている。
- 注意: 既定では `0.0.0.0`（すべてのネットワークインターフェース）で公開されるので、同じ LAN の他のマシンからも届く。自分だけで使うなら `127.0.0.1:8080:80` と書く。

## 7. BuildKit とは

- イメージをビルドするエンジン。Dockerfile を、ステップの依存関係のグラフ（LLB）に変換し、依存の無いステップを並列に実行する。変わっていないステップはキャッシュで飛ばす。
- 昔のビルダーと比べると、並列化、キャッシュの共有（`--cache-to` / `--cache-from`）、マルチプラットフォームのビルド、秘密情報を安全に渡す仕組み（`--mount=type=secret`）が加わった。
- `docker buildx` は BuildKit を使うための CLI。Compose の `build` は、buildx の bake 経由で BuildKit を呼ぶのが基本で、buildx が無いと昔の classic builder を使う（[docker-engine.md の BuildKit / Buildx](./docker-engine.md#buildkit--buildx)）。

## 8. rootless モードとは

- dockerd とコンテナを、**root ではなく一般ユーザーの権限で動かす**モード。user namespace を使い、コンテナの中の root を、ホストでは一般ユーザーの UID（`/etc/subuid` で割り当てた範囲）に対応させる。
- 利点は、コンテナから脱出されたり dockerd に穴があったりしても、奪われるのは一般ユーザーの権限で済むこと。
- 制約もある。1024 番より小さいポートは、設定しないと使えない。ネットワークは RootlessKit 経由になるので遅くなる。リソースの上限には cgroup v2 が必要。使えるストレージドライバや機能にも、一部制限がある。

## 9. Docker を使うときに、セキュリティで気を付けること

- **`docker.sock` へのアクセスは、ホストの root 権限と同じ**。`docker` グループにユーザーを入れることや、コンテナに `/var/run/docker.sock` をマウントすることは、root を渡すのと同じだと考える。Compose の `use_api_socket` もこれに当たる。Engine API を TCP で公開するときは、必ず TLS を使う。
- コンテナの権限は最小にする。
  - `--privileged` は使わない。`cap_add` は必要なものだけにする。
  - イメージの中では `USER` で一般ユーザーにする。
  - `-p` は必要なら `127.0.0.1` に限定する（`0.0.0.0` の公開は ufw などのファイアウォールを素通りすることがある）。
- イメージと秘密情報に気を付ける。
  - 信頼できるイメージを、digest で固定して使い、脆弱性をスキャンする。
  - パスワードを `ENV` や `COPY` でイメージに焼き込まない（レイヤーに残る）。secrets を使う。
  - Docker Engine 自体も更新し続ける。
