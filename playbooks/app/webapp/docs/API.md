# API仕様

## エンドポイント一覧

### コンテナ管理

#### GET `/`
トップページ（`/containers`へリダイレクト）

**レスポンス**: 302 Redirect

---

#### GET `/containers`
コンテナ一覧を表示

**コントローラー**: `ContainerController.Index()`

**レスポンス**: HTML (containers.html)

**データ**:
```go
{
    "page_title": "コンテナ一覧",
    "active_page": "containers",
    "containers": []Container{
        {
            ID: "abc123...",
            Name: "nginx-server",
            Image: "nginx:latest",
            State: "running",
            Status: "Up 2 hours",
            Ports: ["8080:80"]
        },
        // ...
    },
    "flash": "started" | "stopped" | "restarted" | ""
}
```

**エラー**:
- Docker実行失敗時: エラーメッセージを表示

---

#### POST `/containers/:id/start`
コンテナを起動

**コントローラー**: `ContainerController.Start()`

**パラメータ**:
- `id` (path): コンテナID

**レスポンス**: 303 Redirect to `/containers?flash=started`

**エラーレスポンス**: 303 Redirect to `/containers?flash=start_failed`

**実行コマンド**: `docker start {id}`

---

#### POST `/containers/:id/stop`
コンテナを停止

**コントローラー**: `ContainerController.Stop()`

**パラメータ**:
- `id` (path): コンテナID

**レスポンス**: 303 Redirect to `/containers?flash=stopped`

**エラーレスポンス**: 303 Redirect to `/containers?flash=stop_failed`

**実行コマンド**: `docker stop {id}`

---

#### POST `/containers/:id/restart`
コンテナを再起動

**コントローラー**: `ContainerController.Restart()`

**パラメータ**:
- `id` (path): コンテナID

**レスポンス**: 303 Redirect to `/containers?flash=restarted`

**エラーレスポンス**: 303 Redirect to `/containers?flash=restart_failed`

**実行コマンド**: `docker restart {id}`

---

#### GET `/containers/:id/logs`
コンテナのログを表示

**コントローラー**: `ContainerController.Logs()`

**パラメータ**:
- `id` (path): コンテナID

**レスポンス**: HTML (container_logs.html)

**データ**:
```go
{
    "page_title": "コンテナログ",
    "active_page": "containers",
    "container": Container{
        ID: "abc123...",
        Name: "nginx-server",
        // ...
    },
    "logs": "2026-04-12 10:00:00 [info] Server started\n..."
}
```

**エラー**:
- コンテナが存在しない場合: 404ページ

**実行コマンド**: `docker logs --tail 500 {id}`

---

### Playbookインストール

#### GET `/install`
インストール画面を表示（Playbook一覧）

**コントローラー**: `InstallController.Index()`

**レスポンス**: HTML (install.html)

**データ**:
```go
{
    "page_title": "コンテナインストール",
    "active_page": "install",
    "playbooks": []PlaybookInfo{
        {
            Name: "nginx",
            Path: "/path/to/playbooks/containers/nginx/main.yml",
            Description: "Nginxウェブサーバーをコンテナとして起動します。",
            IsLocal: true,
            Variables: []Variable{} // 空の場合もあり
        },
        {
            Name: "nodejs-webapp",
            Path: "/path/to/playbooks/containers/nodejs-webapp/main.yml",
            Description: "Node.js Webアプリケーションをコンテナとして起動します。",
            IsLocal: true,
            Variables: []Variable{
                {
                    Name: "container_name",
                    Label: "コンテナ名",
                    Type: "text",
                    Default: "node-webapp",
                    Required: true,
                    Help: "Dockerコンテナの名前を指定してください"
                },
                // ...
            }
        }
    },
    "error": "",
    "flash": "installed" | ""
}
```

---

#### GET `/install/:name/config`
環境変数設定画面を表示

**コントローラー**: `InstallController.Config()`

