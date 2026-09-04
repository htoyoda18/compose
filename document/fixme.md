## 現存するFIXME（2026-08-24 upstream/main時点で再スキャン）

- go.mod:215（旧160）
  - exclude ブロックの FIXME
  - go.sum上も除外対象のpseudo-versionが今も有効。kustomize側の依存修正待ちで変化なし
- cmd/display/tty.go:59（旧56）
  - dryRun フィールドが未接続。plain/json含め3writer共通の抜け
  - **修正PR作成済み・未マージ**: https://github.com/docker/compose/pull/14053 （Open, レビュー0件, mergeStateStatus: BLOCKED）
- pkg/compose/build_bake.go:393（旧298）
  - bake への fs.read 自動許可。意図的な保留（確認プロンプトを挟むとUX破壊）、変化なし
- pkg/compose/create_test.go:359（旧330）
  - バグではなく、クリーンアップの TODO。変化なし
- pkg/compose/create.go:1383（旧1247）
  - DriverConfig がモデルにない。compose-spec側の制約で変化なし
- .github/workflows/merge.yml:26（新規発見）
  - reusable workflowのinputsで`env`オブジェクトを直接参照できないGitHub Actionsの制約
  - job outputに経由させる回避策で対応済み。upstream側でも対応不可な外部プラットフォーム制約

## 解消済み（リストから削除）

- ~~pkg/compose/create.go:1619（コンテナ削除とanonymous volume継承の競合）~~
  - upstreamの大規模リファクタで`convergence`構造体が削除され`reconcile.go`に置き換わった
  - 同種の問題（名前付きvolume再作成時にanonymous volumeが失われる）は今も残っているが、FIXMEマーカーは外れ、reconcile.go:65-69あたりの設計コメントに格下げされている
- ~~pkg/compose/watch.go:729/787（FIXME .dockerignore）~~
  - **修正・マージ済み**: https://github.com/docker/compose/pull/14117
  - 調査の結果、単なる残骸コメントではなく、2025-01のリファクタ(ed10804e0)でDockerfile/compose.yaml除外matcherが代替なしに削除された実害のあるリグレッションだった
  - 初回PRの後、レビュー対応でフォローアップ修正も追加された(コミット8ddbdc41a: `cli.DefaultFileNames`/`DefaultOverrideFileNames`やカスタムDockerfile名にも対応する、より正確なパターンに改善)
- ~~pkg/e2e/build_test.go:176（ブロッキング Issue moby/buildkit#5558 解消済みで再有効化可能）~~
  - **修正・マージ済み**: https://github.com/docker/compose/pull/14120
  - `TestBuildSSH`内で無効化されていた「間違ったsshキーID」サブテストを再有効化。issueは2024年12月に解消済みで、現行vendor済みbuildkit(v0.26.3)には問題なし