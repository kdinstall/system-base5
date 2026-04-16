# MongoDB コンテナ

MongoDB 7ドキュメント指向データベースをコンテナとして起動します。

## 機能

- MongoDB 7を使用
- ポート27017でリッスン（変更可能）
- `/opt/mongodb/data`にデータを永続化
- root認証を有効化

## 変数

- `container_name`: コンテナ名（デフォルト: mongodb-server）
- `container_port`: ホスト側のポート（デフォルト: 27017）
- `mongo_root_username`: rootユーザー名（デフォルト: root）
- `mongo_root_password`: rootパスワード（デフォルト: mongo_pass_123）
- `data_dir`: データディレクトリ（デフォルト: /opt/mongodb/data）

## 接続方法

```bash
mongosh "mongodb://root:mongo_pass_123@localhost:27017"
```

または

```bash
docker exec -it mongodb-server mongosh -u root -p mongo_pass_123
```

## 使用方法

```bash
ansible-playbook main.yml
```

変数を指定する場合:

```bash
ansible-playbook main.yml -e "mongo_root_password=MySecurePass123"
```
