# コードの読み方

このリポジトリのコードをどの順で読むか。各層の役割は [architecture.md](./architecture.md) を参照。
`upstream/main` 289d8a037（2026-10-16）時点。行番号はすぐに古くなるので、関数名で探す。

← [document 一覧](../README.md)

## Phase 1: エントリーポイントの理解

1. `cmd/main.go`
   ↓ `main()` → プラグイン版／スタンドアロン版の判定 → `RootCommand()`
2. `cmd/compose/compose.go` の `RootCommand`
   ↓ 全サブコマンドを登録。`ProjectOptions` から compose-go でプロジェクトを読み込む
3. `cmd/compose/up.go`（`upCommand`、`runUp`）
   ↓ フラグを `api.UpOptions` に詰め替える
4. `pkg/api/api.go`
   ↓ `Service` インターフェースと各 Options
5. `pkg/compose/compose.go`
   ↓ `Service` インターフェースの実装（`composeService`）

## Phase 2: 1 つのコマンドを完全に理解（`up -d`）

[plan-engine.md](./plan-engine.md) と一緒に読む。

1. `pkg/compose/up.go` の `Up` → `upDetached`
2. `pkg/compose/create.go` の `preparePlan`（イメージの用意、観測）
3. `pkg/compose/observed_state.go` の `collectObservedState`
4. `pkg/compose/reconcile.go` の `reconcile` → `reconcileService` → `planStartPhase`
5. `pkg/compose/executor.go` の `run` → `executor_ops.go`
6. `pkg/compose/create.go` の `createMobyContainer`（compose.yaml の設定が `ContainerCreate` の引数になるところ）

前景の `up` は、その後に `runInteractiveUp`（attach、ログ、monitor、ナビゲーションメニュー）が続く。

## Phase 3: 横展開 - 他のコマンド

優先度 高:
├─ down.go # up の逆。まだ plan engine ではなく InReverseDependencyOrder
├─ start.go # ScopeStart の plan。ラベルから組み立て直す projectFromName もここで使う
├─ build.go / build_bake.go / build_classic.go # bake と classic の 2 経路
├─ logs.go # ストリーミング処理
└─ run.go # 一時コンテナ（oneoff）。依存サービスは create → start

優先度 中:
├─ exec.go # 実行中コンテナ操作
├─ ps.go # 状態取得
├─ scale.go # createAndStart
├─ watch.go # ファイル監視 → sync / rebuild
└─ hook.go / pre_start.go # ライフサイクルフック

## Phase 4: 深い部分を理解

- Docker API 連携: `apiClient()` の呼び出し、`pkg/mocks` の mock でテストする方法
- Compose ファイルのパース: compose-go（別リポジトリ）の loader。バグの原因が compose-go 側にあることも多い
- 並行制御・同期: executor の DAG 実行、`errgroup`、`maxConcurrency`
- ラベル: `pkg/api/labels.go` の定数と、`com.docker.compose.*` が何に使われているか
