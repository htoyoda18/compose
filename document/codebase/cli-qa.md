# cli.md を読んで出た疑問 Q&A

[cli.md](./cli.md) を読んでいて出た疑問への短い回答。

← [document 一覧](../README.md)

## Docker コンテキストで、moby 側の処理はどう変わるのか

Docker 全般の話なので docker-doc に置いた → [docker-engine-qa.md #10](../../../docker-doc/concepts/docker-engine-qa.md#10-docker-コンテキストでmoby-側の処理はどう変わるのか)

## 1. なぜプラグイン版になったのか。どんな利点があったのか

- **経緯**: Compose v1 は Python で書かれた単体のコマンド `docker-compose` だった。v2 で Go で書き直され、`docker` CLI のプラグインとして `~/.docker/cli-plugins` に置く形になった（`README.md:45`、`:92`）。
- **コードで確かめられる利点**: `docker` CLI の初期化と設定を、そのまま使い回せる。
  - `plugin.Run` と `plugin.PersistentPreRunE` が、`docker` 本体と同じ手順で CLI を初期化する。だから、コンテキスト、`-H`、TLS、`~/.docker/config.json`（認証情報など）を、compose 側で実装し直さずに済む（`cmd/main.go:39`、`:55`）。
  - docker/cli の実装も直接呼べる。たとえば `stats` は、docker CLI の `container.RunStats` にそのまま任せている（`cmd/compose/stats.go:73`）。
  - スタンドアロンの `docker-compose` も同じバイナリで動く。引数をプラグインの形に並べ替えてから、同じ処理を通す（`cmd/main.go:86-87`）。だから v1 からの移行の互換を保ったまま、実装は 1 つで済む。
- **確かめられなかったこと**: 「Python への依存をなくす」「`docker` の操作感と揃える」といった、プラグインにすると決めた理由そのものは、このリポジトリのコードや文書には書かれていない。Docker 社の発表などで確認が要る。
