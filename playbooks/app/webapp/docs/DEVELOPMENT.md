# 開発ガイド

## 開発環境のセットアップ

### 前提条件

- Go 1.23以上
- Node.js 18以上
- Docker 20.10以上
- Ansible 2.9以上
- Git
- VSCode（推奨）

### 初回セットアップ

```bash
# 1. リポジトリクローン
git clone <repository-url>
cd system-base5/playbooks/app/webapp

# 2. 依存関係インストール
go mod download
npm install

# 3. 開発サーバー起動
make dev
```

### VSCode設定

**.vscode/launch.json**
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch Webapp",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/src",
      "env": {
        "SERVER_PORT": "58080",
        "PLAYBOOKS_DIR": "${workspaceFolder}/../../../containers"
      },
      "cwd": "${workspaceFolder}"
    }
  ]
}
```

**.vscode/settings.json**
```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "editor.formatOnSave": true,
  "[go]": {
    "editor.defaultFormatter": "golang.go"
  },
  "files.exclude": {
    "**/.git": true,
    "**/node_modules": true
  }
}
```

### 推奨VSCode拡張機能

- **Go**: golang.go
- **YAML**: redhat.vscode-yaml
- **Tailwind CSS IntelliSense**: bradlc.vscode-tailwindcss
- **Prettier**: esbenp.prettier-vscode

## プロジェクト構造の理解

```
webapp/
├── src/
│   ├── main.go                    # エントリーポイント
│   ├── router.go                  # ルーティング定義
│   ├── config/                    # 設定管理
│   │   └── config.go
│   ├── controllers/               # HTTPハンドラ
│   │   ├── containerController.go
│   │   └── installController.go
│   ├── lib/                       # ビジネスロジック
│   │   ├── ansible/
│   │   ├── docker/
│   │   ├── playbook/
│   │   └── template/
│   └── templates/                 # HTMLテンプレート
│       ├── header.html
│       ├── footer.html
│       ├── containers.html
│       ├── install.html
│       └── install_config.html
├── public/                        # 静的ファイル
│   └── assets/
│       ├── css/
│       └── js/
├── docs/                          # ドキュメント
└── tests/                         # テスト（今後追加）
```

## コーディング規約

### Go

#### ファイル命名

- キャメルケース: `containerController.go`
- パッケージ名は小文字のみ: `package ansible`

#### 関数命名

```go
// 公開関数：大文字で開始
func ListContainers() ([]Container, error)

// 非公開関数：小文字で開始
func parseContainerJSON(data string) (*Container, error)
```

#### エラーハンドリング

```go
// ✅ 推奨：エラーをラップして返す
func ReadVariables(dir string) ([]Variable, error) {
    content, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read variables.yml: %v", err)
    }
    // ...
}

// ❌ 避ける：エラーを無視
content, _ := os.ReadFile(path)
```

#### 構造体のタグ

```go
type Variable struct {
    Name     string `yaml:"name"`      // YAMLタグ
    Label    string `yaml:"label"`
    Type     string `yaml:"type"`
    Default  string `yaml:"default"`
    Required bool   `yaml:"required"`
    Help     string `yaml:"help"`
}
```

### HTML/テンプレート

#### テンプレート名

```go
{{ define "containers.html" }}
// テンプレート内容
{{ end }}
```

#### 変数アクセス

```html
<!-- ドット記法 -->
{{ .page_title }}
{{ .containers }}

<!-- 条件分岐 -->
{{- if .error }}
  <div class="error">{{ .error }}</div>
{{- end }}

<!-- ループ -->
{{- range .containers }}
  <div>{{ .Name }}</div>
{{- end }}
```

### CSS（Tailwind）

#### クラスの順序

```html
<!-- レイアウト → サイズ → 色 → その他 -->
<div class="flex items-center gap-x-2 py-2 px-4 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
```

#### カスタムCSS

```css
/* src/style/input.css */
@tailwind base;
@tailwind components;
@tailwind utilities;

