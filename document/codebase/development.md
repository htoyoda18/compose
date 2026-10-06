# 開発・運用

ビルド・テスト・Lint・CI・リリースの手順と仕組み。PR の出し方は [contributing/](../contributing/README.md) を参照。

← [document 一覧](../README.md)

## ビルド

```bash
# 基本ビルド
make build

# クロスコンパイル（11プラットフォーム対応）
make cross

# Dockerビルド
docker buildx bake binary
```

### インストール

```bash
# ローカルインストール（~/.docker/cli-plugins/にインストール）
make install
```

### 対応プラットフォーム

- darwin/amd64, darwin/arm64
- linux/amd64, linux/arm/v6, linux/arm/v7, linux/arm64
- linux/ppc64le, linux/riscv64, linux/s390x
- windows/amd64, windows/arm64

計 11 プラットフォーム

## テスト

```bash
# ユニットテスト
make test

# E2Eテスト（プラグインモード）
make e2e-compose

# E2Eテスト（スタンドアロンモード）
make e2e-compose-standalone

# 全テスト
make e2e
```

### E2E テスト（Scenario DSL）

- `pkg/e2e/scenario.go` の `NewScenario` を使い、「意図・インライン compose 定義・コマンドと期待結果のステップ」を宣言的に書く DSL。
- 状態ベースの検証を優先し、文字列出力の部分一致（`OutputContains`）は最後の手段とする方針（`pkg/e2e/SCENARIO.md` 参照）。
- `E2E_KEEP_FAILED=1` で失敗時の成果物を保持でき、デバッグに使える。

## 開発ワークフロー

```bash
# コード検証
make validate  # lint + vendor-validate + headers + docs

# プリコミットチェック
make pre-commit  # validate + lint + build + test + e2e
```

## Lint / CI（golangci-lint, GitHub Actions）

- `golangci-lint` v2（`.golangci.yml`）で gofumpt・gci 等複数のリンタをまとめて実行し、フォーマットとコード品質を担保する。
- `docker buildx bake lint` で Dockerfile に固定されたバージョンの lint をローカル/CI で再現できる。
- GitHub Actions で以下を実行する CI/CD パイプラインが組まれている:
  - **validate**: lint、vendor、headers、docs の検証
  - **binary**: マルチプラットフォームバイナリビルド
  - **test**: ユニットテストとカバレッジ
  - **e2e**: E2E テスト

## リリース・バージョニング

- タグ付けによるセマンティックバージョニングでリリースする、一般的な OSS の運用。
- クロスコンパイル（`make cross`）で 11 プラットフォーム分のバイナリを一括ビルドし、GitHub Releases 等で配布する。
- スタンドアロン（`docker-compose`）とプラグイン（`docker compose`）の両モードで同じバイナリを配布している。
