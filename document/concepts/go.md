# Go 言語

Compose のコードを読むのに必要な Go の知識。Compose での並行処理の使われ方は
[codebase/architecture.md #並行処理パターン](../codebase/architecture.md#並行処理パターンerrgroup-等) を参照。
読んでいて出た疑問の Q&A は [go-qa.md](./go-qa.md)。

← [document 一覧](../README.md)

## 並行処理

- goroutine
  - Go ランタイムが管理する超軽量スレッド
  - Compose 本体は `up`/`down` 時の複数サービス並列処理などで多用している。
- GOMAXPROCS
  - 同時に実行できる goroutine 数
- goroutine leak
- goroutine の使いどころ
  - I/O 待ち
  - 複数独立タスク
  - イベント購読
  - producer / consumer
- channel
  - goroutine 間で値を安全にやり取りするためのパイプ
  - 送信・受信
    ```go
      ch <- 1
      v := <-ch
    ```
  - buffered/unbuffered の違いや `select` による多重待受を押さえる。
- WaitGroup
  - sync.WaitGroup
    - 全部終わるまで待つ
  - errgroup.Group
    - 最初のエラーを返す
    - エラーが出たら ctx を cancel

### errgroup

- 概要
  - Go で「並行処理 + エラーハンドリング + キャンセル」を安全にまとめるための定番ツール
- errgroup で出来ること
  - 複数 goroutine を起動
  - 最初に起きた error を 1 つ返す
  - エラー発生時に context をキャンセルできる
- 1 つでも失敗したら、全体として失敗させる

## context

- キャンセル伝播・タイムアウト・リクエストスコープの値渡しを行う標準パターン。
- context の役割
  - 「この処理、もうやめていいよ」を安全に伝える
- `context.Context` を関数の第一引数として引き回すのが Go の慣習。
- Compose では Ctrl+C 等の中断シグナルを、起動した子 goroutine やコマンド実行まで伝搬させるのに使われる。

## interface 設計

- 実装ではなく振る舞いに対して小さなインターフェースを定義し、組み合わせて使う Go 流の設計思想。
- `pkg/api` の `Service` インターフェースのように、呼び出し側と実装を疎結合にする境界の切り方。
- モックによるテスト容易性にも直結する。

## error wrapping

- `fmt.Errorf("...: %w", err)` でエラーに文脈を追加し、元のエラーを保持したまま連鎖させる書き方。
- `errors.Is` / `errors.As` で特定のエラー種別を判定する。
- スタックトレースを持たない Go では、呼び出し階層をエラー文脈で辿ることが多い。

## go modules / vendoring

- `go.mod` / `go.sum` によるバージョン付き依存管理の仕組み。
- `vendor/` ディレクトリに依存コードそのものを同梱する vendoring。
- 本リポジトリは vendoring しており、依存差分は `vendor/` にも反映される点に注意。
