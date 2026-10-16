# oci.md を読んで出た疑問 Q&A

[oci.md](./oci.md) を読んでいて出た疑問への短い回答。

← [document 一覧](../README.md)

## 1. manifest とは

- イメージが**どの部品でできているかを書いた目録**（JSON）。中身そのものは入っておらず、config（実行時の設定）と各レイヤーの digest・サイズが書かれている。
- 実行時の設定（Entrypoint、Cmd、Env など）を持っているのは **config** のほう。manifest は config を digest で指しているだけ。
- 複数の CPU 向けのイメージでは、さらに上に image index（manifest の目録）がある。pull のときに、自分の CPU 向けの manifest が選ばれる。

## 2. 「レイヤーの digest リスト」とは

- digest は、**中身から計算したハッシュ値**（`sha256:...`）。manifest の `layers` には、各レイヤーの digest が下の層から順に並んでいる。
- 同じ中身なら必ず同じ digest になるので、共通のレイヤーは 1 回しかダウンロード・保存されない。ダウンロードしたものが壊れていたり改ざんされていたりすれば、ハッシュが合わないので気付ける。
- タグ（`nginx:latest`）は付け替えられるが、digest（`nginx@sha256:...`）は永久に同じイメージを指す。

## 3. コンテナはどう起動されるのか

- containerd が、イメージのレイヤーを overlayfs で重ねて rootfs を用意する。あわせて、イメージの config と compose.yaml などの設定から、`config.json`（OCI Runtime Spec の形式）を作る。
- runc がそれを読み、namespace を作る → cgroup で上限を設定する → `/proc` やボリュームをマウントする → pivot_root で rootfs を `/` にする → capability や seccomp で権限を絞る、の順に準備する。
- 最後に、Entrypoint + Cmd を exec して PID 1 にする。runc はそこで終わり、以後は containerd-shim が親としてプロセスを見守る。

## 4. Dockerfile の 1 行が 1 レイヤーになるのか

- ファイルを変える命令（`RUN`、`COPY`、`ADD`）だけがレイヤーになる。`ENV`、`CMD`、`EXPOSE`、`LABEL` などは config を書き換えるだけで、レイヤーにはならない。`docker history` で SIZE が 0B の行がそれ。
- 後の行でファイルを消しても、下のレイヤーには残るので、イメージは小さくならない。だから `apt-get install` と後片付けは、同じ `RUN` の中で行う。
- マルチステージビルドでは、最後のステージのレイヤーだけが最終イメージに入る。
