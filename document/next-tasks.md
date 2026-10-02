# 次に着手できるタスク候補

[pr-trends.md](./pr-trends.md) の結論（実在する問題か／根本原因を直すか／メンテナの作業とぶつからないか／
メンテナの後押しがあるか）を物差しにして、「まだ誰も着手していない、外部から着手できて、
プロジェクトを本当に良くするタスク」を探した結果。

- 調査日: 2026-10-03
- 基準コード: `upstream/main` 42f48072b（2026-10-02）
- 調べたもの: open issue 全 51 件、open PR 全 37 件、2026-08-15 以降にマージされた PR 約 130 件の本文とレビューコメント、
  エピック [#14074](https://github.com/docker/compose/issues/14074) と [#14081](https://github.com/docker/compose/issues/14081) の未チェック項目、
  新しく追加された TODO と `t.Skip`
- 「検証」列の意味: **実機** = 自分で upstream main をビルドして再現した / **コード** = 現行コードを読んで確認した（実行はしていない）

## 結論（先にやる順）

| 順 | タスク | 種類 | 検証 | 後押し | 衝突リスク | 規模 | 最初の一手 |
| -: | ------ | ---- | ---- | ------ | ---------- | ---- | ---------- |
| 1 | [`ensureModels` のエラー無視とデータ競合](#1-ensuremodels-のエラー無視とデータ競合) | バグ | コード | ◎ メンテナがレビューで「フォローアップに値する」 | 低 | S | そのまま PR（レビューコメントを引用） |
| 2 | [`start -p` で任意の依存サービスが必須扱いになる](#2-start--p-で任意の依存サービスが必須扱いになる) | バグ（未報告） | **実機** | — | 中〜高（#14081 Lot 2） | S | **issue を立てる**（#14081 にも一言） |
| 3 | [`down` の pre_start フックコンテナ後始末の残作業](#3-down-の-pre_start-フックコンテナ後始末の残作業) | バグ | コード | ◎ メンテナが「別 PR で出してよい」 | 低〜中 | S〜M | そのまま PR（レビューコメントを引用） |
| 4 | [classic build の `--push` がプロジェクト全体をサービス数だけ push する](#4-classic-build-の---push-がプロジェクト全体をサービス数だけ-push-する) | バグ | コード | ○ エピック #14074 に 🐛 として記載 | 低 | S | issue を立てる（エピックから切り出す） |
| 5 | [`config -q` が一覧系フラグと組み合わせると出力する](#5-config--q-が一覧系フラグと組み合わせると出力する) | バグ | **実機** | ○ エピック #14074 E に記載 | 中（#14046） | S〜M | issue を立てる |
| 6 | [classic build のイメージに compose ラベルが付かない](#6-classic-build-のイメージに-compose-ラベルが付かない) | バグ | コード | ○ エピック #14074 F | 低 | S | 4 と同じ issue でまとめて相談 |

1 と 3 は、メンテナが「誰かやってね」とレビューで書き残したまま放置されているものなので、**issue を立てなくても受け入れられる見込みが最も高い**。
信頼を積むなら 1 → 3 → 2 の順が良い。2 は唯一、実機で再現できた未報告のバグで価値は最も高いが、メンテナの大改修とぶつかる位置にあるので issue を立てて判断を仰ぐ。

---

## 1. `ensureModels` のエラー無視とデータ競合

- **出典**: ndeloof の [#14177](https://github.com/docker/compose/pull/14177) インラインレビュー（`model.go:49`）
  「Two pre-existing issues here… worth a follow-up」
- **場所**: [pkg/compose/model.go](../pkg/compose/model.go) `ensureModels`
  - `availableModels, err := mdlAPI.ListModels(ctx)` の `err` を一度も見ていない（失敗すると空リスト扱いになり、全モデルを pull しに行く）
  - 各 goroutine が外側の `err` に `err = mdlAPI.PullModel(...)` と代入している → 複数のモデルを pull すると**データ競合**
- **直し方**: `ListModels` の `err` を返す。goroutine 内は `err :=` にする。
- **テスト**: `model_test.go` は存在しない。`plugins_control_test.go` のヘルパープロセス方式を真似る。
- **着手状況**: 「ensureModels」「ListModels」で検索した限り、issue / PR とも無い。このファイルの最終変更は #14177 自身。

## 2. `start -p` で任意の依存サービスが必須扱いになる

- **症状**: `depends_on` に `required: false` を付けた依存サービスが失敗していると、
  ファイルを読む `docker compose start` は警告を出して起動するのに、`docker compose -p <name> start` は失敗して起動しない。
- **実機での再現**（Docker 29.1.3、upstream main ビルド）:
  ```yaml
  services:
    web:
      image: alpine:3.21
      command: sleep infinity
      depends_on:
        init:
          condition: service_completed_successfully
          required: false
    init:
      image: alpine:3.21
      command: sh -c "exit 1"
  ```
  | コマンド（`up -d` → `stop` の後） | 結果 |
  | --------------------------------- | ---- |
  | `docker compose start`            | 警告「optional dependency "init" didn't complete successfully」を出し、`web` は running |
  | `docker compose -p optdep start`  | `service "init" didn't complete successfully: exit 1` で失敗、`web` は exited |
  | `docker compose -p optdep restart` | 正常（`down.go` 側で補正しているため） |
- **根本原因**:
  - `-p` を付けて `-f` を付けないと、compose.yaml があってもファイルを読まない（[cmd/compose/compose.go](../cmd/compose/compose.go) `projectOrName` の `if len(o.ConfigPaths) > 0 || o.ProjectName == ""`）。
    代わりにコンテナのラベルからプロジェクトを組み立て直す（`projectFromName`）。
  - そのラベル `com.docker.compose.depends_on` には `サービス名:condition:restart` しか書かれず、**`required` が保存されない**
    （[pkg/compose/create.go](../pkg/compose/create.go) の `fmt.Sprintf("%s:%s:%t", dep, d.Condition, d.Restart)`）。
  - 読み戻すときに `required := true` と決め打ちしている（[pkg/compose/compose.go](../pkg/compose/compose.go) `projectFromName`）。
  - `stop`/`restart`/`kill`/`down` は `down.go` で `Required = false` に補正しているが、`start` は補正していない。
- **直し方の案**: ラベルに 4 つ目のフィールド `:required` を書く。読み戻すときは、フィールドが無ければ `true`（古いラベルとの互換）。
  あわせてラベルを書く順をソートする（今は map の反復順で毎回変わる）。
- **衝突リスク**: #14081 の Lot 2 に「`compose start` builds a start-only plan — including the label-reconstructed project path (`projectFromName`)」という未着手項目がある。
  修正箇所（`create.go` のラベル書き込みと `compose.go` の読み戻し）は plan engine の外だが、**メンテナに先に聞く**のが筋。
- **着手状況**: issue / PR とも見当たらない（「required false start」「optional dependency start」「depends_on label required」で検索）。

## 3. `down` の pre_start フックコンテナ後始末の残作業

- **出典**: ndeloof の [#14091](https://github.com/docker/compose/pull/14091) レビュー
  「The agreed follow-ups (down scoping vs dependents, list-failure aborting teardown, …) remain open — fine by me to land them separately.」
- **場所**: [pkg/compose/down.go](../pkg/compose/down.go) `removePreStartHookContainers`
  - (a) `down web` のようにサービスを指定すると、フックの後始末は指定サービスだけ。実際のティアダウンは依存先も巻き込むので、依存先のフックコンテナが残る。
  - (b) doc コメントは「失敗してもティアダウンは中断しない」と書いているのに、`ContainerList` が失敗すると `return err` し、
    呼び出し側（`down.go` 120 行付近）も `return err` するので、**ネットワーク・ボリューム・イメージの削除がまるごと飛ぶ**。
  - (c) 任意の改善: `pre_start.go` の `context.Background()` を `context.WithoutCancel(ctx)` に（トレースのスパンを保つため。ndeloof の提案）。
- **衝突リスク**: `down` は #14081 の対象外（明記されている）。#14149 と #14155 が `down.go` を触っているが、別の箇所。
- **やり方**: (b) を最優先にする。小さく、doc コメントとの食い違いという明確な根拠がある。

## 4. classic build の `--push` がプロジェクト全体をサービス数だけ push する

- **出典**: エピック #14074 F「classic `--push` pushes the whole project once per built service」（🐛 印付き、「standalone issue に切り出す価値あり」とされている）
- **場所**: [pkg/compose/build_classic.go](../pkg/compose/build_classic.go) のサービスごとのコールバック内
  `if options.Push { return s.push(ctx, project, api.PushOptions{}) }`。
  `push()`（[pkg/compose/push.go](../pkg/compose/push.go)）は project の全サービスを対象にする。
- **影響**: classic builder（`DOCKER_BUILDKIT=0`、または buildx が無い環境）のときだけ。
  - N 個のサービスをビルドすると N×N 回 push する
  - `build --push web` で、ビルドしていない依存サービス `db` の古いイメージまで push する
  - 並列ビルド中に、まだビルドが終わっていないサービスの push が走りうる
  - bake の経路はターゲットごとに push するので問題ない（`build_bake.go`）
- **直し方**: 走査の後に 1 回だけ、ビルドしたサービス（`serviceToBuild`）に絞って push する。classic build には単体テストが無いので、モックのテストを書くのが主な作業。

## 5. `config -q` が一覧系フラグと組み合わせると出力する

- **実機での再現**: `docker compose config -q --services` / `--volumes` / `--images` / `--hash='*'` / `--variables` は、
  `-q`（「検証だけして何も出力しない」）なのに一覧を出力する。`config -q` 単体は何も出さない。
- **根本原因**: `-q` を `os.Stdout = devnull` という差し替えで実装している（[cmd/compose/config.go](../cmd/compose/config.go) の PreRunE）。
  docker/cli は起動時に stdout を掴んでいるので、`dockerCli.Out()` 経由の出力には効かない。
  素の `config -q` が静かなのは、別の早期 return があるから。`build -q` も同じ実装（[cmd/compose/build.go](../cmd/compose/build.go)）。
- **衝突リスク**: 中。ndeloof の #14046（`config --filter`）が `config.go` を触っている。
- **注意**: 「`-q` と `--services` の同時指定はエラーにすべき」という判断もありうるので、どちらに倒すか issue で聞く。

## 6. classic build のイメージに compose ラベルが付かない

- **出典**: エピック #14074 F
- **場所**: プロジェクト／サービスのラベルを付ける `getImageBuildLabels`（[pkg/compose/build.go](../pkg/compose/build.go)）の呼び出し元は bake だけ。
  classic は `com.docker.compose.image.builder=classic` しか付けない（`build_classic.go`）。
- **影響**: `watch` の再ビルドや `down --rmi` で、classic が残した dangling イメージがプロジェクトのものと認識されず、掃除されない。
- **直し方**: classic の `imageBuildOptions` でも `getImageBuildLabels` を使う。純粋関数なので表形式のテストが書きやすい。4 と同じ issue で相談するとよい。

---

## 次点（先に issue で合意を取ってから）

| タスク | メモ |
| ------ | ---- |
| `swarmEnabled` がパッケージ全体でエラーまで永久にキャッシュする（`pkg/compose/compose.go`） | エピック #14074 E。主に SDK 利用者に影響。隣の `runtimeVersionCache` のようにインスタンス単位にする |
| `pull_refresh_after` が無視されているのに、設定ハッシュには含まれている | ndeloof が #14253 で「別の変更で判断すべき」と書いた。除外・実装・警告のどれにするか issue で聞く |
| [#7431](https://github.com/docker/compose/issues/7431) `push --ignore-push-failures` が認証・ネットワークエラーまで握りつぶす | thaJeztah の RFC。先に `--ignore-missing` を足す案が安全 |
| [#9122](https://github.com/docker/compose/issues/9122) `up --wait` 中にログを流す | 要望は大きい。ただし「draft PR を出して」は特定のパッチ持ちへの返答で、`up` の処理は #14081 Lot 2 と正面衝突する。今は待つ |
| watch のビルドタグ（`pkg/watch/watcher_naive.go`） | Linux/Windows で `-tags fsnotify` がコンパイルできない、コメントが誤り。開発者だけに影響。1 行修正 |
| `docker compose events` がコンテナのイベントしか流さない | [todo.md](./todo.md) 参照。公開 API `api.Event` の設計変更が要る |
| `stop` を 2 回実行すると何もしていないのに「Stopped」と出る | [todo.md](./todo.md) 参照。挙動変更なので合意が要る |

## やらないもの（理由つき）

| 対象 | 理由 |
| ---- | ---- |
| plan engine（`reconcile.go`, `start.go`, executor）全般 | #14081 でメンテナが書き換え中（open PR #14274） |
| provider / relay まわり | ndeloof が毎日のように変更している |
| [#11816](https://github.com/docker/compose/issues/11816) `!reset` が効かない | compose-go 側で ndeloof が draft（compose-go#913）を出している |
| [#13672](https://github.com/docker/compose/issues/13672), [#13809](https://github.com/docker/compose/issues/13809) | すでに PR がある（#14198, #13811） |
| [#14204](https://github.com/docker/compose/issues/14204) フック出力が `-d` で消える | moby 側の機能（moby#53407）待ち |
| [#13602](https://github.com/docker/compose/issues/13602), [#13348](https://github.com/docker/compose/issues/13348) | main で再現しない |
| 使われていない公開 API の削除（エピック #14074 B） | ユーザーへの影響が無く、公開 SDK の破壊的変更。メンテナが自分でやる可能性が高い |
| backend の作り方が 3 通りある、`ProjectOptions` の書き換えなど（エピック #14074 E/F の一部） | 調べた結果、ユーザーから見える影響が無いリファクタ。メンテナが活発に変更中のファイル |
| provider イメージ配布の `SetLimit(0)` で止まる件 | CLI からは起きない（`--parallel` は `> 0` のときしか渡されない）。SDK で明示的に 0 を渡したときだけ |

## 調査中に分かったプロセス上の注意

- **`status/approved` ラベルはリポジトリに存在しない**。CONTRIBUTING.md は「AI を使った PR は `status/approved` の issue に紐づけること」と書いているが、
  実際には付けようがない。だから「メンテナがレビューや issue で明示的に後押ししたもの」を選ぶのが現実的な代わりになる（上の 1 と 3）。
- **AI_POLICY.md:「AI の `Co-Authored-By` トレーラーは付けない」**。開示はコミットではなく PR の説明に書く。
  メンテナのコミットに Claude のトレーラーがあるのは「この規則は外部からのコントリビューションに適用する」という例外のため。
- PR を開く前に `AI_AGENT_DISCLOSURE.md` を、自分でレビューした上で削除する（[pr-trends.md](./pr-trends.md) §8）。
