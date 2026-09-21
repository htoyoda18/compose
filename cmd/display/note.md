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
- tty.go(本パッケージの中核、18KB)
  - `Full`が返す`ttyWriter`が`api.EventProcessor`を実装し、ANSIエスケープでカーソルを
    戻しながら複数タスクの進捗を上書き描画するTUIレンダラー
  - `task`構造体で各リソースの状態(進捗/経過時間/親子関係)を保持し、100msごとの
    `time.Ticker`で再描画。ビルド中(`api.StatusBuilding`)はBuildKit側の表示と
    衝突しないようtickerを一時停止する仕組みも持つ
  - `adjustLineWidth`/`applyPadding`等、ターミナル幅に収まるよう詳細情報→進捗サイズ→
    taskID の順に段階的に切り詰めるロジックが大部分を占める
- json.go
  - `JSON`が返す`jsonWriter`。各イベントを1行1JSONの`jsonMessage`として出力する
    機械可読モード(`docker compose --progress json ...`)
- plain.go
  - `Plain`が返す`plainWriter`。TTY機能を使わず、イベントをそのまま1行ずつ出力する
    (非対話環境やパイプ経由での実行時に使われる)
- quiet.go
  - `Quiet`が返す`quiet`。全メソッドが no-op で、`run --quiet`/`build --quiet`時など
    出力を完全に抑制したい場合に使う
- colors.go
  - `DoneColor`/`WarningColor`/`ErrorColor`等の色付け関数(package-level var)を定義し、
    `NoColor()`で全てを恒等関数に差し替えることでカラー出力を一括無効化する
- dryrun.go
  - `DRYRUN_PREFIX`(` DRY-RUN MODE - `)の定数のみを持つ小さなファイル
- spinner.go
  - `Spinner`。100ms経過ごとに次の文字へ進むアニメーション用の内部状態(Windowsでは
    Unicodeスピナー文字の代わりに`-`にフォールバック)

## 確認できた問題点・改善点

### `dryRun`フィールドが3つのwriter全てで常にfalseのまま(生きていない)

`jsonWriter.dryRun`(json.go)/`plainWriter.dryRun`(plain.go)/`ttyWriter.dryRun`
(tty.go)はいずれも構造体フィールドとして存在し、`if w.dryRun { ... }`で
`DRYRUN_PREFIX`を出し分ける・JSONの`dry-run`フィールドを立てる、という分岐が
書かれているが、**公開コンストラクタ(`Full`/`JSON`/`Plain`/`Quiet`)のどれも
この値を設定する引数を持たない**。呼び出し元の`selectEventProcessor`
(`cmd/compose/compose.go`)を確認しても、`display.Full(...)`/`display.Plain(...)`/
`display.JSON(...)`のいずれの呼び出しにも`dryRun`相当の引数は渡されていない。

つまり`docker compose --dry-run up`のようにdry-runモードで実行しても、
進捗表示には` DRY-RUN MODE - `プレフィックスが一切出ない。`tty.go:76`に

```go
dryRun    bool // FIXME(ndeloof) (re)implement support for dry-run
```

というFIXMEコメントが残っており、メンテナ自身もこれが未実装であることを
把握している。→ ちょうどこの調査時点で自分(htoyoda18)が
[PR #14053](https://github.com/docker/compose/pull/14053)
(`fix(display): wire --dry-run flag into progress writers`)としてこの
FIXMEの解消に取り組み中。

### `tty.go:231` でリソースIDの特別扱いに文字列リテラルを直書きしている

```go
if e.ID == "Compose" {
    _, _ = fmt.Fprintln(w.info, ErrorColor(e.Details))
    continue
}
```

`"Compose"`という特別なリソースID(バックエンド全体のエラー等を表す)は
`pkg/api/event.go`で`const ResourceCompose = "Compose"`としてちゃんと
公開定数が定義されている。ここだけリテラル文字列で比較しており、
`api.ResourceCompose`を使っていない。今は値が一致しているので実害はないが、
将来この定数値が変更された場合ここだけ追従漏れするリスクがある一貫性の問題。

### テストが無いファイルが多い

`json_test.go`/`tty_test.go`は存在するが、`colors.go`/`dryrun.go`/`mode.go`/
`plain.go`/`quiet.go`/`spinner.go`には単体テストが無い。特に`plain.go`は
非対話実行時のデフォルト表示なので、`json_test.go`と同程度の簡単なテスト
(`Event`が期待通りの行を出力するか)を足す価値はありそう。

### `tty.go`の複雑さ

`print`/`printWithDimensions`/`adjustLineWidth`/`applyPadding`まわりは
ターミナル幅に収める行の切り詰めロジックが密結合しており、ファイル全体の
約半分をこの整形処理が占める。ロジック自体は`tty_test.go`で手厚くテストされて
おり動作に疑わしい点は見当たらなかったが、変更を加える際は影響範囲が
広くなりやすい箇所だと分かった。

## 他のnote.mdとの関連

- [document/pr-review-patterns.md](../../document/pr-review-patterns.md)
  テーマ④(出力はアプリの表示基盤を経由すべき)は、まさにこの`cmd/display`が
  提供する`api.EventProcessor`のことを指している。`logrus.Warn`等で直接
  ログ出力すると、ここでまとめた`ttyWriter`のTUI描画を経由しないため
  画面に出ない、という失敗パターンだった([pr/pr-14146.md](../../pr/pr-14146.md)参照)。
