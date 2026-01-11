- 並行処理
  - goroutine
    - Go ランタイムが管理する超軽量スレッド
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
  - context.Context のキャンセル伝播
    - context の役割
      - 「この処理、もうやめていいよ」を安全に伝える
  - WaitGroup
    - sync.WaitGroup
      - 全部終わるまで待つ
    - errgroup.Group
      - 最初のエラーを返す
      - エラーが出たら ctx を cancel
- errgroup
  - 概要
    - Go で「並行処理 + エラーハンドリング + キャンセル」を安全にまとめるための定番ツール
  - errgroup で出来ること
    - 複数 goroutine を起動
    - 最初に起きた error を 1 つ返す
    - エラー発生時に context をキャンセルできる
  - 1 つでも失敗したら、全体として失敗させる
