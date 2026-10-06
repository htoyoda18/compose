# cmd/display メモ

`docker compose`の進捗表示(TTY/Plain/JSON/Quiet)を担うパッケージ。
`pkg/compose`側は`api.EventProcessor`インターフェース(`Start`/`On`/`Done`)経由で
イベントを流すだけで、実際にどう画面に描画するかはこのパッケージが一元管理する。
どのモードを使うかは`cmd/compose/compose.go`の`selectEventProcessor`が
`--progress`/`--ansi`/TTY判定から決定し、`display.Mode`(グローバル変数)に反映する。

## ファイル一覧・責務

- mode.go
  - `Mode`(グローバル変数)と`ModeAuto`/`ModeTTY`/`ModePlain`/`ModeQuiet`/`ModeJSON`の定数定義
  - コマンド開始時に`selectEventProcessor`が一度だけ解決し、以降のコードは`ModeAuto`を見ることはない
  - 例外的に`run`/`build`の`--quiet`だけが PreRun から`ModeQuiet`を上書きする
- colors.go
  - `DoneColor`/`WarningColor`/`ErrorColor`等の色付け関数(package-level var)を定義し、
    `NoColor()`で全てを恒等関数に差し替えることでカラー出力を一括無効化する
  - 実体は`aec`のANSIシーケンス適用関数(`colorFunc = func(string) string`)
- dryrun.go
  - `DRYRUN_PREFIX`(` DRY-RUN MODE - `)の定数のみを持つ小さなファイル
- quiet.go
  - `Quiet`が返す`quiet`。全メソッドが no-op で、`run --quiet`/`build --quiet`時など
    出力を完全に抑制したい場合に使う
- plain.go
  - `Plain`が返す`plainWriter`。TTY機能を使わず、イベントをそのまま1行ずつ出力する
    (非対話環境やパイプ経由での実行時に使われる)
  - カーソル制御・再描画を一切行わないので CI やログ収集向け
- json.go
  - `JSON`が返す`jsonWriter`。各イベントを1行1JSONの`jsonMessage`として出力する
    機械可読モード(`docker compose --progress json ...`)
  - `jsonMessage`が出力スキーマそのもの(id/parent_id/status/text/current/total/percent等)
  - `Resource.StatusText()`で内部の`EventStatus`を文字列化して載せる
- tty.go(TTYレンダラーの調停役)
  - `Full()`が返す`termWriter`。自身は**ロックとライフサイクル管理に専念**し、
    描画の中身は model/layout/screen の3層に委譲する
  - `Start`で100msティックの再描画goroutineを起動し、`Done`または ctx キャンセルで停止。
    停止はすべて context 経由なのでチャネルハンドシェイク待ちでブロックしない
  - BuildKit が同じ端末に描いている間は`suspended`で沈黙し、終わったら
    新しいブロックを下に作り直す(上書きしない)
  - 非detached時の`Starting`/`Started`は、コンテナログを踏まないよう意図的に捨てる
  - オプション`WithDryRun()`で全行に`DRYRUN_PREFIX`を付ける口が用意されている
- tty_model.go
  - 画面と無関係な純粋データモデル。`node`(タスク1件の状態)と`taskTree`
  - `taskTree.apply(event, now)`が唯一の状態遷移で、時刻は外部から注入される
  - 親子リンク・ルート一覧・完了件数をインクリメンタルに維持(1フレームO(1))
  - 進捗値は`max`で単調化し、複数イメージで共有されるレイヤーは
    最初に現れた親(`anchor`)からの更新のみ受け付けてちらつきを防ぐ
- tty_layout.go
  - `(taskTree, operation, layoutOpts)` → 端末行の配列 を返す純粋関数群
  - 唯一の強い不変条件は「全行が端末幅以下」で、`renderSegs`が最終的に強制する
  - 幅計算は go-runewidth でセル単位(CJKは2セル)、色は幅が確定してから適用するので
    ANSIシーケンスが幅計算に混入しない
  - 狭いときの劣化順序が固定(サイズ→ID→ステータス→詳細)、高さ超過は`... N more`
  - スピナー・点字プログレスバー・経過タイマー・ステータス色もここで組み立てる
- tty_screen.go
  - 描画済みブロックの端末領域を所有し、前フレームとの差分再描画を行う
  - 変化のない行は改行だけで飛ばし、1フレームを1回の`Write`にまとめる
  - 幅が縮んだ/外部出力が混ざった場合は`reset()`で位置追跡を捨て、新規ブロック扱いにする
- json_test.go / tty_test.go
  - TTY側は偽クロック+バッファ端末で、幅・高さの不変条件、goroutineリーク、
    差分描画、スナップショットを検証している

