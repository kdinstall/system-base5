# 環境変数の動的入力機能

## 概要

環境変数の動的入力機能は、Ansible Playbookのインストール時にDB接続情報やアプリケーション設定を柔軟に設定できる機能です。各Playbookに`variables.yml`ファイルを用意することで、Web画面から直感的に環境変数を入力できます。

## 主な特徴

- **動的フォーム生成**: variables.ymlの定義から入力フォームを自動生成
- **型安全**: text/password/number型に対応
- **デフォルト値**: 初期値の自動設定
- **バリデーション**: 必須チェック機能
- **後方互換性**: variables.ymlがないPlaybookも従来通り動作
- **セキュリティ**: password型はマスク表示、機密情報の保護

## variables.yml 形式

### 基本構造

```yaml
variables:
  - name: 変数名
    label: 画面に表示するラベル
    type: 入力タイプ（text | password | number）
    default: デフォルト値
    required: 必須フラグ（true | false）
    help: ヘルプメッセージ
```

### フィールド詳細

| フィールド | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| `name` | string | ✓ | Ansible変数名（英数字とアンダースコアのみ） |
| `label` | string | ✓ | 画面に表示するラベル（日本語可） |
| `type` | string | ✓ | 入力タイプ: `text`, `password`, `number` |
| `default` | string | ✓ | デフォルト値（空文字列も可） |
| `required` | boolean | ✓ | 必須入力の場合 `true` |
| `help` | string | ✓ | 入力欄下に表示するヘルプテキスト |

### 完全な例

```yaml
variables:
  # テキスト入力
  - name: container_name
    label: コンテナ名
    type: text
    default: my-app
    required: true
    help: Dockerコンテナの名前を指定してください

  # 数値入力
  - name: container_port
    label: ポート番号
    type: number
    default: 3000
    required: true
    help: ホスト側のポート番号（1-65535）

  # パスワード入力（機密情報）
  - name: db_password
    label: データベースパスワード
    type: password
    default: ""
    required: true
    help: データベース接続用のパスワード（8文字以上推奨）

  # オプション項目
  - name: app_env
    label: アプリケーション環境
    type: text
    default: production
    required: false
    help: production, development, staging のいずれか
```

## 実装の仕組み

### 1. データ構造（Go）

```go
// Variable は環境変数の定義を保持
type Variable struct {
    Name     string `yaml:"name"`
    Label    string `yaml:"label"`
    Type     string `yaml:"type"`
    Default  string `yaml:"default"`
    Required bool   `yaml:"required"`
    Help     string `yaml:"help"`
}

// VariablesConfig はvariables.ymlの構造
type VariablesConfig struct {
    Variables []Variable `yaml:"variables"`
}

// PlaybookInfo にVariablesフィールドを追加
type PlaybookInfo struct {
    Name        string
    Path        string
    Description string
    IsLocal     bool
    Variables   []Variable  // 環境変数定義
}
```

### 2. variables.yml読み込み

```go
func ReadVariables(dir string) ([]Variable, error) {
    variablesPath := filepath.Join(dir, "variables.yml")
    
    // ファイルが存在しない場合は空のスライスを返す
    if _, err := os.Stat(variablesPath); os.IsNotExist(err) {
        return []Variable{}, nil
    }
    
    // YAMLファイル読み込み
    content, err := os.ReadFile(variablesPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read variables.yml: %v", err)
    }
    
    // パース
    var config VariablesConfig
    if err := yaml.Unmarshal(content, &config); err != nil {
        return nil, fmt.Errorf("failed to parse variables.yml: %v", err)
    }
    
    return config.Variables, nil
}
```

### 3. Playbook一覧読み込み時の自動処理

```go
func ListLocalPlaybooks(basePath string) ([]PlaybookInfo, error) {
    // ... ディレクトリスキャン
    
    for _, entry := range entries {
        playbookDir := filepath.Join(basePath, entry.Name())
        
        // main.yml存在確認
        if _, err := os.Stat(filepath.Join(playbookDir, "main.yml")); err == nil {
            // variables.yml読み込み（エラーは無視）
            variables, _ := ReadVariables(playbookDir)
            
            playbooks = append(playbooks, PlaybookInfo{
                Name:        entry.Name(),
                Path:        filepath.Join(playbookDir, "main.yml"),
                Description: readDescription(playbookDir),
                IsLocal:     true,
                Variables:   variables,  // 設定
            })
        }
    }
    
    return playbooks, nil
}
```

### 4. 動的フォーム生成（HTML）

