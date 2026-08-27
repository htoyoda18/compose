# 開発・運用

← [roadmap 一覧](./README.md) ／ [document 一覧](../README.md)

- **E2E テスト（Scenario DSL）**
  - `pkg/e2e/scenario.go` の `NewScenario` を使い、「意図・インライン compose 定義・コマンドと期待結果のステップ」を宣言的に書く DSL。
  - 状態ベースの検証を優先し、文字列出力の部分一致（`OutputContains`）は最後の手段とする方針（`pkg/e2e/SCENARIO.md` 参照）。
  - `E2E_KEEP_FAILED=1` で失敗時の成果物を保持でき、デバッグに使える。
- **Lint / CI（golangci-lint, GitHub Actions）**
  - `golangci-lint` v2（`.golangci.yml`）で gofumpt・gci 等複数のリンタをまとめて実行し、フォーマットとコード品質を担保する。
  - GitHub Actions で `validate`（lint/vendor/headers/docs）・`binary`・`test`・`e2e` 等のジョブを実行する CI/CD パイプラインが組まれている。
  - `docker buildx bake lint` で Dockerfile に固定されたバージョンの lint をローカル/CI で再現できる。
- **リリース・バージョニング**
  - タグ付けによるセマンティックバージョニングでリリースする、一般的な OSS の運用。
  - クロスコンパイル（`make cross`）で 11 プラットフォーム分のバイナリを一括ビルドし、GitHub Releases 等で配布する。
  - スタンドアロン（`docker-compose`）とプラグイン（`docker compose`）の両モードで同じバイナリを配布している。
- **コントリビューションフロー（DCO, PR テンプレート）**
  - コミットに DCO（Developer Certificate of Origin、`Signed-off-by`）を必須とし、`git commit -s` での署名を求める。
  - `CONTRIBUTING.md` にコーディング規約・PR の出し方等がまとめられている。
  - レビュー・CI 通過・メンテナによる approve を経て main にマージされる、典型的な GitHub Flow。
