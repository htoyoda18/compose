[root](../note.md) / [cmd](note.md) / Q&A

# cmd/note.md を読んで出た疑問 Q&A

[cmd/note.md](note.md) を読んでいて出た疑問への短い回答。`upstream/main`（2026-10-22 時点）のコードで確認した。
接続先の話（dockerd）→ コマンド実行前の準備（`PersistentPreRunE`）→ その準備で選ばれる進捗表示（display）→ 対話の入力側（prompt）の順に並べている。

## 1. dockerd は moby/moby のことか

- ほぼそのとおり。**dockerd は、moby/moby リポジトリのコードからビルドされるデーモンのプログラム名**。moby/moby はそのソースコードを持つオープンソースのプロジェクト（2017 年に docker/docker から改名した）。
- 「Docker Engine」は、moby の dockerd に containerd や runc などを組み合わせて Docker 社が配布している製品の名前。moby がソース、dockerd が実行するデーモン、Docker Engine が製品、という関係。
- Compose は dockerd の Engine API を呼ぶ側なので、moby/moby の中のクライアントと型のモジュール（`github.com/moby/moby/client` と `github.com/moby/moby/api`）を依存として使っている。

## 2. `PersistentPreRunE` とは何か

- cobra（Go の CLI フレームワーク）が用意しているフックで、**そのコマンドと、その下のすべてのサブコマンドの本体を実行する直前に呼ばれる**関数。名前の `Persistent` は子コマンドにも引き継がれること、`E` はエラーを返せる版であることを表す。
- Compose では `RootCommand`（`cmd/compose/compose.go`）に定義していて、`up` でも `ps` でも必ず通る。中でやっていることは次のとおり。
  - docker CLI 側の準備（`runParentPreRun`）
  - `.env` の読み込み
  - `--ansi` の解決
  - 進捗表示のモードの選択（`selectEventProcessor`）
  - `--parallel` や `--dry-run` をバックエンドのオプションに追加
- cobra は親の `PersistentPreRunE` を自動では呼ばないので、`runParentPreRun` で明示的に呼んでいる。さらに `cmd/main.go` がこの関数を包み直していて、実際の順番は次のとおり。
  1. docker CLI の初期化（`plugin.PersistentPreRunE`）
  2. OTel のトレース開始（`cmdtrace.Setup`）
  3. 上の Compose 側の準備

## 3. display は、コマンド実行中のイベントをどう画面に描画しているのか

- `pkg/compose` の処理は、画面に直接書かずに**イベントを送るだけ**（「コンテナ web を作成中」「完了」など）。受け取る側は `pkg/api` の `EventProcessor` インターフェース（`Start` / `On` / `Done` の 3 つ）で、その実装が `cmd/display` にある。どの実装を使うかは、上の `selectEventProcessor` が `--progress` と「stderr が端末かどうか」で決める。
- 単純な実装は、受け取ったイベントをすぐ書き出す。Plain（`plain.go`）は 1 行ずつ追記し、JSON（`json.go`）は 1 イベントを 1 行の JSON にする。Quiet は何も出さない。
- TTY（`tty.go`）は 3 層に分かれている。
  - `tty_model.go`: イベントを受けてタスクの木（親子関係・状態・開始と終了の時刻）を更新するだけの、純粋なデータ。
  - `tty_layout.go`: その木を、端末の幅に収まる行の配列に変換する。
  - `tty_screen.go`: 前回描いた内容と比べ、ANSI エスケープでカーソルを戻して、変わった行だけを書き直す。
  - これを 100ms ごとに繰り返すので、スピナーと経過時間が同じ場所で動いて見える。

## 4. 対話端末の survey と、非対話環境でのパイプ経由とは何か

- `down` などの破壊的な操作の前に「本当に実行しますか? [y/N]」と聞く仕組みのこと。`cmd/prompt/prompt.go` の `NewPrompt` が、**stdin が端末かどうか**で 2 つの実装を切り替える。
  - 端末（対話端末）なら `User`: 端末を raw モードにして 1 文字ずつ読み、Backspace や Ctrl+C も自前で処理する。
  - そうでなければ `Pipe`: `echo y | docker compose ...` のようにパイプで渡された入力から 1 行読むだけ。
- **cmd/note.md の「survey を使った確認」は古い情報**。survey（`github.com/AlecAivazis/survey/v2`、対話プロンプトのライブラリ）は 2026-07-06 のコミット 8f1a46c87 で取り除かれ、今は上の自前の実装になっている（依存を減らす流れの一部）。
- 非対話の環境（CI など）で確認を飛ばしたいときは、パイプで入力を渡すか、`-y` などのフラグがあるコマンドではそれを使う。
