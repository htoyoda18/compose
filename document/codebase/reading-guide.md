# コードの読み方

このリポジトリのコードをどの順で読むか。各層の役割は [architecture.md](./architecture.md) を参照。

← [document 一覧](../README.md)

## Phase 1: エントリーポイントの理解

1. cmd/main.go (86 行)
   ↓ main() → pluginMain() → RootCommand()
2. cmd/compose/compose.go:424 (RootCommand)
   ↓ 全サブコマンドを登録
3. cmd/compose/up.go (upCommand)
   ↓ 1 つのコマンドを深堀り
4. pkg/api/api.go
   ↓ Service インターフェース定義を確認
5. pkg/compose/compose.go
   ↓ Service インターフェースの実装

## Phase 2: 1 つのコマンドを完全に理解

docker compose up から始める

## Phase 3: 横展開 - 他のコマンド

優先度 高:
├─ down.go # up の逆、削除ロジック
├─ build.go # Buildkit 連携
├─ logs.go # ストリーミング処理
└─ run.go # 一時コンテナ

優先度 中:
├─ exec.go # 実行中コンテナ操作
├─ ps.go # 状態取得
└─ scale.go # レプリカ制御

## Phase 4: 深い部分を理解

Docker API 連携
Compose ファイルパース
並行制御・同期
