[root](../../note.md) / [pkg](../note.md) / compose

`api.Compose` インターフェースの実装本体(`composeService`)。実装 62 ファイル(テスト除く)のうち、
ノートがあるものと主要ファイルだけを載せる。各ファイルは `cmd/compose` の同名ファイルとほぼ 1 対 1 に対応する。

## 中核

- [compose.go](compose.go.md)
  - `composeService` の定義とコンストラクタ `NewComposeService`、Option 関数群
- projectloader.go
  - `LoadProject` / `Generate`。compose ファイルの読み込みと、既存コンテナからのプロジェクト生成
- dependencies.go
  - サービス依存グラフと、上向き(起動順)/下向き(停止順)の並列走査
- [containers.go](containers.go.md)
  - プロジェクト配下のコンテナ一覧取得と、絞り込み・並べ替え用ヘルパー(`Containers` 型)
- hash.go
  - サービス/ネットワークの config hash 計算。コンテナを作り直すかの判定に使う

## plan engine(観測 → 計画 → 実行)

- observed_state.go
  - `ObservedState`。プロジェクトに属する Docker リソースの現状を集めたもの
- reconcile.go
  - 望ましい状態と `ObservedState` を比較し、差分を埋める操作のプランを組み立てる
- plan.go
  - プランの要素(`OperationType` = Docker API 1 呼び出し、`PlanPhase`)の定義
- executor.go
  - プランを依存順に並列実行する `planExecutor`

## コマンド実装

- [create.go](create.go.md)
  - コンテナ作成に必要な Container/Host/Network 設定の組み立て
- [up.go](up.go.md)
  - `up`。create → start → (attach/watch) の流れ
- start.go / stop.go
  - `start` / `stop`。依存順に起動・停止する
- [down.go](down.go.md)
  - `down`。コンテナ停止・削除と、必要に応じたネットワーク/ボリューム/イメージ削除
- run.go
  - `run`。one-off コンテナの準備と実行
- [build.go](build.go.md)
  - `build`。ビルド対象の決定と依存解決(実際のビルドは build_bake.go / build_classic.go)
- pull.go / push.go / publish.go
  - イメージの pull / push と、compose ファイルを OCI アーティファクトとして publish
- logs.go
  - `logs`。コンテナのログ取得と follow
- [attach.go](attach.go.md)
  - `up` などでのログ attach。コンテナの取得と色割り当て
- [attach_service.go](attach_service.go.md)
  - `attach`。対象コンテナを 1 つに絞って docker/cli の attach を実行
- [wait.go](wait.go.md)
  - `wait`。指定サービスのコンテナ終了を並行に待つ
- [watch.go](watch.go.md)
  - `watch`。watch 設定のルール化、ファイル変更のバッチ化、sync/rebuild/restart/exec の実行
- hook.go / pre_start.go
  - ライフサイクルフック(post_start 等)と pre_start フックの実行
