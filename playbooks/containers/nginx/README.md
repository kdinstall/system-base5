# Nginx コンテナ

Nginxウェブサーバーをコンテナとして起動します。

## 機能

- Nginx最新版を使用
- ポート8080でリッスン（変更可能）
- `/opt/nginx/html`をドキュメントルートとしてマウント
- サンプルのindex.htmlを自動作成

## 変数

- `container_name`: コンテナ名（デフォルト: nginx-server）
- `container_port`: ホスト側のポート（デフォルト: 8080）
- `host_html_dir`: HTMLファイルのディレクトリ（デフォルト: /opt/nginx/html）

## アクセス

```
http://localhost:8080
```

## 使用方法

```bash
ansible-playbook main.yml
```

変数を指定する場合:

```bash
ansible-playbook main.yml -e "container_name=my-nginx container_port=9000"
```