/* カスタムスタイル */
@layer components {
  .btn-primary {
    @apply py-2 px-4 bg-blue-600 text-white rounded-lg hover:bg-blue-700;
  }
}
```

## 開発ワークフロー

### 1. 機能開発

```bash
# 1. フィーチャーブランチ作成
git checkout -b feature/add-domain-management

# 2. コード変更
# 3. テスト実行
go test ./...

# 4. コミット
git add .
git commit -m "Add domain management feature"

# 5. プッシュ
git push origin feature/add-domain-management
```

### 2. フロントエンド開発

```bash
# CSSウォッチモード（自動再ビルド）
make watch

# 別ターミナルでアプリ起動
make run

# ブラウザでライブリロード（手動リフレッシュ）
```

### 3. デバッグ

#### ログ出力

```go
import "log"

// デバッグログ
log.Printf("Debug: extraVars = %v", extraVars)

// エラーログ
log.Printf("Error: %v", err)
```

#### VSCodeデバッガー

1. ブレークポイント設定
2. F5キーでデバッグ開始
3. 変数を確認

### 4. ホットリロード（開発効率化）

現在は手動リフレッシュですが、将来的にair等のツール導入を検討:

```bash
# airのインストール
go install github.com/cosmtrek/air@latest

# 自動リロード
air
```

## テスト

### ユニットテスト

```go
// lib/playbook/manager_test.go
package playbook

import "testing"

func TestReadVariables(t *testing.T) {
    // テストデータ準備
    dir := "testdata/nodejs-webapp"
    
    // 実行
    variables, err := ReadVariables(dir)
    
    // 検証
    if err != nil {
        t.Fatalf("ReadVariables failed: %v", err)
    }
    
    if len(variables) != 11 {
        t.Errorf("Expected 11 variables, got %d", len(variables))
    }
    
    if variables[0].Name != "container_name" {
        t.Errorf("Expected first variable name to be 'container_name', got '%s'", variables[0].Name)
    }
}
```

### 実行

```bash
# 全テスト実行
go test ./...

# 特定パッケージ
go test ./src/lib/playbook

# カバレッジ
go test -cover ./...

# 詳細出力
go test -v ./...
```

### 統合テスト（手動）

```bash
# 1. アプリ起動
make dev

# 2. ブラウザで操作
# - コンテナ一覧表示
# - コンテナ起動/停止
# - Playbookインストール

# 3. ログ確認
# ターミナル出力を確認
```

## トラブルシューティング

### CSSが反映されない

```bash
# キャッシュクリア
rm -rf public/assets/css/style.css

# 再ビルド
make build

# ブラウザのハードリフレッシュ
# Ctrl+Shift+R (Linux/Windows)
# Cmd+Shift+R (Mac)
```

### Goの依存関係エラー

```bash
# go.modをクリーン
go mod tidy

# モジュールキャッシュクリア
go clean -modcache

# 再ダウンロード
go mod download
```

### ポート競合

```bash
# 58080ポートを使用しているプロセスを確認
# Linux/Mac
lsof -i :58080

# Windows
netstat -ano | findstr :58080

# プロセスを停止
kill <PID>
```

### Dockerコマンド失敗

```bash
# Dockerサービスの状態確認
sudo systemctl status docker

# Docker権限確認（Linux）
sudo usermod -aG docker $USER
# ログアウト/ログインが必要
```

## パフォーマンスプロファイリング

### CPU/メモリプロファイル

```go
import (
    "net/http"
    _ "net/http/pprof"
)

func main() {
    // pprofエンドポイント追加
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    
    // 通常のアプリ起動
    // ...
}
```

### プロファイル取得

```bash
# CPUプロファイル
go tool pprof http://localhost:6060/debug/pprof/profile

# メモリプロファイル
go tool pprof http://localhost:6060/debug/pprof/heap

