# 次に着手できるタスク候補

← [document 一覧](../README.md)

[pr-trends.md](../contributing/pr-trends.md) の結論（実在する問題か／根本原因を直すか／メンテナの作業とぶつからないか／
メンテナの後押しがあるか）を物差しにして、「まだ誰も着手していない、外部から着手できて、
プロジェクトを本当に良くするタスク」を探した結果。

- 調査日: 2026-10-03（初回）、2026-10-16（再調査）
- 基準コード: `upstream/main` 7b36d08dd（2026-10-15）。初回は 42f48072b（2026-10-02）で、その間に 106 コミット。最新リリースは v5.6.0（10-02）
- 調べたもの:
  - 初回: open issue 全 51 件、open PR 全 37 件、2026-08-15 以降にマージされた PR 約 130 件の本文とレビューコメント、
    エピック [#14074](https://github.com/docker/compose/issues/14074) と [#14081](https://github.com/docker/compose/issues/14081) の未チェック項目、
    新しく追加された TODO と `t.Skip`
  - 再調査:
    - 初回の候補すべての再検証
    - open issue 全件（10-03 以降の新しい issue は 7 件で、どれもメンテナが対応中・修正済み・エピック・自分の issue のいずれか）
    - 10-02 以降にマージされた PR 49 件のレビュー、open PR 49 件
    - 10-02 以降に失敗した CI ジョブのログ約 20 件
    - `42f48072b..upstream/main` で追加された TODO と `t.Skip`
    - 新しい issue からは候補が出なかったので、**よく使われるコマンドのコードを読んで、未報告のバグを探した**
- 「検証」列の意味:
  - **実機** = 自分で upstream main をビルド・実行して再現した
  - **テスト** = 単体テストやツールの出力で再現した
  - **CI** = CI のログで確認した
  - **コード** = 現行コードを読んで確認した（実行はしていない）
- 下のコードリンクのうち GitHub を指すものは、手元の作業ブランチにまだ無いファイル（provider / relay / inspect の `internal/` への移動などがあったため）

## 結論（先にやる順）

| 順 | タスク | 種類 | 検証 | 後押し | 衝突リスク | 規模 | 最初の一手 | 進捗 |
| -: | ------ | ---- | ---- | ------ | ---------- | ---- | ---------- | ---- |
| 1 | [CI の unit テストで毎回 1 件の偽の FAIL が出る](#1-ci-の-unit-テストで毎回-1-件の偽の-fail-が出る) | CI の衛生 | **テスト** | △ thaJeztah が本物の失敗と取り違えかけた | 低 | XS | そのまま PR | 🟡 [#14351](https://github.com/docker/compose/pull/14351) レビュー待ち |
| 2 | [`run` が依存サービスを作り直すと匿名ボリュームを引き継がない](#2-run-が依存サービスを作り直すと匿名ボリュームを引き継がない) | **データ消失**（未報告） | コード | △ 関連 issue #13769 | 低〜中 | XS（+e2e） | 実機で再現 → issue | 未着手 |
| 3 | [watch の監視中に `.git/` が除外されない](#3-watch-の監視中に-git-が除外されない) | バグ（未報告） | **テスト** | △ glours が #14217 で同種の欠陥を指摘 | 中（`watch.go`） | XS | issue を立てる | 未着手 |
| 4 | [bake のビルドが curl の書式指定などで失敗する](#4-bake-のビルドが-curl-の書式指定などで失敗する) | バグ（未報告） | **テスト** | ○ ndeloof が `${` を同じ形で直した前例 | 低 | S | issue を立てる | 未着手 |
| 5 | [`down` の pre_start フックコンテナ後始末](#5-down-の-pre_start-フックコンテナ後始末) | バグ | コード | ◎ メンテナが「別 PR で出してよい」 | 低〜中 | S〜M | (b) のマージ後に (a)(d) を PR | 🟡 (b) [#14310](https://github.com/docker/compose/pull/14310) APPROVED、CI の実行承認待ち |
| 6 | [e2e `TestWatchSyncIntoSymlinkedDirectory` が不安定](#6-e2e-testwatchsyncintosymlinkeddirectory-が不安定) | 不安定なテスト | CI + コード | — | 低 | S | そのまま PR（失敗した run を引用） | 未着手 |
| 7 | [stats のテストの決定化と取得失敗の debug ログ](#7-stats-のテストの決定化と取得失敗の-debug-ログ) | テスト／診断 | コード | ◎ ndeloof が承認時に「non-blocking suggestions」 | 低 | S | そのまま PR（レビューを引用） | 未着手 |
| 8 | [`ensureModels` のエラー無視とデータ競合](#8-ensuremodels-のエラー無視とデータ競合) | バグ | コード | ◎ メンテナがレビューで「フォローアップに値する」 | 低 | S | apt に新しい docker-model が出たら rebase | 🟡 [#14284](https://github.com/docker/compose/pull/14284) 上流の修正の配布待ち |
| 9 | [`events --until` が必ず EOF エラーで終わる](#9-events---until-が必ず-eof-エラーで終わる) | バグ（未報告） | テスト | — | 低 | XS | issue を立てる | 未着手 |
| 10 | [classic build の `--push` がプロジェクト全体をサービス数だけ push する](#10-classic-build-の---push-がプロジェクト全体をサービス数だけ-push-する) | バグ | コード | ○ エピック #14074 に 🐛 として記載 | 低 | S | issue を立てる（エピックから切り出す） | 未着手 |
| 11 | [provider のメタデータ取得が呼び出し元の ctx を使わない](#11-provider-のメタデータ取得が呼び出し元の-ctx-を使わない) | バグ（小） | コード | ◎ ndeloof が「follow-up に回す」と明言 | 中（provider） | S | そのまま PR（レビューを引用） | 未着手 |
| 12 | [`config -q` が一覧系フラグと組み合わせると出力する](#12-config--q-が一覧系フラグと組み合わせると出力する) | バグ | **実機** | ○ エピック #14074 E に記載 | 中（#14332, #14046） | S〜M | issue を立てる | 未着手 |
| 13 | [`pkg/compose` のテストが Windows でコンパイルできない](#13-pkgcompose-のテストが-windows-でコンパイルできない) | テストのビルド | テスト | ◎ ndeloof がレビューで指摘 | 中（#14305） | XS | upstream main で再確認して PR | 未着手（#14296 は 10-16 にマージ済み） |

**「簡単で、インパクトが大きい」ものは 1〜4**:

- **1** はすべての PR の CI に効き、2 行で直る。メンテナがこのノイズに実際に惑わされた記録があるので、受け入れられやすい。**すぐ出せる**。
- **2** はインパクトが最大。修正は 1 行で、`up` と `watch` がすでにやっていることを `run` に揃えるだけ。
  ただしユーザーのデータに関わるので、**実機で「データが消える」ことを再現してから** issue を立てる。e2e テストも付ける。
- **3** は watch のパスに `.git` を含む（`path: .` など）ほぼすべての利用者に効く。修正はパターン 2 か所だけ。
  ただし `watch.go` を ndeloof の Lot 2 の PR が触っているので、issue で聞いてから。
- **4** は再現が確実（`bake --print`）で、メンテナ自身の前例と同じ形の修正。衝突も無い。

その次に、テストだけの小さな変更（6、7）と、待っているもの（5、8）を進める。9 以降は余裕があるとき。

初回から順位を変えたもの:

- 初回 2 位「`start -p` で任意の依存サービスが必須扱いになる」→ PR がメンテナの方針と合わずクローズしたので、[保留](#保留メンテナの方針と合わなかったもの) へ。
- 初回 6 位「classic build のイメージに compose ラベルが付かない」→ 方針リスクが上がったので、[次点](#次点先に-issue-で合意を取ってから) へ。

---

## 1. CI の unit テストで毎回 1 件の偽の FAIL が出る

- **症状**: `TestRunHandlesKeysUntilContextIsDone` は成功しているのに、gotestsum が
  `=== FAIL: cmd/formatter TestRunHandlesKeysUntilContextIsDone (unknown)` と `DONE … 1 failure` を出す。
  - 終了コードは 0 だが、Unit Test Summary は `failed = 1` になる。
  - main の green な run でも毎回出る（例: [run 38757079565](https://github.com/docker/compose/actions/runs/38757079565)）。
  - `make test` も同じ gotestsum の呼び出しなので、手元でも出る。
- **実害**: thaJeztah が [#14339](https://github.com/docker/compose/pull/14339) で、このノイズ行を見て
  「This failure may be genuine… `=== FAIL: cmd/formatter TestRunHandlesKeysUntilContextIsDone (unknown)`」と書いた（本当の原因は別のテスト）。
  CI の失敗を読むたびに人の時間を取る。
- **原因**:
  - テストが `'d'` キーを送る → `clearNavigationMenu()`（[cmd/formatter/shortcut.go](https://github.com/docker/compose/blob/7b36d08dd0d30da3abc1e646baec8751f0bcb0b2/cmd/formatter/shortcut.go#L348)）が、改行の無い ANSI シーケンスを os.Stdout に書く。
  - その結果、`--- PASS` 行の頭にシーケンスが付き、gotestsum が `-v` の出力を解析できなくなる。
  - 入ったのは #14324（149cf98f5、10-10）。根っこはエピック #14074 の「cmd/formatter が os.Stdout に直書き」と同じ。
- **場所**: [cmd/formatter/shortcut_test.go](https://github.com/docker/compose/blob/7b36d08dd0d30da3abc1e646baec8751f0bcb0b2/cmd/formatter/shortcut_test.go#L32)
- **直し方**: テストの冒頭で `disableAnsi = true` にし、`t.Cleanup` で元に戻す。このフラグは [cmd/formatter/ansi.go](../../cmd/formatter/ansi.go) にすでにある。
- **検証**: upstream/main のコピーで、CI と同じ `gotestsum -- -v -race -cover … -args -test.gocoverdir=` を流した。
  修正前は `1 failure`、修正後は `DONE 12 tests` で失敗 0。別の調査でも、`go test -json -v` で毎回再現した。
- **衝突リスク**: 低。thaJeztah の [#14180](https://github.com/docker/compose/pull/14180) は `shortcut.go` を触るが、テストファイルは触らない。
- **着手状況**: テスト名で検索しても、issue / PR は無かった（10-16）。
- **進捗**（2026-10-17）: 自分の PR [#14351](https://github.com/docker/compose/pull/14351) `test(formatter): keep escape sequences out of the go test output`（OPEN、REVIEW_REQUIRED）
  - 変更は `shortcut_test.go` の 5 行だけ（`disableAnsi = true` と `t.Cleanup`、理由のコメント）。
  - 確認したこと: gotestsum v1.13.0 で、`1 failure` → 失敗 0。`go test ./cmd/...` と `make lint` が通る。
  - 該当する issue が無いので、PR 本文の Related issue には「テストだけの小さな修正である理由」を書いた。
    CONTRIBUTING.md は AI を使った PR に承認済みの issue を求めているので、メンテナに issue を求められたら立てる。
  - **次の一手**: レビューを待つ。fork からの PR なので、CI を動かすのにメンテナの承認が要る。承認後はブランチを触らない（#14310 の教訓）。

## 2. `run` が依存サービスを作り直すと匿名ボリュームを引き継がない

- **症状（推定）**:
  1. db の設定を変える。
  2. `docker compose run app ...` を実行する。
  3. 依存先の db が作り直され、**前のコンテナの匿名ボリュームを引き継がず、新しい匿名ボリュームが付く**。
  - postgres や mysql のように、イメージが `VOLUME` を宣言していて名前付きボリュームを使っていない構成では、データが消えたように見える。
  - `up` なら引き継ぐので、同じ変更でも `run` か `up` かで結果が違う。
- **根本原因**: [pkg/compose/run.go](https://github.com/docker/compose/blob/7b36d08dd0d30da3abc1e646baec8751f0bcb0b2/pkg/compose/run.go#L298-L307) の `startDependencies` が `s.create` に渡す `api.CreateOptions` に `Inherit` が無く、false のままになる。
  - `up` は `Inherit: !noInherit`（`cmd/compose/up.go:285`）、watch の再ビルドは `Inherit: true`（`pkg/compose/watch.go:685`）を渡している。
  - 引き継ぐかどうかは `reconcile.go:1544` の `if r.options.Inherit` で決まる。
- **関連**: [#13769](https://github.com/docker/compose/issues/13769)「container re-created when it shouldn't, if using `docker-compose run`」
  （再作成の後で、db に期待したユーザーがいない、という報告。元は #9897）。
  ndeloof は [#14144](https://github.com/docker/compose/pull/14144) で `run --no-recreate` を足し、10-14 にこの issue をクローズした。
  **再作成を避ける手段はできたが、再作成されたときにデータが消える原因は残っている**。
- **直し方**: `Inherit: true` を 1 行足す（`run` には `-V` 相当のフラグは無いので、`watch` と同じく常に true でよいはず）。
- **テスト**: e2e で「`VOLUME` を宣言したイメージの依存サービスにファイルを書く → 設定を変えて `run` → ファイルが残っている」を確かめる。Scenario DSL で書ける。
- **衝突リスク**: 低〜中。`run.go` は glours の jobs [#14155](https://github.com/docker/compose/pull/14155)（draft）、ndeloof の #14299 と #14092（PoC）も触っているが、どれも別の箇所。
  `run` は #14081 の対象外。
- **最初の一手**: **まだ実機で再現していない**。まず手元で再現し、「`up` では残るのに `run` では消える」を示して issue を立てる（#13769 を引用）。
  データ消失なので、メンテナの反応は早いはず。修正が 1 行なので、issue と同時に PR を出してもよい。

## 3. watch の監視中に `.git/` が除外されない

- **症状**: watch のパスに `.git` が含まれると（`path: .` など）、git の操作のたびに `.git/*` がコンテナに sync されるか、rebuild が走る。
  最初の sync（initial sync）では除外されるので、挙動が食い違っている。
- **根本原因**: [pkg/compose/watch.go](https://github.com/docker/compose/blob/7b36d08dd0d30da3abc1e646baec8751f0bcb0b2/pkg/compose/watch.go#L317)（317 行と 742 行）の
  `watch.NewDockerPatternMatcher("/", []string{".git/"})` は、`/.git`（ファイルシステムのルート直下）にしか一致しない。
  監視中の照合はフルパス（`/proj/.git/index` など）で行うので、一致しない。2023 年からこのまま。
- **直し方**: パターンを `**/.git` にする（2 か所）。表形式の単体テストを足す。
- **検証**: 単体テストで、`Matches("/tmp/x/.git/index")` が false を返すことを確かめた。
- **後押し（間接）**: glours が外部 PR [#14217](https://github.com/docker/compose/pull/14217) のレビューで、「監視中の照合はフルパスなので、パス付きの matcher が要る」と同じ種類の欠陥を指摘している。
- **衝突リスク**: 中。
  - #14217（外部、CHANGES_REQUESTED のまま作者が止まっている）が同じ関数を触っている。
  - ndeloof の Lot 2 の PR（#14300、#14302、#14304、#14305）も `watch.go` を触っている。ただし、どれも別の箇所。
- **最初の一手**: issue を立てる。#14217 にも関わるので、「#14217 とは別に直すか、まとめるか」を聞く。

## 4. bake のビルドが curl の書式指定などで失敗する

- **症状**: 例えば `dockerfile_inline` に `RUN curl -w '%{http_code}' ...` と書くと、bake 経由のビルドが `Invalid template control keyword` で失敗する。classic builder なら通る。
  ラベルの値に `${` を書いた場合も同じ。
- **根本原因**: bake は、渡された JSON の文字列を HCL のテンプレートとして評価する。ところが [pkg/compose/build_bake.go](https://github.com/docker/compose/blob/7b36d08dd0d30da3abc1e646baec8751f0bcb0b2/pkg/compose/build_bake.go#L252) のエスケープが不十分。
  - args（252 行）と dockerfile-inline（290 行）は `${` だけをエスケープしていて、`%{` をしていない。
  - Labels（292 行）は、エスケープそのものをしていない。
- **直し方**: `${` → `$${` と `%{` → `%%{` をまとめて行う関数を作り、3 か所に使う。約 10 行。
- **検証**: 手元の buildx v0.24 の `docker buildx bake --print` で再現した。
- **後押し**: ndeloof が過去に `${` を同じ形で直している（0c5bd16da、a03f2562d）。今回はその直し漏れ。
- **衝突リスク**: 低（`build_bake.go` を触る open PR は無い）。

## 5. `down` の pre_start フックコンテナ後始末

- **出典**: ndeloof の [#14091](https://github.com/docker/compose/pull/14091) レビュー
  「The agreed follow-ups (down scoping vs dependents, list-failure aborting teardown, …) remain open — fine by me to land them separately.」
- **場所**: [pkg/compose/down.go](../../pkg/compose/down.go) `removePreStartHookContainers`
  - (a) `down web` のようにサービスを指定すると、フックの後始末は指定したサービスの分だけ（`down.go:126` は `options.Services` だけを渡し、`:563-577` でサービスごとに絞る）。
    実際のティアダウンは依存元まで巻き込む（`WithRootNodesAndDown` → `getAncestors`）ので、依存元のフックコンテナが残る。
  - (b) 一覧取得に失敗するとティアダウンがまるごと止まる → **#14310 で対応済み（マージ待ち）**。
  - (c) 任意の改善: `pre_start.go` の `context.Background()` を `context.WithoutCancel(ctx)` にする（トレースのスパンを保つため。ndeloof の提案）。
  - (d) フックコンテナは `VolumesFrom` でサービスのボリュームを参照しているのに、サービスコンテナより後に消している
    （フックの削除は `:126`、サービスの `removeContainers`（`rm -v`）は `:514`）。そのため、匿名ボリュームを回収し損ねるかもしれない。
    フックを先に消すのが安全（レビューでも「to verify」扱い）。
- **衝突リスク**: `down` は #14081 の対象外。ただし glours の jobs の draft [#14155](https://github.com/docker/compose/pull/14155) が `down.go` を広く触っている。
- **進捗**（2026-10-16）:
  - (b) は自分の PR [#14310](https://github.com/docker/compose/pull/14310) `fix(down): keep tearing down when listing retained hook containers fails`。
    - ndeloof が 10-13 に APPROVED（auto-rebase も有効にしてくれた）。
    - 自分が main を merge したため、10-15 に auto-merge が外れた。rebase し直して、承認時と同じ 1 コミットに戻した。
    - fork からの PR なので、CI を動かすにはメンテナの承認が要る。PR のコメントで依頼済み。
  - (a)(d) に関する issue / PR は無い（10-16 時点）。
  - **次の一手**: #14310 のマージを待つ。数日動きが無ければ、PR で一度だけ軽く催促する。
    マージ後に (d) → (a) の順で、別々の PR にする。(d) は順序を入れ替えるだけで小さい。(a) は、対象のサービスの集まりをティアダウンと同じにする必要がある。

## 6. e2e `TestWatchSyncIntoSymlinkedDirectory` が不安定

- **症状**: 1 週間で 2 回失敗している（うち 1 回は main 上）。
  - main（10-12）: [run 38226095992](https://github.com/docker/compose/actions/runs/38226095992)
  - PR（10-08）: [run 37741830070](https://github.com/docker/compose/actions/runs/37741830070)
  - どちらも `cat: can't open '/var/sub/hello.txt'` で 60 秒タイムアウトし、`cannot exec in a stopped state` も出ている。
- **原因（コードと CI のログからの推定。手元では実行していない）**:
  - `up --watch` では、`watcher.Start` が `s.start` より先に走る
    （[pkg/compose/up.go](https://github.com/docker/compose/blob/7b36d08dd0d30da3abc1e646baec8751f0bcb0b2/pkg/compose/up.go#L241) の 241 行と 310 行）。
    だから「Watch enabled」はコンテナの起動前に出る。10-08 のログでも、「Syncing…」が「Started」より前に出ている。
  - テストは「Watch enabled」を待ってすぐ `os.Rename` する（[pkg/e2e/watch_test.go](../../pkg/e2e/watch_test.go) の upstream 491-497 行付近）。
    そのため、コンテナの command が `ln -s /var/sub /app/data/sub` を作る前に同期が走ることがある。
  - 同じファイルの 415 行のコメント「"Watch enabled" is logged AFTER the initial up」も実態と合っていない。
- **直し方**: rename の前に、`exec app test -L /app/data/sub` が成功するまで poll する（状態で待つ。[SCENARIO.md](../../pkg/e2e/SCENARIO.md) の方針どおり）。誤っているコメントも直す。
- **衝突リスク**: 低。`watch_test.go` を触る open PR は無い。interactive up の順序を変える #14305 が入っても、状態で待つ形なら影響を受けない。
- **注意**: 不安定なテストを直す PR では、「なぜ失敗するか」を CI のログで示すことが大事。上の 2 つの run を引用する。

## 7. stats のテストの決定化と取得失敗の debug ログ

- **出典**: ndeloof が [#14337](https://github.com/docker/compose/pull/14337) を承認したときのレビュー（10-15）
  「Two non-blocking suggestions: `sampleStats` errors are dropped silently… A `logrus.Debugf` would help diagnose it. /
  `TestCollectContainerStatsBoundsConcurrency` asserts `peak == statsConcurrency`, which depends on scheduling…
  Having each call wait until `statsConcurrency` are in flight (with a timeout) would make it deterministic.」
  - 承認の 30 秒後にマージされ、どちらも手付かずのまま。docker-agent のレビューも同じ点を指摘していた。
- **場所**:
  - [cmd/compose/stats_projects_test.go](https://github.com/docker/compose/blob/7b36d08dd0d30da3abc1e646baec8751f0bcb0b2/cmd/compose/stats_projects_test.go#L184-L208)（20ms の sleep に頼っている）
  - [cmd/compose/stats_projects.go](https://github.com/docker/compose/blob/7b36d08dd0d30da3abc1e646baec8751f0bcb0b2/cmd/compose/stats_projects.go#L197)（`err == nil` のときだけ値を使い、エラーは捨てている）
- **直し方**: テストに、呼び出しが `statsConcurrency` 本そろうまで待つバリア（タイムアウト付き）を入れる。失敗したサンプルは `logrus.Debugf` で出す。
- **ついでに見つかったもの**: 一時停止中のコンテナの使用量が 0 になる（`stats_projects.go:191` が running のときしか取得しない。1 行）。
  未リリースの機能なので、同じ PR で直すか別にするかを相談するとよい。
- **衝突リスク**: 低。ただし作者（glours）が自分でやる可能性はある。出す前に、`stats_projects` を触る PR が無いかを確認する。

## 8. `ensureModels` のエラー無視とデータ競合

- **出典**: ndeloof の [#14177](https://github.com/docker/compose/pull/14177) インラインレビュー（`model.go:49`）
  「Two pre-existing issues here… worth a follow-up」
- **場所**: [pkg/compose/model.go](../../pkg/compose/model.go) `ensureModels`
  - `ListModels` の `err` を見ていない（失敗すると空のリストとして扱い、全モデルを pull しに行く）
  - 各 goroutine が外側の `err` に代入している → 複数のモデルを pull すると**データ競合**
- **進捗**（2026-10-16）: 自分の PR [#14284](https://github.com/docker/compose/pull/14284) `fix(models): check the model listing error and keep pull errors per goroutine`（OPEN）
  - ndeloof が APPROVED（10-05）。その後、e2e の `TestComposeModel` が失敗した（`error unmarshalling available models`）。
  - glours が CHANGES_REQUESTED（10-06）。原因は `docker/model-runner` 側にある（runner の初回インストール時に、進捗表示が stdout に混ざる）。
    上流で直すまでマージしない方針で、修正 [docker/model-runner#1079](https://github.com/docker/model-runner/pull/1079) は 10-06 にマージ済み。
  - **待つ条件の訂正**: CI は docker-model を GitHub のリリースからではなく、
    `test.docker.com` のスクリプトで **apt の test チャネル**から入れている（`ci.yml:232-238`。バージョンは固定していない）。
    - 10-15 の main の CI で使われていたのは **v1.2.6**。GitHub の最新 v1.2.8（08-12）すら、まだ apt に来ていない。
    - #1079 を含むリリースは、GitHub にもまだ無い。
    - したがって、待つ条件は「#1079 を含む `docker-model-plugin` が apt の test チャネルに出ること」。
      確認先: `https://download.docker.com/linux/ubuntu/dists/noble/test/binary-amd64/Packages` の `docker-model-plugin` の Version。
  - その間に glours の 4f9bec66b（10-13）が、`model.go` に dry-run 対応を入れ、`model_test.go` を新しく作った。
    #14284 の diff は upstream/main にそのまま当たる（`git apply --check` で確認）。GitHub 上も MERGEABLE。
  - **次の一手**: 上の条件を満たしたら rebase し、`TestComposeModel` が通るのを確かめて、PR で glours に連絡する（そう約束済み）。それまでは触らない。

## 9. `events --until` が必ず EOF エラーで終わる

- **症状**: `docker compose events --until <時刻>` は、指定した時刻に達してストリームが正常に終わったのに、`EOF` エラーで終了する。
- **根本原因**: [internal/inspect/events.go](https://github.com/docker/compose/blob/7b36d08dd0d30da3abc1e646baec8751f0bcb0b2/internal/inspect/events.go#L77) の `case err := <-res.Err: return err` が、
  正常終了を表す `io.EOF` もエラーとして返している。docker/cli の `events` は `io.EOF` を nil として扱う。2025-07 に `--until` が追加されてから、ずっとこの状態。
- **直し方**: `errors.Is(err, io.EOF)` なら nil を返す。3 行。
- **検証**: moby の client が終端で EOF を送ることを確かめ、単体テストで再現した。
- **影響**: `--until` を使うと毎回失敗する。ただし、このフラグの利用者は少なめ。
- **衝突リスク**: 低（触っている open PR は無い）。

## 10. classic build の `--push` がプロジェクト全体をサービス数だけ push する

- **出典**: エピック #14074 F「classic `--push` pushes the whole project once per built service」（🐛 印付き、「standalone issue に切り出す価値あり」とされている）
- **場所**: [pkg/compose/build_classic.go](../../pkg/compose/build_classic.go) のサービスごとのコールバック内（upstream 103-104 行）
  `if options.Push { return s.push(ctx, project, api.PushOptions{}) }`。
  `push()`（[pkg/compose/push.go](../../pkg/compose/push.go)）は project の全サービスを対象にする。
- **影響**: classic builder（`DOCKER_BUILDKIT=0`、または buildx が無い環境）のときだけ。
  - N 個のサービスをビルドすると N×N 回 push する
  - `build --push web` で、ビルドしていない依存サービス `db` の古いイメージまで push する
  - 並列ビルド中に、まだビルドが終わっていないサービスの push が走りうる
  - bake の経路はターゲットごとに push するので問題ない（`build_bake.go`）
- **直し方**: 走査の後に 1 回だけ、ビルドしたサービス（`serviceToBuild`）に絞って push する。classic build には単体テストが無いので、モックのテストを書くのが主な作業。
- **状況（10-16）**: `build_classic.go` は 10-02 以降変わっていない。`push.go` は `--ignore-missing`（#14341）とレジストリ認証で変わったが、全サービスを回す点は同じ。
  issue / PR は無い。Benehiko の draft [#14336](https://github.com/docker/compose/pull/14336) が `push.go` の別の箇所を触っている。

## 11. provider のメタデータ取得が呼び出し元の ctx を使わない

- **出典**: ndeloof の [#14333 のインラインレビュー](https://github.com/docker/compose/pull/14333#discussion_r4257871410)
  「Propagating the caller's context is a fair improvement and a behavior change, so it is left to a follow-up.」
- **場所**: [internal/provider/plugins.go](https://github.com/docker/compose/blob/7b36d08dd0d30da3abc1e646baec8751f0bcb0b2/internal/provider/plugins.go#L615-L617) `getPluginMetadata`
  - `exec.Command` と `context.Background()` を使っているので、metadata コマンドが遅い・固まる provider だと Ctrl-C が効かない。
- **直し方**: 引数に `ctx` を足し、`exec.CommandContext(ctx, …)` と `PrepareShellOut(ctx, …)` に渡す。
  テストは `plugins_control_test.go` のヘルパープロセス方式で書ける。
- **衝突リスク**: 中。provider は 10-15 に `internal/` へ移ったばかりで、ndeloof が頻繁に触る領域。10-16 時点で `plugins.go` を触る open PR は無い。
  出す直前に、もう一度 open PR を確認する。
- **同じ PR の、より小さい残作業**: `removeRelayLinkNetwork` が最初のエラーで止まる。
  - [r4256695116](https://github.com/docker/compose/pull/14333#discussion_r4256695116) で「`errors.Join` で集めるのは follow-up で」とされている。場所は `internal/relay/relay.go:374-378`。
  - XS だがインパクトは小さいので、11 と同じ PR にはしない。出すなら別にする。

## 12. `config -q` が一覧系フラグと組み合わせると出力する

- **実機での再現**: `docker compose config -q --services` / `--volumes` / `--images` / `--hash='*'` / `--variables` は、
  `-q`（「検証だけして何も出力しない」）なのに一覧を出力する。`config -q` 単体は何も出さない。
- **根本原因**: `-q` を `os.Stdout = devnull` という差し替えで実装している（[cmd/compose/config.go](../../cmd/compose/config.go) の PreRunE、upstream 103-109 行）。
  docker/cli は起動時に stdout を掴んでいるので、`dockerCli.Out()` 経由の出力には効かない。
  素の `config -q` が静かなのは、別の早期 return があるから。`build -q` も同じ実装（[cmd/compose/build.go](../../cmd/compose/build.go)）。
- **衝突リスク**: 中。
  - ndeloof の [#14046](https://github.com/docker/compose/pull/14046)（`config --filter`）は 08-18 から止まっている。
  - 新しく [#14332](https://github.com/docker/compose/pull/14332)（`config --images` の修正、CHANGES_REQUESTED）が `config.go` を触っている。
- **注意**: 「`-q` と `--services` の同時指定はエラーにすべき」という判断もありうるので、どちらに倒すか issue で聞く。

## 13. `pkg/compose` のテストが Windows でコンパイルできない

- **出典**: ndeloof の [#14309 のインラインレビュー](https://github.com/docker/compose/pull/14309#discussion_r4220751191)
  「The package's tests still don't compile on Windows on main because of newStartTestService/serviceContainer from start_test.go, a pre-existing and unrelated issue.」
  - Copilot と docker-agent も、#14303 と #14309 で同じ種類の壊れ方を毎回指摘している。#14341 の `push_test.go` で、壊れる箇所がさらに増えた。
- **原因**: [pkg/compose/start_test.go](../../pkg/compose/start_test.go) の 1 行目が `//go:build !windows`。
  なのに、そこで定義したヘルパー（`newStartTestService` など）を、`down_test.go`、`push_test.go`、`service_containers_test.go` などが使っている。
- **直し方**: build tag を消す（`start_test.go` は Unix 専用の API を使っていない）か、ヘルパーを tag の無いファイルに移す。
  - `GOOS=windows go test` でビルドが通り、darwin でもテストが通ることは、調査のときに確認した。
  - 前例として、外部からの同じ種類の修正 [#13683](https://github.com/docker/compose/pull/13683) がマージされている。
- **衝突リスク**: **今は高い**。ndeloof の [#14296](https://github.com/docker/compose/pull/14296) と #14305 が、`start_test.go` を大きく書き換え中（+8/-286）。
  **#14296 のマージ後に**、そのときの main で状況を確かめてから出す（書き換えで直っている可能性もある）。
- **更新（2026-10-17）**: #14296 は 10-16 にマージされた。upstream main（289d8a037）でも `start_test.go` の 1 行目は `//go:build !windows` のままで、直っていない。
  残る衝突は draft の #14305（と統合ブランチ #14308）だけなので、衝突リスクは「中」に下げた。出す前に `GOOS=windows go vet ./pkg/compose/` で今も壊れていることを確かめる。

---

## 保留（メンテナの方針と合わなかったもの）

### `start -p` で任意の依存サービスが必須扱いになる

- **症状**: `required: false` の依存サービスが失敗していると、ファイルを読む `docker compose start` は警告を出して起動するのに、
  `docker compose -p <name> start` は失敗して起動しない（実機で再現済み）。
- **原因**: `-p` だけのときは、コンテナのラベルからプロジェクトを組み立て直す（`projectFromName`）。
  ところが、ラベル `com.docker.compose.depends_on` には `required` が保存されず、読み戻すときに `required := true` と決め打ちしている。
  `down.go` は補正しているが、`start` は補正していない。
- **経緯**:
  - 自分で issue [#14311](https://github.com/docker/compose/issues/14311) を立て、ラベルに `required` を足す PR [#14312](https://github.com/docker/compose/pull/14312) を出した。
  - ndeloof の返答（10-09）は「not in favor we have more config stored by labels」。
    ファイル無しでプロジェクトを扱いたいなら、モデルそのものを engine に保存する別の方法が要る、という立場。
  - 10-11 に自分で PR をクローズした。issue は「ラベル方式の既知の制約」として open のまま（メンテナの反応は無い）。
  - ラベル → annotation の移行 [#14050](https://github.com/docker/compose/issues/14050) は、10-14 に ndeloof が not_planned でクローズした。
  - #14081 Lot 2 の「start の start-only plan」は [#14296](https://github.com/docker/compose/pull/14296) で進んでいる。ただし `projectFromName` もラベルも触らないので、マージ後もこのバグは残る。
- **今後**: ラベルを増やさない別案として、`start` でも `down.go` と同じく `Required = false` に補正する方法はありうる。
  ただし、必須の依存の失敗まで警告に変わってしまうので、メンテナの合意なしには出せない。
  #14296 がマージされたら、#14311 で「この案はどうか」と一度だけ聞く程度にとどめる。

---

## 次点（先に issue で合意を取ってから）

| タスク | メモ |
| ------ | ---- |
| post_start フック付きのサービスで、`run` の起動が失敗すると永久に止まる（`pkg/compose/run.go:79` の `<-hookErrCh`） | フックの goroutine は start イベントを待つが、起動に失敗すると来ない。`signal.Notify` が全シグナルを奪うので、Ctrl-C も効かない。直し方は「フック用の ctx をキャンセル可能にし、`RunStart` が失敗したら cancel する」で、約 10 行。**コードを読んだだけで、実機での確認が要る**。`run.go` は #14155 も触っている |
| git のリモート include の取得に失敗すると、壊れたキャッシュが残り続ける（`pkg/remote/git.go` の `checkout()`） | 失敗してもディレクトリを消さず、`Load` はディレクトリがあるかしか見ないので、手で消すまで毎回 `not found` になる。OCI 版（`oci.go`）は失敗時に `RemoveAll` している。約 8 行。#14268 と ndeloof の #14085 が同じファイルを触っているので、それらが片付いてから |
| classic build のイメージに compose ラベルが付かない（`getImageBuildLabels` の呼び出し元は `build_bake.go` だけ） | エピック #14074 F。初回の 6 位から下げた。ndeloof は #14050 でこれを「separate pre-existing gap」と認めている。一方で、イメージに project/service ラベルが付くと、そこから `docker run` したコンテナまで compose 管理下に見える問題も挙げている。#14312 の「ラベルを増やしたくない」もあるので、方針を聞いてから |
| `pull_refresh_after` が無視されているのに、設定ハッシュには含まれている（`hash.go`） | ndeloof が #14253 で「別の変更で判断すべき」と書いた。除外・実装・警告のどれにするか issue で聞く。10-16 時点で変化なし |
| `wait --down-project` がヘルプの説明どおりに動かない（`pkg/compose/wait.go`） | ヘルプは「最初のコンテナが止まったら」だが、実際は全部が止まるまで待つ（`eg.Wait()`）。`statusCode` を複数の goroutine が書いていて、データ競合もある。挙動が変わるので合意が要る |
| ライフサイクルの調査（10-17）で気付いた未確認の 3 件 | (1) plan の `StopContainer` は pre_stop を実行しないので、作り直しや scale down では pre_stop が動かない（`executor_ops.go:309-314`）。(2) 前景の `up` の Ctrl-C（`stopApplication`）は timeout を渡さないので、`up -t` が効かないかもしれない。(3) `--abort-on-container-exit` が、restart ポリシーで再起動中のコンテナが死んだときにも止めてしまうかもしれない。**どれも実機で未確認**。(1) は plan engine の領域なので、#14081 の issue で意図を聞くのが先。詳細は [compose-lifecycle.md](../concepts/compose-lifecycle.md) |
| `.env` に書いた `COMPOSE_PROGRESS` が効かない（`cmd/compose/compose.go:234` がフラグの定義時に `os.Getenv` を読む） | ndeloof はこの制約を把握していて、同じ種類の問題を #14139 で直している最中。#14139 のマージ後に、その仕組みに乗せる |
| `push --ignore-missing` で全部スキップされても何も分からない（`push.go:69-71`） | ndeloof が [#14341](https://github.com/docker/compose/pull/14341) の承認時に任意の提案として書いた（「a summary… or a warning when nothing was pushed」）。警告か集計かは設計判断。#14336 が `push.go` を触っている |
| [#14313](https://github.com/docker/compose/issues/14313) bake の並列数の上限（buildx v0.38 の `--jobs` に `--parallel` を渡す） | S〜M。buildx のバージョン判定が要る。`service:` 形式の additional_contexts（buildx でいう linked target）があると、小さい上限値は buildx に拒否される。メンテナの反応はまだ無い |
| [#9122](https://github.com/docker/compose/issues/9122) `up --wait` 中にログを流す | 要望は大きい。ただし、`up` の処理は #14081 Lot 2（#14305 ほか）と正面衝突する。今は待つ |
| 表示まわりの小さなもの | `ps --format "table {{.Health}}"` などでヘッダーが `<no value>` になる（`cmd/formatter/container.go:119-134` に、定義済みなのに使われていない定数がある）。`images --format json` が、結果が空のとき `null` を返し、順序も実行ごとに変わる（`cmd/compose/images.go:108`） |
| watch のビルドタグ（`pkg/watch/watcher_naive.go`） | Linux/Windows で `-tags fsnotify` がコンパイルできず、コメントも誤り。開発者だけに影響。1 行修正。10-16 時点で変化なし |
| `docker compose events` がコンテナのイベントしか流さない | [todo.md](./todo.md) 参照。場所は `internal/inspect/events.go:41` に移った。同じ内容の [#13700](https://github.com/docker/compose/issues/13700) は stale で自動クローズされた。公開 API `api.Event` の設計変更が要る |
| `stop` を 2 回実行すると、何もしていないのに「Stopped」と出る | [todo.md](./todo.md) 参照。`down.go` の `stopContainer` が状態を見ずに `StatusStopped` を出す。挙動変更なので合意が要る |

初回から外したもの:

- `swarmEnabled` のパッケージ全体のキャッシュ: ndeloof が [#14323](https://github.com/docker/compose/pull/14323)（57fd13053）でインスタンス単位にした
- [#7431](https://github.com/docker/compose/issues/7431): glours が [#14341](https://github.com/docker/compose/pull/14341) で `--ignore-missing` を追加してクローズした

## 見守るもの（放置されているが、本物のバグを直している外部 PR）

他人の PR を乗っ取るのは避ける。作者の反応が長く途絶え、メンテナが求めたときだけ「引き継いでよいか」と聞く。

| PR | 内容 | 状態（10-16） |
| -- | ---- | ------------- |
| [#14141](https://github.com/docker/compose/pull/14141) | `EventProcessor.Done` の success が逆（`pkg/compose/progress.go` の `bus.Done(operation, err != nil)`） | 08-27 から誰もレビューしていない。main でも未修正 |
| [#14267](https://github.com/docker/compose/pull/14267) | `volumes_from: [svc:ro]` の `:ro` が落ちる（`resolveVolumeFrom`） | レビュー無し |
| [#14268](https://github.com/docker/compose/pull/14268) | git remote の ref が `feature/main` に誤って解決される | レビュー無し |
| [#14217](https://github.com/docker/compose/pull/14217) | watch のループで Dockerfile を除外していない | 09-23 に glours が CHANGES_REQUESTED、作者は 09-26 から反応が無い。上の 3 と関係する |

## やらないもの（理由つき）

| 対象 | 理由 |
| ---- | ---- |
| plan engine（`reconcile*.go`, `start.go`, `up.go`, `scale.go`, `watch.go`, executor）と、そのテスト（`start_test.go`, `restart_test.go`, `executor_start_test.go` など） | #14081 でメンテナが書き換え中。#14274 は 10-06 にマージされ、Lot 2 の [#14296](https://github.com/docker/compose/pull/14296)、#14305（draft）、#14300、#14302、#14304、#14306、統合ブランチ #14308 が open |
| provider / relay まわり（11 を除く） | 10-15 に `internal/` へ移ったばかりで、ndeloof が毎日のように変更している |
| サービス選択（`cmd/compose/up.go`、`scale.go`、watch の rebuild） | ndeloof が新エピック [#14346](https://github.com/docker/compose/issues/14346)（10-15）を立てた |
| `create.go` | glours の #14342、ndeloof の #14340（umask）、#14299 が open |
| jobs（`cmd/compose/*` と、`down.go`、`run.go` などを広く触る） | glours の draft [#14155](https://github.com/docker/compose/pull/14155) |
| [#11816](https://github.com/docker/compose/issues/11816) `!reset` が効かない | compose-go 側で ndeloof の compose-go#913 が進んでいる（draft を外れ、レビュー待ち） |
| [#13672](https://github.com/docker/compose/issues/13672), [#13809](https://github.com/docker/compose/issues/13809) | すでに PR がある。#13672 には #14198（止まっている）と [#13849](https://github.com/docker/compose/pull/13849)（10-14 更新）の 2 つ。#13809 には #13811 があり、ndeloof が 10-08 にも指摘していて動いている |
| [#14283](https://github.com/docker/compose/issues/14283), [#14288](https://github.com/docker/compose/issues/14288) | ndeloof が #14299 で修正中 |
| [#14204](https://github.com/docker/compose/issues/14204) フック出力が `-d` で消える | moby 側の機能（moby#53407、未マージ）待ち |
| [#13602](https://github.com/docker/compose/issues/13602), [#13348](https://github.com/docker/compose/issues/13348) | main で再現しない |
| executor の multi-hop のテスト（glours の #14274 レビュー） | ndeloof が #14081 に「Lot 1/2 の次の PR で拾う」と書いている |
| 使われていない公開 API の削除（エピック #14074 B） | ユーザーへの影響が無く、公開 SDK の破壊的変更。メンテナが自分で進めている（#14322、#14331、#14333） |
| backend の作り方が 3 通りある、`ProjectOptions` の書き換えなど（エピック #14074 E/F の一部） | 調べた結果、ユーザーから見える影響が無いリファクタ。メンテナが活発に変更中のファイル |
| provider イメージ配布の `SetLimit(0)` で止まる件 | CLI からは起きない（`--parallel` は `> 0` のときしか渡されない）。SDK で明示的に 0 を渡したときだけ |

## エピックの進み具合（10-16）

- **[#14081](https://github.com/docker/compose/issues/14081)（start の plan 化）**: すべて ndeloof が進めている。
  - 10-02 以降にマージ: Lot 1 の #14274 と #14289、Lot 2 の #14303、#14290（detached up が単一の plan で動く）、#14291（scale と watch の rebuild）。
  - 残り: #14296（start の start-only plan）、#14305（interactive up、draft）など。Lot 3（古いコードの削除）は未着手。
- **[#14074](https://github.com/docker/compose/issues/14074)（コードが自分を正しく説明する）**: 本文は 08-28 から編集されていない。
  - チェックが付かないまま実質片付いたもの: B の DownOptions.Images（#14149）、E の swarmEnabled と Watcher の mutex（#14323）。
  - 残っている小さなユーザー向けのバグ:
    - F の classic `--push`（上の 10）と classic のラベル（次点）
    - E の `--quiet` の実装（上の 12）
    - D の WATCH_TAR（ndeloof 自身の #14085 が 08-18 から止まっている）

## 調査中に分かったプロセス上の注意

[contributing/pr-trends.md §8 の補足](../contributing/pr-trends.md#補足-next-tasks-の調査中に分かったプロセス上の注意) に移した。

再調査で新しく分かったこと:

- fork からの PR は、push し直すたびに、CI を動かすのにメンテナの承認が要る。承認後に main を merge すると、auto-merge が外れる（#14310）。
  承認をもらったら、ブランチはできるだけ触らない。
- 外部の依存（docker-model など）が CI にどう入るかは、GitHub のリリースではなく CI のインストール経路（`ci.yml`）で確かめる。
- 新しい issue が少ない時期は、issue から探すより、よく使われるコマンドのコードを読んで「`up` ではやっているのに `run` ではやっていない」のような食い違いを探すほうが、候補が見つかる。
