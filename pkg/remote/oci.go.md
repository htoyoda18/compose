[root](../../note.md) / [pkg](../note.md) / [remote](note.md) / oci.go

- 責務
  - OCI イメージからの Compose プロジェクト読み込み
- ociRemoteLoader
  - OCI レジストリから Compose ファイルをプル
  - Accept
    - oci:// プレフィックスを持つか判定
  - Load
    - OCI アーティファクトをプルしてファイルを取得
  - Dir
    - ローカルキャッシュパスを返す