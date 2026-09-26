# cmd/formatter メモ

`cmd/display`が「Composeの操作進捗(プログレスバー)」を担うのに対し、
`cmd/formatter`は**それ以外のユーザー向け出力**を一手に引き受けるパッケージ。
役割は大きく3系統あり、互いにほぼ独立している。

1. **一覧出力の整形** — `ps`/`ls`/`images`/`config --variables`等の
   table / JSON / Go テンプレート出力(`formatter.go`, `pretty.go`, `json.go`,
   `container.go`, `consts.go`)
2. **コンテナログの整形** — `up`/`logs`でサービス名プレフィックスと色を付けて
   ログ行を流す(`logs.go`, `colors.go`)
3. **アタッチ中の対話メニュー** — `up`実行中に画面下部へ出る
   `v View in Docker Desktop  w Watch  d Detach`のショートカットUI
   (`shortcut.go`, `ansi.go`, `shortcut_unix.go`, `shortcut_windows.go`)

2と3は同じ端末に同時に書き込むため、`logDecorator`(logs.go)が
「ログを出す前にメニューを消し、出した後に描き直す」という形で結合している。
ここが本パッケージで最も複雑かつ壊れやすい部分。

## ファイル一覧・責務

- consts.go
  - 出力フォーマット名の定数(`json`/`table`/`pretty`/`{{json.}}`)のみ
  - `PRETTY`は Deprecated で`TABLE`が後継。`formatter.Print`のswitchで使われる
- formatter.go
  - `Print()`が入口。フォーマット文字列で table / JSON / legacy JSON に振り分ける
  - `reflect`でスライスかどうかを見て、スライスなら要素ごと・単体ならそのまま JSON 化
  - table の場合は整形を呼び出し元の`writerFn`に委ね、自身はヘッダ行だけ面倒を見る
  - 呼び出し元は`cmd/compose`の`config.go`/`list.go`/`images.go`/`bridge.go`の4つ
- pretty.go
  - `text/tabwriter`でヘッダ+本文を桁揃えするだけの薄いラッパー(`PrintPrettySection`)
- json.go
  - `encoding/json`のラッパー。`SetEscapeHTML(false)`でURL中の`&`等を壊さない
  - `ToStandardJSON`はインデント付き、`ToJSON`はインデント指定可
- container.go
  - `docker compose ps`用。`docker/cli`の`formatter.Format`/`SubContext`機構に乗せ、
    Goテンプレート(`{{.Name}}`等)から呼ばれるアクセサを`ContainerContext`として提供
  - `api.ContainerSummary`→表示文字列の変換(ID短縮、ポート整形、サイズ人間可読化等)
  - `NewContainerFormat`が`--format`/`--quiet`/`--size`からテンプレート文字列を組み立てる
  - upstream `docker/cli`の`ContainerContext`をコピーして Compose 固有の
    `Service`/`Project`/`Engine`列を足した派生物
- colors.go
  - サービス名に割り当てる色を管理。`nextColor()`がrainbow配列を巡回して返す
    (mutex保護済み)
  - `SetANSIMode`が`--ansi`(`never`/`always`/`auto`)とTTY判定から
    色を殺すか決め、`ansi.go`の`disableAnsi`も同時に立てる
  - ANSIエスケープの組み立て(`ansiColor`/`ansiColorCode`)もここ
- ansi.go
  - カーソル移動・行消去・OSC8ハイパーリンクなど低レベルANSI操作のヘルパー群
  - すべて`disableAnsi`を見て無効化でき、出力先は**`fmt.Print`(= os.Stdout)固定**
  - `lenAnsi()`はエスケープを剥がしてから長さを測る(行折り返し計算に使う)
- logs.go
  - `up`/`logs`のコンテナログ出力。`api.LogConsumer`の実装`logConsumer`
  - コンテナ名ごとに`presenter`(色+パディング済みプレフィックス)を`sync.Map`で保持し、
    新しい名前が来るたび全体の幅を再計算して全プレフィックスを揃え直す
  - `logDecorator`は`LogConsumer`をラップし、1行ごとに`Before`/`After`フックを挟む
    (`shortcut.go`のメニュー再描画に使われる)