```html
<!-- install.html: Playbook一覧でのボタン条件分岐 -->
{{- if .Variables }}
  <!-- 環境変数設定がある場合 -->
  <a href="/install/{{ .Name }}/config">
    <span>設定してインストール</span>
  </a>
{{- else }}
  <!-- 環境変数設定がない場合 -->
  <form method="POST" action="/install/execute">
    <input type="hidden" name="playbook" value="{{ .Name }}" />
    <button type="submit">インストール</button>
  </form>
{{- end }}
```

```html
<!-- install_config.html: 動的フォーム生成 -->
<form method="POST" action="/install/execute">
  <input type="hidden" name="playbook" value="{{ .playbook_name }}" />
  
  {{- range .variables }}
  <div>
    <label for="env_{{ .Name }}">
      {{ .Label }}
      {{- if .Required }}<span class="text-red-500">*</span>{{- end }}
    </label>
    
    {{- if eq .Type "password" }}
      <input type="password" 
             id="env_{{ .Name }}" 
             name="env_{{ .Name }}" 
             value="{{ .Default }}"
             {{- if .Required }} required{{- end }} />
    
    {{- else if eq .Type "number" }}
      <input type="number" 
             id="env_{{ .Name }}" 
             name="env_{{ .Name }}" 
             value="{{ .Default }}"
             {{- if .Required }} required{{- end }} />
    
    {{- else }}
      <input type="text" 
             id="env_{{ .Name }}" 
             name="env_{{ .Name }}" 
             value="{{ .Default }}"
             {{- if .Required }} required{{- end }} />
    {{- end }}
    
    <p class="help-text">{{ .Help }}</p>
  </div>
  {{- end }}
  
  <button type="submit">インストール実行</button>
</form>
```

### 5. 環境変数の抽出と渡し方（Go）

```go
func (ic *InstallController) Execute(c *gin.Context) {
    playbookName := c.PostForm("playbook")
    
    // env_プレフィックスの環境変数を抽出
    var extraVars []string
    for key, values := range c.Request.PostForm {
        if strings.HasPrefix(key, "env_") && len(values) > 0 {
            varName := strings.TrimPrefix(key, "env_")
            varValue := values[0]
            
            // 空文字列でない場合のみ追加
            if varValue != "" {
                extraVars = append(extraVars, varName+"="+varValue)
            }
        }
    }
    
    // Ansible実行
    playbookPath := playbook.GetPlaybookPath(basePath, playbookName)
    result := ansible.RunPlaybookWithConnection(playbookPath, "local", extraVars)
    
    // ... 結果表示
}
```

### 6. Ansibleへの環境変数渡し

```go
func RunPlaybookWithConnection(playbookPath string, connection string, extraVars []string) *PlaybookResult {
    args := []string{"-i", "localhost,", "--connection", connection, playbookPath}
    
    // 環境変数を -e オプションで追加
    if len(extraVars) > 0 {
        args = append(args, "-e", strings.Join(extraVars, " "))
    }
    
    // 実行: ansible-playbook -i localhost, --connection local main.yml -e "key1=value1 key2=value2"
    cmd := exec.Command("ansible-playbook", args...)
    // ...
}
```

### 7. Playbook側での受け取り

```yaml
# main.yml
- hosts: localhost
  connection: local
  become: false

  vars:
    # デフォルト値（variables.ymlと同じ）
    container_name: "my-app"
    container_port: "3000"
    db_password: ""

  tasks:
    - name: Run container
      command: >
        docker run -d
        --name {{ container_name }}
        -p {{ container_port }}:3000
        -e DB_PASSWORD={{ db_password }}
        my-image:latest
      no_log: "{{ db_password != '' }}"  # 機密情報の非表示化
```

## 実際の使用例

### 例1: Node.js Webアプリケーション

**variables.yml**
```yaml
variables:
  - name: container_name
    label: コンテナ名
    type: text
    default: node-webapp
    required: true
    help: Dockerコンテナの名前を指定してください

  - name: container_port
    label: ポート番号
    type: number
    default: 3000
    required: true
    help: ホスト側のポート番号

  - name: db_host
    label: データベースホスト
    type: text
    default: mysql-server
    required: false
    help: データベースコンテナ名またはホスト名

  - name: db_password
    label: データベースパスワード
    type: password
    default: ""
    required: false
    help: データベース接続用のパスワード（機密情報）
```

**main.yml**
```yaml
- hosts: localhost
  connection: local
  vars:
    container_name: "node-webapp"
    container_port: "3000"
    db_host: "mysql-server"
    db_password: ""

  tasks:
    - name: Create network
      command: docker network create app-network
      failed_when: false

    - name: Run Node.js app
      command: >
        docker run -d
        --name {{ container_name }}
        --network app-network
        -p {{ container_port }}:3000
        -e DB_HOST={{ db_host }}
        -e DB_PASSWORD={{ db_password }}
        node:20-alpine
      no_log: "{{ db_password != '' }}"
```

