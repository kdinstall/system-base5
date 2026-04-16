# MySQL コンテナ

MySQL 8データベースサーバーをコンテナとして起動します。

## 機能

- MySQL 8を使用
- ポート3306でリッスン（変更可能）
- `/opt/mysql/data`にデータを永続化
- 初期データベースとユーザーを自動作成

## 変数

- `container_name`: コンテナ名（デフォルト: mysql-server）
- `container_port`: ホスト側のポート（デフォルト: 3306）
- `mysql_root_password`: rootパスワード（デフォルト: root_password_123）
- `mysql_database`: 作成するデータベース名（デフォルト: myapp）
- `mysql_user`: 作成するユーザー名（デフォルト: myuser）
- `mysql_password`: ユーザーのパスワード（デフォルト: mypass_456）
- `data_dir`: データディレクトリ（デフォルト: /opt/mysql/data）

## 接続方法

```bash
mysql -h localhost -P 3306 -u myuser -pmypass_456 myapp
```

または

```bash
docker exec -it mysql-server mysql -u root -proot_password_123
```

## 使用方法

```bash
ansible-playbook main.yml
```

変数を指定する場合:

```bash
ansible-playbook main.yml -e "mysql_root_password=MySecurePass123"
```