- shortcut.go
  - `up`アタッチ中の対話メニュー本体(`LogKeyboard`)
  - キー入力(`v`/`o`/`l`/`w`/`d`/Ctrl-C/Ctrl-Z/Enter)を`HandleKeyEvents`で処理し、
    Docker Desktopを開く・watchを開始停止する・デタッチする等を実行
  - `printNavigationMenu`/`clearNavigationMenu`が画面最下部にメニューを描画。
    エラーはメニューの1行上に10秒だけ表示される(`KeyboardError`)
  - 呼び出し元は`pkg/compose/up.go`の`setupNavigationMenu`/`runEventLoop`
- shortcut_unix.go / shortcut_windows.go
  - Ctrl-Zの処理だけをOSで出し分け。Unixは`syscall.Kill(0, SIGSTOP)`で
    プロセスグループを停止、Windowsは何もしない
- ansi_test.go / formatter_test.go / container_test.go
  - `OSC8Link`、`Print`の3フォーマット、`Engine`列の有無のみをカバー。
    後述の通り`logs.go`/`shortcut.go`/`colors.go`にはテストが無い

## 確認できた問題点・改善点

| # | 問題点 | 重大度 | 難易度 |
|---|--------|--------|--------|
| 1 | `logConsumer`に data race が3箇所(`-race`で再現確認) | **High** | Low |
| 2 | `LogKeyboard`の状態が無防備に共有されている | Medium | Medium |
| 3 | `ps --format "table {{.Mounts}}"`のヘッダが`<no value>` | Medium | **Low** |
| 4 | `{{.Labels}}`の並び順が非決定的(upstreamはソート済み) | Medium | **Low** |
| 5 | `up --timestamps`が「表示時刻」を出している | Medium | Medium |
| 6 | `Print(nil, ...)`が panic する | Low | Low |
| 7 | JSON出力の形がスライス/非スライスで不統一 | Low | Medium |
| 8 | `ansi.go`が`fmt.Print`固定でテスト不能 | Medium | High |
| 9 | ポート番号の`uint16`への無検査キャスト | Low | Low |
| 10 | `logs.go`/`shortcut.go`/`colors.go`にテストが無い | Medium | Medium |
| 11 | `keyboardError`がエラーごとにgoroutineとTimerを作る | Low | Low |

---

### 1. `logConsumer`に data race が3箇所ある

**問題内容**: `presenters`は`sync.Map`で守られているが、その周辺の素のフィールドが
無防備。`go test -race`で実際に検出できる(8コンテナ×50行で再現):

- `l.width`(int) — `computeWidth()`(logs.go:147)が書き、
  `register()`のRangeループ(logs.go:92)が`p.setPrefix(l.width)`で読む
- `p.prefix`(string) — `setPrefix()`(logs.go:161)が書き、
  `write()`(logs.go:127)が`fmt.Fprintf`で読む
- `register()`自体が Load→Store の非アトミックな read-modify-write

```
WARNING: DATA RACE
Write at 0x... by goroutine 24: (*logConsumer).computeWidth() logs.go:147
Previous write at 0x... by goroutine 26: (*logConsumer).computeWidth() logs.go:147
```

**なぜ問題か**: これは理論上の話ではなく、**通常の`docker compose up`で必ず通る経路**。
`pkg/compose/logs.go:47-52`がコンテナ1つにつき`eg.Go`を1本立て、
全goroutineが同じ`consumer`を共有する。新しいコンテナのログが初めて届くたびに
`register()`が走り、その裏で他のコンテナがログを書いている。
実害としてはプレフィックスの桁ずれや、Goのメモリモデル上は
任意の壊れ方(最適化次第)が許される。CIが`-race`付きなら踏む可能性もある。

**重大度: High** — 正真正銘のデータ競合で、再現経路が日常的。
ただしログ本文が壊れるわけではなく、観測される被害は表示崩れに留まる見込み。

**解決難易度: Low** — `sync.Map`をやめて`sync.Mutex`+ただの`map`にし、
`register`/`getPresenter`/`computeWidth`/`write`のプレフィックス読み出しを
1つのロックで囲むのが素直。`write()`は`p.prefix`をロック下でローカルにコピーしてから
I/Oすればよい(I/O中にロックを握らない)。再現テストも上記の形で簡単に書ける。

### 2. `LogKeyboard`の状態が無防備に共有されている

**問題内容**: `lk.Watch.Watching`と`lk.kError`が、同期なしに複数goroutineから
読み書きされる。

- 書き側: `ToggleWatch`が**別goroutine内で**`lk.Watch.Watching`を更新
  (shortcut.go:295-305)、`keyboardError`→`addError`が`kError`を更新
- 読み側: `PrintKeyboardInfo`→`navigationMenu()`が`lk.Watch.Watching`を、
  `createBuffer`/`printError`が`kError`を読む