### 例2: WordPress

**variables.yml**
```yaml
variables:
  - name: site_title
    label: サイトタイトル
    type: text
    default: My WordPress Site
    required: true
    help: WordPressサイトのタイトル

  - name: admin_user
    label: 管理者ユーザー名
    type: text
    default: admin
    required: true
    help: WordPress管理画面のユーザー名

  - name: admin_password
    label: 管理者パスワード
    type: password
    default: ""
    required: true
    help: WordPress管理画面のパスワード（8文字以上）

  - name: admin_email
    label: 管理者メールアドレス
    type: text
    default: admin@example.com
    required: true
    help: 管理者用のメールアドレス
```

## ベストプラクティス

### 1. 変数の命名規則

✅ **推奨**
```yaml
- name: container_name
- name: db_host
- name: api_key
```

❌ **避けるべき**
```yaml
- name: containerName    # キャメルケース
- name: db-host          # ハイフン
- name: 2nd_port         # 数字から開始
```

### 2. デフォルト値の設定

✅ **推奨**
```yaml
# 合理的なデフォルト値を設定
- name: container_port
  default: 3000
  
# パスワードは空文字列
- name: db_password
  default: ""
```

❌ **避けるべき**
```yaml
# パスワードにデフォルト値を設定
- name: db_password
  default: password123  # セキュリティリスク
```

### 3. 機密情報の扱い

✅ **推奨**
```yaml
# type: password でマスク表示
- name: secret_key
  label: シークレットキー
  type: password
  default: ""
  help: アプリケーション用のシークレットキー
```

```yaml
# Playbook側でno_log設定
tasks:
  - name: Run app
    command: docker run -e SECRET={{ secret_key }} ...
    no_log: "{{ secret_key != '' }}"
```

### 4. ヘルプメッセージの書き方

✅ **推奨**
```yaml
- name: container_port
  help: ホスト側のポート番号（1-65535）。他のコンテナと重複しないようにしてください
```

❌ **避けるべき**
```yaml
- name: container_port
  help: ポート  # 情報が少なすぎる
```

### 5. 必須項目の設定

```yaml
# 必ず指定が必要な項目
- name: container_name
  required: true

# オプション項目（デフォルト値で動作可能）
- name: custom_config
  required: false
```

## トラブルシューティング

### variables.ymlが読み込まれない

**症状**: Playbook一覧で「設定してインストール」ボタンが表示されない

**原因と対策**:
1. **ファイル名の確認**: `variables.yml`（複数形）が正しいか
2. **配置場所**: Playbookディレクトリ直下にあるか
   ```
   playbooks/containers/my-app/
   ├── main.yml
   └── variables.yml  ← ここ
   ```
3. **YAML構文エラー**: YAMLが正しくパースできるか
   ```bash
   # 検証
   python3 -c "import yaml; yaml.safe_load(open('variables.yml'))"
   ```

### 環境変数がAnsibleに渡されない

**症状**: Playbookで変数が空になる

**確認方法**:
```go
// installController.goでデバッグログ追加
log.Printf("extraVars: %v", extraVars)
```

**原因**:
- フォームのname属性に`env_`プレフィックスがない
- 入力値が空文字列（空の場合は渡されない仕様）

### パスワードがログに表示される

**原因**: Playbook側で`no_log`が設定されていない

**対策**:
```yaml
tasks:
  - name: Run container
    command: docker run -e PASSWORD={{ db_password }} ...
    no_log: "{{ db_password != '' }}"
```

## セキュリティ考慮事項

### 1. 入力値のサニタイズ

現在の実装では基本的なフィルタリングのみ実施。将来的に強化推奨:

```go
// 今後実装予定
func sanitizeVarValue(value string) (string, error) {
    // シェルメタ文字のエスケープ
    // SQLインジェクション対策
    // XSS対策
}
```

### 2. 環境変数の保存

**現在の仕様**: インストール時のみ使用、保存しない

**将来の拡張**: 履歴機能を追加する場合は暗号化必須

### 3. HTTPS化

本番環境では必ずHTTPSを使用:
- リバースプロキシ（Nginx）でSSL終端
- Let's Encrypt証明書の自動更新

## まとめ

環境変数の動的入力機能により、以下が実現できます:

✅ Playbookの柔軟な設定
✅ DB接続情報の安全な入力
✅ 再利用可能なPlaybookテンプレート
✅ ユーザーフレンドリーなインストール体験
