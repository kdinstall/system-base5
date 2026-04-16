# PostgreSQL コンテナ

PostgreSQL 16データベースサーバーをコンテナとして起動します。

## 機能

- PostgreSQL 16を使用
- ポート5432でリッスン（変更可能）
- `/opt/postgresql/data`にデータを永続化
- 初期データベースとユーザーを自動作成

## 変数

- `container_name`: コンテナ名（デフォルト: postgresql-server）
- `container_port`: ホスト側のポート（デフォルト: 5432）
- `postgres_password`: パスワード（デフォルト: postgres_pass_123）
- `postgres_user`: ユーザー名（デフォルト: postgres）
- `postgres_db`: データベース名（デフォルト: myapp）
- `data_dir`: データディレクトリ（デフォルト: /opt/postgresql/data）

## 接続方法

```bash
psql -h localhost -p 5432 -U postgres -d myapp
```

または

```bash
docker exec -it postgresql-server psql -U postgres -d myapp
```

## 使用方法

```bash
ansible-playbook main.yml
```

変数を指定する場合:

```bash
ansible-playbook main.yml -e "postgres_password=MySecurePass123"
```
