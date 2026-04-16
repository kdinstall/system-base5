# Node.js Webアプリケーション コンテナ

Node.js Webアプリケーションをコンテナとして起動します。

## 機能

- Node.js 20 Alpine（軽量イメージ）を使用
- Dockerネットワークを作成し、他のコンテナと通信可能
- データベース接続情報を環境変数として設定
- アプリケーション環境（production/development）の指定
- シークレットキーの設定

## 環境変数

このPlaybookは以下の環境変数をサポートしています（Web画面から設定可能）:

- `container_name`: コンテナ名（デフォルト: node-webapp）
- `container_port`: ホスト側のポート（デフォルト: 3000）
- `docker_network`: Dockerネットワーク名（デフォルト: app-network）
- `docker_image`: 使用するDockerイメージ（デフォルト: node:20-alpine）
- `db_host`: データベースホスト（デフォルト: mysql-server）
- `db_port`: データベースポート（デフォルト: 3306）
- `db_name`: データベース名（デフォルト: myapp）
- `db_user`: データベースユーザー（デフォルト: dbuser）
- `db_password`: データベースパスワード
- `app_env`: アプリケーション環境（デフォルト: production）
- `secret_key`: シークレットキー

## Dockerネットワーク

このPlaybookは`app-network`という名前のDockerネットワークを自動作成します。
データベース等の他のコンテナも同じネットワークに接続することで、コンテナ名で名前解決が可能になります。

例：
```bash
# MySQLコンテナを同じネットワークで起動
docker run -d --name mysql-server --network app-network \
  -e MYSQL_ROOT_PASSWORD=root123 mysql:8

# Node.jsアプリからDB_HOST=mysql-serverで接続可能
```

## アクセス

```
http://localhost:3000
```

## 使用方法

Webインターフェースから「インストール」を選択し、必要な環境変数を設定してください。

手動で実行する場合:

```bash
ansible-playbook main.yml \
  -e "container_name=my-app" \
  -e "db_password=secret123" \
  -e "secret_key=mysecret"
```

## 注意事項

- このサンプルはデモ用です。実際のアプリケーションでは適切なDockerイメージとコマンドを指定してください
- 機密情報（パスワード、シークレットキー）は安全に管理してください
- 本番環境ではボリュームマウントでアプリケーションコードを配置することを推奨します
