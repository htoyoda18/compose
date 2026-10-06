# コントリビューション

docker/compose に PR を出すための情報。

← [document 一覧](../README.md)

| ファイル | 内容 |
| --- | --- |
| [checklist.md](./checklist.md) | **PR を出す前のチェックリスト**（下の 2 つの調査から引き出したもの） |
| [pr-trends.md](./pr-trends.md) | 直近のクローズ済み PR 150 件と main のコミット履歴の調査。どんな PR がマージされ／閉じられているか、コミットの積み方、CONTRIBUTING / AI_POLICY のルール |
| [pr-review-patterns.md](./pr-review-patterns.md) | 自分が出した PR のレビュー指摘の傾向分析と対策 |

着手するタスクの候補は [tracking/next-tasks.md](../tracking/next-tasks.md) を参照。

## OSS コントリビューションの基本作法

- **Git / GitHub ワークフロー**
  - フォーク・ブランチ運用、Conventional Commits 的な粒度でのコミット分割。
  - DCO（Signed-off-by）や CLA など、OSS プロジェクトごとのコントリビューション規約。
  - PR テンプレート、レビュー対応、CI（golangci-lint・テスト）通過までの一連の流れ。
- **このリポジトリのコントリビューションフロー（DCO, PR テンプレート）**
  - コミットに DCO（Developer Certificate of Origin、`Signed-off-by`）を必須とし、`git commit -s` での署名を求める。
  - `CONTRIBUTING.md` にコーディング規約・PR の出し方等がまとめられている。
  - レビュー・CI 通過・メンテナによる approve を経て main にマージされる、典型的な GitHub Flow。
  - 実際の運用の詳細は [pr-trends.md](./pr-trends.md)（§5 コミットの積み方 / §6 PR の説明文 / §7 レビューの流れ / §8 ルール）
