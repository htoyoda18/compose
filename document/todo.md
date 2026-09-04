## TODO コメント調査状況

- `cmd/cmdtrace/cmd_span.go:113` ✅ 対応済み（[PR #14152](https://github.com/docker/compose/pull/14152)、Merged）
  - OTel コンポーネントのデバッグログを有効にする環境変数が無い、という機能追加系 TODO。バグではなく実害は無い。当初は専用の`COMPOSE_OTEL_DEBUG`環境変数を実装したが、レビュー指摘を受けて既存の`--debug`/`-D`（logrus経由）に統合。あわせて`docker/cli`の`plugin.Run()`が自パッケージの`otel.ErrorHandler`初期化を上書きしていた問題も解消。
- `cmd/compose/options.go:95`
  - `applyPlatforms`（`buildForSinglePlatform=true`）で呼ばれる `create`/`up`/`run`/`config`/`watch` が対象。`DOCKER_DEFAULT_PLATFORM`未設定・`service.platform`未指定・`build.platforms`に 2 つ以上指定、の 3 条件が揃うと `Build.Platforms` が `nil` にクリアされ、ビルダーが選ぶプラットフォームが宣言済みリストに含まれているかの検証が行われないまま素通りする実バグ。`build_classic.go:234`と`build_bake.go:346`のガードは`len(Platforms)>1`しか見ておらず`nil`は防げないことを確認済み。根本修正にはビルダーが実際に選ぶプラットフォームの事前予測が必要で難易度は中〜高（ホストの GOOS/GOARCH や daemon 情報で代用するヒューリスティックが現実的な落とし所）。
- `internal/oci/push.go:125` ✅ 対応済み（[PR #14146](https://github.com/docker/compose/pull/14146)、OPEN）
  - OCI 1.1→1.0 フォールバック時に警告を出したいが、`internal/oci`パッケージは CLI 層の`logrus`に依存させたくない設計のため出せていなかった。`PushManifest`が実際に使われた`api.OCIVersion`を返すように変更し、呼び出し元の`pkg/compose/publish.go`（既に`logrus`に依存）でフォールバック時に`logrus.Warn`するよう修正。
- `pkg/compose/build_classic.go:335`
  - `thaJeztah`（docker/cli メンテナ）による設計上の疑問で、`docker/cli`の PR ディスカッションにリンクされている。Auth/IdentityToken/RegistryToken をレガシービルダー用 authconfig に含めるべきか未確定であり、upstream 側の結論待ち。このリポジトリ単独では判断・対応不可。
- `pkg/compose/create.go:575`
  - `docker/cli`からコピペしたセキュリティオプションのパース処理で、共通化する方法を探したいという DRY 違反の指摘。現状動作に問題はないが保守負債。共有するには`docker/cli`側が該当ロジックを公開 API として切り出す必要があり、技術難易度よりもリポジトリを跨いだ調整コストが高い。