**なぜ問題か**: 読み側はログ出力のたびに`logDecorator.After`から呼ばれる。
つまり`pkg/compose/up.go:447`のログ配信goroutine群から。
書き側は`runEventLoop`(pkg/compose/up.go:311)のキーボードgoroutineと、
そこからさらに派生したgoroutine。**明確に別スレッド**で、ロックが一切ない。
`keyboardError`が張るタイマーgoroutine(shortcut.go:250)も`printNavigationMenu`を
呼ぶので、描画自体が3方向から並行に走りうる。

**重大度: Medium** — 競合はするが、壊れるのは端末表示のみ(メニューの二重描画、
カーソル位置のずれ)。データや終了コードには影響しない。

**解決難易度: Medium** — フィールドにmutexを足すだけでは、
描画そのものが複数goroutineから同時に走る構造は直らない。
本筋は「描画要求をチャネルに送り、単一の描画goroutineが処理する」形への変更で、
そこまでやると`shortcut.go`のかなりの範囲に手が入る。

### 3. `ps --format "table {{.Mounts}}"`のヘッダが`<no value>`になる

**問題内容**: `NewContainerContext()`(container.go:117-136)のヘッダマップに
`Mounts`/`LocalVolumes`/`Networks`のエントリが無い。
一方で`Mounts()`/`LocalVolumes()`/`Networks()`メソッドは実装済みで、
`mountsHeader`/`localVolumes`/`networksHeader`という定数も**定義されているのに
どこからも使われていない**(container.go:42-44)。

実際に`ContainerWrite`を叩いて確認:

```
FORMAT "table {{.Mounts}}\t{{.Name}}"
<no value>   NAME
m1           c1
```

`SubHeaderContext`は`map[string]string`で、テンプレートがマップの欠損キーを
参照すると`<no value>`が出る。`Health`/`ExitCode`/`Publishers`も同様に未登録。

**なぜ問題か**: ユーザーが`--format`で指定できる公開フィールドなのに、
出力が明らかに壊れて見える。upstream `docker/cli`の同じ構造体には
`"Mounts": mountsHeader`等がきちんと入っており(cli/command/formatter/container.go:122-124)、
Compose側にコピーしてくる際に落ちたものと読める。未使用定数がその痕跡。

**重大度: Medium** — ユーザー可視の明確な不具合だが、
デフォルトのtable形式では使われない列なので遭遇頻度は低い。

**解決難易度: Low** — ヘッダマップに3〜6行足すだけ。
`container_test.go`に既にある`ContainerWrite`のテストがそのまま雛形になる。

### 4. `{{.Labels}}`の並び順が非決定的

**問題内容**: `ContainerContext.Labels()`(container.go:241-251)がマップを
そのままrangeして`strings.Join`している。Goのマップ反復順はランダム。

実際の出力(同じ入力で実行するたび順序が変わる):
```
NAME      LABELS
c1        b=2,a=1,c=3,d=4,e=5
```

**なぜ問題か**: `docker compose ps --format '{{.Labels}}'`の出力が
実行ごとに変わるため、スクリプトでの差分検知やテストのゴールデンファイルが作れない。
upstream `docker/cli`の同じ関数は`slices.Sort(joinLabels)`で明示的にソートしている
(cli/command/formatter/container.go:295-300)。**Compose側のコピーだけソートが抜けている**。

**重大度: Medium** — 出力の再現性が壊れている。upstreamとの挙動差でもある。

**解決難易度: Low** — `slices.Sort`を1行足すだけ。upstreamに合わせる形なので
レビューも通しやすい。

### 5. `up --timestamps`が「コンテナが出力した時刻」ではなく「CLIが表示した時刻」を出す

**問題内容**: `logConsumer.write()`(logs.go:122)が`time.Now()`をその場で取っている。
そして`up`と`logs`でタイムスタンプの出し方が食い違っている:

| コマンド | formatterの`timestamp`引数 | APIの`Timestamps` |
|---|---|---|
| `compose logs --timestamps` | `false` (cmd/compose/logs.go:99) | `opts.timestamps` (同108行) |
| `compose up --timestamps` | `upOptions.timestamp` (cmd/compose/up.go:306) | **設定なし** |

つまり`logs`はDocker APIに付けさせた**コンテナ由来の**タイムスタンプを表示し、
`up`はCLIが行を印字した瞬間の**ローカル時刻**を表示する。