## 確認できた問題点・改善点

重大度は「ユーザーに実害が出るか」、難易度は「修正の波及範囲と判断の重さ」で付けた。

| # | 問題点 | 重大度 | 難易度 | 一言 |
|---|--------|--------|--------|------|
| 1 | `dryRun`が全writerで死んでいる | **中** | 低 | 唯一のユーザー可視な機能欠落。修正は機械的 |
| 2 | `Start`が前サイクルのgoroutineを待たない | 低 | 中 | Compose本体からは到達不能。直し方に落とし穴あり |
| 3 | `rowSegs`が引数`row`を破壊 | 低 | **極低** | 現状バグ無しの潜在的な罠。1行 |
| 4 | json の`omitempty`で0値が落ちる | 中 | 中 | 実害は外部ツール側。出力互換性の判断が要る |
| 5 | json の marshal エラー握りつぶし | 低 | 極低 | 現状到達不能 |
| 6 | plain の出力スペース崩れ | 低 | 低 | 見た目のみ。e2eへの影響なしを確認済み |
| 7 | `DoneColor`未使用 | 極低 | 極低 | 削除するだけ |
| 8 | グローバル可変状態(`Mode`/色) | 低 | **高** | 設計課題。26箇所に波及 |
| 9 | `Starting`/`Started`のドロップ | 低 | 中 | 意図的な挙動。変えるなら e2e 確認が要る |
| 10 | テスト欠如 | 低 | 低 | `plain_test.go`を足すだけ |
| 11 | 命名・未使用引数 | 極低 | 極低 | 参照3箇所 |

