# cmd/compatibility メモ

`docker-compose`(standalone バイナリ)として起動された場合に、旧来の
`docker-compose` の引数を `docker compose`(CLIプラグイン)向けの引数列に
変換するための互換レイヤー。ファイルは実質1つ。

## convert.go

- **Convert** (L56): `cmd/main.go` から `plugin.RunningStandalone()` 時に
  呼ばれ、`os.Args` を書き換える中心関数。引数を1つずつ見て、
  ①completionコマンドの並べ替え、②`--verbose`→`--debug`、`-h`→`--help`、
  `--version`/`-v`→`version`サブコマンドへの変換、③グローバルフラグ
  (bool/string)を`compose`サブコマンドの前に括り出す、④それ以外は
  サブコマンド以降の引数としてそのまま残す、という処理を行う。
- **getCompletionCommands / getBoolFlags / getStringFlags** (L28-53):
  それぞれのカテゴリのフラグ/コマンド名リストを返すだけの小さなヘルパー。

## convert_test.go

`Convert` に対するテーブル駆動テスト。実際にあった過去issue
(`issues/1962`, `issues/8648`, `issues/12`など)を回帰テストとして
固定化しているのが特徴的。フラグの引数が渡されない異常系
(`--log-level`のみ)は`os.Exit(1)`するため、サブプロセスを起動して
終了コードを検証する手法(`BE_CRASHER`パターン)でテストしている。

## 気になった点・改善提案

1. **getCompletionCommands()/getBoolFlags()/getStringFlags() がループ内で
   毎回呼ばれ、都度スライスを生成している**
   `Convert`のループ内(引数の数だけ)で`getCompletionCommands()`と
   `getBoolFlags()`が毎回呼び出され、そのたびに新しいスライスがアロケート
   される。起動時に1回しか呼ばれないCLIなので実害はほぼないが、素直に
   package-level の `var`(もしくは `map[string]struct{}` にして
   `slices.Contains`のO(n)探索をO(1)にする)にした方がシンプルかつ効率的。

2. **-h書き換えのロジックとissue #1962のテストがギリギリ整合している**
   `-h`は無条件に`--help`へ変換されず(L75)、`getStringFlags`に`-H`は
   あってもソース上"-h"を除外する明示的な仕組みはなく、実際には`switch`文の
   `case "-h":`が「引数全体が`-h`のときだけ」発火するため、
   `exec mongo -h postgres`のような「サブコマンド以降に出てくる`-h`」には
   触れない、という設計になっている(`arg[0] != '-'`のチェックでサブコマンド
   以降はループを抜けるため)。意図通り動いてはいるが、コメントが「なぜ安全か」
   を薄く説明しているだけなので、依存しているロジック(ループを抜ける条件)
   への言及がなく、少し読み解きにくい。

3. **エラー処理が os.Exit(1) 直書き**
   L91-92で引数不足のとき`fmt.Fprintf` + `os.Exit(1)`しているが、`Convert`
   自体はテスト容易性を考えると`error`を返す設計の方が本来望ましい(実際、
   テストが子プロセスを`exec`して終了コードを確認するという回りくどい方法を
   取っているのはこれが原因)。ただし`main.go`から見て`Convert`は「起動直後の
   引数パース」という性質上、CLIとして即座に落として良い処理であり、実害が
   あるバグではない。

バグと呼べるものは見当たらなかった。

## convert.go は最新の Go でシンプルにできるか

`go.mod`は Go 1.26.3 指定なので、かなり新しい機能が使える。すでに
`slices.Contains`(1.21+)や`strings.Cut`(1.18+)は使われているが、もう
一歩シンプルにできる箇所がある。

### strings.CutPrefix(Go 1.20+)でprefix+分割を1行に統合

現状(L97-103):
```go
if strings.HasPrefix(arg, flag) {
    _, val, found := strings.Cut(arg, "=")
    if found {
        rootFlags = append(rootFlags, flag, val)
        continue ARGS
    }
}
```
これは実質「`flag=`で始まっているか」を`HasPrefix`+`Cut`の2段階でチェック
しているが、`strings.CutPrefix`を使えば1回で書ける:
```go
if val, ok := strings.CutPrefix(arg, flag+"="); ok {
    rootFlags = append(rootFlags, flag, val)
    continue ARGS
}
```
こちらの方が「`flag=`という接頭辞を切り出す」という意図がそのままコードに
表れて読みやすく、`Cut(arg, "=")`が万一 flag と無関係な位置の`=`に
マッチする(理論上は起きないが)可能性の心配も無くなる。

### helper関数をpackage-level varに

```go
var (
    completionCommands = []string{"__complete", "__completeNoDesc"}
    boolFlags           = []string{"--debug", "-D", "--verbose", "--tls", "--tlsverify"}
    stringFlags         = []string{"--tlscacert", "--tlscert", "--tlskey", "--host", "-H", "--context", "--log-level"}
)
```
とすれば関数呼び出し自体が不要になり、`slices.Contains(boolFlags, arg)`の
ように直接参照できる。3つの`get*`関数を消せる分、行数も減る。

### range-over-int は使えない

メインの`for i := 0; i < l; i++`ループは、値を消費するために`i`を手動で
インクリメントする必要があるため、Go 1.22の`for i := range l`
(range-over-int)には置き換えられない。ラベル付き`continue ARGS`も、
Goのイディオムとしてはこのままで妥当。

全体としては大きな書き直しが必要なほどの技術的負債ではなく、
`strings.CutPrefix`への置き換えとヘルパー関数のvar化、という2点の
小さなリファクタで十分シンプルになる。
