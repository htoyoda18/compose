# cmd/cmdtrace メモ

Docker Compose CLI の各コマンド実行を OpenTelemetry でトレース(計測)するためのディレクトリ。
ファイルは `cmd_span.go` の1つのみ。

## cmd_span.go

- **Setup** (L50): `PersistentPreRunE` から呼ばれるエントリポイント。
  `internal/tracing.InitTracing` でトレーサーを初期化し、実行されたコマンド名・フラグ・
  TTY判定などの属性を持つルート span を開始する。
- **wrapRunE** (L77): cobra の `RunE`(または `Run`)を差し替え、元の処理の前後に
  span の終了処理・エラー記録・exporter の flush(shutdown)を挟むラッパー。
  `PersistentPostRunE` だとエラー時に呼ばれないため、あえてこの方式を採用している。
- **commandName** (L132): span名生成のため、cobraコマンドの親を辿ってコマンドパス
  (`docker compose` 部分を除く)を配列にし、"逆アルファベット順"にソートして返す。
- **getFlags** (L144): 実際にユーザーが指定したフラグ名一覧を取得するヘルパー。

## cmd_span_test.go

`getFlags` と `commandName` に対するテーブル駆動のユニットテスト。
`wrapRunE` や `Setup` 自体はテストされていない。

## 気になった点・改善提案

1. **commandName のソートロジックが脆弱** (L140)
   親を辿るループは既に leaf→root(例: `[watch, alpha]`)の順で結果を作っており、
   これはコマンド階層をそのまま反映した決定的な順序。しかしその後
   `sort.Sort(sort.Reverse(...))` で**アルファベット降順**に並べ替えている。
   これが階層順と一致するのは、現状の `alpha` サブコマンド群
   (`viz`, `publish`, `generate`)がすべてアルファベット的に `"alpha"` より
   後ろに来る文字列だから、という偶然に過ぎない。
   将来 `alpha` より辞書順で前に来る名前のサブコマンド(例: `abort` のような名前)を
   追加すると、`[abort, alpha]` が既に正しい順序なのに、逆アルファベットソートで
   `[alpha, abort]` と誤って入れ替わり、span の名前(`cli/alpha-abort`)が実際の
   呼び出し階層と食い違ってしまう。
   → ソートを行わずループで得た順序をそのまま使うか、少なくとも「なぜソートが
   必要か」をコメントで明示すべき(現状のコメントは「一貫性のため」としか
   書かれておらず、根拠が薄い)。

2. **wrapRunE 内の `cmdSpan != nil` チェックが実質デッドコード** (L91)
   `cmdSpan` は `Setup` 内で `otel.Tracer("").Start(...)` から返された値で、
   OTEL の仕様上 noop でも非nilの `trace.Span` が返るため、このnilチェックは
   常に true。バグではないが、意味のない防御コードで読み手を惑わせる可能性がある。

3. **tracing shutdown のタイムアウトが100msで固定** (L113)
   OTLPエクスポート先がリモート(ネットワーク越し)の場合、100msでは span の
   フラッシュが間に合わず、サイレントにトレースが失われる可能性がある。
   エラー自体は `logrus.Debug` 止まりで気づきにくい設計。意図的なトレードオフ
   (CLI終了を遅らせたくない)だとは思うが、値が短めなのでコメントで根拠を残すか、
   環境変数で調整可能にすると良さそう。

4. **wrapRunE / Setup のテストが無い**
   現状のテストは `getFlags` と `commandName` のみで、エラー発生時に
   `exitCode` や span ステータスが正しく設定されるかなど、`wrapRunE` の分岐ロジックは
   カバーされていない。

バグと呼べるほど致命的なものはないが、1 が最も実害が出やすい
(将来のコマンド追加で顕在化しうる)。