[root](../../note.md) / [cmd](../note.md) / [compose](note.md) / up.go Q&A

# up.go.md を読んで出た疑問 Q&A

[up.go.md](up.go.md) を読んでいて出た疑問への短い回答。手元の `upstream/main` e18ab3d87 と、go.mod が指す compose-go（v2.16.2-0.20261015044941-5515644a24af）で確認した。
`runUp` が処理する順（リモートの確認 → job の確認）に並べている。

## 1. リモートの compose ファイルとは

- **手元のファイルではなく、git リポジトリや OCI レジストリに置いた compose ファイルを `-f` で直接指定して動かす機能**。`-f` の値が `oci://` で始まるか（`pkg/remote/oci.go:124-126`）、git の参照として解釈できれば（`pkg/remote/git.go:68-71`）リモートとみなす。取得は [pkg/remote](../../pkg/remote/note.md) の 2 つのローダーが担う（`cmd/compose/compose.go:384-394`）。
  - それぞれ環境変数 `COMPOSE_EXPERIMENTAL_GIT_REMOTE` / `COMPOSE_EXPERIMENTAL_OCI_REMOTE` で無効にできる（`pkg/remote/git.go:41-51`、`pkg/remote/oci.go:45`）。`--offline` のときは使わない（`compose.go:388-390`）。
- **他人が書いた設定を実行することになるので、`up` の前に確認を挟む**のが `checksForRemoteStack`（`cmd/compose/options.go:124-135`）。最初の `-f` がリモートのときだけ動く（`options.go:110-121`）。
  1. 中でさらにリモートのファイルを `include` していれば、「悪意のある可能性がある」と警告して確認を取る（`options.go:264-289`）
  2. 取得したファイルの置き場所を表示する（`options.go:257-262`）
  3. 変数の展開に使う値を一覧で見せて、確認を取る（`options.go:148-`）
  - `--yes` なら、1 と 3 の確認を飛ばす。

## 2. compose の job とは

- **1 回実行して終わる処理（マイグレーション、バッチなど）を、`services` とは別に `jobs:` として宣言する仕組み**。中身はサービスと同じコンテナの定義（`ContainerSpec` + `WorkloadSpec`）に、いつ起動するかを表す `triggers` を足したもの（compose-go `types/jobs.go:20-29`）。
  - `triggers.manual`: 手動での起動を許すか。省略すると許す。`false` で禁止（`types/jobs.go:31-36`）
  - `triggers.schedule`: cron 式による定期実行（`types/jobs.go:37`、`:41-49`）
- **`up` は job を起動しない**。手動の job は `docker compose run <job>` で起動するものなので、警告だけ出す（`up.go:359-372`）。`run` は job をサービスの形に変換して実行する（`run.go:463`〜）。
- **定期実行は、このバージョンではまだ実装されていない**。黙って無視すると期待を裏切るので、`schedule` のある job があれば `up` をエラーで止める（`up.go:374-389`）。この扱いはコミット 124cbb8eb（2026-08-19）で入った。
- compose-spec の仕様書側の定義は、このリポジトリでは確かめていない（確認したのは compose-go の型とコメントだけ）。