**パラメータ**:
- `name` (path): Playbook名

**レスポンス**: HTML (install_config.html)

**データ**:
```go
{
    "page_title": "環境変数設定",
    "active_page": "install",
    "playbook_name": "nodejs-webapp",
    "variables": []Variable{
        {
            Name: "container_name",
            Label: "コンテナ名",
            Type: "text",
            Default: "node-webapp",
            Required: true,
            Help: "Dockerコンテナの名前を指定してください"
        },
        {
            Name: "db_password",
            Label: "データベースパスワード",
            Type: "password",
            Default: "",
            Required: false,
            Help: "データベース接続用のパスワード（機密情報）"
        }
        // ...
    ]
}
```

**エラー**:
- Playbookが存在しない場合: 404ページ
- variables.yml読み込み失敗: エラーメッセージ表示

---

#### POST `/install/execute`
Playbookを非同期実行してコンテナをインストール

**コントローラー**: `InstallController.Execute()`

**リクエストボディ** (application/x-www-form-urlencoded):

**パターン1: ローカルPlaybook（直接実行）**
```
playbook=nginx
```

**パターン2: ローカルPlaybook（環境変数あり）**
```
playbook=nodejs-webapp
env_container_name=my-webapp
env_container_port=3000
env_db_host=mysql-server
env_db_password=secret123
env_secret_key=mysecretkey
```

**パターン3: URLからダウンロード**
```
download_url=https://github.com/user/repo.git
download_type=git
```

**レスポンス**: 303 Redirect to `/install/jobs/:id`

**処理フロー**:
1. 実行中のジョブがないか確認（1つのジョブのみ実行可能）
2. Playbook名またはダウンロードURL取得
3. URL指定の場合、Git/HTTPでダウンロード
4. `env_` プレフィックスの環境変数を抽出
5. ジョブIDを生成（UUID）
6. ジョブを作成してジョブマネージャーに登録
7. 非同期で `ansible-playbook` コマンド実行
   ```bash
   ansible-playbook -i localhost, --connection local \
     /path/to/main.yml \
     -e "container_name=my-webapp" \
     -e "db_password=secret123"
   ```
8. ジョブ詳細画面へリダイレクト

**エラー**:
- 実行中のジョブがある場合: 422 Unprocessable Entity（エラーメッセージ表示）
- Playbookが見つからない: 422 Unprocessable Entity
- ダウンロード失敗: 422 Unprocessable Entity

---

### ジョブ管理

#### GET `/install/jobs`
インストールジョブ一覧を表示

**コントローラー**: `InstallController.JobList()`

**レスポンス**: HTML (job_list.html)

**データ**:
```go
{
    "page_title": "ジョブ一覧",
    "active_page": "install",
    "jobs": []Job{
        {
            ID: "550e8400-e29b-41d4-a716-446655440000",
            Name: "nginx",
            Status: "completed",
            StartTime: time.Time{...},
            EndTime: time.Time{...},
            Logs: []string{...},
            Error: ""
        },
        // ...
    }
}
```

**ステータス値**:
- `pending`: 待機中
- `running`: 実行中
- `completed`: 完了
- `failed`: 失敗

---

#### GET `/install/jobs/:id`
ジョブ詳細とリアルタイムログ表示

**コントローラー**: `InstallController.JobDetail()`

**パラメータ**:
- `id` (path): ジョブID（UUID）

**レスポンス**: HTML (job_detail.html)

**データ**:
```go
{
    "page_title": "ジョブ詳細",
    "active_page": "install",
    "job": Job{
        ID: "550e8400-e29b-41d4-a716-446655440000",
        Name: "nginx",
        Status: "running",
        StartTime: time.Time{...},
        EndTime: time.Time{...},
        Logs: []string{
            "PLAY [localhost] ***...",
            "TASK [Gathering Facts] ***...",
            // ...
        },
        Error: ""
    }
}
```

