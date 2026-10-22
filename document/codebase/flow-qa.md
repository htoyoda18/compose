# flow.md を読んで出た疑問 Q&A

[flow.md](./flow.md) を読んでいて出た疑問への短い回答。`upstream/main` e18ab3d87（2026-10-22）時点のコードで確認した。
flow.md の流れに沿って、「共通の流れ」→「up」→「build」→「run」の順に並べている。

← [document 一覧](../README.md)

## 1. `compose.NewComposeService` でバックエンドを作るとは、具体的に何をしているのか

- `pkg/compose/compose.go` の `NewComposeService` は、`composeService` という構造体を 1 つ作って、`api.Compose` インターフェースとして返すだけ。Docker への接続やコンテナの操作はまだ何もしない。
- 中身は「これからの処理で使う道具の詰め合わせ」。docker/cli が初期化済みの `dockerCli`（Engine API クライアントや設定ファイルの入り口）、並列数の上限（既定は無制限の `-1`）、レジストリ認証の取得方法を持たせる。その後 `WithEvents`（進捗表示の送り先）、`WithPrompt`（確認プロンプト）、`WithDryRun`、`WithStreams` などのオプションで差し替える。
- 進捗表示の送り先やプロンプトが渡されなければ、何もしない既定値を入れる。だから SDK として他のプログラムから使うときも、最低限 `dockerCli` だけ渡せば動く。

## 2. Go では `internal` ディレクトリを使うことが多いのか

- 多い。`internal` は Go 言語そのものの仕組みで、**`internal/` の下のパッケージは、その親ディレクトリより下のコードからしか import できない**（コンパイラが禁止する）。外に公開したくない実装を置く定番の場所。
- このリポジトリでは、`pkg/`（`pkg/api` や `pkg/compose`）が SDK として外部から使われる公開部分、`internal/`（`desktop`, `oci`, `projectloader`, `provider`, `relay`, `sync`, `tracing` など 16 個）が Compose の中だけで使う部品、という分け方になっている。
- 公開すると互換性を守る約束が生まれるので、外部に使わせる理由がないものは `internal` に置いておくと、後から自由に変更できる。

## 3. plan engine は最近できたものなのか

- 比較的最近。ndeloof が「観測 → 計画（Plan）→ 実行（Executor）」の仕組みを PR #13830「Reconciliation plan」で入れ、**2026-06-08 にマージ、v5.2.0（2026-06-23）で初めてリリース**された。それ以前は、コマンドごとに依存順にたどって直接 API を呼ぶ命令的な実装だった。
- その後も拡張が続いている。ネットワークやボリュームの作り直しを Plan に載せたのが 7 月（#13962, #13966）。エピック #14081 で `start` の処理を Plan に移す作業が 9 月以降に進んでいる（#14200 など）。
- ただし、まだ全部のコマンドが通っているわけではない。flow.md のとおり、`down` は今も plan engine を通らず、`InReverseDependencyOrder` で直接削除している。

## 4. BuildKit は並列で処理するので、`doBuildClassic` より速いのか

- 速いことが多いが、理由は「classic は直列だから」ではない。`doBuildClassic` も `InDependencyOrder` でサービスをたどり、依存関係のないサービスは `--parallel` の上限まで**並列にビルドする**。flow.md の「1 つずつビルド」は、正確には「サービスごとに 1 回ずつビルド要求を出す」という意味。
- 本当の違いはビルドの中身。classic は Engine API の旧ビルダー（`BuilderV1`）を使い、Dockerfile の各命令を上から順に実行する。BuildKit は Dockerfile を依存グラフ（LLB）にして、マルチステージの独立したステージを並列に実行したり、使わないステージを飛ばしたりする。キャッシュの効き方も賢い。
- さらに `doBuildBake` は全サービスを 1 回の `docker buildx bake` にまとめて渡すので、サービスの間で共通のステージやキャッシュも共有できる。classic では、secrets・ssh・マルチプラットフォームなど BuildKit 専用の機能も使えない。

## 5. one-off コンテナとは何か

- `docker compose run <service>` で作る**使い捨てのコンテナ**のこと。サービス定義を土台にして、そのとき限りのコマンドを実行するために、通常のサービスのコンテナとは別に 1 つ作る（例: `docker compose run web npm test`）。
- 通常のコンテナとはラベルと名前で区別される。`com.docker.compose.oneoff=True` が付き、レプリカ番号（`container-number`）は付かない。名前は `<project>-<service>-run-<ランダムな ID>` になる。そのため `ps` などの通常の一覧からは原則として外れ、`up` が管理するレプリカの数にも入らない。
- 終了後もコンテナは残るので、`--rm` を付けると自動で消える。残ったものは `down --remove-orphans` などで片付ける。
