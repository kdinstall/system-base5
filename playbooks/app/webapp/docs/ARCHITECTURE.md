# アーキテクチャ設計

## ディレクトリ構造

```
webapp/
├── docs/                    # ドキュメント
│   ├── README.md
│   ├── OVERVIEW.md
│   ├── ARCHITECTURE.md     # このファイル
│   ├── API.md
│   ├── VARIABLES.md
│   ├── PLAYBOOK_GUIDE.md
│   └── DEPLOYMENT.md
├── src/
│   ├── main.go             # エントリーポイント
│   ├── router.go           # ルーティング定義
│   ├── config/
│   │   └── config.go       # 環境変数読み込み
│   ├── controllers/
│   │   ├── containerController.go    # コンテナ操作
│   │   └── installController.go      # Playbook実行、ジョブ管理
│   ├── lib/
│   │   ├── ansible/
│   │   │   └── ansible.go            # Ansible CLI wrapper（同期・非同期）
│   │   ├── docker/
│   │   │   └── docker.go             # Docker CLI wrapper
│   │   ├── job/
│   │   │   └── manager.go            # ジョブ管理（シングルトン）
│   │   ├── playbook/
│   │   │   └── manager.go            # Playbook管理
│   │   └── template/
│   │       ├── template.go           # テンプレート読み込み
│   │       └── data.go               # 共通データ
│   ├── templates/
│   │   ├── header.html               # 共通ヘッダー
│   │   ├── footer.html               # 共通フッター
│   │   ├── 404.html                  # エラーページ
│   │   ├── containers.html           # コンテナ一覧
│   │   ├── container_logs.html       # ログ表示
│   │   ├── install.html              # インストール画面
│   │   ├── install_config.html       # 環境変数設定画面
│   │   ├── job_list.html             # ジョブ一覧
│   │   └── job_detail.html           # ジョブ詳細（リアルタイムログ）
│   └── style/
│       └── input.css                 # Tailwind CSS入力
├── public/
│   └── assets/
│       ├── css/
│       │   └── style.css             # ビルド済CSS
│       └── js/
│           └── preline.js            # Preline UI
├── go.mod                            # Go依存関係
├── package.json                      # Node.js依存関係
├── Makefile                          # ビルドタスク
└── README.md
```

## レイヤー構造

### 1. プレゼンテーション層（Presentation Layer）

**責務**: HTTPリクエスト/レスポンスの処理、HTMLレンダリング

#### Controllers
- **ContainerController**: コンテナ操作のエンドポイント
  - `Index()`: コンテナ一覧表示
  - `Start()`: コンテナ起動
  - `Stop()`: コンテナ停止
  - `Restart()`: コンテナ再起動
  - `Logs()`: ログ表示

- **InstallController**: Playbookインストールのエンドポイント
  - `Index()`: Playbook一覧表示
  - `Config()`: 環境変数設定画面表示
  - `Execute()`: Playbook非同期実行
  - `JobList()`: ジョブ一覧表示
  - `JobDetail()`: ジョブ詳細・ログ表示
  - `JobLogs()`: ジョブログ増分取得（JSON）
  - `JobStatus()`: ジョブステータス取得（JSON）
  - `GetRunningJob()`: 実行中ジョブ取得（JSON）

#### Templates
- Go標準の`html/template`を使用
- 共通レイアウト（header/footer）の再利用
- データバインディングでの動的コンテンツ生成

### 2. ビジネスロジック層（Business Logic Layer）

**責務**: ドメインロジック、外部コマンド実行の抽象化

#### Libraries

**job パッケージ**
```go
type Job struct {
    ID         string
    Name       string
    Status     string // pending, running, completed, failed
    StartTime  time.Time
    EndTime    time.Time
    Logs       []string
    Error      string
    CancelFunc context.CancelFunc
    mu         sync.Mutex
}

type JobManager struct {
    jobs map[string]*Job
    mu   sync.RWMutex
}

func GetManager() *JobManager // シングルトン
func (jm *JobManager) CreateJob(id, name string) *Job
func (jm *JobManager) GetJob(id string) (*Job, bool)
func (jm *JobManager) GetAllJobs() []*Job
func (jm *JobManager) GetRunningJob() *Job
func (jm *JobManager) DeleteJob(id string)
func (j *Job) AppendLog(line string)
func (j *Job) UpdateStatus(status string)
func (j *Job) SetError(err string)
func (j *Job) GetLogs(offset int) []string
```

**ansible パッケージ**
```go
type PlaybookResult struct {
    Success bool
    Output  string
    Error   string
}

func RunPlaybookWithConnection(
    playbookPath string, 
    connection string, 
    extraVars []string
) *PlaybookResult

func RunPlaybookAsync(
    jobID string,
    playbookPath string,
    connection string,
    extraVars []string
) error
```

