# plan engine（観測 → 計画 → 実行）

← [document 一覧](../README.md)

`docker compose up` などが「望ましい状態」と「実際の状態」の差を埋める仕組み。
2026 年に、それまでの convergence engine（`convergence.go`）から、この plan engine に置き換わった。
`convergence.go` はコミット e801d5fc7「rename convergence.go, whose convergence engine no longer exists」で改名され、もう存在しない。

- 基準コード: `upstream/main` 289d8a037（2026-10-16）
- コードへのリンクは GitHub の固定コミットを指す（手元の作業ブランチには、まだ無いファイルがあるため）
- 「望ましい状態に近づける」という考え方そのものは [concepts/compose-concepts.md](../concepts/compose-concepts.md#収束処理convergenceと望ましい状態への到達) を参照

## 一言でいうと

```
compose.yaml ──(compose-go)──▶ types.Project（望ましい状態）
                                     │
dockerd ──(ラベルで一覧)──▶ ObservedState（実際の状態のスナップショット）
                                     │
                  reconcile()  ◀─────┘   ← API を一切呼ばない純粋な計算
                       │
                       ▼
                 Plan（操作の DAG）       ← 1 ノード = Docker API 呼び出し 1 回
                       │
              planExecutor.run()         ← DAG を並列に実行
                       │
                       ▼
                Docker Engine API
```

3 段に分けた理由は、**判断（何をするか）と実行（どうやるか）を分ける**ため。

- `reconcile()` は API を呼ばない。だから、`ObservedState` を手で組み立てるだけで、Docker 無しに単体テストできる。
- executor は判断をしない（[executor.go](https://github.com/docker/compose/blob/289d8a037874a5197c54cfa055d28cc5d980593f/pkg/compose/executor.go#L33-L35) の doc コメント「It carries no decision logic」）。
- 古い仕組みでは、create と start がそれぞれ dockerd に問い合わせていた。そのため、start は create がいま何をしたかを知らなかった（[up.go の upDetached の doc コメント](https://github.com/docker/compose/blob/289d8a037874a5197c54cfa055d28cc5d980593f/pkg/compose/up.go#L64-L70)）。

## 1. 観測: `preparePlan` と `ObservedState`

入口は [create.go の `preparePlan`](https://github.com/docker/compose/blob/289d8a037874a5197c54cfa055d28cc5d980593f/pkg/compose/create.go#L103-L180)。
コンテナ・ネットワーク・ボリュームには、まだ何も手を付けない。順に次のことをする。

1. `WithoutUnresolvedOptionalDependencies()` で、profiles やサービスの指定のせいで宙に浮いた、任意（`required: false`）の `depends_on` を消す
2. コンテナ名の重複を確かめる
3. **plan の外で用意するもの**を先にそろえる
   - イメージ（pull / build: `ensureImagesExists`）
   - AI モデル（`ensureModels`）
   - external なネットワーク・ボリュームの存在確認
4. `collectObservedState` で dockerd に問い合わせ、`ObservedState` を作る
5. orphan（compose.yaml から消えたサービスのコンテナ）があれば警告する
6. `reconcile()` を呼んで Plan を作る

[`ObservedState`](https://github.com/docker/compose/blob/289d8a037874a5197c54cfa055d28cc5d980593f/pkg/compose/observed_state.go#L37-L57) は、プロジェクトのラベルが付いたリソースのスナップショット。

| フィールド | 中身 |
| ---------- | ---- |
| `Containers` | サービス名 → コンテナ。各コンテナはラベルを解析済み（`ConfigHash`、`ImageDigest`、`Number` など） |
| `Orphans` | 対応するサービスが無いコンテナ |
| `Networks` / `Volumes` | compose のキー → **同じキーを持つ実体すべて**（名前変更の残骸などで複数ありうる。どれを使うかは reconciler が決める） |
| `HookContainers` | pre_start フックを実行するためのコンテナ |

ここで集めた状態は、Compose が自前のデータベースを持たず、ラベルだけで状態を管理している（[compose-concepts.md](../concepts/compose-concepts.md#ラベルによるリソース管理状態-db-を持たない設計)）ことの具体的な姿でもある。

## 2. 計画: `reconcile()`

[reconcile.go の `reconcile`](https://github.com/docker/compose/blob/289d8a037874a5197c54cfa055d28cc5d980593f/pkg/compose/reconcile.go#L166-L217) は、望ましい状態と観測した状態を比べて Plan を返す。

### スコープ（どのフェーズを計画するか）

[`ReconcileScope`](https://github.com/docker/compose/blob/289d8a037874a5197c54cfa055d28cc5d980593f/pkg/compose/reconcile.go#L49-L61) で、計画するフェーズを選ぶ。

| スコープ | Create フェーズ | Start フェーズ | 使う人 |
| -------- | :-: | :-: | ------ |
| `ScopeCreate` | ○ | — | `create`、前景の `up` の前半、`run` の依存サービス |
| `ScopeCreateStart` | ○ | ○ | `up -d`、`scale`、watch の rebuild |
| `ScopeStart` | — | ○ | `start`、前景の `up` の後半、`run` の依存サービスの起動 |

- **Create フェーズ**: ネットワーク・ボリューム・コンテナを望ましい形にする（作る、作り直す、減らす）。
- **Start フェーズ**: コンテナを running にする（依存の条件待ち → pre_start → 起動 → post_start）。

### Create フェーズの中身

`reconcile()` は次の順に計画する。

1. `reconcileNetworks`: 無いネットワークを作る。設定が変わったネットワークは作り直す（つないでいるコンテナを止め、新しいネットワークにつなぎ直す）
2. `reconcileVolumes`: 無いボリュームを作る。設定が変わったボリュームは、**データが消えるので `prompt` でユーザーに確認してから**作り直す
3. `reconcileContainers`: サービスを依存順にたどり、`reconcileService` で各サービスを処理する
4. `--remove-orphans` のときは orphan を消す

[`reconcileService`](https://github.com/docker/compose/blob/289d8a037874a5197c54cfa055d28cc5d980593f/pkg/compose/reconcile.go#L767-L950) がサービスごとにすること:

- **provider サービス**なら、`OpRunProvider` を 1 つ置くだけ
- 既存のコンテナを、作り直すべきものから順に並べる（`sortContainers`）
- コンテナの数が多すぎれば、`Stop` → `Remove` で減らす（scale down）
- [`mustRecreate`](https://github.com/docker/compose/blob/289d8a037874a5197c54cfa055d28cc5d980593f/pkg/compose/reconcile.go#L1021-L1048) が true なら作り直す。判定は次の順:
  1. 方針が `never`（`--no-recreate`）なら作り直さない。`force`（`--force-recreate`）なら作り直す
  2. 名前空間やボリュームを共有している親サービス（`network_mode: service:x` や `volumes_from` など）が作り直されるなら、作り直す
  3. ラベル `config-hash` が、今の設定から計算したハッシュと違えば、作り直す
  4. イメージの digest、image volume の digest が違えば、作り直す
  5. 動いているコンテナのネットワークの接続が違う、またはボリュームが違えば、作り直す
- それ以外の、最新の状態のコンテナは何もしない。paused や dead なら `OpStartContainer` を置く
- コンテナの数が足りなければ、`OpCreateContainer` を足す（scale up）

方針は、`up` で指定したサービスには `Recreate`、それ以外（依存先）には `RecreateDependencies` が使われる。

### 作り直しは 4 つの操作に分かれる

[`planRecreateContainer`](https://github.com/docker/compose/blob/289d8a037874a5197c54cfa055d28cc5d980593f/pkg/compose/reconcile.go#L1615-L1700) は、作り直しを 4 つの操作に分ける。

```
CreateContainer(一時的な名前) → StopContainer(古い) → RemoveContainer(古い) → RenameContainer(新しい → 正式な名前)
```

- **「新しいコンテナができてから、古いものを消す」**のが不変条件。新しいコンテナを作れなかったら、古いコンテナはそのまま残る。
- 一時的な名前は `<古いIDの先頭12文字>_<正式な名前>`。
- `Inherit` が true のときだけ、古いコンテナを `Inherited` として渡す。新しいコンテナは、その匿名ボリュームを引き継ぐ（`up` は `-V` が無ければ true。`run` の依存サービスでは渡されていない → [next-tasks.md の 2](../tracking/next-tasks.md#2-run-が依存サービスを作り直すと匿名ボリュームを引き継がない)）。
- 作り直すサービスに依存しているサービスは、先に止める（`planStopDependents`）。

### Start フェーズの中身

[`planStartPhase`](https://github.com/docker/compose/blob/289d8a037874a5197c54cfa055d28cc5d980593f/pkg/compose/reconcile.go#L1140-L1159) は、サービスを依存順にたどって、次のような鎖を作る。

```
[依存の条件待ち] → RunPreStart（1 回だけ） → StartContainer → RunPostStart → (次のレプリカ) StartContainer → ...
```

- `depends_on` の `condition: service_started` には、待ちのノードを作らない。依存先の鎖の最後のノードへの辺を張るだけ。
- `service_healthy` と `service_completed_successfully` は、`OpWaitCondition` ノードになる（[`waitConditionNode`](https://github.com/docker/compose/blob/289d8a037874a5197c54cfa055d28cc5d980593f/pkg/compose/reconcile.go#L1353-L1387)）。
  - 同じ（サービス, 条件）を待つ依存元が何個あっても、ノードは 1 つにまとめる。
  - `required: false` なら `BestEffort` にする（失敗しても plan を止めない）。必須の依存元が 1 つでもあれば、必須に格上げする。
  - **health は計画のときではなく、実行のときに確かめる**。plan に書くのは「何を待つか」だけ。
- レプリカは 1 つずつ順に起動する（以前の仕組みと同じ順序を保つため）。

## 3. Plan のデータ構造

[plan.go](https://github.com/docker/compose/blob/289d8a037874a5197c54cfa055d28cc5d980593f/pkg/compose/plan.go) がすべて。

- `Plan`: `PlanNode` の配列。**トポロジカル順**（依存先が先）に並ぶ。
- `PlanNode`: `ID`、`Operation`、`DependsOn`（前提のノード）、`Group`（イベント表示用のまとまり。例 `recreate:web:1`）、`Phase`（Create か Start）。
- `Operation`: 種類と対象、理由（`Cause`）、操作ごとに必要なデータ。

| 種類 | 実行時の API 呼び出し（[executor_ops.go](https://github.com/docker/compose/blob/289d8a037874a5197c54cfa055d28cc5d980593f/pkg/compose/executor_ops.go)） |
| ---- | ---- |
| `CreateNetwork` / `RemoveNetwork` / `ConnectNetwork` / `DisconnectNetwork` | `NetworkCreate` / `NetworkRemove` / `NetworkConnect` / `NetworkDisconnect` |
| `CreateVolume` / `RemoveVolume` | `VolumeCreate` / `VolumeRemove` |
| `CreateContainer` | `ContainerCreate`（`createMobyContainer`） |
| `StartContainer` | secrets と configs をコンテナに注入してから `ContainerStart` |
| `StopContainer` / `RemoveContainer` / `RenameContainer` | `ContainerStop` / `ContainerRemove` / `ContainerRename` |
| `CreateHookContainer` | pre_start フックを実行するコンテナを作る |
| `WaitCondition` | health / 終了を、実行時に観測して待つ |
| `RunPreStart` / `RunPostStart` | フックを実行する |
| `RunProvider` | provider のプラグインを実行する |

その他の性質:

- `BestEffort`: 失敗しても plan を止めない操作（ネットワーク名を変えたときの古いネットワークの削除、古いフックコンテナの掃除、任意の依存の待ち）。
- `Plan.String()` は、人が読める形で plan を出力する。**テストはこの文字列を丸ごと比べる**。

```
[] -> #1 network:default, CreateNetwork, not found
[1] -> #2 service:web:1, CreateContainer, no existing container
[2] -> #3 service:web:1, StopContainer, replaced by #2 [recreate:web:1]
```

1 行の読み方は `[前提ノードのID] -> #自分のID 対象, 操作, 理由 [グループ]`。

## 4. 実行: `planExecutor`

[executor.go の `run`](https://github.com/docker/compose/blob/289d8a037874a5197c54cfa055d28cc5d980593f/pkg/compose/executor.go#L120-L220) が DAG を実行する。

- **ノードごとに done チャネル**を作る。各ノードは自分の goroutine で、`DependsOn` のチャネルがすべて閉じるのを待ってから動く。
- **Create フェーズ → Start フェーズの順に、2 回に分けて実行する**。Create フェーズがすべて成功するまで、Start フェーズのノードは 1 つも動かない（関係の無いサービスの作成中に別のサービスが起動し始める、ということが無い）。
  Create のノードが Start のノードに依存していたら、デッドロックするので即エラーにする。
- 並列度は `maxConcurrency`（`--parallel` / `COMPOSE_PARALLEL_LIMIT`）で制限する。
- 失敗の扱い:
  - 失敗したノードを記録する。依存元はそれを見て、自分を実行しない。
  - そのフェーズの context をキャンセルして、兄弟のノードを止める。
  - 報告するエラーは、最初に失敗した本来のエラー。キャンセルで巻き込まれたノードの `context.Canceled` ではない。
- ノードの結果（作ったコンテナの ID など）は `reconciliationContext` に入れる。後のノードは `CreateNodeID` で参照する（Rename は、Create が作ったコンテナの ID を知る必要がある）。
- 進捗の表示は `groupTracker`（[executor_events.go](https://github.com/docker/compose/blob/289d8a037874a5197c54cfa055d28cc5d980593f/pkg/compose/executor_events.go)）が、グループ単位でまとめる。
  - 作り直し: `Recreate` → `Recreated`
  - 起動の鎖（pre_start、start、post_start）: 1 行の `Starting` → `Started`
- `--wait-timeout` は、`OpWaitCondition` ノードだけを、**それぞれ独立した時間枠で**制限する。plan 全体の締め切りにすると、関係の無い作業が遅いだけで「healthy にならなかった」と誤って失敗するため（[up.go](https://github.com/docker/compose/blob/289d8a037874a5197c54cfa055d28cc5d980593f/pkg/compose/up.go#L82-L97) のコメント）。

## 5. どのコマンドが plan engine を通るか（2026-10-17 時点）

| コマンド | 経路 |
| -------- | ---- |
| `up -d` | `upDetached` → `preparePlan(ScopeCreateStart)` → executor（1 つの plan） |
| `up`（前景） | `create`（`ScopeCreate`）→ attach、ログ、monitor の準備 → `start`（`ScopeStart`）。**2 つの plan に分かれている**（[up.go:310](https://github.com/docker/compose/blob/289d8a037874a5197c54cfa055d28cc5d980593f/pkg/compose/up.go#L310)） |
| `create` | `ScopeCreate` |
| `start` | `ScopeStart`。プロジェクトはファイルか、ラベルから組み立て直したもの（`projectFromName`）。#14296 で 10-16 に plan に移った |
| `scale`、watch の rebuild | `createAndStart`（`ScopeCreateStart`） |
| `run` | 依存サービスに `create` → `start`。oneoff のコンテナ自体は plan の外 |
| `stop`、`restart`、`down`、`kill` | **まだ古い仕組み**。`InDependencyOrder` / `InReverseDependencyOrder`（[dependencies.go](https://github.com/docker/compose/blob/289d8a037874a5197c54cfa055d28cc5d980593f/pkg/compose/dependencies.go#L85-L100)）でサービスを順にたどり、直接 API を呼ぶ |
| classic build | `InDependencyOrder` |

## 6. 移行の状況（エピック [#14081](https://github.com/docker/compose/issues/14081)）

ndeloof が一人で進めている。

- **Lot 1**（完了）: executor の土台。#14274、#14289。
- **Lot 2**（進行中）:
  - 済み: `up -d` を 1 つの plan で動かす（#14290）、`scale` と watch の rebuild（#14291）、`start` の start-only plan（#14296、10-16 マージ）
  - 残り: 前景の `up` を 1 つの plan にする（#14305 ほか、draft）
- **Lot 3**（未着手）: 古い経路の削除。

このあたりのファイル（`reconcile*.go`、`executor*.go`、`plan.go`、`observed_state.go`、`up.go`、`start.go`）は毎日のように変わるので、**外部から PR を出すのは避ける**（[next-tasks.md のやらないもの](../tracking/next-tasks.md#やらないもの理由つき)）。
ただし、読んで理解しておくと、ほかの場所のバグの原因（どのオプションが plan に渡っているか、など）を見つけるのに役立つ。

## 7. 読むとき・テストするときのコツ

- **読む順番**:
  1. `plan.go`（データ構造）
  2. `create.go` の `preparePlan`
  3. `reconcile.go` の `reconcile` → `reconcileService` → `planStartPhase`
  4. `executor.go` の `run`
  5. `executor_ops.go`
- **reconcile のテスト**（`reconcile_test.go`、`reconcile_start_test.go`）は、`types.Project` と `ObservedState` を手で組み立てて、`reconcile()` の結果を `plan.String()` と比べる。Docker は要らない。
  plan の形がどう変わるかを確かめたいときは、まずここに表形式のケースを足すのが早い。

  ```go
  plan, err := reconcile(t.Context(), project, observed, defaultReconcileOptions(), noPrompt)
  assert.NilError(t, err)
  assert.Equal(t, plan.String(), strings.TrimSpace(`
  [] -> #1 network:backend, CreateNetwork, not found
  [] -> #2 network:frontend, CreateNetwork, not found
  `)+"\n")
  ```

- **executor のテスト**（`executor_*_test.go`）は、mock の API クライアントで、呼ばれる順番とイベントを確かめる。
- doc コメントがとても詳しい。「なぜこうなっているか」は、ほぼコメントとレビューの議論（PR 番号がコメントに書かれている）で追える。
