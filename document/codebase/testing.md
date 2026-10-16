# テストと CI

← [document 一覧](../README.md)

docker/compose で、メンテナに受け入れられるテストを書くための作法と、CI の実際。
コマンドの一覧は [development.md](./development.md)、PR を出す前の確認は [contributing/checklist.md](../contributing/checklist.md) を参照。

- 基準コード: `upstream/main` 289d8a037（2026-10-16）。`file:line` は upstream の行番号
- 最初に読むもの: [pkg/e2e/SCENARIO.md](../../pkg/e2e/SCENARIO.md)（e2e の規約）、CONTRIBUTING.md の 345-390 行あたり（テストとコメントのルール）

## 1. ユニットテストの作法

| 項目 | このリポジトリのやり方 |
| ---- | ---------------------- |
| アサーション | **gotest.tools/v3 だけ**。`assert`、`assert/cmp`（`is` という別名で import）、`poll`。testify は `.golangci.yml` の depguard で**禁止** |
| context | **`t.Context()`**。テストで `context.Background()` / `context.TODO()` を使うと、forbidigo で lint エラー |
| 形 | 表形式（`[]struct` + `t.Run`）が多い。ループ変数は `tt` か `tc` |
| 名前 | `TestXxx` か `TestXxx_何を確かめるか`。テストの上に「どの不変条件を守っているか」のコメントを書くのが慣習（コメントは経緯ではなく、今の振る舞いを書く） |
| 並列 | ユニットテストでは `t.Parallel()` はほとんど使わない |
| race | CI のユニットテストは常に `-race` 付き |
| 時間 | fake clock は使わない。Go 標準の `testing/synctest` を使う（`pkg/watch/debounce_test.go`、`pkg/compose/watch_test.go`） |
| TTY | `github.com/creack/pty` で偽の TTY を作る。そのファイルは `//go:build !windows` になる |

```go
// 表形式の例（cmd/compose/compose_progress_test.go のような形）
for _, tc := range []struct {
	name string
	ansi string
	want string
}{ ... } {
	t.Run(tc.name, func(t *testing.T) {
		got := ...
		assert.Equal(t, got, tc.want)
	})
}
```

## 2. mock（gomock）で Docker 無しにテストする

### 生成

