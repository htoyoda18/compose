# 学習ロードマップ

段階ごとに「概念 → その概念が現れるコード」の順で読む。概念ドキュメント(`concepts/`)、コードの解説(`codebase/`)、
ディレクトリ・ファイル単位のコードノート(`note.md` / `xxx.go.md`)を混ぜて並べている。
(Q&A)は直前のファイルを読んで疑問が出たときに開けばよい。

← [document 一覧](./README.md) / [コードノート入口](../note.md)

## Stage 1: 全体像を掴む

Compose が何をするツールで、リポジトリがどう分かれているかを掴む。

1. [concepts/intro.md](./concepts/intro.md) — そもそも Docker / Compose とは
   - (Q&A)[concepts/intro-qa.md](./concepts/intro-qa.md)
2. [concepts/how-compose-works.md](./concepts/how-compose-works.md) — ラベルと差分埋めで動く仕組みを例で追う
3. [codebase/overview.md](./codebase/overview.md) — このリポジトリの概要
4. [コードノート入口(note.md)](../note.md) — ルートディレクトリの役割と依存関係の図
5. [codebase/architecture.md](./codebase/architecture.md) — レイヤー構成(cmd / pkg/api / pkg/compose / internal)

## Stage 2: コマンドが本体に届くまで

`docker compose up` を打ってから `pkg/compose` が呼ばれるまでを追う。Go の基礎もここで押さえる。

1. [concepts/go.md](./concepts/go.md) — goroutine / context / interface など、コードを読むのに必要な Go
   - (Q&A)[concepts/go-qa.md](./concepts/go-qa.md)
2. [codebase/flow.md の「共通の流れ」](./codebase/flow.md#共通の流れ) — main.go から pkg/compose まで
3. [cmd/note.md](../cmd/note.md) → [cmd/main.go.md](../cmd/main.go.md) — CLI 層とエントリーポイント
4. [codebase/cli.md](./codebase/cli.md) — cobra、プラグイン版 / スタンドアロン版、Docker コンテキスト
5. [cmd/compose/note.md](../cmd/compose/note.md) → [cmd/compose/up.go.md](../cmd/compose/up.go.md) — サブコマンドの実装
6. [pkg/note.md](../pkg/note.md) → [pkg/api/note.md](../pkg/api/note.md) → [pkg/api/api.go.md](../pkg/api/api.go.md) — CLI と本体をつなぐインターフェース

## Stage 3: up と差分埋め(Compose の中核)

「望ましい状態との差分を埋める」が、実装でどうなっているかを追う。

1. [concepts/compose-concepts.md](./concepts/compose-concepts.md) — ラベル、収束処理、依存グラフ
2. [concepts/compose-lifecycle.md](./concepts/compose-lifecycle.md) — depends_on、作り直しの判定と config hash、ラベル一覧
3. [codebase/plan-engine.md](./codebase/plan-engine.md) — 観測 → 計画 → 実行
4. [pkg/compose/note.md](../pkg/compose/note.md) → [up.go.md](../pkg/compose/up.go.md) → [create.go.md](../pkg/compose/create.go.md) → [containers.go.md](../pkg/compose/containers.go.md)
5. [codebase/reading-guide.md](./codebase/reading-guide.md) — `up -d` をコードで最後まで読む順番
6. [pkg/api/labels.go.md](../pkg/api/labels.go.md) — ラベル定数

## Stage 4: Docker の下回り

Compose が呼んでいる Docker Engine 側と、その周辺の部品を理解する。

1. [concepts/docker-engine.md](./concepts/docker-engine.md) — dockerd / containerd / runc、Engine API、BuildKit
   - (Q&A)[concepts/docker-engine-qa.md](./concepts/docker-engine-qa.md)
2. [concepts/container-tech.md](./concepts/container-tech.md) — namespace / cgroup / overlayfs
   - (Q&A)[concepts/container-tech-qa.md](./concepts/container-tech-qa.md)
3. [concepts/networking.md](./concepts/networking.md) — bridge / veth / DNS。Compose のネットワークの実体
   - (Q&A)[concepts/networking-qa.md](./concepts/networking-qa.md)
4. [concepts/oci.md](./concepts/oci.md) — イメージ、レイヤー、レジストリ
   - (Q&A)[concepts/oci-qa.md](./concepts/oci-qa.md)
5. [internal/note.md](../internal/note.md) → [codebase/internal.md](./codebase/internal.md) — OCI レジストリ操作、Desktop 連携、PID ファイルなど
6. [pkg/remote/note.md](../pkg/remote/note.md) — `oci://` / `git://` の compose ファイルを取ってくる仕組み

## Stage 5: 他のコマンドへ横展開

up で掴んだ形を、ほかのコマンドに当てはめていく。

1. [codebase/flow.md の「コマンド別の流れ」](./codebase/flow.md#コマンド別の流れpkgcompose-内) — down / build / run / logs / watch
2. [pkg/compose/down.go.md](../pkg/compose/down.go.md)、[build.go.md](../pkg/compose/build.go.md)、[watch.go.md](../pkg/compose/watch.go.md)、[attach.go.md](../pkg/compose/attach.go.md)、[wait.go.md](../pkg/compose/wait.go.md)
3. [cmd/display/note.md](../cmd/display/note.md)、[cmd/formatter/note.md](../cmd/formatter/note.md) — 進捗表示と出力の整形
4. [pkg/bridge/note.md](../pkg/bridge/note.md)、[relay/note.md](../relay/note.md) — bridge と provider の relay

## Stage 6: 開発して PR を出す

1. [codebase/development.md](./codebase/development.md) — ビルド・Lint・CI・リリース
2. [codebase/testing.md](./codebase/testing.md) — テストの作法
3. [pkg/e2e/SCENARIO.md](../pkg/e2e/SCENARIO.md) — e2e テストの Scenario DSL
4. [contributing/README.md](./contributing/README.md)、[CONTRIBUTING.md](../CONTRIBUTING.md) と [AI_POLICY.md](../AI_POLICY.md)
5. [contributing/pr-trends.md](./contributing/pr-trends.md) → [contributing/pr-review-patterns.md](./contributing/pr-review-patterns.md) → [contributing/checklist.md](./contributing/checklist.md)
6. [tracking/next-tasks.md](./tracking/next-tasks.md)、[tracking/issues.md](./tracking/issues.md)、[tracking/todo.md](./tracking/todo.md)、[tracking/fixme.md](./tracking/fixme.md) — 次にやるタスクを探す
