# PR を出す前のチェックリスト

[pr-trends.md](./pr-trends.md)（プロジェクト全体の PR の傾向）と
[pr-review-patterns.md](./pr-review-patterns.md)（自分の PR に付いたレビュー指摘）から引き出した、自分用のチェック項目。

← [contributing 一覧](./README.md) ／ [document 一覧](../README.md)

## 出す前（何を直すか）

[pr-trends.md](./pr-trends.md) の調査から。

- [ ] 対象の issue に `status/approved` が付いているか（無ければ PR を出す前に issue で相談する）
- [ ] 過去に同じ提案が却下されていないか（issue を検索して重複を確かめる）
- [ ] **最新の `upstream/main` で**バグを再現できたか。再現手順とエラーの実物を説明文に貼れるか
- [ ] 症状を抑えるのではなく、根本原因の場所を直しているか（compose-go や docker/cli 側で直すべきものではないか）
- [ ] plan engine / provider など、メンテナが今まさに書き換えている領域とぶつかっていないか

## コードとテスト

[pr-review-patterns.md](./pr-review-patterns.md) の ①〜⑧ に対応。

- [ ] ① この変更が効く/効かないコードパスを、呼び出し元まで遡って確認したか
- [ ] ② 同じ問題を解決している既存の仕組み・定数・慣習が無いか探したか
- [ ] ③ 戻り値は「呼び出し側が本当に欲しい答え」になっているか
- [ ] ④ 追加した出力/警告は、周辺コードと同じ経路(イベントバス/`dockerCli.Out()`/`logrus`)を使っているか
- [ ] ⑤ 修正対象の値がユーザー設定で上書き可能な場合、その一般形まで対応したか
- [ ] ⑥ 追加したテストケースは、それぞれ別の分岐を通っているか
- [ ] ⑦ リファクタ後、不要になった引数・戻り値・分岐が残っていないか
- [ ] ⑧linter 対応のコードは、本当に意味のあるエラーハンドリングになっているか
- [ ] 不具合の再現は e2e の `Scenario` DSL で書いたか（[pkg/e2e/SCENARIO.md](../../pkg/e2e/SCENARIO.md)）

## コミットと PR

[pr-trends.md](./pr-trends.md) の調査から。

- [ ] コミットは論理単位ごとに squash したか（基本は 1 コミット）。`rebase main` で追従したか
- [ ] コミットメッセージは `type(scope): 修正後の振る舞い` の形か。本文に「なぜ」を書いたか。`-s` で sign-off したか
- [ ] AI の `Co-Authored-By` トレーラーを付けていないか（AI_POLICY.md。AI 利用の開示は PR の説明に書く）
- [ ] PR テンプレートを埋めたか（Fixes #、Testing done、AI tool used）
- [ ] `AI_AGENT_DISCLOSURE.md` を、自分でレビューした上で削除したか
- [ ] docker-agent / copilot のレビューコメントにも対応したか