**docker パッケージ**
```go
type Container struct {
    ID      string
    Name    string
    Image   string
    State   string
    Status  string
    Ports   []string
}

func ListContainers() ([]Container, error)
func StartContainer(id string) error
func StopContainer(id string) error
func RestartContainer(id string) error
func GetContainerLogs(id string, lines int) (string, error)
```

**playbook パッケージ**
```go
type Variable struct {
    Name     string
    Label    string
    Type     string
    Default  string
    Required bool
    Help     string
}

type PlaybookInfo struct {
    Name        string
    Path        string
    Description string
    IsLocal     bool
    Variables   []Variable
}

func ListLocalPlaybooks(basePath string) ([]PlaybookInfo, error)
func ReadVariables(dir string) ([]Variable, error)
func ValidatePlaybookExists(basePath, name string) error
```

### 3. インフラストラクチャ層（Infrastructure Layer）

**責務**: 外部システムとの実際の通信

#### Docker CLI Integration
```go
// os/exec.Commandでdockerコマンドを実行
cmd := exec.Command("docker", "ps", "-a", "--format", "{{json .}}")
```

#### Ansible CLI Integration
```go
// ansible-playbookコマンドを実行
cmd := exec.Command("ansible-playbook", 
    "-i", "localhost,", 
    "--connection", "local", 
    playbookPath,
    "-e", "key=value")
```

## データフロー

### コンテナ一覧表示の流れ

```
1. ブラウザ
   ↓ GET /containers
2. Gin Router
   ↓ 
3. ContainerController.Index()
   ↓
4. docker.ListContainers()
   ↓ exec.Command("docker ps -a")
5. Docker Engine
   ↓ JSON出力
6. Container構造体にパース
   ↓
7. テンプレートにデータ渡し
   ↓ containers.html
8. HTML生成
   ↓ HTTP Response
9. ブラウザ表示
```

### Playbookインストールの流れ（非同期・環境変数あり）

```
1. ブラウザ
   ↓ GET /install
2. InstallController.Index()
   ↓
3. playbook.ListLocalPlaybooks()
   ├─ main.yml読み込み
   └─ variables.yml読み込み (ReadVariables)
   ↓
4. install.html表示（Playbook一覧）
   ↓ ユーザーが「設定してインストール」クリック
5. GET /install/:name/config
   ↓
6. InstallController.Config()
   ↓ playbook.ReadVariables()
7. install_config.html表示（入力フォーム）
   ↓ ユーザーが環境変数入力
8. POST /install/execute
   ↓ 
9. InstallController.Execute()
   ├─ 実行中ジョブチェック (GetRunningJob)
   ├─ ジョブID生成 (UUID)
   ├─ ジョブ作成 (jobManager.CreateJob)
   ├─ extraVars配列構築
   └─ ansible.RunPlaybookAsync()
      ├─ ゴルーチン起動
      ├─ context.WithCancel()
      ├─ exec.Command("ansible-playbook -e key=value")
      ├─ StdoutPipe/StderrPipe読み取り
      └─ ログを1行ずつ job.AppendLog()
10. 303 Redirect → /install/jobs/:id
11. job_detail.html表示
    ├─ 初期ログ表示
    └─ JavaScript 1秒ごとポーリング
       ├─ GET /install/jobs/:id/status
       ├─ GET /install/jobs/:id/logs?offset=N
       └─ ステータスが完了/失敗まで継続
12. Ansible Engine (バックグラウンド)
    ├─ Playbookタスク実行
    ├─ Dockerコンテナ作成
    └─ ジョブステータス更新 (completed/failed)
```

### ジョブ管理フロー

```
1. ジョブ一覧画面アクセス
   ↓ GET /install/jobs
2. InstallController.JobList()
   ↓ jobManager.GetAllJobs()
3. job_list.html表示
   ├─ ジョブID、名前、ステータス
   ├─ 開始・終了時刻
   └─ 詳細リンク

4. ジョブ詳細画面アクセス
   ↓ GET /install/jobs/:id
5. InstallController.JobDetail()
   ↓ jobManager.GetJob(id)
6. job_detail.html表示
   ├─ 初期ログレンダリング（{{ range .job.Logs }}）
   └─ JavaScript自動更新
      ├─ 1秒ごとにステータス取得
      ├─ 新規ログを増分取得（offset指定）
      ├─ DOM更新（ログ追記）
      └─ completed/failed で停止

7. 実行中ジョブ確認
   ↓ GET /install/jobs/running
8. InstallController.GetRunningJob()
   ↓ jobManager.GetRunningJob()
9. JSON レスポンス
   ├─ running: true/false
   └─ job: { id, name, status }

10. インストール画面
    ├─ JavaScript 3秒ごとにポーリング
    ├─ 実行中ジョブがあればボタン無効化
    └─ 警告メッセージ表示
```

