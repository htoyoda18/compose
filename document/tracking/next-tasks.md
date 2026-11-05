# 次に着手できるタスク候補

← [document 一覧](../README.md)

[pr-trends.md](../contributing/pr-trends.md) の物差し（実在する問題か／根本原因を直すか／メンテナの作業とぶつからないか／メンテナの後押しがあるか）で選んだ、外部から着手できる候補。
探し方と調査の注意は [investigation.md](./investigation.md) にまとめた。

- 最終更新: 2026-11-06（「進行中」は 11-05 時点）
- 最後の定点観測: 2026-11-05（v5.7.0）。上位の候補より先にやるべき新しい問題は無かった。N8、N10、N11 は v5.7.0 でも残っている（コードで確認）
- 基準コード: N1〜N17 の行番号は `upstream/main` 289d8a037（2026-10-16）。1〜13 の GitHub へのリンクは 7b36d08dd。GitHub を指すリンクは、手元の作業ブランチにまだ無いファイル
- 運用: PR を出したら、候補の詳細は `compose/pr/pr-XXXX.md` に移す。片付いたものは [完了](#完了) に 1 行だけ残す（経緯は git の履歴にもある）
- 「検証」列: **実機** = 自分で upstream main をビルド・実行して再現した／**実機（調査）** = 調査のエージェントが実機で再現し、自分はコードで原因を確かめた／**テスト** = 単体テストやツールの出力で再現した／**CI** = CI のログで確認した／**コード** = コードを読んだだけ

## 進行中（自分の PR と issue）

2026-11-05 時点。

| 対象 | 状態 | 次の一手 | 詳細 |
| ---- | ---- | -------- | ---- |
| PR [#14449](https://github.com/docker/compose/pull/14449) `down -v <service>` が依存先のボリュームを消す（N7） | APPROVED（ndeloof、11-04）、CI pass。急がない指摘 2 つ | 指摘に返信する。本文の `AI Tool:` 欄を埋める | [pr-14449.md](../../pr/pr-14449.md) |
| PR [#14434](https://github.com/docker/compose/pull/14434) 名前付き context の builder（N5） | 11-04 に agammann の指摘を直して push。REVIEW_REQUIRED、CI は実行承認待ち。issue #14433 に status/approved は無い | 再レビューを待つ。承認されたら 1 コミットにまとめる | [pr-14434.md](../../pr/pr-14434.md) |
| issue [#14447](https://github.com/docker/compose/issues/14447) 許可リスト型の `.dockerignore`（N6） | status/approved（11-04）。前提の [#14441](https://github.com/docker/compose/pull/14441) はレビュー待ち。xixi-box の [#14450](https://github.com/docker/compose/pull/14450) は #14441 を含まない | 自分では実装しない。#14450 に「#14441 が先でないと #12236 が再発する」と伝えるか決める | [N6](#n6-許可リスト型の-dockerignore-で-watch-が無視のルールを全部捨てる) |
| PR [#14310](https://github.com/docker/compose/pull/14310) down のフック後始末 | APPROVED。10-27 に rebase。CI の実行承認とマージ待ち | 待つ。ブランチは触らない | [pr-14310.md](../../pr/pr-14310.md) |
| PR [#14284](https://github.com/docker/compose/pull/14284) `ensureModels` | CHANGES_REQUESTED。10-27 の返信に返事無し | 待つ | [pr-14284.md](../../pr/pr-14284.md) |
| issue [#14372](https://github.com/docker/compose/issues/14372) 起動の失敗が無関係なサービスの起動まで止める（N1） | ndeloof が報告の仕方を [#14376](https://github.com/docker/compose/pull/14376) で直した（v5.7.0）。issue は open | 閉じてよいか聞く（11-05 の T3b で、起動されなかったサービスが「not started」と報告されるのを確認済み） | issue |
| PR [#14244](https://github.com/docker/compose/pull/14244) `run --env-from-file` の解析エラーを握りつぶす | REVIEW_REQUIRED（10-15 から動き無し）。関連の issue と AI の開示欄が無い | 今のルール（承認済みの issue が要る）に合わせるか決める | — |
| PR [#14053](https://github.com/docker/compose/pull/14053) `--dry-run` を進捗表示に渡す | REVIEW_REQUIRED（09-26 から動き無し）。CI の実行承認待ちが 10-26 に失敗扱いになった。関連の issue が無い | 同上 | [pr-14053.md](../../pr/pr-14053.md) |
| issue [#14311](https://github.com/docker/compose/issues/14311) `start -p` | ndeloof「ラベルに設定を足す方向は受けない。既知の制限として open のまま」（10-29） | [保留](#保留メンテナの方針と合わなかったもの)のまま | — |
| issue [#14453](https://github.com/docker/compose/issues/14453) `down --rmi <service>`（ndeloof が作成） | #14449 のイメージ版。Aghayenoor が [#14458](https://github.com/docker/compose/pull/14458) を出した | 手を出さない | — |

## 候補（まだ着手していないもの）

### 上位（効果が大きく、今すぐ動く価値があるもの）

| 順 | タスク | 種類 | 誰に効くか | 検証 | 規模 | 衝突 | 最初の一手 |
| -: | ------ | ---- | ---------- | ---- | ---- | ---- | ---------- |
| 1 | [S1 セキュリティに関わる発見](#s1-セキュリティに関わる発見) | セキュリティ | （非公開） | 実機（調査）+ コード | XS | — | security@docker.com に**非公開で**報告 |
| 2 | [N8 plan engine で止めるコンテナに pre_stop が走らない](#n8-plan-engine-で止めるコンテナに-pre_stop-が走らない) | 退行（v5.2.0〜） | pre_stop の利用者 | 実機（調査。v5.1.4 と比較）+ コード | S | 高（plan engine） | issue |
| 3 | [N9 同じ依存先を必須と任意の両方から待つと任意側も起動されない](#n9-同じ依存先を必須と任意の両方から待つと任意側も起動されない) | 退行（v5.7.0） | 同じ依存先を `required: true` と `false` の両方から待つ構成 | 実機（調査） | S | 高（plan engine） | 保留（意図どおりの可能性。詳細を参照） |
| 4 | [N10 `scale` も匿名ボリュームを引き継がない](#n10-scale-も匿名ボリュームを引き継がない) | データ消失系 | `scale` で依存先が作り直されるとき | 実機（調査）+ コード | XS | 低〜中 | #14357 に追記して相談 |
| 5 | [N11 pre_stop のあるサービスで `restart` と `stop` が壊れる](#n11-pre_stop-のあるサービスで-restart-と-stop-が壊れる) | バグ | pre_stop の利用者 | 実機（調査）+ コード | XS | 低〜中 | issue → PR |
| 6 | [13 `pkg/compose` のテストが Windows でコンパイルできない](#13-pkgcompose-のテストが-windows-でコンパイルできない)（＋ CI のチェック） | CI | メンテナ（毎回レビューで人が手当てしている） | テスト（修正の試作まで） | XS + S | 中 | issue（ndeloof の発言を引用）→ PR |

### その次（余裕があるとき）

| 順 | タスク | 種類 | 規模 | 最初の一手 |
| -: | ------ | ---- | ---- | ---------- |
| 7 | [N12 `ps` がコンテナごとに inspect する](#n12-ps-がコンテナごとに-inspect-する) | 性能（リモート、ssh） | S | 数字を付けて issue → PR |
| 8 | [N14 unhealthy で失敗したときに healthcheck の出力を出す](#n14-unhealthy-で失敗したときに-healthcheck-の出力を出す) | 使い勝手（CI） | S | issue → PR |
| 9 | [N13 ssh の Docker で接続が多すぎて `stop` や `down` が失敗する](#n13-ssh-の-docker-で接続が多すぎて-stop-や-down-が失敗する) | 信頼性 | S（既定値の変更） | issue |
| 10 | [N15 `publish` の機密データの検査が push する中身を見ていない](#n15-publish-の機密データの検査が-push-する中身を見ていない) | 安全性 | S | 自分でも確かめてから issue |
| 11 | [N16 watch の sync がボリュームの中身を消す](#n16-watch-の-sync-がボリュームの中身を消す) | データ消失（開発時） | S | 自分でも確かめてから issue |
| 12 | [N17 CI の基盤の小さな改善](#n17-ci-の基盤の小さな改善) | CI | XS ずつ | (a) は PR、(b) は issue |
| 13 | [4 bake のテンプレート評価](#4-bake-のビルドが-curl-の書式指定などで失敗する) | バグ | S | issue |
| 14 | [6 `TestWatchSyncIntoSymlinkedDirectory`](#6-e2e-testwatchsyncintosymlinkeddirectory-が不安定) | 不安定なテスト | S | PR |
| 15 | [7 stats のテスト](#7-stats-のテストの決定化と取得失敗の-debug-ログ) | テスト | S | PR（レビューを引用） |
| 16 | [10 classic build の `--push`](#10-classic-build-の---push-がプロジェクト全体をサービス数だけ-push-する) | バグ | S | issue |
| 17 | [11 provider の ctx](#11-provider-のメタデータ取得が呼び出し元の-ctx-を使わない) | バグ（小） | S | PR（レビューを引用） |
| 18 | [12 `config -q`](#12-config--q-が一覧系フラグと組み合わせると出力する) | バグ | S〜M | issue |

### マージ済みの PR のフォローアップ

どれも急がない指摘。

| 元の PR | 内容 | 規模 |
| ------- | ---- | ---- |
| [#14358](../../pr/pr-14358.md) | `pkg/api/api.go` の `RunOptions.CreateOptions` のコメントが古い（`Inherit` は無視される、と書いたまま）。testdata の `init: true` と `COMPOSE_IGNORE_ORPHANS=True` の理由を書く | XS |
| [#14351](../../pr/pr-14351.md) | テストの cleanup で `false` に戻すより、元の値（`old := disableAnsi`）に戻す（任意。`ansi_test.go` の書き方とは合っている） | XS |
| [#14310](../../pr/pr-14310.md) | マージ後に、#14091 の残り (d) → (a) を別々の PR に | XS、S |

### 進め方

- **S1 だけは急ぐ**。セキュリティなので、必ず非公開の窓口で。
- **それ以外は、どれも「まだ報告されていない問題」**。必ず「issue → 承認 → PR」の順（[pr-trends.md の §8](../contributing/pr-trends.md#補足-next-tasks-の調査中に分かったプロセス上の注意)）。
- **メンテナの注意は限られている**。自分の PR は 6 本開いている（11-05 時点）。新しい issue は 1〜2 件ずつ、反応を見ながら出す。
- **plan engine の中のもの（N8、N9）は、issue に再現手順と原因の場所を書くところまで**。直し方は ndeloof の判断に任せる。
- 「その次」の表は、上位が片付いてから。

## 調査のタスク

小さなタスクに分け、1 つずつやる（15〜30 分）。狙う場所と道具は [investigation.md](./investigation.md)。終わったものは [完了](#終わった調査のタスク) へ。

| ID | 内容 | 完了の条件 |
| -- | ---- | ---------- |
| T3c | Ctrl-C の後もアプリが動き続ける: T3b で、起動のフェーズに Ctrl-C すると、終了コード 130 で返った後も api、db などが running のまま残った（負荷の高い並列で 24 回中 2 回）。Ctrl-C を、db が Started / Waiting になった時点に合わせて送り、v5.6.1 と v5.7.0 で比べる | 再現の頻度、退行かどうか、どこで止め損ねているか |
| T3d | 作成中の Ctrl-C で `created` のコンテナが残る: T3b で、`up` が作成の完了を待たずに終わり、直後の `down` の後に残った（v5.7.0 は 40 回中 12 回、v5.6.1 は 0 回。次の `up` で片付く）。どのコミットで入ったか（#14305 など）を確かめる | 原因のコミットと場所、影響の大きさ |
| T4 | plan engine の機能の棚卸し: v5.1.4（plan engine の前）の作成・起動・停止の経路にあった処理を、コードから一覧にする | 機能の一覧（T5 の材料） |
| T5〜 | T4 の一覧の機能ごとに、v5.1.4 と v5.7.0 で同じシナリオを動かして比べる（1 機能 1 タスク） | 機能ごとの、違いの有無 |
| T6 | `-race` でビルドした compose（main）で、e2e の up / down / start / watch を流す | 競合の報告の有無と場所 |
| T7 | 遅延の注入: 中継ソケットを作り直し、`up` と `down` を 50〜100 ms の遅延で回す | 固まる、失敗する、API の呼び出しが増える、があるか |
| T8 | 信頼できない入力: remote include（OCI / git）の取得と展開を、信頼されない側から見る | 境界を越えるものの有無。あれば非公開で報告し、詳細はこのファイルに書かない |

## 新機能の候補（2026-10-22）

👍 の多い要望の上位約 40 件のうち 39 件は、実装済みか方針で却下で閉じている。だから「👍 の数」より「**メンテナが歓迎と書いているか**」と「**メンテナの作業とぶつからないか**」で選ぶ。新機能は**必ず issue で合意を取ってから**。

| 順 | 候補 | 後押し | 衝突 | 規模 | 最初の一手と注意 |
| -: | ---- | ------ | ---- | ---- | ---------------- |
| F1 | ターミナルのタブやタスクバーに進捗を出す（[#14365](https://github.com/docker/compose/issues/14365)） | **強い**（ndeloof が 10-21 に「歓迎」。数は `cmd/display/tty_model.go` の `taskTree.counts()` にある） | 低 | S | issue で、OSC 9;4 と 7501 のどちらから始めるかと、オプトアウトの方法を決める。TTY のときだけ出す。9;4 が通知の OSC 9 と重なる問題（nihaals の指摘）がある |
| F2 | `publish --annotation`（[#13739](https://github.com/docker/compose/issues/13739)） | **強い**（[#13852](https://github.com/docker/compose/pull/13852) は ndeloof が承認したまま stale で閉じた） | 低 | S | 元の作者に続けるか聞く。返事が無ければ Co-authored-by を残して作り直す。publish は glours の担当範囲 |
| F3 | サービスが healthy になった後に走るフック（[#13203](https://github.com/docker/compose/issues/13203)） | 中（ndeloof「`post_start` に条件を足すより、新しい種類のフックで」） | 中 | M | chris3713 の `post_healthy` の提案と組めるか聞き、spec → compose-go → compose の順に。#10596 の後処理の需要にも効く |
| F4 | `pre_start` の `per_replica`（[#14259](https://github.com/docker/compose/issues/14259)） | 中（ndeloof 自身の issue。設計の案つき） | **高**（plan engine） | M | spec は受け付けるのに実行でエラーになる（`pre_start.go:117`、`reconcile.go:998`）。#14305 がマージされるまで待ち、始める前に issue で「やります」と書く |
| F5 | `down -v` で消すボリュームを確かめる（roadmap [#665](https://github.com/docker/roadmap/issues/665)） | 弱い（返答は「`--volumes` がある」だけ） | 低 | S | TTY のときだけ確認、`-y` で飛ばす案。挙動が変わるので、#13554 のように却下される見込みが高い。issue で温度を確かめるところまで |
| F6 | `ls` / `images` の `--format` を Go テンプレートにする（[#12948](https://github.com/docker/compose/issues/12948)） | 中（ndeloof「docker/cli と同じやり方で」） | 低 | S | #12949、#13952 は stale で閉じた。issue で、今もその方向でよいか確かめる |

### 大きいが、今は手を出さないもの

| 候補 | 👍 | 理由 |
| ---- | -: | ---- |
| bind mount の所有者や権限を指定する（roadmap [#398](https://github.com/docker/roadmap/issues/398)） | 43 | Engine 側の機能（idmapped mount）が要る。Docker の担当者の返事は無い |
| リモートの Docker で `file:` の configs / secrets を使う（[#11867](https://github.com/docker/compose/issues/11867)） | 30 | [次点](#次点先に-issue-で合意を取ってから)と同じ。設計の提案から（L） |
| `up --wait` 中にログを流す（[#9122](https://github.com/docker/compose/issues/9122)） | 23 | [次点](#次点先に-issue-で合意を取ってから)と同じ。衝突の小さい部分は [N14](#n14-unhealthy-で失敗したときに-healthcheck-の出力を出す) |
| 定義済みの変数を compose.yaml の中で使う（compose-spec [#248](https://github.com/compose-spec/compose-spec/issues/248)） | 23 | 3 か所に変更が要るのに、メンテナの設計の返事が無い。UID/GID の変数は方針で却下（#7853） |
| bake ファイルを build の定義に使う（[#12581](https://github.com/docker/compose/issues/12581)） | 10 | [次点](#次点先に-issue-で合意を取ってから)と同じ。設計が決まっていない（L） |
| レプリカを 1 つずつ起動する（[#14232](https://github.com/docker/compose/issues/14232)） | 1 | ndeloof が `start_policy: sequential` の形を示した。起動は plan engine の中なので、エピックが片付いてから |
| スケジュール実行の jobs | — | glours の draft [#14155](https://github.com/docker/compose/pull/14155) が進行中 |
| イメージの署名の検証 | — | docker/cli が v29.9 で Content Trust を取り除き、ndeloof は「`docker pull` と同じ挙動に」として #13933 を not planned で閉じた |
| 外部の secrets ストア | — | #13821 は「docker/secrets-engine を使って」として not planned |
| `config` の出力の秘密情報を伏せる | — | [#14356](https://github.com/docker/compose/issues/14356) に ndeloof は「not convinced」。安全側は、新機能より既存の穴を塞ぐほうが筋が良い（N15、次点のリモートの compose の平文表示など） |

## 非公開で報告するもの

### S1. セキュリティに関わる発見

- 1 件ある。**このファイルは公開リポジトリ（自分の fork）に置かれるので、内容はここに書かない**。
- CONTRIBUTING.md の「Reporting security issues」に従い、security@docker.com に非公開で報告する。
- 修正の規模は XS だが、報告への返答で方針が決まるまで、修正の PR も出さない。

## 次点（先に issue で合意を取ってから）

| 症状 | 場所 | なぜ待つか |
| ---- | ---- | ---------- |
| post_start フック付きのサービスで、`run` の起動が失敗すると永久に止まる（Ctrl-C も効かない） | `pkg/compose/run.go:79` の `<-hookErrCh`。起動に失敗すると start イベントが来ない。直し方は約 10 行 | コードを読んだだけで、実機で未確認。`run.go` は #14155 も触る |
| git のリモート include の取得に失敗すると、壊れたキャッシュが残り続ける | `pkg/remote/git.go` の `checkout()`。OCI 版は失敗時に `RemoveAll` している。約 8 行 | #14268 と ndeloof の #14085 が同じファイルを触る |
| classic build のイメージに compose ラベルが付かない（エピック #14074 F） | `getImageBuildLabels` は `build_bake.go` からしか呼ばれない | ndeloof は「ラベルを増やしたくない」（#14312、#14050）。方針を聞いてから |
| `pull_refresh_after` が無視されているのに、設定ハッシュに含まれている | `hash.go` | 除外・実装・警告のどれにするか（ndeloof が #14253 で「別の変更で判断すべき」） |
| `wait --down-project` が、最初のコンテナでなく全部が止まるまで待つ。終了コードも実行ごとに揺れる | `pkg/compose/wait.go:70`（`eg.Wait()`、`statusCode` のデータ競合を `-race` で確認） | 挙動が変わるので合意が要る |
| `.env` の `COMPOSE_PROGRESS`・`COMPOSE_STATUS_STDOUT` が効かない。サブディレクトリから実行すると、`.env` の `COMPOSE_*` が反映されない | `cmd/compose/compose.go:234` がフラグの定義時に `os.Getenv` を読む。`setEnvWithDotEnv` が親をさかのぼらない | 同じ種類を直す ndeloof の #14139（止まっている）のマージ後に、その仕組みに乗せる |
| `push --ignore-missing` で全部スキップされても何も分からない | `push.go:69-71` | 警告か集計かは設計判断（ndeloof が #14341 の承認時に提案）。#14336 が `push.go` を触る |
| [#9122](https://github.com/docker/compose/issues/9122) `up --wait` 中にログを流す（👍23） | — | ndeloof の [#14306](https://github.com/docker/compose/pull/14306) が逆向きに固めようとしていて、#14305 が書き換え中。衝突の小さい部分は [N14](#n14-unhealthy-で失敗したときに-healthcheck-の出力を出す) |
| [#10596](https://github.com/docker/compose/issues/10596) `up --wait` が、一回きりのコンテナの exit 0 で失敗する（👍29） | — | フラグを足す PR は 2 回却下。方向は init container の正式化（`pre_start`、jobs、F3）。出すならエラーに案内を足す程度（XS） |
| [#11867](https://github.com/docker/compose/issues/11867) / [#9648](https://github.com/docker/compose/issues/9648) リモートの Docker で `file:` の configs / secrets が使えない（👍 約 60） | — | 設計で止まっている（実装の PR 3 本が stale で閉じた）。出すなら設計提案のコメントから（L） |
| [#12581](https://github.com/docker/compose/issues/12581) build で bake の機能を使いたい（👍10） | — | glours「後方互換なら歓迎」だが、設計の合意が無い（L） |
| [#13757](https://github.com/docker/compose/issues/13757) pull のサイズ表示が、ほぼ最後まで出ない | `cmd/display/tty_layout.go:155-171` の `aggregateProgress`（total=0 の層が 1 つでもあると全部消す） | 表示の変え方（分かっている分だけ出す）を issue で決める。S、触る open PR なし |
| Ctrl-C / SIGTERM を 1 回送っただけで、まれに即座に kill される（112 回中 3 回、v2.24.0 から） | 同じシグナルを `cmd/compose/compose.go:112-115` と `pkg/compose/up.go:187` の 2 か所で受ける | #14305 と衝突（中） |
| `down -v --remove-orphans` が、orphan のコンテナの匿名ボリュームを残す | `down.go:120` が `volumes` に `false` を直書き（XS） | #13465 が「明示的な指示無しに消さない」で not planned。`-v` は明示なので筋は通るが、確認してから |
| 作り直しのたびに、匿名ボリュームが漏れる（`up -V`、scale down、`VOLUME` の変更） | 匿名マウントにプロジェクトのラベルが無いので、`down -v` で回収できない（M） | この漏れのおかげで #14357 のデータを取り戻せたので、単純に消すのは危険。優先度は低い |
| リモートの compose（OCI / git）が、参照する環境変数の値を平文で表示する | `cmd/compose/options.go:148-176`、`:232-252` | #13900 が入ると CI のログに毎回出るので、それより先に（XS〜S） |
| `cache_from` / `cache_to` の `type=local` の相対パスが、実行したディレクトリ基準になる | compose-go の `paths/resolve.go` | compose-go 側の PR（S） |
| 同じ名前のディレクトリにある別のプロジェクトが衝突する（#6898 など既知） | `working_dir` のラベルが違えば警告する（S） | 優先度は低い |
| watch が、シンボリックリンクの先のディレクトリをたどって同期する | `internal/sync/tar.go:276-295` | 新しい境界は越えないので、優先度は低い |
| `ps --format "table {{.Health}}"` のヘッダーが `<no value>`。`images --format json` が空で `null`、順序も揺れる | `cmd/formatter/container.go:119-134`、`cmd/compose/images.go:108` | 小さいので、まとめて出す |
| `-tags fsnotify` が Linux/Windows でコンパイルできない（開発者だけ） | `pkg/watch/watcher_naive.go`（1 行） | 影響が小さい |
| `docker compose events` がコンテナのイベントしか流さない（[todo.md](./todo.md)） | `internal/inspect/events.go:41` | 公開 API `api.Event` の設計変更が要る。#13700 は stale で閉じた |
| `stop` を 2 回実行すると、何もしていないのに「Stopped」と出る（[todo.md](./todo.md)） | `down.go` の `stopContainer` | 挙動の変更なので合意が要る |

## 保留（メンテナの方針と合わなかったもの）

### `start -p` で任意の依存サービスが必須扱いになる

- **症状**: `required: false` の依存サービスが失敗していると、ファイルを読む `start` は警告を出して起動するのに、`-p <name> start` は失敗する（実機で再現済み）。
- **原因**: `-p` だけのときは、ラベルからプロジェクトを組み立て直す（`projectFromName`）。ラベル `com.docker.compose.depends_on` には `required` が無く、読み戻すときに `true` と決め打ちしている。`down.go` は補正しているが、`start` はしていない。
- **経緯**: issue [#14311](https://github.com/docker/compose/issues/14311) と、ラベルに `required` を足す PR #14312 を出したが、ndeloof は「ラベルに設定を増やしたくない」。PR は閉じた。issue は既知の制限として open のまま（10-29）。
- **今後**: `start` でも `down.go` と同じく `Required = false` に補正する案はありうる。ただし必須の依存の失敗まで警告に変わるので、メンテナの合意なしには出さない。

## 見守るもの（放置されているが、本物のバグを直している外部 PR）

他人の PR を乗っ取るのは避ける。作者の反応が長く途絶え、メンテナが求めたときだけ「引き継いでよいか」と聞く。

| PR | 内容 | 状態（10-16） |
| -- | ---- | ------------- |
| [#14141](https://github.com/docker/compose/pull/14141) | `EventProcessor.Done` の success が逆（`pkg/compose/progress.go` の `bus.Done(operation, err != nil)`） | 08-27 から誰もレビューしていない |
| [#14267](https://github.com/docker/compose/pull/14267) | `volumes_from: [svc:ro]` の `:ro` が落ちる（`resolveVolumeFrom`） | レビュー無し |
| [#14268](https://github.com/docker/compose/pull/14268) | git remote の ref が `feature/main` に誤って解決される | レビュー無し |
| [#14217](https://github.com/docker/compose/pull/14217) | watch のループで Dockerfile を除外していない | 09-23 に CHANGES_REQUESTED、作者は 09-26 から反応が無い |

## やらないもの（理由つき）

| 対象 | 理由 |
| ---- | ---- |
| plan engine（`reconcile*.go`, `start.go`, `up.go`, `scale.go`, `watch.go`, executor）と、そのテスト | #14081 でメンテナが書き換え中（Lot 2 は v5.7.0 に入った。#14305 など残りと Lot 3 が続く） |
| provider / relay まわり（11 を除く） | 10-15 に `internal/` へ移ったばかりで、ndeloof が毎日のように変更している |
| サービス選択（`cmd/compose/up.go`、`scale.go`、watch の rebuild） | ndeloof のエピック [#14346](https://github.com/docker/compose/issues/14346) |
| `create.go` | glours の #14342、ndeloof の #14340（umask）、#14299 |
| jobs（`cmd/compose/*` と、`down.go`、`run.go` などを広く触る） | glours の draft [#14155](https://github.com/docker/compose/pull/14155) |
| [#11816](https://github.com/docker/compose/issues/11816) `!reset` が効かない | compose-go 側で ndeloof の compose-go#913 が進んでいる |
| [#13672](https://github.com/docker/compose/issues/13672), [#13809](https://github.com/docker/compose/issues/13809) | すでに PR がある（#14198、[#13849](https://github.com/docker/compose/pull/13849)、#13811） |
| [#14283](https://github.com/docker/compose/issues/14283), [#14288](https://github.com/docker/compose/issues/14288) | ndeloof が #14299 で修正中 |
| [#14204](https://github.com/docker/compose/issues/14204) フック出力が `-d` で消える | moby 側の機能（moby#53407、未マージ）待ち |
| [#13602](https://github.com/docker/compose/issues/13602), [#13348](https://github.com/docker/compose/issues/13348) | main で再現しない |
| executor の multi-hop のテスト | ndeloof が #14081 に「Lot 1/2 の次の PR で拾う」と書いている |
| 使われていない公開 API の削除、backend の作り方や `ProjectOptions` の整理（エピック #14074 B/E/F の一部） | ユーザーへの影響が無い。メンテナが自分で進めている |
| provider イメージ配布の `SetLimit(0)` で止まる件 | CLI からは起きない（SDK で明示的に 0 を渡したときだけ） |
| 一度 unhealthy になったら、回復を待たずにすぐ失敗する（`--wait`、`service_healthy`） | 設計どおり。ndeloof が「Compose はオーケストレーターではない」として 2 回クローズ（[#11030](https://github.com/docker/compose/issues/11030)、[#12424](https://github.com/docker/compose/issues/12424)） |
| `--abort-on-container-exit` が、restart ポリシーで再起動中のコンテナの終了でも止まる | 設計どおり（10-20 に実機で確認。v5.1.4 でも同じ。ndeloof も #10310 で同じ立場） |
| 不安定なテストへの一律の再試行（`gotestsum --rerun-fails` など） | 9〜10 月の不安定なテストの上位 8 本のうち 4 本は、製品側の本物の競合だった。再試行はそれを隠す |
| CI の高速化 | 中央値 10.8 分でキューの待ちも無い。時間を食っているのは再実行（16%） |
| CLI の起動時間と、compose ファイルの読み込みの速さ | 実測で問題なし（`version` 22 ms、空の `ps` 約 30 ms、1000 サービスの読み込みで 0.35 秒） |
| コンテナの起動のロック（`startMx`）を単純に外す | ssh の Docker で接続が増えて失敗する。外すなら [N13](#n13-ssh-の-docker-で接続が多すぎて-stop-や-down-が失敗する) の上限と一緒に、メンテナの判断で |
| 要望は大きいが、方針で却下されているもの | [#4249](https://github.com/docker/compose/issues/4249) 環境変数で配列、[#7853](https://github.com/docker/compose/issues/7853) UID/GID や OS の変数、#13554 `--progress quiet` で完全に無音、#13465 `down` で孤立ボリュームを消す、#12497 include の glob、#13993 profile だけを `down`、#13716 health の待機中に detach、#11003 watch で compose.yaml の変更を反映、#14326 exec だけの watch、#13807 ボリュームの自動確認のラベル（代わりに #14273 の `x-disposable`） |

## 完了

### マージされた自分の PR

| 日付 | issue / PR | 内容 | 記録 |
| ---- | ---------- | ---- | ---- |
| 10-27 | [#14384](https://github.com/docker/compose/issues/14384) / [#14385](https://github.com/docker/compose/pull/14385) | 依存の待ちが必ず 500 ミリ秒かかる（N3） | [pr-14385.md](../../pr/pr-14385.md) |
| 10-27 | [#14394](https://github.com/docker/compose/issues/14394) / [#14395](https://github.com/docker/compose/pull/14395) | `up -t` が Ctrl-C と `--abort-on-container-exit` で効かない（N4） | [pr-14395.md](../../pr/pr-14395.md) |
| 10-26 | [#14373](https://github.com/docker/compose/issues/14373) / [#14374](https://github.com/docker/compose/pull/14374) | `down -v` が他のプロジェクトのボリュームを消す（N2） | [pr-14374.md](../../pr/pr-14374.md) |
| 10-22 | [#14351](https://github.com/docker/compose/pull/14351) | CI の unit テストで毎回 1 件の偽の FAIL が出る（1） | [pr-14351.md](../../pr/pr-14351.md) |
| 10-20 | [#14357](https://github.com/docker/compose/issues/14357) / [#14358](https://github.com/docker/compose/pull/14358) | `run` が依存サービスを作り直すと匿名ボリュームを引き継がない（2） | [pr-14358.md](../../pr/pr-14358.md) |
| 10-20 | [#14362](https://github.com/docker/compose/issues/14362) / [#14363](https://github.com/docker/compose/pull/14363) | watch の監視中に `.git/` が除外されない（3） | [pr-14363.md](../../pr/pr-14363.md) |

### 他の人の変更で片付いた候補

| 対象 | どう片付いたか |
| ---- | -------------- |
| 9 `events --until` が必ず EOF エラーで終わる | [#14444](https://github.com/docker/compose/pull/14444) で直った（11-03） |
| [#14313](https://github.com/docker/compose/issues/14313) bake の並列数の上限 | maxproske の [#14352](https://github.com/docker/compose/pull/14352)（10-28、v5.7.0） |
| `swarmEnabled` のパッケージ全体のキャッシュ | ndeloof の [#14323](https://github.com/docker/compose/pull/14323) でインスタンス単位になった |
| [#7431](https://github.com/docker/compose/issues/7431) `push` で存在しないイメージを飛ばす | glours の [#14341](https://github.com/docker/compose/pull/14341) で `--ignore-missing` が入った |

### 終わった調査のタスク

どれも 11-05。そこから引き出した教訓は [investigation.md](./investigation.md#調べるときの注意) に移した。

| ID | 内容 | 結果 |
| -- | ---- | ---- |
| T1 | v5.6.1 と v5.7.0 で、e2e の全 fixture の `config --hash '*'` を比べる | 差は `umask` と pre_start + env_file（#14293）の 2 つで、どちらも意図どおり |
| T2 | v5.6.1 で作った環境を v5.7.0 で操作する（アップグレード） | 作り直しは意図どおりの 2 種類だけ（#14293、#14402）。データの消失なし |
| T3 | 前景の `up` の 12 シナリオを、並列 4 本 × 10 回 | 480 回で、固まる・終了コードの揺れ・リソースの残りは 0 件 |
| T3b | 依存の鎖を詰め込んだ 1 プロジェクトの 5 シナリオを、並列 4 本 × 10 回（約 130 回で打ち切り） | 固まる回は 0 件。Ctrl-C で気になる結果が 2 つ → T3c、T3d |

---

## 候補の詳細

番号の N1〜N17 は 10-20 の第 3 回の調査、1〜13 はそれより前の調査で付けたもの（片付いた番号は欠番）。PR を出したら、詳細は `compose/pr/` に移す。

### N6. 許可リスト型の dockerignore で watch が無視のルールを全部捨てる

- **状態**: issue [#14447](https://github.com/docker/compose/issues/14447) を出した（status/approved）。自分では実装しない（[進行中](#進行中自分の-pr-と-issue)）。
- **症状**: `.dockerignore` を「`*` で全部除外し、`!src` などで必要なものだけ許す」形で書くと、watch は除外のルールを 1 つも当てはめない。
  - ビルドコンテキストの外のファイルや、`node_modules`・`.env`・`dist` の変更でも再ビルドが走る。
  - watch の対象の中に書き込みがあると（bind mount したデータ、`cache_to` のローカルキャッシュなど）、再ビルドが止まらなくなる（45 秒で 23〜42 回。拒否リスト型なら 1 回）。
- **原因（自分でも確認）**: `pkg/watch/dockerignore.go:136-143` の `NewDockerPatternMatcher` が、パターンに `*` があると、`!` 以外を全部捨てる。
  #12236 の対策として #12239（2024-10）で入った。#12236 の本当の原因は、初回の同期がファイル名だけで照合していること（`watch.go:791`、`:814`）。
- **直し方**: 特例を消し、初回の同期をフルパスで照合する（Dockerfile と compose ファイル名の固定パターンは名前のまま）。S〜M。#12236 の回帰テストも付ける。

### N8. plan engine で止めるコンテナに pre_stop が走らない

- **種類**: 退行（v5.2.0 から。v5.7.0 も同じ）。

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
- **最初の一手**: issue で、退行であること（上の表）と、作り直しでも走らせるかを聞く。

### N9. 同じ依存先を必須と任意の両方から待つと任意側も起動されない

- **種類**: 退行（v5.7.0 に入った）。
- **症状**: 同じ依存先を、あるサービスは `required: false`、別のサービスは `required: true` で待っていて、依存先が失敗した場合:
  - v5.6.0: 任意側は警告を出して起動する。
  - v5.7.0: 任意側も起動されず `created` のまま残る。警告も出ない。
- **再現（調査、実機）**: dep（必ず unhealthy）、opt（`required: false`）、req（`required: true`）。`service_healthy` で 3/3 回、`service_completed_successfully` で 2/2 回。
- **原因**: `reconcile.go:1357-1386` の `waitConditionNode` が、待ちのノードを（依存先、条件）ごとに 1 つにまとめ、必須の依存元が 1 つでもあると全体を必須にする。待ちが失敗すると、依存元がすべて飛ばされる（`executor.go:356-363`）。
  エピック #14081 の本文は「`required: false` は局所的に吸収する」と書いている。
- **保留の理由（11-05）**: `waitConditionNode` のコメントが、この動きを意図どおりと書いている。N1 で ndeloof が示した方針（失敗したらフェーズ全体を止め、報告の仕方を直す）とも合う。この構成で、起動されなかったサービスが #14376 の形で報告されるかは未確認。
- **直し方（出すなら）**: まとめる単位に `required` を加える。S。

### N10. scale も匿名ボリュームを引き継がない

- **種類**: データ消失系（v5.1.4 からずっと）。`run` の #14357 / #14358 と同じバグ。
- **症状**: 依存先の設定を変えてから `docker compose scale web=2` を実行すると、依存先が新しい匿名ボリュームで作り直される。`up -d` なら引き継ぐ。
- **原因（自分でも確認）**: `scale.go:38` が `api.CreateOptions{Services: services}` だけを渡し、`Inherit` が false のまま。
- **直し方**: `Inherit: true` を 1 行（XS）。実機で直ることを確認（調査）。
- **衝突**: 低〜中（エピック #14346 が ScaleOptions を作り直す予定）。
- **最初の一手**: #14357 に「`scale` でも同じ」とコメントし、別の PR にするかをメンテナに聞く。

### N11. pre_stop のあるサービスで restart と stop が壊れる

- **症状**:
  - (1) 止まっているサービスに `restart` すると、`container ... is not running` で失敗して起動しない（v2.40.3 / v5.0.0 から）。
  - (2) 一時停止中（paused）のサービスに `stop` すると、終了コード 0 なのに、paused のまま止まらない（v2.38.0 から）。
- **原因（自分でも確認）**:
  - (1) `restart.go:115-121` が、状態を見ずにフックを実行し、失敗すると return する。
  - (2) `down.go:439-441` が、Conflict を受けると `return nil` で抜け、停止そのものをしない。
- **直し方**: どちらも「running のときだけフックを走らせる」（XS）。実機で直ることを確認（調査）。

### N12. ps がコンテナごとに inspect する

- **症状**: `ps` は、Health と終了コードを得るためだけに、コンテナごとに inspect を呼ぶ（`-q` や `--services` でも）。20 コンテナで API 呼び出しが 22 回、接続が 20 本。
- **原因（自分でも確認）**: `internal/inspect/ps.go:45-64`。API 1.52 以降は、コンテナの一覧に Health が含まれる。
- **直し方**: 一覧の Health を使い、終了したコンテナだけ inspect する。API 1.52 未満では今のまま。S（約 20 行）。
- **数字（調査、実測）**: 22 → 2 回。ssh で 50 コンテナ 1.5 秒 → 0.28 秒。JSON の出力は同じ。
- **同じ種類で、もっと小さいもの**: `images` が同じイメージを何度も inspect する（20 サービスで 40 回 → 1 回にできる、XS）。`up` などで `/_ping` が 1 回多い（XS）。

### N13. ssh の Docker で接続が多すぎて stop や down が失敗する

- **症状**: ssh では、API の接続ごとに ssh のセッションを張る。並列数は既定で無制限なので、50 サービスで `stop` / `down` が `ssh … dial-stdio exited with status 255` で失敗した（sshd の MaxStartups の既定 10 を超える）。
- **直し方の案**: ssh の endpoint で、並列数の指定が無いときの既定値を約 8 にする（`cmd/compose/compose.go:520-527` 付近）。`--wait` の最後の確認も同じ上限に入れる。S。既定値を変えるので合意が要る。
- **数字（調査、実測）**: 上限 8 で失敗がゼロ。読み取り系のコマンドは 50〜80% 速くなる。
- **既存**: #8191、#8856、#9448、#9870、#10117、#11677（どれも stale / not planned）。ndeloof は #10117 で「compose での短期的な修正は無い」。既定の上限という案は、まだ出ていない。
- **関連（plan engine の中心なので、issue に数字を書くだけ）**:
  - コンテナの起動が、1 本のロック（`startMx`、`service_containers.go:424-435`）で直列になっている。50 サービスの `start` が約 12 秒 → 約 3.5 秒になる。ただし ssh の上限と一緒でないと失敗が増える。
  - コンテナの作成が `depends_on` の段ごとに待つ（`reconcile.go` の `infrastructureDeps`）。4 段の鎖で、作成に 0.8 秒（1 段なら 0.2 秒）。

### N14. unhealthy で失敗したときに healthcheck の出力を出す

- **症状**: `--wait` や `service_healthy` が unhealthy で失敗すると、`container X is unhealthy` としか出ない。Docker が持っている healthcheck の出力は捨てている。
- **原因（自分でも確認）**: `service_containers.go:578-612` の `isServiceHealthy` は、inspect の結果を持っているのに、`State.Health.Log` を使わない。
- **直し方**: 最後の probe の終了コードと出力を、短く切ってエラーに付ける。追加の API 呼び出しは要らない。S。
- **需要**: #9122（👍23）と #10596（👍29）の議論、外部の記事（Zenn「docker compose up --wait の落とし穴」）。

### N15. publish の機密データの検査が push する中身を見ていない

- **症状**: `docker compose publish` は、読み込んだモデルを YAML に戻したものを検査するが、push するのは元のファイルそのもの。
  コメントに書いたパスワードや、`command`・healthcheck・`x-*` の値が、検査をすり抜けて公開される（調査、ローカルの registry。5.0.1 も同じ）。確認の文面には、検出した値が平文で出る。
- **原因**: `publish.go:761-777` と `:860-870` が `MarshalYAML()` の結果を検査し、`:261-262` が `os.ReadFile` の中身を公開する。
- **直し方**: push する中身そのものを検査し、確認では値を伏せる。S。公開の issue でよい。

### N16. watch の sync がボリュームの中身を消す

- **症状**: sync が `./src → /app`、名前付きボリュームが `/app/uploads` の構成で、ホストの `src/uploads` を消すと、コンテナで `rm -rf /app/uploads` が走り、ボリュームの中身が消える（警告が 1 行出るだけ）。
- **原因**: `internal/sync/tar.go:76-92` が、ホストに無いパスを無条件に消す。`watch.go:148-186` は、それがボリュームのマウント先かを確かめない。
- **直し方**: コンテナ側のパスがボリュームや tmpfs のマウント先（とその下）なら、削除を飛ばして警告する。S。
- **衝突**: 中（`watch.go`）。

### N17. CI の基盤の小さな改善

8 週間分（857 run、536 ジョブ）の集計から:

- (a) Dockerfile の `go mod download` に再試行を入れる（XS）。proxy.golang.org の HTTP/2 エラーで、週に約 2 回の手動の再実行が出ている。
- (b) **main の Codecov へのアップロードが、2026-01-22 から毎回失敗している**（[run 38929564017](https://github.com/docker/compose/actions/runs/38929564017) に「Token required because branch is protected」。ジョブは green なので誰も気付いていない）。
  Codecov 上の main は 01-21 のまま 53%、実際は約 71%。直すには token か OIDC の設定（メンテナの作業）が要るので、issue で知らせる。`pkg/mocks` を ignore に足す提案も一緒に。
- (c) standalone の e2e で、`GOCOVERDIR` の警告が毎回 stderr に出る（失敗メッセージの 1 行目がこの警告になる）。XS〜S。
- (d) main の push の 22% が、次の push で cancel されて CI の結果を持たない（`cancel-in-progress: true`）。PR だけ cancel する案（XS。方針なので聞く）。
- (e) unit テストが root で動くので、権限のテスト 2 本（`TestSync_StatPermissionError`、`TestWatchRecursivelySkipsUnreadableDir`）が常に skip されている。

### 4. bake のビルドが curl の書式指定などで失敗する

- **症状**: `dockerfile_inline` に `RUN curl -w '%{http_code}' ...` と書くと、bake 経由のビルドが `Invalid template control keyword` で失敗する。classic builder なら通る。ラベルの値の `${`、`build: "./dir%{x}"` も同じ（調査、実機）。
- **原因**: bake は、渡された JSON の**すべての文字列**を HCL のテンプレートとして評価する。ところが [pkg/compose/build_bake.go](https://github.com/docker/compose/blob/7b36d08dd0d30da3abc1e646baec8751f0bcb0b2/pkg/compose/build_bake.go#L252) は、args（252 行）と dockerfile-inline（290 行）で `${` だけをエスケープし、Labels（292 行）や他の文字列はエスケープしていない。
- **直し方**: `${` → `$${` と `%{` → `%%{` をまとめて行う関数を作り、context・contexts・dockerfile・tags・secret・ssh・cache などすべての文字列に当てる。S。
- **後押し**: ndeloof が過去に `${` を同じ形で直している（0c5bd16da、a03f2562d）。今回はその直し漏れ。
- **衝突**: 低（`build_bake.go` を触る open PR は無い）。

### 6. e2e `TestWatchSyncIntoSymlinkedDirectory` が不安定

- **症状**: 8 週間分の CI の集計（857 run）で、**残っている唯一の未修正の不安定なテスト**。`cat: can't open '/var/sub/hello.txt'` で 60 秒タイムアウトし、`cannot exec in a stopped state` も出る（main の [run 38226095992](https://github.com/docker/compose/actions/runs/38226095992)、PR の [run 37741830070](https://github.com/docker/compose/actions/runs/37741830070)）。
- **原因（コードと CI のログからの推定）**:
  - `up --watch` では、`watcher.Start` が `s.start` より先に走る（[pkg/compose/up.go](https://github.com/docker/compose/blob/7b36d08dd0d30da3abc1e646baec8751f0bcb0b2/pkg/compose/up.go#L241) の 241 行と 310 行）。だから「Watch enabled」はコンテナの起動前に出る。
  - テストは「Watch enabled」を待ってすぐ `os.Rename` する（[pkg/e2e/watch_test.go](../../pkg/e2e/watch_test.go) の 491-497 行付近）。コンテナが `ln -s /var/sub /app/data/sub` を作る前に、sync が exec しようとして失敗し、その変更は再試行されずに捨てられる。
  - 同じファイルの 415 行のコメント「"Watch enabled" is logged AFTER the initial up」も実態と合っていない。
- **直し方**: rename の前に、`exec app test -L /app/data/sub` が成功するまで poll する（状態で待つ。[SCENARIO.md](../../pkg/e2e/SCENARIO.md) の方針どおり）。誤っているコメントも直す。
- **注意**: PR では、上の 2 つの run を引用して「なぜ失敗するか」を示す。製品側にも「起動直後の変更が黙って失われる」余地がある（推定）が、テストの PR では気付いたこととして書き添える程度にする。

### 7. stats のテストの決定化と取得失敗の debug ログ

- **出典**: ndeloof が [#14337](https://github.com/docker/compose/pull/14337) を承認したときのレビュー（10-15）。承認の 30 秒後にマージされ、どちらも手付かず。
  「`sampleStats` errors are dropped silently… A `logrus.Debugf` would help diagnose it. / `TestCollectContainerStatsBoundsConcurrency` asserts `peak == statsConcurrency`, which depends on scheduling… Having each call wait until `statsConcurrency` are in flight (with a timeout) would make it deterministic.」
- **場所**: [cmd/compose/stats_projects_test.go](https://github.com/docker/compose/blob/7b36d08dd0d30da3abc1e646baec8751f0bcb0b2/cmd/compose/stats_projects_test.go#L184-L208)（20ms の sleep に頼っている）、[cmd/compose/stats_projects.go](https://github.com/docker/compose/blob/7b36d08dd0d30da3abc1e646baec8751f0bcb0b2/cmd/compose/stats_projects.go#L197)（エラーを捨てている）
- **直し方**: テストに、呼び出しが `statsConcurrency` 本そろうまで待つバリア（タイムアウト付き）を入れる。失敗したサンプルは `logrus.Debugf` で出す。
- **ついでに見つかったもの**: 一時停止中のコンテナの使用量が 0 になる（`stats_projects.go:191` が running のときしか取得しない。1 行）。同じ PR にするかは相談。
- **衝突**: 低。作者（glours）が自分でやる可能性はあるので、出す前に `stats_projects` を触る PR を確かめる。

### 10. classic build の `--push` がプロジェクト全体をサービス数だけ push する

- **出典**: エピック #14074 F「classic `--push` pushes the whole project once per built service」（🐛 印付き、「standalone issue に切り出す価値あり」）
- **場所**: [pkg/compose/build_classic.go](../../pkg/compose/build_classic.go) のサービスごとのコールバック内（103-104 行）の `if options.Push { return s.push(ctx, project, api.PushOptions{}) }`。`push()` は project の全サービスを対象にする。
- **影響**: classic builder（`DOCKER_BUILDKIT=0`、または buildx が無い環境）のときだけ。N 個のサービスで N×N 回 push する。`build --push web` で、ビルドしていない `db` の古いイメージまで push する。bake の経路は問題ない。
- **直し方**: 走査の後に 1 回だけ、ビルドしたサービスに絞って push する。classic build には単体テストが無いので、モックのテストを書くのが主な作業。
- **衝突**: Benehiko の draft [#14336](https://github.com/docker/compose/pull/14336) が `push.go` の別の箇所を触っている。

### 11. provider のメタデータ取得が呼び出し元の ctx を使わない

- **出典**: ndeloof の [#14333 のインラインレビュー](https://github.com/docker/compose/pull/14333#discussion_r4257871410)「Propagating the caller's context is a fair improvement and a behavior change, so it is left to a follow-up.」
- **場所**: [internal/provider/plugins.go](https://github.com/docker/compose/blob/7b36d08dd0d30da3abc1e646baec8751f0bcb0b2/internal/provider/plugins.go#L615-L617) の `getPluginMetadata`。`exec.Command` と `context.Background()` を使っているので、metadata コマンドが固まる provider だと Ctrl-C が効かない。
- **直し方**: 引数に `ctx` を足し、`exec.CommandContext(ctx, …)` と `PrepareShellOut(ctx, …)` に渡す。テストは `plugins_control_test.go` のヘルパープロセス方式で書ける。
- **衝突**: 中。provider は ndeloof が頻繁に触る。出す直前に open PR を確かめる。
- **同じレビューの、より小さい残作業**: `removeRelayLinkNetwork` が最初のエラーで止まる（`internal/relay/relay.go:374-378`。[r4256695116](https://github.com/docker/compose/pull/14333#discussion_r4256695116) で「`errors.Join` は follow-up で」）。XS だがインパクトが小さいので、出すなら別の PR。

### 12. `config -q` が一覧系フラグと組み合わせると出力する

- **症状（実機）**: `config -q --services` / `--volumes` / `--images` / `--hash='*'` / `--variables` は、`-q`（検証だけして何も出力しない）なのに一覧を出力する。
- **原因**: `-q` を `os.Stdout = devnull` という差し替えで実装している（[cmd/compose/config.go](../../cmd/compose/config.go) の PreRunE、103-109 行）。docker/cli は起動時に stdout を掴んでいるので、`dockerCli.Out()` 経由の出力には効かない。`build -q` も同じ実装。
- **衝突**: 中。ndeloof の [#14046](https://github.com/docker/compose/pull/14046)（止まっている）と [#14332](https://github.com/docker/compose/pull/14332) が `config.go` を触る。
- **注意**: 「`-q` と `--services` の同時指定はエラーにすべき」という判断もありうるので、issue で聞く。

### 13. `pkg/compose` のテストが Windows でコンパイルできない

- **出典**: ndeloof の [#14309 のインラインレビュー](https://github.com/docker/compose/pull/14309#discussion_r4220751191)「The package's tests still don't compile on Windows on main because of newStartTestService/serviceContainer from start_test.go, a pre-existing and unrelated issue.」
  Copilot と docker-agent も、#14303 と #14309 で毎回同じ指摘をしている。
- **原因**: [pkg/compose/start_test.go](../../pkg/compose/start_test.go) の 1 行目が `//go:build !windows` なのに、そこで定義したヘルパーを `down_test.go:1345/1351` と `push_test.go:88/147` が使う。
  - 壊れているのは、テストのある 26 パッケージ中 `pkg/compose` だけ。エラーは 4 件、原因のシンボルは 3 つ。windows/arm64 も同じ。
  - 本当に pty が要るのは `hook_test.go` と `cmd/compose/compose_progress_test.go` だけ。`start_test.go`・`restart_test.go`・`executor_start_test.go` は、`hook_test.go` の `writeStdcopyFrame` を使っているせいで、タグが連鎖して付いている。
- **最小の修正（調査で検証済み、+38/-23）**: `writeStdcopyFrame` をタグの無いファイルに移し、3 つのファイルからタグを外す。windows/amd64 と arm64 で vet が通り、darwin で `go test ./pkg/compose` と lint が通る。Windows で実際にテストが通るかは未確認。前例として外部の [#13683](https://github.com/docker/compose/pull/13683) がマージされている。
- **CI に足すもの**: validate の matrix に `GOOS=windows go vet -tags "$BUILD_TAGS" ./...` のチェックを足す（約 20 行、1 ジョブ約 1 分）。golangci-lint でなく vet なのは、Windows では `hook_test.go` が除外されて `unparam` の警告が余分に出るため。
- **背景**: Windows と macOS でテストを実行していた merge.yml のジョブは、2026-04-15（9fd9dc7ca）に消された。それ以降、Linux 以外でテストを動かす場所が無い。
- **衝突**: 中。#14302、#14305 / #14308（draft）が同じファイルを触るが、変えるのは 1 行目のタグとヘルパーの場所だけ。出す前に `GOOS=windows go vet ./pkg/compose/` で今も壊れていることを確かめる。
