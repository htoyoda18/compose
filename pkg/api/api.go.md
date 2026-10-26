[root](../../note.md) / [pkg](../note.md) / [api](note.md) / api.go

# pkg/api/api.go

Compose の公開 API の本体(918 行)。`Compose` インターフェースと、そのメソッドに渡すオプション・返ってくる結果の型を定義する。
実装は無い。実装は `pkg/compose` の `composeService` で、`compose.NewComposeService` で手に入る(`api.go:92-93`)。

## ファイルの中の並び

| 行 | 中身 |
| -- | ---- |
| 43-90 | プロジェクトの読み込み: `LoadListener`、`ProjectLoadOptions`、`OCIOptions` |
| 92-206 | **`Compose` インターフェース**と、それを構成する 6 つのインターフェース |
| 208-221 | `UnsupportedAttribute`(Swarm 以外では効かない compose ファイルの属性の報告) |
| 223-808 | 各メソッドの `XxxOptions` と結果の型 |
| 810-832 | 定数(スタックの状態、作り直しの方針) |
| 834-890 | `Stack`、ログとコンテナのイベント(`LogConsumer`、`ContainerEvent`) |
| 892-918 | イメージ名のヘルパー |

## Compose インターフェース

`Compose` は、用途ごとの 6 つのインターフェースを束ねたもの(`api.go:102-109`)。メソッドの一覧は [note.md の表](note.md#compose-インターフェースの中身) を参照。

メソッドの第 2 引数は 2 種類ある。

| 引数 | 使うメソッド | 意味 |
| ---- | ------------ | ---- |
| `project *types.Project` | Create / Up / Scale / Build / Push / Pull / Publish / Watch / Viz / RunOneOffContainer | compose ファイルを読み込んだ結果が必要な操作 |
| `projectName string` | Start / Stop / Down / Ps / Logs / Exec など | 既にあるコンテナを、ラベルのプロジェクト名で探せば済む操作。オプションの `Project` は nil でもよい(`docker compose -p NAME down` のように compose ファイル無しで動かせる) |

## 読むときに大事な約束

コメントに書かれている、型だけでは分からない約束。

1. **操作の対象は、渡したプロジェクト全体**(`api.go:95-101`)
   - 一部のサービスだけを対象にするなら、呼ぶ前にプロジェクトを絞る。読み込みのときに `ProjectLoadOptions.Services` で絞るか、`WithSelectedServices` を使う(`api.go:57-61`)
   - **オプションの `Services` フィールドは絞り込みではない**。利用者が名前を挙げたサービスを示すだけ。例えば `CreateOptions.Services` は、挙げたサービスに `Recreate`、それ以外に `RecreateDependencies` の方針を使い分けるためにある(`api.go:341-346`)
2. **`UpOptions` は `CreateOptions` と `StartOptions` を並べただけ**(`api.go:431-434`)。`up` = create + start
3. **`StartOptions` の多くのフィールドは、`Up` の前景の実行でしか使われない**(`api.go:366-398`)
   - `Attach` が nil かどうかで、`Up` が前景で動くか(ログを流し続ける)、起動したらすぐ戻るか(`-d`)が決まる
   - `OnExit` / `ExitCodeFrom` / `Watch` / `NavigationMenu` は `Up` の前景でだけ効き、`Start` は無視する
   - `AttachTo` は 2 つの意味を持つ。`Start` では、プロジェクトをラベルから作り直すときのサービス名。`Up` の前景では、ログを流すサービス
4. **`RunOptions` は `Run` と `Exec` で共用している**(`api.go:539-570`)。`Exec` が見るのはその一部(Service、Index、Command、Environment、WorkingDir、User、Privileged、Interactive、Tty、Detach)だけ。埋め込みの `CreateOptions` も、依存サービスを作るときに Build / IgnoreOrphans / RemoveOrphans / QuietPull しか使わない

## 主な型

### オプション(`XxxOptions`)

フラグとほぼ 1 対 1 で、`cmd/compose` がフラグを詰めて渡す。特に読む価値があるものだけ挙げる。

| 型 | 要点 |
| -- | ---- |
| `ProjectLoadOptions` | 読み込みの設定。`Services` で対象を絞る唯一の場所。`OnUnsupportedAttribute` を設定したときだけ、効かない属性の検出が走る |
| `BuildOptions` | ビルドの設定。`Apply` でプロジェクトに反映する: `DOCKER_DEFAULT_PLATFORM` をサービスの platform にし、`--pull` / `--no-cache` を各サービスの build に足す。image も build も無いサービスはエラー(`api.go:308-336`) |
| `CreateOptions` | 作り直しの方針(`Recreate` / `RecreateDependencies`)、匿名ボリュームを引き継ぐか(`Inherit`)、停止のタイムアウト |
| `StartOptions` | 上の「約束 3」を参照 |
| `DownOptions` | `Volumes`(ボリュームも消す)、`Images`(`ImagePruneMode`。"" / "local" / "all" 以外は、何かを消す前にエラー) |
| `ScaleOptions` | `Replicas` が台数。`Services` は作り直してよいサービスで、それ自体は台数を変えない |
| `RunOptions` | 上の「約束 4」を参照 |

### 結果の型

| 型 | 返すメソッド | 中身 |
| -- | ------------ | ---- |
| `ContainerSummary` | Ps | コンテナの ID・名前・イメージ・状態・ヘルス・公開ポート(`PortPublishers`)・ラベルなど |
| `PortPublisher(s)` | Ports、Ps | 公開ポート 1 つ(`80/tcp -> 0.0.0.0:8080` の形で表示)。`PortPublishers` は並べ替えのための型 |
| `Stack` | List | プロジェクト 1 つの名前・状態・compose ファイル |
| `ImageSummary` / `ContainerProcSummary` / `VolumesSummary` | Images / Top / Volumes | それぞれのコマンドの 1 行分 |

### ログとイベント

- `LogConsumer`: コンテナのログを受け取る側のインターフェース(`Log` = 標準出力、`Err` = 標準エラー、`Status`)。`cmd/formatter` が実装して画面に出す
- `ContainerEvent`: 前景の `up` が監視するコンテナの出来事。`Type` は次のどれか(`api.go:871-890`)

| 定数 | 意味 |
| ---- | ---- |
| `ContainerEventLog` / `ContainerEventErr` | 標準出力 / 標準エラーの 1 行(`Line` に入る) |
| `ContainerEventCreated` / `ContainerEventStarted` / `ContainerEventRestarted` | 作成 / 起動 / 再起動 |
| `ContainerEventStopped` / `ContainerEventExited` | 停止 / 終了(`ExitCode` に終了コード) |
| `ContainerEventRecreated` | 作り直しのために止まった |
| `HookEventLog` | フックのログ |

## 定数

| 定数 | 値 | 使い道 |
| ---- | -- | ------ |
| `RecreateDiverged` / `RecreateForce` / `RecreateNever` | "diverged" / "force" / "never" | `CreateOptions.Recreate` の値。設定が変わったときだけ作り直す / 必ず作り直す / 作り直さない |
| `CascadeIgnore` / `CascadeStop` / `CascadeFail` | 0 / 1 / 2 | `StartOptions.OnExit`。コンテナが終わったときにほかを止めるか |
| `STARTING` / `RUNNING` / `UPDATING` / `REMOVING` / `UNKNOWN` / `FAILED` | "Starting" など | `Stack.Status` |
| `OCIVersion1_0` / `OCIVersion1_1` | "1.0" / "1.1" | `publish` のマニフェストの形式。ECR には 1.0、それ以外は 1.1 を自動で使う(利用者は選べない) |
| `WatchLogger` | "#watch" | watch のログの送り主に使う予約名 |

## ヘルパー関数

- `GetImageNameOrDefault`: サービスのイメージ名。`image:` が無ければ `<プロジェクト名>-<サービス名>`(ビルドしたイメージのタグになる)
- `GetDependentImages`: サービス本体のイメージ以外に必要なイメージ。今は `pre_start` フックが別のイメージを使う場合だけ