## コンポーネント設計

### 1. Router (router.go)

**責務**: URLパスとコントローラーのマッピング

```go
func initRouter() *gin.Engine {
    router := gin.Default()
    
    // テンプレートロード
    t, _ := tmpl.LoadTemplates("src/templates")
    router.SetHTMLTemplate(t)
    
    // 静的ファイル
    router.Static("/assets", "public/assets")
    
    // ルーティング
    registerContainerRouter(router)
    registerInstallRouter(router)
    
    return router
}
```

### 2. Config (config/config.go)

**責務**: 環境変数の読み込みと検証

```go
type Env struct {
    AppName      string
    ServerPort   string
    PlaybooksDir string
}
```

デフォルト値:
- `AppName`: "Docker管理"
- `ServerPort`: "58080"
- `PlaybooksDir`: "../containers"

### 3. Template Loader (lib/template/template.go)

**責務**: テンプレートの一括読み込み

```go
func LoadTemplates(dir string) (*template.Template, error) {
    pattern := filepath.Join(dir, "*.html")
    return template.ParseGlob(pattern)
}
```

### 4. Data Merger (lib/template/data.go)

**責務**: ページデータと共通データのマージ

```go
func BaseData() gin.H {
    return gin.H{
        "app_name": config.GetEnv().AppName,
        "g_year":   time.Now().Year(),
    }
}

func MergeData(data gin.H) gin.H {
    base := BaseData()
    for k, v := range data {
        base[k] = v
    }
    return base
}
```

## セキュリティ設計

### 1. 入力検証

**Playbook名のサニタイズ**
```go
// ディレクトリトラバーサル対策
if strings.Contains(playbookName, "..") {
    return error
}
```

**環境変数の検証**
```go
// env_プレフィックスでフィルタリング
for key, values := range c.Request.PostForm {
    if strings.HasPrefix(key, "env_") {
        // 処理
    }
}
```

### 2. 機密情報の保護

**Ansibleログでの非表示化**
```yaml
# main.yml
tasks:
  - name: Run container
    command: docker run ...
    no_log: "{{ db_password != '' or secret_key != '' }}"
```

**HTMLでのマスク表示**
```html
<input type="password" name="env_db_password" />
```

### 3. コマンドインジェクション対策

**exec.Commandの安全な使用**
```go
// ❌ 危険: シェル経由での実行
cmd := exec.Command("sh", "-c", "docker ps | grep " + input)

// ✅ 安全: 引数を分離
cmd := exec.Command("docker", "ps", "-a", "--filter", "name="+input)
```

## エラーハンドリング

### エラー処理の層別戦略

**1. Infrastructure層**
- エラーを返すのみ、ログ出力は上位層
```go
func ListContainers() ([]Container, error) {
    output, err := cmd.CombinedOutput()
    if err != nil {
        return nil, fmt.Errorf("docker ps failed: %v", err)
    }
    // ...
}
```

**2. Business Logic層**
- エラーメッセージの加工、リトライロジック
```go
func ReadVariables(dir string) ([]Variable, error) {
    if _, err := os.Stat(variablesPath); os.IsNotExist(err) {
        return []Variable{}, nil  // エラーでなく空配列を返す
    }
    // ...
}
```

**3. Presentation層**
- エラーの表示、ユーザーへのフィードバック
```go
func (ic *InstallController) Index(c *gin.Context) {
    playbooks, err := playbook.ListLocalPlaybooks(basePath)
    if err != nil {
        c.HTML(http.StatusOK, "install.html", tmpl.MergeData(gin.H{
            "error": "Playbook一覧の取得に失敗しました: " + err.Error(),
        }))
        return
    }
    // ...
}
```

## パフォーマンス最適化

### 1. テンプレートの事前読み込み
- アプリ起動時に全テンプレートをパース
- リクエストごとのパース不要

### 2. 静的ファイルの効率的配信
- Ginの`router.Static()`で直接配信
- CSSは本番ビルドで最小化

### 3. Dockerコマンドの最適化
```go
// JSON形式で一度に全情報取得
docker ps -a --format '{{json .}}'
```

## スケーラビリティ

### 現在の制限
- シングルプロセス
- 同時実行制御なし
- セッション管理なし

### 将来の拡張案
- ワーカープール（Playbook並列実行）
- Redis（セッション管理）
- データベース（履歴管理）
- Prometheus（メトリクス収集）
