# デプロイメント

## 開発環境セットアップ

### 必要な環境

- **OS**: Linux, macOS, Windows (WSL2)
- **Go**: 1.23以上
- **Node.js**: 18以上（Tailwind CSSビルド用）
- **Docker**: 20.10以上
- **Ansible**: 2.9以上
- **Make**: GNU Make

### セットアップ手順

#### 1. リポジトリのクローン

```bash
git clone <repository-url>
cd system-base5/playbooks/app/webapp
```

#### 2. 依存関係のインストール

```bash
# Go依存関係
go mod download

# Node.js依存関係（Tailwind CSS）
npm install
```

#### 3. 環境変数の設定（オプション）

```bash
# .env ファイルを作成（開発環境用）
cat > .env << 'EOF'
SERVER_PORT=58080
PLAYBOOKS_DIR=../../../containers
EOF
```

> **注意**: 開発環境では `playbooks/app/webapp` から見た相対パス `../../../containers` を使用します。  
> 本番環境では `/opt/kdinstall/containers` を使用します（環境変数未設定時は `../containers` が自動設定され、これは `/opt/kdinstall/webapp` から見て `/opt/kdinstall/containers` になります）。

#### 4. ビルドと起動

```bash
# 開発モード（CSSビルド + アプリ起動）
make dev

# または個別に実行
make build   # CSSビルド
make run     # アプリ起動
```

#### 5. ブラウザでアクセス

開発環境では通常HTTPモードで起動します：

```
http://localhost:58080
```

HTTPSモードで開発する場合は、テスト用の自己署名証明書を生成して環境変数を設定してください：

```bash
# テスト証明書生成
mkdir -p certs
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout certs/server.key -out certs/server.crt \
  -subj "/C=JP/ST=Tokyo/L=Tokyo/O=kdinstall/CN=localhost"

# HTTPS起動
ENABLE_SSL=true SSL_CERT_PATH=./certs/server.crt SSL_KEY_PATH=./certs/server.key make run
```

その後、`https://localhost:58080` でアクセス（ブラウザの自己署名証明書警告を承認）

### Makeコマンド一覧

| コマンド | 説明 |
|---------|------|
| `make dev` | CSSビルド + アプリビルド + 起動（開発用） |
| `make build` | Tailwind CSSをビルド |
| `make watch` | CSSをウォッチモードでビルド（自動再ビルド） |
| `make run` | Goアプリを起動（CSSビルド済み前提） |
| `make clean` | ビルド成果物を削除 |

### ディレクトリ構成の確認

開発環境では相対パスでPlaybookディレクトリを参照します:

```
system-base5/
├── playbooks/
│   ├── app/
│   │   └── webapp/          ← 作業ディレクトリ
│   │       ├── src/
│   │       ├── public/
│   │       └── Makefile
│   └── containers/          ← Playbookディレクトリ
│       ├── nginx/
│       ├── mysql/
│       └── ...
```

## 本番環境デプロイ

### デプロイ先の環境

- **OS**: Ubuntu 20.04以上（推奨）
- **配置先**: `/opt/kdinstall/webapp`
- **サービス**: systemd (`kdinstall-webapp.service`)
- **実行ユーザー**: root（Dockerコマンド実行のため）

### 自動デプロイ（Ansible）

プロジェクトルートから実行:

```bash
# デプロイPlaybookを実行
ansible-playbook playbooks/app/main.yml
```

このPlaybookは以下を実行します:

1. 必要なパッケージのインストール（Go, Node.js, Docker, Ansible, openssl）
2. 自己署名SSL証明書の生成（初回のみ、10年有効）
3. アプリケーションのビルド
4. `/opt/kdinstall/webapp`へのデプロイ
5. systemdサービスの登録と起動（HTTPS有効化）

### SSL証明書について

本番環境では自己署名SSL証明書が自動生成されます：

- **証明書**: `/opt/kdinstall/certs/server.crt`
- **秘密鍵**: `/opt/kdinstall/certs/server.key`
- **有効期間**: 10年（3650日）
- **サブジェクト**: `/C=JP/ST=Tokyo/L=Tokyo/O=kdinstall/CN=localhost`

証明書は初回デプロイ時のみ生成され、既に存在する場合は再生成されません。証明書を再生成する場合：

```bash
# 既存証明書を削除
sudo rm /opt/kdinstall/certs/server.{crt,key}

# Playbookを再実行（証明書が再生成される）
ansible-playbook playbooks/app/main.yml
```

本番環境でLet's Encrypt等の正式な証明書を使用する場合は、環境変数を設定してください：

```bash
# systemdサービスファイルを編集
sudo systemctl edit kdinstall-webapp

# 以下を追加
[Service]
Environment=SSL_CERT_PATH=/etc/letsencrypt/live/example.com/fullchain.pem
Environment=SSL_KEY_PATH=/etc/letsencrypt/live/example.com/privkey.pem

# サービスを再起動
sudo systemctl restart kdinstall-webapp
```