# 可視化
go tool pprof -http=:8081 <profile-file>
```

## Git運用

### ブランチ戦略

- `main`: 本番環境
- `develop`: 開発環境
- `feature/*`: 新機能開発
- `bugfix/*`: バグ修正
- `hotfix/*`: 緊急修正

### コミットメッセージ

```
<type>: <subject>

<body>

<footer>
```

**Type:**
- `feat`: 新機能
- `fix`: バグ修正
- `docs`: ドキュメント
- `style`: フォーマット
- `refactor`: リファクタリング
- `test`: テスト追加
- `chore`: ビルド設定等

**例:**
```
feat: Add environment variable configuration screen

- Add variables.yml support
- Create install_config.html template
- Implement dynamic form generation

Closes #123
```

## コードレビューチェックリスト

### Go

- [ ] エラーハンドリングが適切
- [ ] 公開API（大文字関数）にコメントがある
- [ ] 構造体にYAML/JSONタグがある
- [ ] fmt.Errorfでエラーをラップしている
- [ ] ロックが必要な場合は適切に使用

### HTML/テンプレート

- [ ] XSS対策（エスケープ）がされている
- [ ] Tailwind CSSクラスが適切
- [ ] レスポンシブデザイン対応
- [ ] アクセシビリティ（ARIA）対応

### 全般

- [ ] テストが追加されている
- [ ] ドキュメントが更新されている
- [ ] ログ出力が適切
- [ ] パフォーマンスへの影響が最小限

## リリースプロセス

### 1. バージョン番号の決定

セマンティックバージョニング: `MAJOR.MINOR.PATCH`

- MAJOR: 破壊的変更
- MINOR: 新機能追加（後方互換性あり）
- PATCH: バグ修正

### 2. CHANGELOG更新

```markdown
# Changelog

## [1.1.0] - 2026-04-12

### Added
- 環境変数の動的入力機能
- variables.yml サポート

### Changed
- Playbook一覧画面のUIを改善

### Fixed
- コンテナログ表示のバグ修正
```

### 3. タグ作成

```bash
git tag -a v1.1.0 -m "Release version 1.1.0"
git push origin v1.1.0
```

### 4. デプロイ

```bash
# 本番環境へデプロイ
ansible-playbook playbooks/app/main.yml
```

## よくある質問（FAQ）

### Q: Tailwind CSSのクラスが適用されない

A: 以下を確認してください:
1. `make build`でCSSをビルド
2. `public/assets/css/style.css`が生成されているか確認
3. ブラウザキャッシュをクリア

### Q: Playbookが一覧に表示されない

A: 以下を確認してください:
1. `main.yml`が存在するか
2. `PLAYBOOKS_DIR`環境変数が正しいか
3. アプリを再起動したか

### Q: 環境変数設定画面が表示されない

A: 以下を確認してください:
1. `variables.yml`が存在するか
2. YAML形式が正しいか
3. `install_config.html`が存在するか

## 今後の開発予定

### 短期（v1.1 - v1.2）

- [ ] ドメイン管理機能
- [ ] リバースプロキシ自動設定
- [ ] SSL証明書管理

### 中期（v1.3 - v2.0）

- [ ] ユーザー認証/認可
- [ ] RBAC（ロールベースアクセス制御）
- [ ] WebSocketでのリアルタイムログ
- [ ] REST API提供

### 長期（v2.0+）

- [ ] Kubernetes対応
- [ ] クラスタ管理
- [ ] メトリクス収集と可視化
- [ ] マルチテナント対応

## リソース

### 公式ドキュメント

- [Go Documentation](https://golang.org/doc/)
- [Gin Framework](https://gin-gonic.com/docs/)
- [Tailwind CSS](https://tailwindcss.com/docs)
- [Ansible Documentation](https://docs.ansible.com/)

### 関連プロジェクト

- [Docker Documentation](https://docs.docker.com/)
- [Preline UI](https://preline.co/)

### コミュニティ

- GitHub Issues: バグ報告・機能要望
- GitHub Discussions: 質問・ディスカッション

## サポート

問題が発生した場合:

1. [トラブルシューティング](#トラブルシューティング)を確認
2. ログを確認（`journalctl -u kdinstall-webapp`）
3. GitHub Issuesで報告

---

Happy Coding! 🚀