[Makefile](https://github.com/docker/compose/blob/289d8a037874a5197c54cfa055d28cc5d980593f/Makefile#L102-L112) の `make mocks` で生成する。`//go:generate` は使わない。

- mock の対象は 3 つ。
  - docker CLI の `command.Cli`（`pkg/mocks/mock_docker_cli.go`）
  - moby の `client.APIClient`（`pkg/mocks/mock_docker_api.go`）
  - `pkg/api/api.go` のインターフェース
- mockgen は v0.4.0 に固定されている。v0.5 では、moby の引数名 `_` のせいで壊れたコードが生成されるため。
- CI の `validate-mocks` が `make mocks` を再実行して差分を確かめる。**インターフェースを変えたら `make mocks` を忘れない**。

### composeService の組み立て方

共通のヘルパーは無い。テストファイルごとに小さなヘルパーを作るのが流儀。

```go
// 昔からある形（pkg/compose/down_test.go の prepareMocks）
mockCtrl := gomock.NewController(t)
api := mocks.NewMockAPIClient(mockCtrl)
cli := mocks.NewMockCli(mockCtrl)
cli.EXPECT().Client().Return(api).AnyTimes()
tested, err := NewComposeService(cli)
```

- 新しい形として、`executor_test.go` の `newTestService` と `start_test.go` の `newStartTestService` がある。
  - `Ping` や `ClientVersion` のような、毎回呼ばれる API を `AnyTimes()` で先に埋めておく。
  - イベントを記録する processor を渡す。
- **期待（`EXPECT()`）していない API 呼び出しはテスト失敗になる**。だから、「何も EXPECT しない」こと自体が「この経路は daemon に触らない」というアサーションになる（`start_test.go` のコメント）。
- 期待の書き方:
  - 引数は、具体的な Options の構造体で照合するのが基本（例: `client.ContainerRemoveOptions{Force: true}`）。
  - 何でもよい引数は `gomock.Any()`。
  - 順番は `gomock.InOrder` か `.After(call)` で指定する。

### plan engine のテストはもっと単純

`reconcile()` は API を呼ばないので、mock すら要らない。`types.Project` と `ObservedState` を手で作り、`plan.String()` を比べる（[plan-engine.md の 7](./plan-engine.md#7-読むときテストするときのコツ)）。

## 3. その他のパターン

- **ヘルパープロセス**: 外部の実行ファイルを呼ぶコードは、テストのバイナリ自身を子プロセスとして再実行してテストする（`internal/provider/plugins_control_test.go`）。

  ```go
  cmd := exec.Command(os.Args[0], "-test.run=TestHelperProviderPull")
  cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
  // ...
  func TestHelperProviderPull(t *testing.T) {
  	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
  		t.Skip("helper process for ...")
  	}
  	// 子プロセスとして振る舞う
  }
  ```

- **golden ファイル**: 使っていない（BUILDING.md の golden の更新手順は古い）。ユニットテストの testdata は `pkg/compose/testdata/` だけ。
- **build tag**:
  - `e2e`: e2e のテスト（`pkg/e2e/*_test.go`）
  - `standalone`: e2e を standalone モードで動かす
  - `!windows`: pty を使うテスト（`hook_test.go` など）。それ以外にも、`start_test.go` などに連鎖して付いている（下の 7 を参照）
  - `fsnotify`: watch の実装の切り替え

## 4. e2e（`pkg/e2e`）

### Scenario DSL（新しいテストは必ずこれ）

規約は [SCENARIO.md](../../pkg/e2e/SCENARIO.md) にある。要点:

- **新しいテストは `NewScenario` で書く**。古い `NewCLI` 形式のテストに書き足さない。
- 1 つのテストで確かめる不変条件は 1 つ。intent は「X must Y」の形で書く。
- Step の名前は振る舞いを説明する文にする。コマンドの繰り返しにしない。
- compose ファイルは **`pkg/e2e/testdata/<TestName>/compose.yaml`** に置く。テスト同士で共有しない。どのテストにも使われていないディレクトリがあると、`TestTestdataDirsAreOwned` が失敗する。
  （AGENTS.md / CLAUDE.md には「inline compose model」と書いてあるが、実装と SCENARIO.md は testdata 方式。**こちらに従う**）
- **状態で確かめる**: `ServiceState`、`NotRecreated`、`LabelSet` など。出力の文字列一致（`OutputContains`）は最後の手段。
- `time.Sleep` は禁止。`Eventually(check, timeout)` を使う（中身は `poll.WaitOn`、500ms ごとに状態を取り直す）。
- 新しいチェックは、汎用的な名前で `checks.go` に足す。
- 回帰テストには、テストの上のコメントに issue のリンクと、どう壊れていたかを書く。

```go
// pkg/e2e/restart_test.go（compose ファイルは testdata/TestRestart/compose.yaml）
NewScenario(t, "restart must bring an exited service back up, restarting the same container").
	Step("up starts the service, whose first run exits at once",
		ComposeCmd("up", "-d"),
		Eventually(ServiceState("app", "exited"), 10*time.Second)).
	Step("restart brings the service back up, reusing the container",
		ComposeCmd("restart"),
		Eventually(ServiceState("app", "running"), 10*time.Second),
		NotRecreated("app"))
```

その他の機能:

- `NewScenario` はデフォルトで並列に実行される。daemon 全体の状態を変えるシナリオは `Serial()` にする。
- プロジェクト名はテスト名から自動で決まる（例: `TestFooBar` → `e2e-foo-bar`）。前後に `down -v --remove-orphans` を実行する。
- 条件を満たさないときに skip する: `Requires(ContainerdImageStore)` / `Requires(EngineVersionAtLeast(n))`。
- `Env`、`Defer`、Action の修飾子（`MayFail()`、`WithEnv`、`WithStdin`、`Within(timeout)`）。

### 失敗したとき

- `artifacts: <dir>` に、compose.yaml、各 step の出力、コンテナ一覧、events、ログ、状態のスナップショットが残る。
- `E2E_KEEP_FAILED=1` を付けて実行し直すと、後片付けをしない。調べ終わったら `docker compose --project-name <p> down -v --remove-orphans` で消す。

### 手元で 1 本だけ動かす

```bash
make build                                              # bin/build/docker-compose を作る（無いと WARNING が出て動かない）
go test -tags e2e ./pkg/e2e/ -run TestRestart -v        # plugin モード
make build-and-e2e-compose E2E_TEST=TestRestart         # 同じことを make で
```

- テストのバイナリは `COMPOSE_E2E_BIN_PATH`、無ければ `bin/build/docker-compose` から探す。
- buildx と docker-model のプラグインは、`~/.docker/cli-plugins` などから探す。見つからなくても WARNING だけで、テストは走る。
- provider 系のテストには `make example-provider` が要る。

## 5. make のターゲット

| 目的 | コマンド | 実体 |
| ---- | -------- | ---- |
| ユニットテスト（CI と同じ） | `make test` | `docker buildx bake test` → Dockerfile の中で `gotestsum -- -v -race -cover ...`（e2e パッケージは除く）。**Docker と buildx が要る** |
| ユニットテスト（手早く） | `go test ./pkg/... ./cmd/...` | daemon は要らない |
| e2e | `make e2e-compose` / `make e2e-compose-standalone` | `-tags e2e -parallel=4 -timeout 20m` |
| lint | `make lint` | `docker buildx bake lint`（golangci-lint v2.13.2 に固定）。手元の golangci-lint が古い Go でビルドされていると動かないので、こちらが確実 |
| フォーマット | `make fmt` | `gofumpt -w .` |
| mock | `make mocks` | 上の 2 |
| ドキュメント | `make docs` | 下の 7 |
| まとめて検証 | `make validate` | go.mod、ライセンスヘッダー、docs、mocks |

- gotestsum のバージョン（v1.13.0）は、Makefile と Dockerfile の両方に書かれている。
- lint のルールで、CLAUDE.md の「Max cyclomatic complexity: 16」は古い。今は gocyclo が無効で、代わりに **gocognit（min-complexity 30）**。早期 return は数えず、ネストを重く数える。

## 6. CI（`.github/workflows/ci.yml`）

| ジョブ | 中身 |
| ------ | ---- |
| `validate` | lint、go.mod、ヘッダー、docs、mocks、relay の lint とテスト |
| `binary` | 11 プラットフォームのクロスビルド（Windows を含む） |
| `test` | ユニットテスト（`make test` と同じ）とカバレッジ |
| `e2e` | 6 本の matrix。mode（plugin / standalone）× Engine（29 stable で graphdriver と containerd image store、28 oldstable で graphdriver）。OS は ubuntu だけ |
| `coverage` | ユニットと e2e のカバレッジをまとめて Codecov に送る |
| `ci-ok` | ブランチ保護用の、ひとまとめの必須チェック |

知っておくこと:

- **fork からの PR は、メンテナが承認するまで CI が動かない**（`action_required` になる）。push し直すたびに承認が要る（#14310、#14351 で経験済み）。
- e2e の Engine は `get-docker.sh --version N` で入れ直す。docker-model は apt の `docker-model-plugin` で入る。バージョンは固定されておらず、apt の test チャネルの最新になる（#14284 で経験済み）。
- **Windows のテストはコンパイルも実行もしない**。バイナリのクロスビルドだけ。だから、Windows でテストが壊れても CI は気付かない（下の 7）。
- `ai-disclosure-gate.yml`: PR に `AI_AGENT_DISCLOSURE.md` が含まれていると、マージをブロックする。CONTRIBUTING.md の指示は「自分でレビューしたら、PR を出す前に消す」。
- DCO はワークフローではなく、外部のステータスとしてチェックされる（GitHub App と推定）。
- fork の PR には、AI のレビュー（`pr-review.yml`）が付く。

## 7. ドキュメントの生成（フラグを変えたら）

- `docs/reference/` の yaml と md は、cobra のコマンド定義から生成される（`docs/yaml/main/generate.go`）。
- md の `<!---MARKER_GEN_START-->` から `<!---MARKER_GEN_END-->` の間は自動生成。外側の説明文は手書きで、生成しても残る。
- **フラグを足したり説明を変えたりしたら、`make docs` を実行して、`docs/reference/` の差分もコミットする**。忘れると CI の `validate-docs` が「Docs result differs」で失敗する。

## 8. 落とし穴

- **Windows でテストがコンパイルできない**: `start_test.go` は `//go:build !windows` なのに、そこで定義したヘルパー（`newStartTestService`）を、タグの無い `down_test.go` などが使っている。
  - upstream main でも `GOOS=windows go vet ./pkg/compose/` が `undefined: newStartTestService` で失敗する。
  - 教訓: **`!windows` のファイルに置いたヘルパーを、タグの無いファイルから使わない**（[next-tasks.md の 13](../tracking/next-tasks.md#13-pkgcompose-のテストが-windows-でコンパイルできない)）。
- **ANSI がテストの出力に混ざる**: os.Stdout にエスケープシーケンスを書くコードをテストすると、gotestsum の集計が壊れる。テストでは `disableAnsi = true` にする（#14351）。e2e で出力を比べるときは `--ansi=never` を付ける。
- **watch の e2e は不安定になりやすい**:
  - 初回の起動は CI で 10 秒を超えることがあるので、最初の待ちだけ 2 分にしている（`pkg/e2e/watch_test.go`）。
  - 「Watch enabled」が出ても、コンテナの起動が終わったとは限らない（[next-tasks.md の 6](../tracking/next-tasks.md#6-e2e-testwatchsyncintosymlinkeddirectory-が不安定)）。
- **`make lint` の build tag は `e2e` だけ**: `standalone` タグのファイルは lint されない（推定）。
