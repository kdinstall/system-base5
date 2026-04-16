# 1行でDockerサーバ環境 + Go Webアプリ構築

サーバに`root`ログインし１行のコマンドを実行するだけでDocker環境とGo製Webアプリ環境が構築できるスクリプトです。

## 対象OS

- Ubuntu 24

## ライセンス

[![MIT license](https://img.shields.io/badge/License-MIT-blue.svg)](https://lbesson.mit-license.org/)

# 内容

`script/start.sh` は Ansible を用い、**`playbooks/docker` → `playbooks/app`** の順で playbook を実行します。

1. **Docker**（`geerlingguy.docker` ロール、`zip` / `unzip`）
2. **Webアプリ** — **Docker管理Webアプリ**（[Gin](https://github.com/gin-gonic/gin) + HTMLテンプレート + [Tailwind CSS v4](https://tailwindcss.com/)）を systemd サービスとして導入

アプリの機能:

- Dockerコンテナの一覧表示・起動・停止・再起動・ログ表示
- Ansible Playbook経由でのコンテナインストール（Nginx, MySQL, PostgreSQL, MongoDB, Redis, Node.js Webアプリ）
- インストール時の環境変数設定（DB接続情報等を動的に入力）
- URL指定でのカスタムPlaybookダウンロード・インストール
- 静的アセットは `/assets`、コンテナ一覧は `/containers`、インストール画面は `/install`（`/` は `/containers` へリダイレクト）

Goアプリは `playbooks/app/webapp` を単独プロジェクトとして管理し、playbook 実行時に `/opt/kdinstall/webapp` へ配備します。デプロイ時に **Node.js で Tailwind をビルド**（`npm run build`）したうえで `go build` します（`go.mod` の Go バージョン要件に合わせてビルドします）。

## インストールモジュール

- `geerlingguy.docker` 8.0.0（Ansible Galaxy、`playbooks/docker/requirements.yml`）で Docker をインストール
- `zip`, `unzip` をインストール
- `golang-go`, `curl`, `git`, `ca-certificates`, `gnupg` など Webアプリ実行・ビルド用パッケージをインストール
- NodeSource リポジトリ経由で **Node.js** をインストール（デプロイ時の CSS ビルド用）
- **Ansible** をインストール（コンテナインストール用）
- `root` ユーザーで `kdinstall-webapp` サービスを実行（Docker コマンド実行のため）
- `kdinstall-webapp` を有効化し、既定では `http://<server>:8080` で待ち受け（リッスンポートは `playbooks/app/main.yml` の `app_port` で変更し、再デプロイで反映）

# 使い方

新規にOSをインストールしたサーバに`root`でログインし、以下の１行のコマンドをそのままコピーして実行します。

## 実行コマンド

最新のリリースタグを使用して実行します。

```bash
curl -fsSL https://raw.githubusercontent.com/kdinstall/system-base5/master/script/start.sh | bash
```

> **注意:** デフォルトでは GitHub の最新リリースタグが自動的に取得・使用されます。  
> 開発中の最新コードを使いたい場合は、後述のテスト実行コマンドを使用してください。

オプション（`bash -s --` 経由で渡す）:

| オプション | 説明 |
|---|---------|
| `-test` | 最新の `master` ブランチを使用して実行 |
| `--help` | ヘルプを表示 |

## テスト実行

最新の master ブランチを使用してテスト実行する場合は、テスト用スクリプトを使用します。

```bash
curl -fsSL https://raw.githubusercontent.com/kdinstall/system-base5/master/test/start.sh | bash
```

## 導入後の確認

以下のコマンドで Docker と Webアプリの導入状態を確認できます。

```bash
systemctl status docker --no-pager
systemctl status kdinstall-webapp --no-pager
curl -fsSL -o /dev/null -w "%{http_code}\n" http://localhost:8080/containers
```

- `kdinstall-webapp` が `active` であれば、作業ディレクトリは既定で `/opt/kdinstall/webapp` です
- `/containers` はコンテナ一覧の HTML を返します（HTTP 200 を想定）

### Webブラウザからのアクセス

デプロイ完了後、Webブラウザから以下のURLで画面を確認できます。

サーバのホスト名やIPアドレスが `example.com` または `192.168.1.100` の場合:

- **コンテナ一覧（トップ相当）:** http://example.com:8080/containers または http://192.168.1.100:8080/containers  
  （`http://...:8080/` にアクセスすると `/containers` へリダイレクトされます）
- **インストール画面:** `.../install`
- **コンテナログ表示:** `.../containers/<id>/logs`

## Goアプリの管理

- アプリ本体は `playbooks/app/webapp` で管理します（Go は主に `src/`、テンプレートは `src/templates/`、Tailwind 入力は `src/style/input.css`、生成 CSS は `public/assets/css/style.css`）
- デプロイ時は `playbooks/app/main.yml` が上記を `/opt/kdinstall/webapp` へコピーし、プロジェクト直下で `node_modules` と `package-lock.json` を削除したうえで `npm install --include=optional` → `npm run build` → `go build -o /opt/kdinstall/bin/webapp ./src` を実行します（配備ファイルに変更があったとき、またはバイナリがまだ無いときにビルドが走ります）
- Playbookディレクトリのパスは環境変数 `PLAYBOOKS_DIR` で変更できます（未設定時は `/opt/kdinstall/containers`）。実行時の環境変数は systemd ユニットで渡す場合は `systemctl edit kdinstall-webapp` などで追記してください
- 変更を反映する場合は1行コマンドを再実行してください