**機能**:
- JavaScript で1秒ごとにステータスとログをポーリング
- 新しいログを増分取得して表示
- ステータスが `completed` または `failed` になるまで自動更新

**エラー**:
- ジョブが存在しない場合: 404ページ

---

#### GET `/install/jobs/:id/logs`
ジョブのログを増分取得（JSON）

**コントローラー**: `InstallController.JobLogs()`

**パラメータ**:
- `id` (path): ジョブID
- `offset` (query): 取得開始位置（デフォルト: 0）

**レスポンス**: JSON

```json
{
  "logs": [
    "TASK [Install Docker] ***...",
    "ok: [localhost]",
    "PLAY RECAP ***..."
  ]
}
```

**使用例**:
```
GET /install/jobs/550e8400-e29b-41d4-a716-446655440000/logs?offset=10
```

offset=10 の場合、11行目以降のログを返す（JavaScriptで既に表示したログをスキップ）

**エラー**:
- ジョブが存在しない場合: 404 Not Found

---

#### GET `/install/jobs/:id/status`
ジョブのステータスを取得（JSON）

**コントローラー**: `InstallController.JobStatus()`

**パラメータ**:
- `id` (path): ジョブID

**レスポンス**: JSON

```json
{
  "status": "running",
  "start_time": "2026-04-20T10:30:00Z",
  "end_time": null,
  "error": ""
}
```

**ステータスが completed の例**:
```json
{
  "status": "completed",
  "start_time": "2026-04-20T10:30:00Z",
  "end_time": "2026-04-20T10:32:15Z",
  "error": ""
}
```

**ステータスが failed の例**:
```json
{
  "status": "failed",
  "start_time": "2026-04-20T10:30:00Z",
  "end_time": "2026-04-20T10:31:30Z",
  "error": "ansible-playbook command failed with exit code 2"
}
```

**エラー**:
- ジョブが存在しない場合: 404 Not Found

---

#### GET `/install/jobs/running`
現在実行中のジョブを取得（JSON）

**コントローラー**: `InstallController.GetRunningJob()`

**レスポンス**: JSON

**実行中のジョブがある場合**:
```json
{
  "running": true,
  "job": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "mysql",
    "status": "running"
  }
}
```

**実行中のジョブがない場合**:
```json
{
  "running": false,
  "job": null
}
```

**用途**:
- インストール画面でインストールボタンの有効/無効を制御
- 設定画面で実行中警告を表示

---

### 静的ファイル

#### GET `/assets/*filepath`
CSSやJavaScriptファイルを配信

**ディレクトリ**: `public/assets/`

**例**:
- `/assets/css/style.css`
- `/assets/js/preline.js`

---

### エラーページ

#### 404 Not Found
存在しないURLへのアクセス

**レスポンス**: HTML (404.html)

**データ**:
```go
{
    "page_title": "Not Found"
}
```

---

## データ構造

### Container
```go
type Container struct {
    ID      string   // コンテナID（先頭12桁）
    Name    string   // コンテナ名（例: "nginx-server"）
    Image   string   // イメージ名（例: "nginx:latest"）
    State   string   // 状態（"running" | "exited" | "paused" 等）
    Status  string   // ステータス詳細（例: "Up 2 hours"）
    Ports   []string // ポートマッピング（例: "8080:80"）
}
```

### PlaybookInfo
```go
type PlaybookInfo struct {
    Name        string     // Playbook名（ディレクトリ名）
    Path        string     // main.ymlへのフルパス
    Description string     // README.mdから抽出した説明
    IsLocal     bool       // ローカルPlaybook判定（常にtrue）
    Variables   []Variable // 環境変数定義（variables.ymlから読み込み）
}
```

### Variable
```go
type Variable struct {
    Name     string // 変数名（例: "container_name"）
    Label    string // 表示ラベル（例: "コンテナ名"）
    Type     string // 入力タイプ（"text" | "password" | "number"）
    Default  string // デフォルト値
    Required bool   // 必須フラグ
    Help     string // ヘルプメッセージ
}
```

