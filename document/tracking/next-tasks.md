# 次に着手できるタスク候補

← [document 一覧](../README.md)

[pr-trends.md](../contributing/pr-trends.md) の結論（実在する問題か／根本原因を直すか／メンテナの作業とぶつからないか／
メンテナの後押しがあるか）を物差しにして、「まだ誰も着手していない、外部から着手できて、
プロジェクトを本当に良くするタスク」を探した結果。

- 調査日: 2026-10-03（初回）、2026-10-16（再調査）、2026-10-20（第 3 回）
- 基準コード: `upstream/main` 289d8a037（2026-10-16）。最新リリースは v5.6.0（10-02）で、そこから main までに**未リリースのコミットが 127 件**ある（plan engine の大きな変更を含む）
  - 前回までの節（1〜13）のコードリンクは、当時の 7b36d08dd を指したまま。第 3 回の節（N1〜N17）の行番号は 289d8a037 のもの
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
  - 第 3 回: 「大きなインパクト」を狙い、これまで見ていなかった 7 つの方向を、調査のエージェントで並行して調べた。どれも実機での確認が中心
    1. 次のリリースの退行: v5.6.0 と main を、約 45 のシナリオで実際に動かして比べた
    2. 性能: API 呼び出しを記録する中継ソケットを挟み、コマンドごとの呼び出しの数を数えた。50 ms の遅延の模擬と、本物の ssh でも測った
    3. 利用者の困りごと: リアクション順の issue、stale で閉じられた要望、フォーラム、Hacker News など
    4. 正しさ: フラグ → Options → 実装の対応を全コマンドで確認、静的解析、前回「未確認」だった疑いの実機での確認
    5. 他のコンポーネントとの境界: bake、buildx、compose-go、`.env`
    6. CI: 8 週間分の 857 run、536 ジョブのログを集計
    7. 安全性: データの消失、破壊的な操作の範囲、秘密情報の漏れ、リソースの漏れ
    - 重要な発見は、自分（メインの会話）でもコードを読むか、実機で再現して確かめた
- 「検証」列の意味:
  - **実機** = 自分で upstream main をビルド・実行して再現した
  - **実機（調査）** = 調査のエージェントが実機で再現した。自分はコードを読んで原因を確かめた
  - **テスト** = 単体テストやツールの出力で再現した
  - **CI** = CI のログで確認した
  - **コード** = 現行コードを読んで確認した（実行はしていない）
- 下のコードリンクのうち GitHub を指すものは、手元の作業ブランチにまだ無いファイル（provider / relay / inspect の `internal/` への移動などがあったため）

## 進行中（自分の PR と issue）

