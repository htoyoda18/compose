# リポジトリのアーキテクチャ

← [roadmap 一覧](./README.md) ／ [document 一覧](../README.md)

- **レイヤー構成（cmd / pkg/api / pkg/compose / internal）**
  - `cmd` は CLI のコマンド定義・引数パース・表示ロジックを担当するエントリーポイント層。
  - `pkg/api` は Compose の機能をインターフェースとして定義する層で、外部の Go プログラムからも SDK として利用可能。
  - `pkg/compose` が実際のロジック（収束処理・Docker Engine API 呼び出し等）を持つ実装層。
  - `internal` はこのリポジトリ専用の非公開パッケージ（Desktop 連携・トレーシング・OCI 操作等）を置く。
- **インターフェース駆動設計**
  - `pkg/api/api.go` の `Service` インターフェースが、CLI 層と実装層（`pkg/compose`）の境界になっている。
  - テストではこのインターフェースをモック化し、Docker Engine に接続せずにロジックを検証できる。
  - SDK としての利用時も、この抽象化のおかげで実装詳細を意識せずに Compose を組み込める。
- **並行処理パターン（errgroup 等）**
  - `golang.org/x/sync/errgroup` を使い、複数 goroutine のエラーをまとめて収集しつつ並列実行する。
  - サービスの並列起動・並列ビルド等、依存グラフ上で並列実行可能な単位ごとにこのパターンが使われる。
  - 1 つでもエラーが出れば context をキャンセルし、他の goroutine に中断を伝える設計。
- **プログレス表示・イベントシステム**
  - `up` / `build` 実行中の進捗を、TTY / plain / JSON 形式でリアルタイムに描画する仕組み。
  - 内部的にはイベント（開始・進捗・完了等）を発行し、writer 実装がそれを購読して描画する構成。
  - この仕組みにより、ターミナル出力だけでなく外部 UI（Docker Desktop 等）にも同じイベントを流用できる。
- **OpenTelemetry によるトレーシング**
  - 各コマンド実行の span を生成し、内部処理の所要時間や呼び出し関係を可視化できるようにする。
  - OTLP エクスポーター（gRPC/HTTP）経由で外部のトレーシングバックエンドに送信可能。
  - `internal/tracing` にラッパーがあり、コマンドや API 呼び出しの主要ポイントに計装されている。