**なぜ問題か**: 同じ`--timestamps`という名前のフラグで意味が違う。
ログのバッファリングや配信の詰まりがあると両者は簡単にずれ、
`up --timestamps`の値はデバッグ用途で当てにできない。
さらに複数行メッセージでは全行に同じ時刻が付く。

**重大度: Medium** — ユーザーが誤った情報を受け取りうる。
ただし気付きにくく、致命的な誤動作ではない。

**解決難易度: Medium** — `up`側でもAPIの`Timestamps`を立てて
formatter側の自前スタンプをやめるのが筋だが、
`up`のログ経路は`streamContainerLogs`(pkg/compose/up.go:455)を通るので
そこまで引き回す必要がある。挙動が変わるのでe2eの確認も要る。

### 6. `Print(nil, ...)`が panic する

**問題内容**: `formatter.go:53`の`reflect.TypeOf(toJSON).Kind()`は、
`toJSON`が nil インターフェースだと`reflect.TypeOf`が nil を返し、
そこに`.Kind()`を呼んで nil ポインタ参照になる。実際に確認済み:

```
panic: runtime error: invalid memory address or nil pointer dereference
  .../cmd/formatter.Print(...) formatter.go:53
```

**なぜ問題か**: 現在の呼び出し元4箇所(`config.go`/`list.go`/`images.go`/`bridge.go`)は
いずれも非nilのスライスやマップを渡すので今は到達しない。
ただし`Print`はパッケージ公開APIで、`any`を受ける以上 nil は普通に来うる形。
`[]T(nil)`のような**型付きnil**は正常に`null`を出力するので、
「nilでも動くことがある」のがかえって罠になっている。

**重大度: Low** — 現状到達不能。

**解決難易度: Low** — 先頭で`if toJSON == nil`を弾くか、
`reflect.ValueOf(toJSON).Kind()`(nilなら`Invalid`を返す)に変えるだけ。

### 7. JSON出力の形がスライスか否かで揃っていない

**問題内容**: `Print`のJSON分岐が、スライスなら`ToJSON`(インデントなし・改行1つ)、
それ以外なら`ToStandardJSON`+`Fprintln`(4スペースインデント・**末尾に空行**)と
別々の形を出す。実測:

```
[]s{{A:"x"},{A:"y"}}  -> "[{\"A\":\"x\"},{\"A\":\"y\"}]\n"
s{A:"x"}              -> "{\n    \"A\": \"x\"\n}\n\n"
```

`encoding/json.Encoder.Encode`が既に改行を付けるので、
その上の`Fprintln`が余分な空行を作っている。

**なぜ問題か**: `compose ls --format json`はスライス、
`compose config --variables --format json`はマップを渡している(config.go:759)ため、
**同じ`--format json`でコマンドによって整形とトレーリング改行が変わる**。
下流でパースする側にとって扱いづらい。

**重大度: Low** — パース自体は通る(JSONとしては両方valid)。見た目と一貫性の問題。

**解決難易度: Medium** — 揃える方向はすぐ決まるが、
どちらに揃えるかは**既存の出力互換性**の判断になる。upstreamに出すなら要相談。
なお同じ`config --variables`の table 出力も`for name := range variables`で
マップを直接回しており行順が非決定的(cmd/compose/config.go:760)。
こちらは`cmd/formatter`の外だが、合わせて直す価値がある。

### 8. `ansi.go`が`fmt.Print`固定で、メニュー描画がテストできない

**問題内容**: `saveCursor`/`moveCursor`/`clearLine`等すべてが
`fmt.Print`= プロセスの`os.Stdout`に直書きする。`io.Writer`を受け取らない。

**なぜ問題か**:

- `shortcut.go`のメニュー描画は全てこの関数群の上に建っているため、
  **出力を検証する単体テストが一切書けない**(実際テストが無い)
- `logConsumer`はコンストラクタで`stdout`/`stderr`を受け取る設計なのに、
  同じ端末に書くメニュー側だけがグローバルに書く。
  `dockerCli.Out()`をリダイレクトしてもメニューだけは素通りする不整合
- `cmd/display`側は`screen{out: io.Writer}`と`termWriter`への注入で
  同じ問題をきちんと解いており、パッケージ間で設計水準が揃っていない

**重大度: Medium** — 現時点でのバグではないが、
問題2(競合)や表示崩れを直そうにも**検証手段が無い**ことが
そのまま改善の足かせになっている。