着手するなら **3 → 5 → 7 → 11 → 10 → 6**(どれも局所・低リスク)、
次に **1**(PR #14053)、**4** は互換性の合意を取ってから、
**2 / 9 / 8** は挙動・設計に踏み込むので単独PRにすべき。

### `dryRun`が3つのwriter全てで常にfalseのまま(生きていない)

`DRYRUN_PREFIX`を出し分ける分岐は書かれているのに、その値を立てる経路が無い。

- TTY側には`display.WithDryRun()`というオプションが用意されている(`tty.go:60`)が、
  **どこからも呼ばれていない**(`grep`でヒットするのは定義のみ)
- `plainWriter.dryRun`/`jsonWriter.dryRun`はフィールドだけ残り、
  コンストラクタ`Plain(out)`/`JSON(out)`に設定する引数が無い
- 呼び出し元の`selectEventProcessor`(`cmd/compose/compose.go:714-748`)も
  `dryRun`相当の値を一切渡していない

つまり`docker compose --dry-run up`のようにdry-runモードで実行しても、
進捗表示には` DRY-RUN MODE - `プレフィックスが一切出ない。
→ [PR #14053](https://github.com/docker/compose/pull/14053)
(`fix(display): wire --dry-run flag into progress writers`)で対応中。
TTY側は引数追加ではなく`Full(..., display.WithDryRun())`というオプション渡しに
寄せるのが今のコードに素直な形。

**重大度: 中** — ここに挙げた中で唯一、ユーザーに直接見える機能欠落。
`--dry-run`の目的は「何が起きるはずだったか」の提示なので、
実行ログと見分けが付かないのは誤解を招きうる。ただし dry-run 自体は正しく働いており、
実リソースが触られるわけではないので「壊れる」類ではない。

**難易度: 低** — 変更は機械的。`Plain`/`JSON`にも`WithDryRun`相当を足し、
`selectEventProcessor`にdryRunを引き回すだけ(呼び出し元は`cmd/compose`内に閉じている)。
`json_test.go`と同形の単体テストで検証できる。

### `Start`が前サイクルの再描画goroutineを待たず、`Done`の保証が破れうる

`termWriter.Start`は`Done`を挟まずに再度呼ばれると前サイクルを`stopTicks()`で
キャンセルするが、**`ticksExited`を新しいチャネルで上書きして古い方を捨てている**
(`tty.go:102-119`)。`Done`は最新サイクル分しか`<-exited`で待たないため、
退役済みgoroutineが`Done`復帰後に最後の1フレームを描く余地が残る。

これは`Done`のコメント(`tty.go:150-153`)が主張する
「`Done`が返れば以後何も再描画しないので、直後の出力(プロンプトやコンテナログ)が
壊れることはない」という保証と食い違う。`Start`側のコメントは
「最後の1回の再描画は無害(同じモデル・ロック下)」と書いているが、
問題は内容ではなく**タイミング**(Done後に書き込まれること)なので、
古い`ticksExited`も保持して両方待つ形にした方が整合する。

**重大度: 低** — Compose本体からは到達しない。イベントの発火は
`pkg/compose/progress.go`の`Run()`に一本化されていて、`Start`→`pf`→`Done`を必ず対にする。
`Up`も内部では小文字の`s.create`/`s.start`(Runを呼ばない版)を使っており、
`Run`のネストは意図的に避けられている(`pkg/compose/up.go:46-54`)。
つまり現状 `Done`なしの再`Start`は起きず、ライブラリとして`Full()`を直接叩いた場合のみの話。
ただし`Full`は公開APIで、コード側もその誤用を想定して書かれている(コメント参照)ので、
「想定はしているのに保証しきれていない」状態ではある。

**難易度: 中** — 直し方に落とし穴がある。`refresh`goroutineは`w.mu`を取るため、
**`Start`のロック保持中に`<-exited`で待つと即デッドロックする**。
`ticksExited`をスライスにして`Done`側でまとめて待つか、
`Start`でロックを解いてから待ち合わせる必要があり、素直な1行修正にはならない。
`tty_test.go`の`goleak`+偽クロックの枠組みは使えるが、再現テストの作り込みは要る。

### `rowSegs`が「純粋」を謳いながら入力`row`を書き換えている

`tty_layout.go:236`で`r.id = truncateCells(r.id, idBudget)`と引数を破壊している。
現状は`buildRows`が毎フレーム`row`を作り直すため実害は出ていないが、
ファイル冒頭の「no mutation」というドキュメントに反しており、
同じ`row`に2回適用するとIDが二重に切り詰められる。ローカル変数に受けるべき。

**重大度: 低** — 今は`buildRows`が毎フレーム`row`を作り直し、
`rowSegs`も1フレーム1回しか呼ばれないので実害ゼロ。将来の踏み台としての危険のみ。

**難易度: 極低** — `id := truncateCells(r.id, idBudget)`とローカルに受けて
以降そちらを使うだけ。挙動は一切変わらない。

### json.goのスキーマが`omitempty`で0値を落とす

`Current`/`Total`/`Percent`に`omitempty`が付いているため、
「進捗0%」「ダウンロード開始直後(current=0)」といった正当な値がフィールドごと消える。
機械可読ストリームとして、受け手が「未報告」と「値が0」を区別できないのは扱いづらい。
また`Tail`は常に`false`固定で、出力されることがない事実上のデッドフィールド。

**重大度: 中** — `--progress json`は機械可読であることが存在理由で、
これはその契約の欠陥。ただし実害を被るのは Compose 自身ではなく下流のツールなので、
CLI を普通に使っている限り気付かない。

**難易度: 中** — コード上は`omitempty`を外す1行。
ただし**出力フォーマットの変更**になるため、既存の消費者への影響判断が要る。
無害寄りに倒すならポインタ型+`omitempty`で「未報告=null / 0=0」を表現する手もあるが、
それも形が変わることに変わりはない。`Tail`の削除も同様に互換性の話。
upstream に出すなら事前に issue で意図を確認した方がよい種類の変更。

### json.goがmarshalエラーを握りつぶす

```go
marshal, err := json.Marshal(message)
if err == nil {
    _, _ = fmt.Fprintln(w.out, string(marshal))
}
```

現在の`jsonMessage`は必ずmarshal可能なので実害はないが、
「エラーなら黙って1レコード消える」形はJSONストリームでは最も避けたい壊れ方。

**重大度: 低** — `jsonMessage`は bool/string/int のみの平坦な構造体で、
`json.Marshal`が失敗する余地が無い。現状は到達不能コード。

**難易度: 極低** — エラー時に stderr へ落とすか、少なくとも
「失敗しえない」ことをコメントで明示する。数行。

### plain.goの出力フォーマットが崩れている

```go
_, _ = fmt.Fprintln(p.out, prefix, e.ID, e.Text, e.Details)
```

`Fprintln`はオペランド間に必ずスペースを入れるため、

- dry-run無効時: `prefix`が空文字列なので先頭に不要なスペースが1つ入る
- dry-run有効時: `DRYRUN_PREFIX`が末尾にスペースを含むので二重スペースになる
- `Details`が空でも行末にスペースが残る

`fmt.Fprintf`で組み立てる方が素直。

**重大度: 低** — 見た目だけの問題。ただし plain は CI ログや
`grep`の対象になる出力なので、地味に効く。
dry-runが有効化される(問題1が直る)と二重スペースが実際に顕在化する点は連動する。

**難易度: 低** — e2e は`"Container xxx Started"`のような部分一致で見ており
(`pkg/e2e/restart_test.go:70`, `pkg/e2e/up_test.go:155`ほか)、
ID とテキストの間を1スペースに保てば壊れないことは確認済み。

### `DoneColor`が未使用

`colors.go:30`の`DoneColor`はパッケージ外含めどこからも参照されていない
(完了表示は`SuccessColor`)。エクスポート済みなのでlinterは拾わないが、削除候補。

**重大度: 極低** / **難易度: 極低** — `cmd/`配下なので外部からimportされることはなく、
`colors.go`の2行を消すだけ。

### グローバル可変状態(`Mode`、色変数)

`Mode`と各`*Color`はパッケージレベルの可変変数で、`NoColor()`は不可逆な破壊的変更。
テストが並列実行できず(`cmd/compose/compose_progress_test.go`も保存/復元で回避)、
ライブラリとして再利用しにくい制約になっている。

**重大度: 低** — 実行時バグにはならない。CLIは1プロセス1コマンドで、
`Mode`も`NoColor()`もコマンドセットアップ時に一度書かれるだけなのでレースも起きない。
効いてくるのはテストの書きやすさと再利用性。

**難易度: 高** — `display.Mode`だけで26箇所から参照されており、
色変数も`cmd/display`内の複数ファイルが直接読んでいる。
パレットを`termWriter`のフィールドに落とすリファクタは広範囲に波及するうえ、
ユーザーから見た振る舞いは何も変わらないので、upstream に価値を説明しづらい。
費用対効果が最も悪く、実質「直さない」判断が妥当。

### `Starting`/`Started`を捨てるとモデルが古い状態のまま残る

`tty.go:164-171`で、非detached時の`Starting`/`Started`を`handle`より**前**に落としている。
描画の意図(コンテナログを上書きしない)は妥当だが、`taskTree`にも届かないため
ノードのテキストが`Created`のまま取り残され、最終フレームの表示と
`[+] <op> done/total`のカウントが実態からずれる。
「モデルには入れるが描画しない」方が一貫する。

**重大度: 低** — 実際に確認すると、アタッチ`up`では operation が`"up"`
(`pkg/compose/up.go:54`)なのでこの分岐に入り、コンテナ行は`Created`のまま止まる。
`Started`の描画抑制自体は意図通りで、コンテナログの直前を踏まないための設計。
カウントも`createdEvent`が Done 状態なので数字が減るわけではない。
つまり見え方の一貫性の問題であって、壊れているわけではない。

**難易度: 中** — 「モデルには入れて描画側で落とす」形にすると、
`layoutFrame`/`buildRows`に抑制の概念を持ち込むことになり、
「レイアウトは純粋関数」という現在の設計思想にどう収めるかの判断が要る。
ユーザーから見える出力が変わるので、e2e(`restart_test.go`/`ps_test.go`等が
`Started`の有無を見ている)への影響確認も必須。
そもそも upstream の意図的な選択なので、直す前に issue で確認すべき部類。

### テストが無いファイルが多い

`json_test.go`/`tty_test.go`は存在するが、`colors.go`/`dryrun.go`/`mode.go`/
`plain.go`/`quiet.go`には単体テストが無い。特に`plain.go`は非対話実行時の
デフォルト表示なので、`json_test.go`と同程度の簡単なテストを足す価値はある。

**重大度: 低** — 現状これが原因のバグは見つかっていない。
ただし問題1(dry-run)を直すなら、plain の回帰テストは同時に欲しい。

**難易度: 低** — `plain_test.go`は`bytes.Buffer`に吐かせて中身を見るだけで、
数十行。既存の`json_test.go`がそのまま雛形になる。

### 細かい点

- `DRYRUN_PREFIX`はGoの命名規約から外れる(`DryRunPrefix`が自然)。
  **重大度: 極低 / 難易度: 極低** — 参照は`plain.go`/`tty_layout.go`/`dryrun.go`の
  3箇所のみで、`cmd/`配下なので外部からは使われない。改名は機械的。
  (前に「広く参照されている」と書いたが、実際に数えると3箇所だった)
- `plain.go`/`json.go`の`Start(ctx context.Context, operation string)`は
  引数未使用なのに名前付き。`quiet.go`に倣って`_`にすると意図が明確。
  **重大度: 極低 / 難易度: 極低**

## 他のnote.mdとの関連

- [document/contributing/pr-review-patterns.md](../../document/contributing/pr-review-patterns.md)
  テーマ④(出力はアプリの表示基盤を経由すべき)は、まさにこの`cmd/display`が
  提供する`api.EventProcessor`のことを指している。`logrus.Warn`等で直接
  ログ出力すると、ここでまとめたTUI描画を経由しないため
  画面に出ない、という失敗パターンだった([pr/pr-14146.md](../../pr/pr-14146.md)参照)。