### 手動デプロイ

#### 1. ビルド

```bash
cd playbooks/app/webapp

# CSSビルド
npm install
npm run build

# Goバイナリビルド
go build -o webapp src/main.go
```

#### 2. ファイル転送

```bash
# サーバーにディレクトリを作成
ssh user@server "sudo mkdir -p /opt/kdinstall/webapp"

# ファイルをコピー
scp -r . user@server:/tmp/webapp
ssh user@server "sudo mv /tmp/webapp/* /opt/kdinstall/webapp/"
```

#### 3. systemdサービス設定

**kdinstall-webapp.service**
```ini
[Unit]
Description=kdinstall-webapp HTTPS web application
After=network.target docker.service

[Service]
Type=simple
User=kdi
Group=kdi
WorkingDirectory=/opt/kdinstall/webapp
ExecStart=/opt/kdinstall/bin/webapp
Environment=SERVER_PORT=58080
Environment=PLAYBOOKS_DIR=/opt/kdinstall/containers
Environment=ENABLE_SSL=true
Environment=SSL_CERT_PATH=/opt/kdinstall/certs/server.crt
Environment=SSL_KEY_PATH=/opt/kdinstall/certs/server.key
Restart=on-failure
RestartSec=3

[Install]
WantedBy=multi-user.target
```

#### 4. サービスの起動

```bash
sudo systemctl daemon-reload
sudo systemctl enable kdinstall-webapp
sudo systemctl start kdinstall-webapp
```

#### 5. 動作確認

```bash
# サービス状態確認
sudo systemctl status kdinstall-webapp

# HTTPSアクセステスト（自己署名証明書を信頼）
curl -k https://localhost:58080/containers

# 証明書情報確認
openssl x509 -in /opt/kdinstall/certs/server.crt -text -noout | grep -E "Validity|Subject"
```

ブラウザからアクセスする場合は `https://localhost:58080` で、自己署名証明書の警告が表示されます。警告を承認すると正常にアクセスできます。

[Service]
Type=simple
User=root
WorkingDirectory=/opt/kdinstall/webapp
Environment="SERVER_PORT=58080"
Environment="PLAYBOOKS_DIR=/opt/kdinstall/containers"
ExecStart=/opt/kdinstall/webapp/webapp
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

#### 4. サービス起動

```bash
# サービスファイルを配置
sudo cp kdinstall-webapp.service /etc/systemd/system/

# サービスを有効化・起動
sudo systemctl daemon-reload
sudo systemctl enable kdinstall-webapp
sudo systemctl start kdinstall-webapp

# ステータス確認
sudo systemctl status kdinstall-webapp
```

### 環境変数の設定

本番環境では環境変数でカスタマイズします:

```bash
# /etc/systemd/system/kdinstall-webapp.service を編集
[Service]
Environment="SERVER_PORT=58080"
Environment="PLAYBOOKS_DIR=/opt/kdinstall/containers"

# 変更を反映
sudo systemctl daemon-reload
sudo systemctl restart kdinstall-webapp
```

## リバースプロキシ設定（Nginx）

本番環境ではNginxをリバースプロキシとして使用することを推奨します。

### Nginx設定例

**/etc/nginx/sites-available/kdinstall**
```nginx
server {
    listen 80;
    server_name your-domain.com;

    # HTTPSへリダイレクト（SSL設定後）
    # return 301 https://$host$request_uri;

    location / {
        proxy_pass http://localhost:58080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket対応（将来の拡張用）
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    # 静的ファイルのキャッシュ
    location /assets/ {
        proxy_pass http://localhost:58080;
        proxy_cache_valid 200 1d;
        expires 1d;
        add_header Cache-Control "public, immutable";
    }
}
```

### SSL/TLS設定（Let's Encrypt）

```bash
# Certbotインストール
sudo apt install certbot python3-certbot-nginx

# 証明書取得
sudo certbot --nginx -d your-domain.com

# 自動更新設定（既に含まれている）
sudo systemctl status certbot.timer
```

## アップデート手順

### アプリケーションの更新

```bash
# 1. 最新コードを取得
cd /path/to/source
git pull

# 2. ビルド
cd playbooks/app/webapp
npm run build
go build -o webapp src/main.go

# 3. デプロイ
sudo cp webapp /opt/kdinstall/webapp/
sudo cp -r public /opt/kdinstall/webapp/

# 4. サービス再起動
sudo systemctl restart kdinstall-webapp
```

### Playbookの追加・更新

```bash
# 新しいPlaybookを追加
sudo cp -r new-playbook /opt/kdinstall/containers/

# サービス再起動（Playbook一覧を再読み込み）
sudo systemctl restart kdinstall-webapp
```

## モニタリング

### ログ確認

```bash
# サービスのログ
sudo journalctl -u kdinstall-webapp -f

# 最新100行
sudo journalctl -u kdinstall-webapp -n 100

# 特定期間のログ
sudo journalctl -u kdinstall-webapp --since "2026-04-12 10:00:00"
```

