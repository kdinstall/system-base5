# Redis コンテナ

Redisキャッシュサーバーをコンテナとして起動します。

## 機能

- Redis最新版を使用
- ポート6379でリッスン（変更可能）
- `/opt/redis/data`にデータを永続化（AOFモード有効）
- 自動再起動設定

## 変数

- `container_name`: コンテナ名（デフォルト: redis-server）
- `container_port`: ホスト側のポート（デフォルト: 6379）
- `data_dir`: データディレクトリ（デフォルト: /opt/redis/data）

## 接続方法

```bash
redis-cli -h localhost -p 6379
```

または

```bash
docker exec -it redis-server redis-cli
```

## 使用方法

```bash
ansible-playbook main.yml
```

変数を指定する場合:

```bash
ansible-playbook main.yml -e "container_port=6380"
```
