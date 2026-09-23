- relay/(独立したGoモジュール。`go.mod`/`Dockerfile`を持ち`pkg/compose/relay.go`とは別ビルド)
  - `compose-relay`。`provider:`型サービス(外部クラウドDB等)をComposeネットワーク上の
    通常コンテナのように見せるためのTCPゲートウェイ。`docker/compose-relay`イメージとして
    `pkg/compose/relay.go`からコンテナ起動される
- main.go
  - `RELAY_ROUTES=port=host:port[,...]`環境変数からルートを読み、ポートごとに
    listenして`upstream`へTCP転送するだけのミニマムなバイナリ(`scratch`イメージ)
  - `parseRoutes`はloopback/unspecifiedなhostを`host.docker.internal`に書き換え、
    providerがホスト視点で報告したアドレスをコンテナから到達可能にする
  - `forward`は双方向`io.Copy`+`CloseWrite`で半クローズを伝播し、SIGTERM時は
    listenerだけ閉じてin-flightの接続はhalfCloseIdleTimeout(60s)まで排出を続ける