### ヘルスチェック

```bash
# サービス状態
sudo systemctl status kdinstall-webapp

# HTTPエンドポイント確認
curl http://localhost:58080/containers

# プロセス確認
ps aux | grep webapp
```

## バックアップ

### バックアップ対象

1. **アプリケーションディレクトリ**: `/opt/kdinstall/webapp`
2. **Playbookディレクトリ**: `/opt/kdinstall/containers`
3. **systemd設定**: `/etc/systemd/system/kdinstall-webapp.service`
4. **Nginx設定**: `/etc/nginx/sites-available/kdinstall`

### バックアップスクリプト例

```bash
#!/bin/bash
BACKUP_DIR="/backup/kdinstall"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR

# アプリケーションとPlaybookをバックアップ
tar -czf $BACKUP_DIR/kdinstall-$DATE.tar.gz \
  /opt/kdinstall/webapp \
  /opt/kdinstall/containers \
  /etc/systemd/system/kdinstall-webapp.service \
  /etc/nginx/sites-available/kdinstall

# 古いバックアップを削除（30日以上前）
find $BACKUP_DIR -name "kdinstall-*.tar.gz" -mtime +30 -delete

echo "Backup completed: $BACKUP_DIR/kdinstall-$DATE.tar.gz"
```

### リストア

```bash
# バックアップから復元
tar -xzf /backup/kdinstall/kdinstall-20260412_120000.tar.gz -C /

# サービス再起動
sudo systemctl daemon-reload
sudo systemctl restart kdinstall-webapp
sudo systemctl reload nginx
```

## トラブルシューティング

### サービスが起動しない

**確認1: ログを確認**
```bash
sudo journalctl -u kdinstall-webapp -n 50
```

**確認2: ポートが使用可能か**
```bash
sudo netstat -tulpn | grep 58080
```

**確認3: 実行ファイルの権限**
```bash
ls -la /opt/kdinstall/webapp/webapp
sudo chmod +x /opt/kdinstall/webapp/webapp
```

### Playbookディレクトリが見つからない

```bash
# 環境変数を確認
sudo systemctl show kdinstall-webapp | grep PLAYBOOKS_DIR

# ディレクトリの存在確認
ls -la /opt/kdinstall/containers
```

### Dockerコマンドが実行できない

```bash
# rootユーザーで実行されているか確認
sudo systemctl show kdinstall-webapp | grep User

# Dockerソケットの権限確認
ls -la /var/run/docker.sock

# Dockerサービスの状態
sudo systemctl status docker
```

## セキュリティ強化

### ファイアウォール設定

```bash
# UFWを使用する場合
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
sudo ufw enable

# 直接58080ポートへのアクセスは拒否（Nginx経由のみ許可）
sudo ufw deny 58080/tcp
```

### アクセス制限（Nginx）

```nginx
# IP制限
location / {
    allow 192.168.1.0/24;  # 特定ネットワークのみ許可
    deny all;
    proxy_pass http://localhost:58080;
}

# Basic認証
location / {
    auth_basic "Restricted Access";
    auth_basic_user_file /etc/nginx/.htpasswd;
    proxy_pass http://localhost:58080;
}
```

### Basic認証の設定

```bash
# htpasswdファイル作成
sudo apt install apache2-utils
sudo htpasswd -c /etc/nginx/.htpasswd admin

# Nginx再起動
sudo systemctl reload nginx
```

## パフォーマンスチューニング

### Goアプリケーション

現在の設定で十分ですが、大規模環境では以下を検討:

```go
// main.go
gin.SetMode(gin.ReleaseMode)  // 本番モード
```

### Nginxキャッシュ

```nginx
# キャッシュディレクトリ設定
proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=my_cache:10m max_size=1g inactive=60m;

server {
    location /assets/ {
        proxy_cache my_cache;
        proxy_cache_valid 200 1d;
        proxy_cache_use_stale error timeout updating;
    }
}
```

## チェックリスト

デプロイ前の確認:

- [ ] ビルドが正常に完了した
- [ ] 環境変数が正しく設定されている
- [ ] Playbookディレクトリが正しいパスに配置されている
- [ ] systemdサービスが有効化されている
- [ ] Nginxリバースプロキシが設定されている
- [ ] SSL証明書が取得されている（本番環境）
- [ ] ファイアウォールが設定されている
- [ ] バックアップスクリプトが設定されている
- [ ] ログローテーションが設定されている
- [ ] モニタリングが設定されている

## 本番環境の推奨構成

```
インターネット
    ↓
[Nginx] (:80, :443)
  ├─ SSL終端
  ├─ リバースプロキシ
  └─ 静的ファイルキャッシュ
    ↓
[Webアプリ] (:58080)
  ├─ Docker管理
  └─ Playbook実行
    ↓
[Docker Engine]
  └─ コンテナ群
```

この構成により、セキュリティとパフォーマンスを確保できます。