| 対象 | 状態（2026-10-20） | 次の一手 | 詳細 |
| ---- | ------------------ | -------- | ---- |
| PR [#14310](https://github.com/docker/compose/pull/14310) down のフック後始末 | APPROVED。CI の実行承認待ち | 待つ。ブランチは触らない | [5](#5-down-の-pre_start-フックコンテナ後始末) |
| PR [#14351](https://github.com/docker/compose/pull/14351) CI の偽の FAIL | REVIEW_REQUIRED（10-17 から動き無し） | 待つ。本文の AI Tool の欄と、Related issue に残った `Fixes #` を直す | [1](#1-ci-の-unit-テストで毎回-1-件の偽の-fail-が出る) |
| issue [#14357](https://github.com/docker/compose/issues/14357) + PR [#14358](https://github.com/docker/compose/pull/14358) `run` の匿名ボリューム | REVIEW_REQUIRED | 本文の AI Tool の欄を埋める。`scale` の同じバグ（[N10](#n10-scale-も匿名ボリュームを引き継がない)）を issue に追記するか考える | [2](#2-run-が依存サービスを作り直すと匿名ボリュームを引き継がない) |
| issue [#14362](https://github.com/docker/compose/issues/14362) + PR [#14363](https://github.com/docker/compose/pull/14363) watch の `.git` | REVIEW_REQUIRED | 本文の AI Tool の欄を埋める | [3](#3-watch-の監視中に-git-が除外されない) |
| PR [#14284](https://github.com/docker/compose/pull/14284) `ensureModels` | CHANGES_REQUESTED（上流の修正の配布待ち） | apt に新しい docker-model が出たら rebase | [8](#8-ensuremodels-のエラー無視とデータ競合) |
| issue [#14311](https://github.com/docker/compose/issues/14311) `start -p` | open、反応無し | [保留](#保留メンテナの方針と合わなかったもの) | — |

## 結論（2026-10-20 時点で先にやる順）

### 上位（効果が大きく、今すぐ動く価値があるもの）

| 順 | タスク | 種類 | 誰に効くか | 検証 | 規模 | 衝突 | 最初の一手 |
| -: | ------ | ---- | ---------- | ---- | ---- | ---- | ---------- |
| 1 | [N1 起動の失敗が無関係なサービスの起動まで止める](#n1-起動の失敗が無関係なサービスの起動まで止める) | **未リリースの退行** | 次のリリースを使う全員。ポートの衝突などで、直前まで動いていたサービスが止まったまま残る | **実機** | 修正は設計判断 | 高（plan engine） | **次のリリースの前に issue で報告** |
| 2 | [S1 セキュリティに関わる発見](#s1-セキュリティに関わる発見) | セキュリティ | （非公開） | 実機（調査）+ コード | XS | — | security@docker.com に**非公開で**報告。公開の issue や PR にしない |
| 3 | [N2 `down -v` が同じ名前を持つ別のプロジェクトのボリュームを消す](#n2-down--v-が同じ名前を持つ別のプロジェクトのボリュームを消す) | **データ消失** | 同じリポジトリを複数の worktree やディレクトリで動かす人など | **実機** | XS〜S | 低〜中 | issue → 承認後に PR |
| 4 | [N3 依存の待ちが必ず 500 ミリ秒かかる](#n3-依存の待ちが必ず-500-ミリ秒かかる) | 性能 | `depends_on` の healthy / completed を使う全員。`up -d` 0.8 → 0.2 秒、`up --wait` 約 −1 秒 | 実機（調査）+ コード | XS | 中（#14308） | 数字を付けて issue → PR |
| 5 | [N4 `up -t` が Ctrl-C と `--abort-on-container-exit` で効かない](#n4-up--t-が-ctrl-c-と---abort-on-container-exit-で効かない) | バグ | CI の `--abort-on-container-exit --exit-code-from` の利用者 | 実機（調査）+ コード | XS | 中（#14305） | issue → PR |
| 6 | [N5 名前付き context で選んだ buildx の builder が無視される](#n5-名前付き-context-で選んだ-buildx-の-builder-が無視される) | バグ | Docker Desktop などで builder を選ぶ人。Docker Build Cloud の公式の手順 | 実機（調査）+ コード | S | 低 | issue（過去の往復を踏まえ、3 ケースの表を付ける） |
| 7 | [N6 許可リスト型の `.dockerignore` で watch が無視のルールを全部捨てる](#n6-許可リスト型の-dockerignore-で-watch-が無視のルールを全部捨てる) | バグ（再ビルドが止まらない） | `*` + `!src` 型の `.dockerignore` と watch を使う人（Node / Python で一般的） | 実機（調査。2 つの調査が別々に発見）+ コード | S〜M | 低 | issue → PR（#14363 が通ってから出すと話が早い） |
| 8 | [N7 `down -v` でサービスを指定すると依存先の名前付きボリュームが消える](#n7-down--v-でサービスを指定すると依存先の名前付きボリュームが消える) | **データ消失** | `down -v <svc>` を使う人（依存先のコンテナが無いとき） | **実機** | S〜M | 低〜中 | issue（直し方を 2 案から選んでもらう） |
| 9 | [N8 plan engine で止めるコンテナに pre_stop が走らない](#n8-plan-engine-で止めるコンテナに-pre_stop-が走らない) | 退行（v5.2.0〜） | pre_stop の利用者（scale down、`restart: true`、ネットワークの作り直し） | 実機（調査。v5.1.4 と比較）+ コード | S | 高（plan engine） | issue |
| 10 | [N9 同じ依存先を必須と任意の両方から待つと任意側も起動されない](#n9-同じ依存先を必須と任意の両方から待つと任意側も起動されない) | 未リリースの退行 | 同じ依存先を `required: true` と `false` の両方から待つ構成 | 実機（調査） | S | 高（plan engine） | issue（N1 と同じ時期に） |
| 11 | [N10 `scale` も匿名ボリュームを引き継がない](#n10-scale-も匿名ボリュームを引き継がない) | データ消失系 | `scale` で依存先が作り直されるとき | 実機（調査）+ コード | XS | 低〜中 | #14357 に追記して相談 |
| 12 | [N11 pre_stop のあるサービスで `restart` と `stop` が壊れる](#n11-pre_stop-のあるサービスで-restart-と-stop-が壊れる) | バグ | pre_stop の利用者 | 実機（調査）+ コード | XS | 低〜中 | issue → PR |
| 13 | [13 `pkg/compose` のテストが Windows でコンパイルできない](#13-pkgcompose-のテストが-windows-でコンパイルできない)（＋ CI のチェック） | CI | メンテナ（毎回レビューで人が手当てしている） | テスト（調査で修正の試作まで） | XS + S | 中 | issue（ndeloof の発言を引用）→ PR |

### その次（余裕があるとき）

| 順 | タスク | 種類 | 規模 | 最初の一手 |
| -: | ------ | ---- | ---- | ---------- |
| 14 | [N12 `ps` がコンテナごとに inspect する](#n12-ps-がコンテナごとに-inspect-する) | 性能（リモート、ssh） | S | 数字を付けて issue → PR |
| 15 | [N14 unhealthy で失敗したときに healthcheck の出力を出す](#n14-unhealthy-で失敗したときに-healthcheck-の出力を出す) | 使い勝手（CI） | S | issue → PR |
| 16 | [N13 ssh の Docker で接続が多すぎて `stop` や `down` が失敗する](#n13-ssh-の-docker-で接続が多すぎて-stop-や-down-が失敗する) | 信頼性 | S（既定値の変更） | issue |
| 17 | [N15 `publish` の機密データの検査が push する中身を見ていない](#n15-publish-の機密データの検査が-push-する中身を見ていない) | 安全性 | S | 自分でも確かめてから issue |
| 18 | [N16 watch の sync がボリュームの中身を消す](#n16-watch-の-sync-がボリュームの中身を消す) | データ消失（開発時） | S | 自分でも確かめてから issue |
| 19 | [N17 CI の基盤の小さな改善](#n17-ci-の基盤の小さな改善) | CI | XS ずつ | (a) は PR、(b) は issue |
| 20 | [4 bake のテンプレート評価](#4-bake-のビルドが-curl-の書式指定などで失敗する)（範囲がすべての文字列に広がった） | バグ | S | issue |
| 21 | [6 `TestWatchSyncIntoSymlinkedDirectory`](#6-e2e-testwatchsyncintosymlinkeddirectory-が不安定) | 不安定なテスト | S | PR |
| 22 | [9 `events --until`](#9-events---until-が必ず-eof-エラーで終わる) | バグ | XS | issue |
| 23 | [7 stats のテスト](#7-stats-のテストの決定化と取得失敗の-debug-ログ) | テスト | S | PR（レビューを引用） |
| 24 | [10 classic build の `--push`](#10-classic-build-の---push-がプロジェクト全体をサービス数だけ-push-する) | バグ | S | issue |
| 25 | [11 provider の ctx](#11-provider-のメタデータ取得が呼び出し元の-ctx-を使わない) | バグ（小） | S | PR（レビューを引用） |
| 26 | [12 `config -q`](#12-config--q-が一覧系フラグと組み合わせると出力する) | バグ | S〜M | issue |

### 進め方

- **1 と 2 だけは急ぐ**。1 は次のリリース（v5.6.0 の次）に入る前に知らせる価値が最も大きい。2 はセキュリティなので、必ず非公開の窓口で。
- **3〜13 は、どれも「まだ報告されていない問題」**。AI_POLICY.md の「No contribution farming」は、未報告の問題について AI を使った PR を出す前に、issue でメンテナに必要性を確かめてもらい、承認を得ることを求めている。だから順番は必ず「issue → 承認 → PR」。
- **メンテナの注意は限られている**。今も自分の PR が 5 本開いている。新しい issue は 1〜2 件ずつ、反応を見ながら出す。規模に対する効果が最も大きいのは、データ消失の N2 と、性能の N3。
- **plan engine の中のもの（N1、N8、N9）は、issue に再現手順と原因の場所を書くところまで**。直し方は ndeloof の判断に任せる。
- 「その次」の表は、上位が片付いてから。

### 前回（10-16）から変わったこと

- 次点の「ライフサイクルの調査で気付いた未確認の 3 件」を実機で確かめた:
  (1) pre_stop → [N8](#n8-plan-engine-で止めるコンテナに-pre_stop-が走らない)（v5.2.0 からの退行と確定）、(2) `up -t` → [N4](#n4-up--t-が-ctrl-c-と---abort-on-container-exit-で効かない)（確定）、(3) `--abort-on-container-exit` → 設計どおり（[やらないもの](#やらないもの理由つき)）。
- 13（Windows）は、壊れているのが `pkg/compose` だけと分かり、修正の試作と CI のチェックの案ができたので、順位を上げた。
- 6（watch の symlink のテスト）は、CI 全体の集計で、残っている唯一の未修正の不安定なテストと分かった。
- 4（bake）は、テンプレートとして評価されるのが args などに限らず、ビルドのすべての文字列だと分かった。

---

## 非公開で報告するもの

### S1. セキュリティに関わる発見

- 1 件ある。**このファイルは公開リポジトリ（自分の fork）に置かれるので、内容はここに書かない**。
- CONTRIBUTING.md の「Reporting security issues」に従い、公開の issue にせず、security@docker.com に非公開で報告する。
- 修正の規模は XS だが、報告への返答で方針が決まるまで、修正の PR も出さない。

---

## 第 3 回（2026-10-20）で見つけた候補の詳細

## N1. 起動の失敗が無関係なサービスの起動まで止める

- **種類**: **未リリースの退行**。v5.6.0 では起きず、main で起きる。このままだと次のリリースに入る。
- **症状**: 起動のフェーズで 1 つでも失敗すると（ポートの衝突、フックの失敗、必須の依存の unhealthy、`--wait-timeout`）、同時に進んでいた**無関係な**サービスの起動、ヘルスチェックの待ち、pre_start まで打ち切られる。
  - 設定を変えて作り直したサービスは、古いコンテナが消えた後に起動されず、`created`（止まった状態）で残る。つまり、直前まで動いていたサービスが止まる。
  - そのサービスについてのメッセージは出ない。進捗の表示には `Waiting` などの行が、終わりの表示の無いまま残る。
  - 実行中の pre_start フックのコンテナは kill されて消える。
- **再現（自分でも確認）**: db（healthcheck あり）、api（db を `service_healthy` で待つ）、web（ポートを公開）の 3 サービス。
  `up -d --wait` → `stop web` → 別のコンテナで web のポートを取る → api の環境変数を変えて `up -d`。
  - 5.0.1: exit 1、`api=running`（2/2 回）
  - main（289d8a037）: exit 1、`api=created`（2/2 回。調査では 3/3 回）
  - ほかの再現例（調査）: 遅い pre_start と、すぐ失敗する pre_start を並べると、main では遅いほうのフックのコンテナが kill される（2/2 回）
- **原因**: 起動のフェーズのノードは、1 つのキャンセルの範囲で動く（`executor.go:218` の `newPhaseCancel`）。最初の失敗で `phase.fail` がフェーズ全体の ctx をキャンセルする（`executor.go:276-282`、`:392`）。
  - v5.6.0 の `start()` は、traversal の ctx ではなく呼び出し元の ctx を `startService` に渡していたので、始まった起動は最後まで走った（v5.6.0 の `start.go:68-77`。自分でも確認）。
  - 影響が出始めたコミット: `up -d` は #14290（4e9477da3）、`start`・前景の `up`・`run` の依存は #14296（7509431e0）、`scale`・watch の rebuild は #14291（c4e6d9962）。
- **意図した変更ではないと判断した根拠**: 関連するコミットの説明が気にしているのは、Create と Start のフェーズの間の影響だけ。#14290 / #14296 のレビューにも、エピック #14081 にも、この違いの議論が無い。
- **修正の試作（調査）**: 起動のフェーズでだけ兄弟をキャンセルしないようにすると、v5.6.0 と同じ結果になる。
  ただし `TestExecutePlanConcurrentPhaseFailureKeepsOriginatingError` が、キャンセルで待ちを解く前提のテストなので、終わらなくなる。どう直すかはメンテナの判断。
- **既存**: 無し（10-02 以降の issue / PR を検索）。
- **衝突**: 高（executor は ndeloof の作業の中心。#14302 / #14305 が open）。
- **最初の一手**: 次のリリースの前に、上の再現手順と原因の場所を付けて issue で報告する。PR は出さない。

## N2. down -v が同じ名前を持つ別のプロジェクトのボリュームを消す

- **種類**: データ消失（5.0.1 と main の両方）。
- **症状**: 2 つのプロジェクトが、同じ `name:` のボリューム（`external` 無し）を宣言していると、片方の `down -v` が、もう片方が作ったボリュームを消す。
  - `up` のほうは「別のプロジェクトのもの」と警告して、作り直しもしない。`up` と `down` で扱いが食い違っている。
  - 起きやすい場面: 同じリポジトリを複数の git worktree やディレクトリで動かす（compose.yaml の `name:` が同じになる）。`external: true` の書き忘れ。
- **再現（自分でも確認、5.0.1）**: A で `up` → データを書く → `down`。B で `up`（警告が出る）→ `down -v` → `Volume vm-shared-data Removed`。A のデータが消える。
- **原因（自分でも確認）**: `down.go:383-398` の `removeVolume` が、`VolumeInspect` の結果を捨てて、名前だけで消している。
  - ネットワークの削除（`removeNetwork`）は、プロジェクトのラベルで絞っている。
  - `up` 側は、`observed_state.go` の `ownedConfigHash` で持ち主を確かめている。
- **直し方**: すでに取っている inspect の結果で、`com.docker.compose.project` のラベルが自分のプロジェクトかを確かめ、違えば警告して飛ばす。XS〜S。ラベルの無いボリュームの扱いは要検討。
- **既存**: 無し。近いものに #9313 / #9172（external のボリュームが消えた件。2022 年に修正済み）。
- **衝突**: 低〜中（`down.go` は glours の jobs の draft #14155 と、自分の #14310 が触るが、別の箇所）。
- **最初の一手**: 再現手順を付けて issue → 承認後に PR。

## N3. 依存の待ちが必ず 500 ミリ秒かかる

- **種類**: 性能（`depends_on` の `service_healthy` / `service_completed_successfully` を使う全員）。
- **症状**: 依存先がすでに healthy でも、最初の確認まで 500 ミリ秒待つ。`up --wait` は、最後の確認でもう一度払う。
- **数字（調査、実測 3 回）**:
  - 何も変わらない `up -d`: 0.81 → 0.20 秒（ローカル）、1.03 → 0.56 秒（50 ms の遅延の模擬）
  - 何も変わらない `up -d --wait`: 1.77 → 0.76 秒（50 ms の遅延の模擬）。本物の ssh でも同程度
  - 遅延の大きさに関係なく効く。
- **原因（自分でも確認）**: `service_containers.go:215-221` の `waitDependency` が、ticker を先に待ってから状態を確かめる。2021 年（49e7f2d45）からこの形。
- **直し方**: ループの先頭で 1 回確かめてから ticker を待つ。約 10 行（XS）。`TestWaitDependencyDeadline` は「tick が先」を前提にしているので、mock の期待を 1 つ足す。
- **既存**: 無し。
- **衝突**: 中。ndeloof が 10-06 にこの関数を 2 回変えている（7849a5091、ebe25e323）。統合ブランチ #14308 が関数の引数を変える。差分は数行。
- **最初の一手**: 数字を付けて issue → PR。

## N4. up -t が Ctrl-C と --abort-on-container-exit で効かない

- **種類**: バグ（v2.0.0 から一度も効いていない。v1 では効いていた）。
- **症状**: 前景の `up` で Ctrl-C を押したときと、`--abort-on-container-exit` / `--exit-code-from` で止めるときに、`-t/--timeout` を無視して、`stop_grace_period`（未指定ならエンジンの既定）だけ待つ。ヘルプには「when attached」と書かれている。
  - CI でよく使う `--abort-on-container-exit --exit-code-from tests` の後片付けが、指定より遅くなる。
- **数字（調査、実機）**: TERM を無視するコンテナ（`stop_grace_period: 8s`）で、`up -t 2` + SIGINT が 8.2 秒（比較の `stop -t 2` は 2.3 秒）。abort の例は全体で 11.2 秒 → 修正後 3.8 秒（終了コードは同じ）。
- **原因（自分でも確認）**: `up.go:423-426` の `stopApplication` と、`:506-509` の `stopLateStarter` が、`api.StopOptions` に `Timeout` を渡していない。
- **直し方**: `Timeout: u.options.Create.Timeout` を 2 か所に足す（XS）。
- **既存**: 無し。
- **衝突**: 中（draft の #14305 が `up.go` を大きく変えるが、この 2 つの関数は触らない）。
- **最初の一手**: issue → PR。

## N5. 名前付き context で選んだ buildx の builder が無視される

- **種類**: バグ（v2.39.3 から）。
- **症状**: Docker Desktop の `desktop-linux` など、default 以外の context で `docker buildx use X`（`--global` なし）や `docker buildx create --use` で選んだ builder を、`docker compose build` が使わない。
  さらに、ビルドのたびに buildx の選択が消え、その後の `docker buildx build` も既定の builder に戻る。
  - Docker Build Cloud の公式の手順（`docker buildx use cloud-...` → `docker compose build`）が、そのまま壊れる（Build Cloud そのものでは未確認。仕組みは同じ）。
- **再現（調査、実機）**: 5.0.1 と main で同じ。
- **原因（自分でも確認）**: `shellout.go` の `propagateDockerEndpoint` が、buildx を呼ぶときに、`DOCKER_CONTEXT` と一緒に `DOCKER_HOST` を必ず渡す。
  docker CLI は `DOCKER_HOST` があると "default" の context を選ぶ。buildx は、選んだ builder を context の名前をキーに保存しているので、キーが合わずに選択を消してしまう。
- **経緯**: この箇所は往復している。#12961（常に渡す）→ TLS の context が壊れた #13015 → #13007（条件付き）→ リモートでビルドできない #13134 → #13182（常に渡すに戻す）→ 今回の不具合。
- **直し方**: 名前付き context のときは `DOCKER_CONTEXT` だけを渡し、`DOCKER_HOST` と TLS の証明書は、`-H` / `DOCKER_HOST` / default context のときだけ渡す。S。
  過去の 2 つの不具合を再発させないことを、3 ケース（名前付き context、`-H`、TLS の名前付き context）の表形式のテストで示す。
- **既存**: 無し。
- **衝突**: 低。
- **最初の一手**: 経緯と 3 ケースの表を付けて issue → 合意後に PR。

## N6. 許可リスト型の dockerignore で watch が無視のルールを全部捨てる

- **種類**: バグ（2 つの調査が、別々の方法で見つけた）。
- **症状**: `.dockerignore` を「`*` で全部除外し、`!src` などで必要なものだけ許す」形で書くと、watch は `.dockerignore` の除外のルールを 1 つも当てはめない。
  - ビルドコンテキストの外のファイルや、`node_modules`・`.env`・`dist` の変更でも再ビルドが走る。
  - watch の対象の中に書き込みがあると（bind mount したデータ、`cache_to` のローカルキャッシュなど）、再ビルドが止まらなくなる（45 秒で 23〜42 回。拒否リスト型なら 1 回）。
- **原因（自分でも確認）**: `pkg/watch/dockerignore.go:136-143` の `NewDockerPatternMatcher` が、パターンに `*` があると、`!` 以外を全部捨てる。
  #12236 の対策として #12239（2024-10）で入った。#12236 の本当の原因は、初回の同期がファイル名だけで照合していること（`watch.go:791`、`:814`）。
- **直し方**: 特例を消し、初回の同期をフルパスで照合する（Dockerfile と compose ファイル名の固定パターンは名前のまま）。S〜M。#12236 の回帰テストも付ける。`dockerignore_test.go` の期待値が変わる。
- **既存**: 無し。
- **衝突**: 低（`pkg/watch/dockerignore.go` を触る open PR は無い）。
- **最初の一手**: issue → PR。#14362 / #14363（`.git`）と同じ「watch が止まらない」種類なので、そちらがレビューされてからのほうが話が早い。

## N7. down -v でサービスを指定すると依存先の名前付きボリュームが消える

- **種類**: データ消失（5.0.1 と main）。
- **症状**: compose ファイルがある状態で `down -v web` を実行すると、止めるコンテナは web だけなのに、消すボリュームは web と**その依存先**（db など）のもの全部になる。
  依存先のコンテナが動いていれば「使用中」で守られるが、無ければ（全体を `down` した後など）消える。
- **再現（自分でも確認、5.0.1）**: `up -d` → db にデータを書く → `down` → `down -v web` → `Volume vm-scope_dbdata Removed`。
- **原因（自分でも確認）**: `cmd/compose/down.go:80` の `projectOrName(services...)` から、`internal/projectloader/loader.go:150` の `WithSelectedServices` と `:157` の `WithoutUnnecessaryResources` で、モデルが「指定したサービスとその依存先」に絞られる。
  `pkg/compose/down.go` の `ensureVolumesDown` は、絞られたモデルのボリュームを全部消す。
  - 補足: `-p` だけで実行したときは、依存元も止める（`WithRootNodesAndDown`）。指定の仕方で挙動が違う（[compose-lifecycle.md の 1](../concepts/compose-lifecycle.md#1-depends_on) も参照）。
- **直し方の案**:
  - (a) 実際に止めるサービスだけが使うボリュームに限る（ネットワークと `--rmi` も同じ）。S〜M。
  - (b) サービスを指定したときは、`-v` と `--rmi` を拒否する。XS だが挙動が変わる（#10552 で milas が出した案）。
- **既存**: 無し。#10552（2023）のレビューで milas が「Volumes and --rmi flag are problematic」と指摘したが、採られなかった。
- **衝突**: 低〜中。
- **最初の一手**: issue で、(a) と (b) のどちらにするかを聞く。

## N8. plan engine で止めるコンテナに pre_stop が走らない

- **種類**: 退行（v5.2.0 から。v5.6.0 と main も同じ）。

  | 場面 | v5.1.4 | v5.2.0 以降 |
  | ---- | ------ | ----------- |
  | scale down（`up --scale`、`scale`） | 走る | 走らない |
  | `restart: true` の依存元を止めるとき | 走る | 走らない |
  | ネットワークの作り直し | 走る | 走らない（v5.4.0 から） |
  | 作り直し（recreate）そのもの | 走らない | 走らない（退行ではない） |

- **原因（自分でも確認）**: plan の `OpStopContainer` を実行する `executor_ops.go:309-314` の `execStopContainer` が、`ContainerStop` を呼ぶだけで pre_stop を実行しない。古い経路（`down.go` の `stopContainer`）は実行している。
- **直し方（調査の試作）**: 実行中のコンテナなら、サービスの PreStop を走らせる（約 12 行）。
  ただし、作り直しでも pre_stop が走るようになる（v5.1.4 にも無かった新しい挙動）ので、入れるかは設計判断。
- **既存**: 無し。
- **衝突**: 高（plan engine）。
- **最初の一手**: issue で、退行であること（上の表）と、作り直しでも走らせるかを聞く。

## N9. 同じ依存先を必須と任意の両方から待つと任意側も起動されない

- **種類**: 未リリースの退行（main だけ）。
- **症状**: 同じ依存先を、あるサービスは `required: false`、別のサービスは `required: true` で待っていて、依存先が失敗した場合:
  - v5.6.0: 任意側は警告を出して起動する。
  - main: 任意側も起動されず `created` のまま残る。警告も出ない。
- **再現（調査、実機）**: dep（必ず unhealthy）、opt（`required: false`）、req（`required: true`）。`service_healthy` で 3/3 回、`service_completed_successfully` で 2/2 回。
- **原因**: `reconcile.go:1357-1386` の `waitConditionNode` が、待ちのノードを（依存先、条件）ごとに 1 つにまとめ、必須の依存元が 1 つでもあると全体を必須にする。待ちが失敗すると、依存元がすべて飛ばされる（`executor.go:356-363`）。
  エピック #14081 の本文は「`required: false` は局所的に吸収する」と書いている。
- **直し方**: まとめる単位に `required` を加える。S。
- **既存**: 無し。
- **衝突**: 高（plan engine）。
- **最初の一手**: N1 と同じ時期に issue。1 つにまとめるか別にするかは、メンテナの負担を考えて決める。

## N10. scale も匿名ボリュームを引き継がない

- **種類**: データ消失系（v5.1.4 からずっと）。`run` の #14357 / #14358 と同じバグ。
- **症状**: 依存先の設定を変えてから `docker compose scale web=2` を実行すると、依存先が新しい匿名ボリュームで作り直される。`up -d` なら引き継ぐ。
- **原因（自分でも確認）**: `scale.go:38` が `api.CreateOptions{Services: services}` だけを渡し、`Inherit` が false のまま。
- **直し方**: `Inherit: true` を 1 行（XS）。実機で直ることを確認（調査）。
- **衝突**: 低〜中（エピック #14346 が ScaleOptions を作り直す予定）。
- **最初の一手**: #14357 に「`scale` でも同じ」とコメントし、#14358 に含めるか別にするかをメンテナに聞く。

## N11. pre_stop のあるサービスで restart と stop が壊れる

- **種類**: バグ。
- **症状**:
  - (1) 止まっているサービスに `restart` すると、`container ... is not running` で失敗して起動しない（v2.40.3 / v5.0.0 から）。
  - (2) 一時停止中（paused）のサービスに `stop` すると、終了コード 0 なのに、paused のまま止まらない（v2.38.0 から）。
- **原因（自分でも確認）**:
  - (1) `restart.go:115-121` が、状態を見ずにフックを実行し、失敗すると return する。
  - (2) `down.go:439-441` が、Conflict を受けると `return nil` で抜け、停止そのものをしない。
- **直し方**: どちらも「running のときだけフックを走らせる」（XS）。実機で直ることを確認（調査）。
- **衝突**: 低〜中。
- **最初の一手**: issue → PR。

## N12. ps がコンテナごとに inspect する

- **種類**: 性能（特にリモートの Docker、ssh）。
- **症状**: `ps` は、Health と終了コードを得るためだけに、コンテナごとに inspect を呼ぶ（`-q` や `--services` でも）。20 コンテナで API 呼び出しが 22 回、接続が 20 本。
- **原因（自分でも確認）**: `internal/inspect/ps.go:45-64`。API 1.52 以降は、コンテナの一覧に Health が含まれる。
- **直し方**: 一覧の Health を使い、終了したコンテナだけ inspect する。API 1.52 未満では今のまま。S（約 20 行）。
- **数字（調査、実測）**: 22 → 2 回。ssh で 50 コンテナ 1.5 秒 → 0.28 秒。JSON の出力は同じ。
- **衝突**: 低。
- **最初の一手**: 数字を付けて issue → PR。
- **同じ種類で、もっと小さいもの**: `images` が同じイメージを何度も inspect する（20 サービスで 40 回 → 1 回にできる、XS）。`up` などで `/_ping` が 1 回多い（XS）。

## N13. ssh の Docker で接続が多すぎて stop や down が失敗する

- **種類**: 信頼性と性能（`DOCKER_HOST=ssh://` の利用者）。
- **症状**: ssh では、API の接続ごとに ssh のセッションを張る。並列数は既定で無制限なので、50 サービスで `stop` / `down` が `ssh … dial-stdio exited with status 255` で失敗した（sshd の MaxStartups の既定 10 を超える）。既存の #10117 などを再現した。
- **直し方の案**: ssh の endpoint で、並列数の指定が無いときの既定値を約 8 にする（`cmd/compose/compose.go:520-527` 付近）。`--wait` の最後の確認も同じ上限に入れる。S。既定値を変えるので合意が要る。
- **数字（調査、実測）**: 上限 8 で失敗がゼロ。読み取り系のコマンドは 50〜80% 速くなる。
- **既存**: #8191、#8856、#9448、#9870、#10117、#11677（どれも stale / not planned でクローズ）。ndeloof は #10117 で「compose での短期的な修正は無い」と書いた。既定の上限という案は、まだ出ていない。
- **最初の一手**: issue で、既定値の案と数字を出す。
- **関連（plan engine の中心なので、issue に数字を書くだけ）**:
  - コンテナの起動が、プロセス全体の 1 本のロック（`startMx`、`service_containers.go:424-435`）で直列になっている。50 サービスの `start` が約 12 秒 → 約 3.5 秒になる。ただし ssh の上限と一緒でないと、ssh で失敗が増える。
  - コンテナの作成が `depends_on` の段ごとに待つ（`reconcile.go` の `infrastructureDeps`）。4 段の鎖で、作成に 0.8 秒（1 段なら 0.2 秒）。

## N14. unhealthy で失敗したときに healthcheck の出力を出す

- **種類**: 使い勝手（CI での原因調査）。
- **症状**: `--wait` や `service_healthy` が unhealthy で失敗すると、`container X is unhealthy` としか出ない。Docker が持っている healthcheck の出力（例: 「5432 番に接続できない」）は捨てている。
- **原因（自分でも確認）**: `service_containers.go:578-612` の `isServiceHealthy` は、inspect の結果を持っているのに、`State.Health.Log` を使わない。
- **直し方**: 最後の probe の終了コードと出力を、短く切ってエラーに付ける。追加の API 呼び出しは要らない。S。
- **需要**: #9122（👍23）と #10596（👍29）の議論、外部の記事（Zenn「docker compose up --wait の落とし穴」）。「CI でログに原因が出ない」という不満に直接効く。
- **衝突**: 低（`service_containers.go` を触る open PR は、draft の #14308 と、別の箇所の #14267 だけ）。
- **最初の一手**: issue → PR。

## N15. publish の機密データの検査が push する中身を見ていない

- **種類**: 安全性（公開の issue でよい）。
- **症状**: `docker compose publish` は、読み込んだモデルを YAML に戻したものを検査するが、push するのは元のファイルそのもの。
  コメントに書いたパスワードや、`command`・healthcheck・`x-*` の値が、検査をすり抜けて公開される（調査、ローカルの registry で確認。5.0.1 も同じ）。確認の文面には、検出した値が平文で出る。
- **原因**: `publish.go:761-777` と `:860-870` が `MarshalYAML()` の結果を検査し、`:261-262` が `os.ReadFile` の中身を公開する。
- **直し方**: push する中身そのものを検査し、確認では値を伏せる。S。
- **検証**: 実機（調査）。自分はまだ確かめていない。
- **最初の一手**: 自分でも確かめてから issue。

## N16. watch の sync がボリュームの中身を消す

- **種類**: データ消失（開発時）。
- **症状**: sync が `./src → /app`、名前付きボリュームが `/app/uploads` の構成で、ホストの `src/uploads` を消すと、コンテナで `rm -rf /app/uploads` が走り、ボリュームの中身が消える（警告が 1 行出るだけ）。
- **原因**: `internal/sync/tar.go:76-92` が、ホストに無いパスを無条件に消す。`watch.go:148-186` は、それがボリュームのマウント先かを確かめない。
- **直し方**: コンテナ側のパスがボリュームや tmpfs のマウント先（とその下）なら、削除を飛ばして警告する。S。
- **検証**: 実機（調査、main）。自分はまだ確かめていない。
- **衝突**: 中（`watch.go`）。
- **最初の一手**: 自分でも確かめてから issue。

## N17. CI の基盤の小さな改善

8 週間分（857 run、536 ジョブ）の集計から:

- (a) Dockerfile の `go mod download` に再試行を入れる（XS）。proxy.golang.org の HTTP/2 エラーで、27 ジョブ・16 回の試行が落ちていた（週に約 2 回の手動の再実行）。
- (b) **main の Codecov へのアップロードが、2026-01-22 から毎回失敗している**。
  - 自分でも確認した: 10-16 の main の [run 38929564017](https://github.com/docker/compose/actions/runs/38929564017) の coverage ジョブに「Token required because branch is protected」。ジョブは green なので、誰も気付いていない。
  - Codecov 上の main は 01-21 のまま 53%。実際は約 71%（mock を除くと約 76%）。
  - 直すには token か OIDC の設定（メンテナの作業）が要るので、issue で知らせる。`pkg/mocks` を ignore に足す提案も一緒に。
- (c) standalone の e2e で、`GOCOVERDIR` の警告が毎回 stderr に出る（失敗メッセージの 1 行目がこの警告になる）。XS〜S。
- (d) main の push の 22% が、次の push で cancel されて CI の結果を持たない（`cancel-in-progress: true`）。PR だけ cancel する案（XS。方針なので聞く）。
- (e) unit テストが root で動くので、権限のテスト 2 本（`TestSync_StatPermissionError`、`TestWatchRecursivelySkipsUnreadableDir`）が常に skip されている。
- 参考: 不安定なテストは、メンテナが 9〜10 月にほぼ片付けた（e2e が失敗した試行の割合は 35〜45% → 3.7%）。一律の再試行（`--rerun-fails`）は、本物のバグを隠すので勧めない（上位 8 本のうち 4 本は、製品側の本物の競合だった）。

---

## 前回までの候補の詳細（番号は前回のまま）

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
- **進捗**（2026-10-19）: issue [#14357](https://github.com/docker/compose/issues/14357)（status/0-triage）と、PR [#14358](https://github.com/docker/compose/pull/14358) `fix(run): keep a recreated dependency's anonymous volumes`（OPEN、REVIEW_REQUIRED、コメント無し）を出した。
  - 実機で再現した: リリース版の 5.0.1 と upstream main の両方。古いボリュームは dangling で残るので、データは手で取り戻せる。
  - 変更: `run.go` に `Inherit: true` の 1 行（とコメント）、e2e の `TestComposeRunInheritsAnonymousVolumes`。修正前は失敗し、修正後は通ることを確認した。
  - 詳しくは [pr/pr-14358.md](../../pr/pr-14358.md)。
  - **次の一手**: レビューを待つ。PR 本文の AI Tool の欄が空なので、編集して埋める。承認後はブランチを触らない。
  - 2026-10-20 時点: 変化なし（PR も issue もメンテナの反応はまだ無い。AI Tool の欄も空のまま）。

## 3. watch の監視中に `.git/` が除外されない

- **症状**: watch のパスに `.git` が含まれると（`path: .` など）、git の操作のたびに `.git/*` がコンテナに sync されるか、rebuild が走る。
  最初の sync（initial sync）では除外されるので、挙動が食い違っている。
  - **rebuild ではさらに深刻**: ソースを 1 回編集するだけで、再ビルドが止まらなくなる（30 秒で 30〜40 回）。
    buildx がビルドのたびに来歴情報のため `.git` を書き換え、それを watch が新しい変更として拾うため。
    `BUILDX_GIT_INFO=0` か、`.dockerignore` に `.git` を書くと止まる（報告が少ないのは、後者を書いているプロジェクトが多いからと推定）。
- **根本原因**: [pkg/compose/watch.go](https://github.com/docker/compose/blob/7b36d08dd0d30da3abc1e646baec8751f0bcb0b2/pkg/compose/watch.go#L317)（317 行と 742 行）の
  `watch.NewDockerPatternMatcher("/", []string{".git/"})` は、`/.git`（ファイルシステムのルート直下）にしか一致しない。
  監視中の照合はフルパス（`/proj/.git/index` など）で行うので、一致しない。2023 年からこのまま。
- **直し方**: パターンを `**/.git` にする。表形式の単体テストを足す。
  742 行（`initialSync`）は、ファイル名だけで照合するので、今のままで `.git` を除外できている（upstream main で確認済み）。直すのは 317 行（`getWatchRules`）の 1 か所だけ。
- **検証**: 単体テストで、`Matches("/tmp/x/.git/index")` が false を返すことを確かめた。
- **後押し（間接）**: glours が外部 PR [#14217](https://github.com/docker/compose/pull/14217) のレビューで、「監視中の照合はフルパスなので、パス付きの matcher が要る」と同じ種類の欠陥を指摘している。
- **衝突リスク**: 中。
  - #14217（外部、CHANGES_REQUESTED のまま作者が止まっている）が同じ関数を触っている。
  - ndeloof の Lot 2 の PR（#14300、#14302、#14304、#14305）も `watch.go` を触っている。ただし、どれも別の箇所。
- **最初の一手**: issue を立てる。#14217 にも関わるので、「#14217 とは別に直すか、まとめるか」を聞く。
- **進捗**（2026-10-20）: issue [#14362](https://github.com/docker/compose/issues/14362)（status/0-triage、調査の経緯を説明するコメントを 1 件追加）と、PR [#14363](https://github.com/docker/compose/pull/14363) `fix(watch): ignore .git at any depth inside the watched path`（OPEN、REVIEW_REQUIRED、反応はまだ無い）を出した。
  - 実機で再現した: リリース版の 5.0.1 と upstream main の両方。buildx の古い版（v0.24.0）と新しい版（v0.30.1）の両方で起きる。
  - 変更: `watch.go` のパターンの 1 行、`watch_test.go` の表形式のテスト `TestWatchRules_IgnoreGitDirectoryAtAnyDepth`。修正前は `.git` の 4 ケースが失敗し、修正後は通る。
  - 実機での確認: 1 回の編集で再ビルドは 1 回、空のコミットで 0 回（修正前は 30 秒で 40 回）。watch の e2e 7 本も通る。
  - 検証で 1 度誤った: watch のログを監視対象のディレクトリの中に書くと、それ自体が再ビルドを起こす。ログは必ず監視の外に書く。
  - **次の一手**: レビューを待つ。PR 本文の AI Tool の欄が空なので、編集して埋める。#14217 が先にマージされたら rebase する。

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
- **更新（2026-10-20、第 3 回の調査）**: テンプレートとして評価されるのは、args・labels・dockerfile_inline に限らず、ビルドに渡す**すべての文字列**だった。
  例: `build: "./dir%{x}"` と書くと、`Invalid template control keyword` で失敗する（調査、実機）。
  直すなら、context・contexts・dockerfile・tags・secret・ssh・cache など、すべての文字列に同じエスケープを当てる。規模は S のまま。
  同じ調査で、`cache_from` / `cache_to` の `type=local` の相対パスが、プロジェクトではなく実行したディレクトリ基準になることも分かった（compose-go 側の修正。[次点](#次点先に-issue-で合意を取ってから)）。

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
- **更新（2026-10-20、第 3 回の調査）**: 8 週間分の CI の集計（857 run）で、**残っている唯一の未修正の不安定なテスト**だと分かった（それ以外は 9〜10 月にメンテナが直した）。
  - 流れ: 「Watch enabled」がコンテナの起動前に出る → テストがその直後に rename する → コンテナがまだ動いていないうちに sync が exec しようとして `cannot exec in a stopped state` で失敗する → その変更は再試行されずに捨てられる。
  - 製品側にも「起動直後の変更が黙って失われる」余地がある（推定）。テストの修正の PR では、そこまでは直さず、気付いたこととして書き添える程度にする。

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
- **更新（2026-10-20、第 3 回の調査）**: 規模がはっきりし、修正の試作もできたので、順位を上げた。
  - **壊れているのは `pkg/compose` の 1 パッケージだけ**（テストのある 26 パッケージ中）。エラーは 4 件、原因のシンボルは 3 つで、どれも `start_test.go` で定義されている（`down_test.go:1345/1351` と `push_test.go:88/147` が使う）。windows/arm64 も同じ。
  - 本当に pty が必要なのは `hook_test.go` と `cmd/compose/compose_progress_test.go` だけ。`start_test.go`・`restart_test.go`・`executor_start_test.go` は OS 固有の API を使わず、`hook_test.go` の `writeStdcopyFrame` を使っているせいで、タグが連鎖して付いている。
  - **最小の修正（調査で検証済み、+38/-23）**: `writeStdcopyFrame` をタグの無いファイルに移し、3 つのファイルからタグを外す。windows/amd64 と arm64 で vet が通り、darwin で `go test ./pkg/compose` と lint が通る。Windows で実際にテストが通るかは未確認（推定では通る）。
  - **CI に足すもの**: validate の matrix に、`GOOS=windows go vet -tags "$BUILD_TAGS" ./...` を実行するチェックを足す（約 20 行、1 ジョブ約 1 分）。golangci-lint ではなく vet を使うのは、Windows では `hook_test.go` が除外されて `unparam` の警告が余分に出るため。
  - **背景**: Windows と macOS でテストを実行していた merge.yml のジョブは、2026-04-15（9fd9dc7ca）に「runner がもう無い」として消された。それ以降、Linux 以外でテストを動かす場所が無い。Windows のテストを CI に入れる提案は、issue にも PR にも無い。
  - **衝突**: #14302（open）、#14305 / #14308（draft）が同じファイルを触る。変えるのは 1 行目のタグとヘルパーの場所だけなので、衝突しても軽い。

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
| post_start フック付きのサービスで、`run` の起動が失敗すると永久に止まる（`pkg/compose/run.go:79` の `<-hookErrCh`） | フックの goroutine は start イベントを待つが、起動に失敗すると来ない。`signal.Notify` が全シグナルを奪うので、Ctrl-C も効かない。直し方は「フック用の ctx をキャンセル可能にし、`RunStart` が失敗したら cancel する」で、約 10 行。**コードを読んだだけで、実機での確認が要る**。`run.go` は #14155 も触っている。CI の調査（10-20）で、この経路（`runPostStartHooksOnEvent`）のテストの coverage が 0% と分かった |
| git のリモート include の取得に失敗すると、壊れたキャッシュが残り続ける（`pkg/remote/git.go` の `checkout()`） | 失敗してもディレクトリを消さず、`Load` はディレクトリがあるかしか見ないので、手で消すまで毎回 `not found` になる。OCI 版（`oci.go`）は失敗時に `RemoveAll` している。約 8 行。#14268 と ndeloof の #14085 が同じファイルを触っているので、それらが片付いてから |
| classic build のイメージに compose ラベルが付かない（`getImageBuildLabels` の呼び出し元は `build_bake.go` だけ） | エピック #14074 F。初回の 6 位から下げた。ndeloof は #14050 でこれを「separate pre-existing gap」と認めている。一方で、イメージに project/service ラベルが付くと、そこから `docker run` したコンテナまで compose 管理下に見える問題も挙げている。#14312 の「ラベルを増やしたくない」もあるので、方針を聞いてから |
| `pull_refresh_after` が無視されているのに、設定ハッシュには含まれている（`hash.go`） | ndeloof が #14253 で「別の変更で判断すべき」と書いた。除外・実装・警告のどれにするか issue で聞く。10-16 時点で変化なし |
| `wait --down-project` がヘルプの説明どおりに動かない（`pkg/compose/wait.go`） | ヘルプは「最初のコンテナが止まったら」だが、実際は全部が止まるまで待つ（`eg.Wait()`）。`statusCode` を複数の goroutine が書いていて、データ競合もある。挙動が変わるので合意が要る。**10-20 に確認済み（調査）**: `go test -race` が `wait.go:70` を検出し、返る終了コードが実行ごとに 3 と 5 で変わる。実機でも、先に終わったコンテナから 15 秒たっても待ち続け、Down もしない |
| （解決）ライフサイクルの調査（10-17）で気付いた未確認の 3 件 | 10-20 に実機で確かめた。(1) pre_stop → [N8](#n8-plan-engine-で止めるコンテナに-pre_stop-が走らない)（v5.2.0 からの退行）。(2) `up -t` → [N4](#n4-up--t-が-ctrl-c-と---abort-on-container-exit-で効かない)。(3) `--abort-on-container-exit` と restart ポリシー → 設計どおり（[やらないもの](#やらないもの理由つき)） |
| `.env` に書いた `COMPOSE_PROGRESS` が効かない（`cmd/compose/compose.go:234` がフラグの定義時に `os.Getenv` を読む） | ndeloof はこの制約を把握していて、同じ種類の問題を #14139 で直している最中（10-20 時点で OPEN、CONFLICTING、08-31 から更新なし）。#14139 のマージ後に、その仕組みに乗せる。**同じ種類（10-20 の調査）**: `COMPOSE_STATUS_STDOUT` は `init()` で読むので、`.env` に書いても絶対に効かない。サブディレクトリから実行すると、プロジェクトの `.env` の `COMPOSE_*` がプロセスの環境変数に反映されない（`setEnvWithDotEnv` が compose ファイルを親までさかのぼって探さない）。`COMPOSE_PROJECT_NAME` と `COMPOSE_PROFILES` は効くので、変数によって挙動が違う |
| `push --ignore-missing` で全部スキップされても何も分からない（`push.go:69-71`） | ndeloof が [#14341](https://github.com/docker/compose/pull/14341) の承認時に任意の提案として書いた（「a summary… or a warning when nothing was pushed」）。警告か集計かは設計判断。#14336 が `push.go` を触っている |
| [#14313](https://github.com/docker/compose/issues/14313) bake の並列数の上限（buildx v0.38 の `--jobs` に `--parallel` を渡す） | S〜M。buildx のバージョン判定が要る。`service:` 形式の additional_contexts（buildx でいう linked target）があると、小さい上限値は buildx に拒否される。メンテナの反応はまだ無い |
| [#9122](https://github.com/docker/compose/issues/9122) `up --wait` 中にログを流す | 要望は大きい（👍23、重複 #12039）。ndeloof も 2025-12 に「wait の間はデフォルトで attach してよい」と方向に賛成している。ただし、ndeloof 自身の open な [#14306](https://github.com/docker/compose/pull/14306) は逆に「`Wait` と `Attach` は API で同時に指定できない」と固めようとしている。触るコードは draft の #14305 が書き換え中。**今は待つ**。衝突の小さい部分（失敗の理由を出す）は [N14](#n14-unhealthy-で失敗したときに-healthcheck-の出力を出す) |
| [#10596](https://github.com/docker/compose/issues/10596) `up --wait` が、一回きりのコンテナ（migration など）の exit 0 で失敗する | 👍29。upstream main で再現（調査）。フラグを足す形の PR は 2 回却下された（#11649、#12172）。メンテナの方向は「init container を正式に扱う」（`pre_start` フック、jobs）。誰からも依存されない後処理は、healthy の後に走るフック（#13203）待ち。コードの変更は出さない。出すなら、`exited (0)` のエラーに「`pre_start` にするか、`service_completed_successfully` で依存させる」案内を足す程度（XS、合意が先） |
| [#11867](https://github.com/docker/compose/issues/11867) / [#9648](https://github.com/docker/compose/issues/9648) リモートの Docker で `file:` の configs / secrets が使えない | 👍 合わせて約 60。止まっている理由は設計（実装の PR が 3 本、stale でクローズ）。ndeloof の案は「engine 側の native secrets」か「明示的に copy する pre_start フック」。出すなら設計提案のコメントから（L） |
| [#12581](https://github.com/docker/compose/issues/12581) build で bake の機能を使いたい（bake ファイルの指定、custom output） | 👍10。glours「後方互換を保つ PR は歓迎」。ただし設計の合意が無い（L）。設計提案から |
| [#13757](https://github.com/docker/compose/issues/13757) pull のサイズ表示が、ほぼ最後まで出ない | `cmd/display/tty_layout.go:155-171` の `aggregateProgress` が、層が 1 つでも「Working かつ total=0」だと表示をまるごと消す。分かっている分だけ表示する（例 `120MB / 450MB+`）。S。触る open PR は無い |
| Ctrl-C / SIGTERM を 1 回送っただけで、まれに即座に kill される | 同じシグナルを 2 か所で受け取っている（`cmd/compose/compose.go:112-115` と `pkg/compose/up.go:187`）ため、1 回目が 2 回目と見なされることがある。実機で 112 回中 3 回（調査）。v2.24.0 から。メニューが無効な前景の `up`（CI、パイプ、IDE、プロセス管理ツール）で起きる。S。#14305 と衝突（中） |
| `down -v --remove-orphans` が、orphan のコンテナの匿名ボリュームを残す | `down.go:120` が `volumes` に `false` を直書きしている。1 行（XS）。#13465 は「明示的な指示無しにボリュームを消さない」で not planned になったが、今回は `-v` が明示的な指示なので筋は通る。issue で確認してから |
| 作り直しのたびに、匿名ボリュームが漏れる | `up -V` は 1 回ごとに「匿名マウント数 × レプリカ数」（postgres なら 1 回 38.6MB）、scale down、イメージの `VOLUME` のパスの変更でも漏れる。ラベルが `com.docker.volume.anonymous` だけなので、`down -v` でも回収できず、`docker volume prune` でしか消えない（調査、実機）。直し方は匿名マウントにプロジェクトのラベルを付けること（M）。**ただし、この漏れのおかげで #14357 のデータは取り戻せた**ので、単純に削除するのは危険。優先度は低い |
| リモートの compose（OCI / git）が、参照している環境変数の値を平文で表示する | TTY が無く `-y` も無いとき、変数の値の表を出してから中断する（`cmd/compose/options.go:148-176`、`:232-252`）。どの変数を表示させるかは、リモートのファイルの作者が決められる。open の #13900 が入ると、中断せずに先へ進むので、CI で毎回この表がログに出る。値を伏せる（XS〜S）。#13900 より先に入れたい |
| `cache_from` / `cache_to` の `type=local` の相対パスが、実行したディレクトリ基準になる | compose-go の `paths/resolve.go` が、この 2 つを絶対パスに直す対象に入れていない。`-f`、`--project-directory`、`include`、サブディレクトリからの実行で、キャッシュが黙って外れ、カレントディレクトリにゴミが残る（調査、実機）。compose-go に PR（S） |
| 同じ名前のディレクトリにある別のプロジェクトが衝突する | 既知（#6898、#7969、#3966）。`working_dir` のラベルが違えば警告する改善（S）。優先度は低い |
| watch が、シンボリックリンクの先のディレクトリをたどって同期する | `internal/sync/tar.go:276-295`。docker build も docker cp もリンクをたどらない。compose ファイル自体が信頼される前提なので、新しい境界は越えない。優先度は低い |
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
| 一度 unhealthy になったら、回復を待たずにすぐ失敗する（`--wait`、`service_healthy`） | 設計どおり。ndeloof が「Compose はオーケストレーターではない」として 2 回クローズしている（[#11030](https://github.com/docker/compose/issues/11030)、[#12424](https://github.com/docker/compose/issues/12424)） |
| `--abort-on-container-exit` が、restart ポリシーで再起動中のコンテナの終了でも止まる | 設計どおり（10-20 に実機で確認）。v5.1.4 でも同じで、ndeloof も #10310 で「理由を問わず、終了で止まることを期待する利用者がいる」という立場 |
| 不安定なテストへの一律の再試行（`gotestsum --rerun-fails` など） | 9〜10 月の不安定なテストの上位 8 本のうち 4 本は、製品側の本物の競合だった。一律の再試行は、それを隠してしまう |
| CI の高速化 | 全体の中央値は 10.8 分で、キューの待ちも無い。本当に時間を食っているのは再実行（16%）なので、効果が小さい |
| CLI の起動時間と、compose ファイルの読み込みの速さ | 実測で問題なし（`version` 22 ms、空の `ps` 約 30 ms で `docker ps` と同じ。1000 サービスの読み込みで 0.35 秒）。プラグイン版の +44 ms は docker/cli 側の処理 |
| コンテナの起動のロック（`startMx`）を単純に外す | ssh の Docker で接続が増えて失敗する（調査、実測）。外すなら [N13](#n13-ssh-の-docker-で接続が多すぎて-stop-や-down-が失敗する) の上限と一緒に、メンテナの判断で |
| 要望は大きいが、方針で却下されているもの | [#4249](https://github.com/docker/compose/issues/4249) 環境変数で配列（パーサーの作り直しが要る）、[#7853](https://github.com/docker/compose/issues/7853) UID/GID や OS の変数（ポータビリティの方針）、#13554 `--progress quiet` で完全に無音、#13465 `down` で孤立ボリュームを消す、#12497 include の glob、#13993 profile だけを `down`、#13716 health の待機中に detach、#11003 watch で compose.yaml の変更を反映、#14326 exec だけの watch、#13807 ボリュームの自動確認のラベル（代わりに #14273 の `x-disposable`） |

## エピックの進み具合（10-20）

- **[#14081](https://github.com/docker/compose/issues/14081)（start の plan 化）**: すべて ndeloof が進めている。
  - 10-02 以降にマージ: Lot 1 の #14274 と #14289、Lot 2 の #14303、#14290（detached up が単一の plan で動く）、#14291（scale と watch の rebuild）、#14296（start の start-only plan、10-16）。
  - 残り: #14305（interactive up、draft）、#14302、#14306 など。Lot 3（古いコードの削除）は未着手。
  - **v5.6.0 には Lot 2 がまだ入っていない**。v5.6.0 と main の比較で、退行が 2 件見つかった（[N1](#n1-起動の失敗が無関係なサービスの起動まで止める)、[N9](#n9-同じ依存先を必須と任意の両方から待つと任意側も起動されない)）。それ以外の約 45 のシナリオでは、意図しない違いは無かった。
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

第 3 回の調査で新しく分かったこと:

- **未リリースのコミットが多い時期は、リリース版と main を同じシナリオで比べるのが効く**。次のリリースに入る前に見つけられれば、影響を受ける人がゼロのうちに直せる（N1、N9）。
- **同じ種類のバグには兄弟がいる**。`run` の `Inherit` の渡し漏れ（#14358）を見つけた後、同じ形を探すと `scale` にもあった（N10）。設定の渡し漏れや、無視のパターンの誤りを見つけたら、同じ形を grep で探す。
- **2 つの調査が別々の方法で同じものを見つけたら、確度が高い**（N6）。
- **セキュリティの発見は、公開リポジトリに置くメモに詳細を書かない**（S1）。報告は CONTRIBUTING.md の「Reporting security issues」に従って security@docker.com へ。
- **AI_POLICY.md の「No contribution farming」**: まだ報告されていない問題について AI を使った PR を出すときは、先に issue でメンテナに必要性を確かめてもらい、承認を得る。問題を確かめずに出す PR や、1 つの変更を細かく分けた PR も、閉じられうる。
- **並行の調査で共有する設定に注意**。一時的な Docker の設定（`DOCKER_CONFIG`）を複数の調査で共有すると、ある調査が builder の選択を変えたときに、他の調査の結果に影響しうる。調査ごとに分ける。
- **性能の調査は、時間ではなく API 呼び出しの数を主な指標にする**。同じ Docker を他の作業と共有していると、時間は大きくぶれる。呼び出しの数は、何度測っても同じだった。
