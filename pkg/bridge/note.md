[root](../../note.md) / [pkg](../note.md) / bridge

- [convert.go](convert.go.md)
  - `docker compose bridge convert`。compose プロジェクトを transformer イメージに渡し、Kubernetes manifest 等の別形式へ変換する
- convert_test.go
  - テスト
- [transformers.go](transformers.go.md)
  - `docker compose bridge transformations`。transformer イメージの雛形(templates + Dockerfile)生成と一覧取得