**解決難易度: High** — `ansi.go`の全関数に`io.Writer`を足し、
`LogKeyboard`にそれを持たせ、`shortcut.go`の全呼び出しを書き換える必要がある。
`disableAnsi`というパッケージグローバルも同時に始末したいところで、
変更は`cmd/formatter`ほぼ全域に及ぶ。ユーザーから見た挙動は変わらないので、
upstreamにこの規模のリファクタの価値を説明するのが難しい部類。

### 9. ポート番号を無検査で`uint16`にキャストしている

**問題内容**: `ContainerContext.Ports()`(container.go:232-233)が
`uint16(publisher.TargetPort)` / `uint16(publisher.PublishedPort)`と
`int`から切り詰めている。65535超なら黙って巻き戻る(65536→0)。

**なぜ問題か**: `api.PortPublishers`のポートは`int`で、
compose-specのパース段階で範囲検証されている前提なら実害は無い。
ただしコード上その保証は読み取れず、`gosec`のG115が本来拾う形。
異常値が来たときに「エラー」ではなく「もっともらしい別のポート番号」を
表示してしまうのが良くない。

**重大度: Low** — 到達には不正なポート定義が必要で、上流で弾かれている公算が大きい。

**解決難易度: Low** — 範囲外を`0`やエラー扱いにするガードを足すだけ。

### 10. `logs.go` / `shortcut.go` / `colors.go` にテストが無い

**問題内容**: テストがあるのは`ansi_test.go`(OSC8リンクのみ)、
`formatter_test.go`(`Print`の3形式)、`container_test.go`(`Engine`列)だけ。
**行数・状態量ともに最も大きい`shortcut.go`(9.6KB)と`logs.go`(4.3KB)が未カバー**。

**なぜ問題か**: 問題1のdata raceが今まで残っていたのは、
並行性を突くテストが1つも無かったことと直結している。
`colors.go`の`nextColor()`もグローバルな`currentIndex`を進めるので、
テストを足す際は順序依存に注意が要る(現状テストが無いので露見していない)。

**重大度: Medium** — 単体では実害ゼロだが、他の問題の検出漏れの原因になっている。

**解決難易度: Medium** — `logs.go`は`io.Writer`を注入できる設計なので
テストを書きやすく、難易度は低い(問題1の回帰テストがそのまま使える)。
`shortcut.go`は問題8(出力先がグローバル)を先に解かないと書けないため、
そちらに引きずられて中程度。

### 11. `keyboardError`がエラーのたびにgoroutineとTimerを作る

**問題内容**: `keyboardError`(shortcut.go:247-255)が
毎回`time.NewTimer`+goroutineを起こし、11秒後に再描画して終わる。
`timer.Stop()`は呼ばれず、goroutineの本数にも上限が無い。

**なぜ問題か**: キーを連打すると短時間に何十本も溜まりうる。
それぞれが`printNavigationMenu()`を呼ぶので、問題2の競合を悪化させる。
11秒で必ず終了するのでリークではないが、無駄かつ不安定要因。

**重大度: Low** — 実害は限定的。

**解決難易度: Low** — タイマーを1本に持ち替える(既存があれば`Reset`)だけ。

---

## まとめ

- **まず直すべきは 1(data race)**。再現が確実で、修正範囲も`logs.go`に閉じており、
  回帰テストも書ける。次いで **3 と 4** は upstream `docker/cli` に正解があるため
  数行で直せてレビューも通しやすい。
- **5 と 7** は出力仕様の変更を伴うので、実装より「どう揃えるか」の合意が本体。
- **2 と 8 と 10** は一体の問題。`shortcut.go`は
  「グローバルな`fmt.Print`」「ロックなしの共有状態」「テスト不在」が
  互いを強化し合っている。手を入れるなら 8 → 10 → 2 の順で、
  まず注入可能にしてテストを敷いてから並行性を直すのが安全。
- 6・9・11 は局所的で低リスク。ついでに片付ける類。

## 他のnote.mdとの関連

- [cmd/display/note.md](../display/note.md)
  `cmd/display`が進捗(プログレス)、`cmd/formatter`がそれ以外の出力、という分担。
  両者は同じ端末に書くため、`cmd/display`の`termWriter`が
  「非detachedの`Starting`/`Started`を描かない」のは、
  ここ`cmd/formatter`の`logConsumer`が流すコンテナログを踏まないための配慮。
  端末出力先の注入(`screen{out}`)を`cmd/display`は済ませていて
  `cmd/formatter`の`ansi.go`は済ませていない、という対比も参考になる(問題8)。