### PlaybookResult
```go
type PlaybookResult struct {
    Success bool   // 実行成功フラグ
    Output  string // 標準出力
    Error   string // 標準エラー出力
}
```

---

## HTTPステータスコード

| コード | 説明 | 使用箇所 |
|--------|------|----------|
| 200 OK | 正常なレスポンス | GET /containers, GET /install 等 |
| 302 Found | 一時リダイレクト | GET / → /containers |
| 303 See Other | POST後のリダイレクト | POST /containers/:id/start 等 |
| 404 Not Found | リソースが見つからない | 存在しないコンテナのログ表示 |
| 422 Unprocessable Entity | 処理不可能なリクエスト | Playbook未検出、ダウンロード失敗 |
| 500 Internal Server Error | サーバーエラー | Docker/Ansible実行失敗（一部） |

---

## フラッシュメッセージ

クエリパラメータ`?flash=`でアラートを表示

### コンテナ操作
| 値 | 表示メッセージ | 色 |
|----|----------------|-----|
| `started` | "コンテナを起動しました。" | 緑 |
| `stopped` | "コンテナを停止しました。" | 黄 |
| `restarted` | "コンテナを再起動しました。" | 青 |
| `start_failed` | （未実装） | 赤 |
| `stop_failed` | （未実装） | 赤 |
| `restart_failed` | （未実装） | 赤 |

### Playbook実行
| 値 | 表示メッセージ | 色 |
|----|----------------|-----|
| `installed` | "インストールが完了しました。" | 緑 |

---

## エラーハンドリング

### Dockerコマンド失敗時
```go
containers, err := docker.ListContainers()
if err != nil {
    c.HTML(http.StatusInternalServerError, "containers.html", tmpl.MergeData(gin.H{
        "error": "コンテナ一覧の取得に失敗しました: " + err.Error(),
        "containers": nil,
    }))
    return
}
```

### Playbook実行失敗時
```go
result := ansible.RunPlaybookWithConnection(playbookPath, "local", extraVars)
// result.Success == false でも画面表示（エラーログを表示）
c.HTML(http.StatusOK, "install.html", tmpl.MergeData(gin.H{
    "result": result,
    "show_result": true,
}))
```

---

## セキュリティ考慮

### パスインジェクション対策
```go
// Playbook名にディレクトリトラバーサル文字列が含まれないかチェック
if strings.Contains(playbookName, "..") || strings.Contains(playbookName, "/") {
    return error
}
```

### 環境変数のサニタイズ
```go
// env_プレフィックスでフィルタリング
for key, values := range c.Request.PostForm {
    if strings.HasPrefix(key, "env_") && len(values) > 0 {
        varName := strings.TrimPrefix(key, "env_")
        // varNameにシェルメタ文字が含まれないか検証（推奨）
    }
}
```

### CSRFトークン
現バージョンでは未実装。将来的に追加推奨。

---

## パフォーマンス

### キャッシング
- テンプレートは起動時に一度だけパース
- 静的ファイルはブラウザキャッシュ推奨（将来実装）

### タイムアウト
- Dockerコマンド: デフォルトタイムアウトなし
- Ansibleコマンド: デフォルトタイムアウトなし
- 長時間実行されるPlaybookに注意

---

## 将来の拡張

### REST APIエンドポイント（計画中）
```
GET    /api/v1/containers          # JSON形式でコンテナ一覧取得
POST   /api/v1/containers/:id/start # JSON APIでコンテナ起動
GET    /api/v1/playbooks            # JSON形式でPlaybook一覧
POST   /api/v1/playbooks/execute    # JSON APIでPlaybook実行
```

### WebSocket（検討中）
```
WS /ws/logs/:id  # リアルタイムログストリーミング
WS /ws/install   # Playbook実行のリアルタイム進捗
```
