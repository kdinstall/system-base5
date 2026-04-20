# Docker管理Webアプリケーション

Dockerコンテナを管理するためのWebアプリケーション。Ansible Playbook経由でのコンテナインストール、起動停止、ログ表示機能を提供します。

## 技術スタック

- **Go**: Gin フレームワーク
- **フロントエンド**: Tailwind CSS v4 + Preline UI
- **テンプレート**: html/template
- **Docker**: Docker CLI（os/exec経由）
- **Ansible**: ansible-playbook CLI（os/exec経由）

## 機能

- Dockerコンテナ一覧表示・起動停止
- Ansible Playbook経由でのコンテナインストール（**非同期実行**）
- **ジョブ管理**: インストール履歴の一覧表示とステータス確認
- **リアルタイムログ表示**: JavaScriptポーリングで自動更新
- デフォルトPlaybookテンプレート（Nginx, MySQL, Redis, PostgreSQL, MongoDB, Node.js Webapp）
- URL指定でのPlaybookダウンロード・インストール
- 再インストール機能（Playbook再実行）
- コンテナログ表示
- **同時実行制限**: 1つのジョブのみ実行可能（リソース競合防止）

## 開発

```bash
# 依存関係のインストールとビルド、起動
make dev

# CSS のみビルド
make build

# CSS のウォッチモード
make watch

# アプリ起動（CSS ビルド済み前提）
make run
```

## デプロイ

`playbooks/app/main.yml` で `/opt/kdinstall/webapp` へ配備されます。

```bash
# Ansible経由でデプロイ
ansible-playbook playbooks/app/main.yml
```

systemd サービス `kdinstall-webapp` として起動します。

## 環境変数

- `SERVER_PORT`: リッスンポート（デフォルト: 58080）
- `PLAYBOOKS_DIR`: Playbookディレクトリパス（デフォルト: ../containers、本番環境では `/opt/kdinstall/containers`）
- `SSL_CERT_PATH`: SSL証明書ファイルパス（デフォルト: `/opt/kdinstall/certs/server.crt`）
- `SSL_KEY_PATH`: SSL秘密鍵ファイルパス（デフォルト: `/opt/kdinstall/certs/server.key`）

## HTTPS対応

**HTTPSのみをサポート**します（HTTPポートは無効）。自己署名SSL証明書を使用してHTTPS配信します（ポート58080）。証明書は初回デプロイ時に自動生成され、10年間有効です。
