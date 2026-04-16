# Playbook作成ガイド

## 概要

このガイドでは、Docker管理Webアプリケーションで使用するAnsible Playbookの作成方法を説明します。Playbookを作成することで、さまざまなコンテナを自動でインストールできます。

## Playbookの基本構造

### 必須ファイル

```
playbooks/containers/your-app/
├── main.yml          # Ansibleタスク（必須）
├── README.md         # 説明文（推奨）
└── variables.yml     # 環境変数定義（オプション）
```

### 最小構成のPlaybook

**main.yml**
```yaml
- hosts: localhost
  connection: local
  become: false

  vars:
    container_name: "my-app"
    container_port: "8080"

  tasks:
    - name: Run container
      command: >
        docker run -d
        --name {{ container_name }}
        -p {{ container_port }}:80
        --restart unless-stopped
        nginx:latest
```

このPlaybookは、ブラウザの「インストール」ボタンから直接実行できます。

## 環境変数定義の追加

環境変数をWeb画面から設定可能にするには、`variables.yml`を追加します。

**variables.yml**
```yaml
variables:
  - name: container_name
    label: コンテナ名
    type: text
    default: my-app
    required: true
    help: Dockerコンテナの名前を指定してください

  - name: container_port
    label: ポート番号
    type: number
    default: 8080
    required: true
    help: ホスト側のポート番号（例: 8080）
```

詳細は[VARIABLES.md](./VARIABLES.md)を参照してください。

## README.mdの書き方

一覧画面に表示される説明文を記載します。

**README.md**
```markdown
# アプリケーション名

最初の段落がPlaybook一覧に表示されます（100文字まで）。

## 機能

- 機能1
- 機能2

## 使用方法

...
```

**表示例**:
- 1行目が見出しとして表示
- 最初の段落（空行まで）が説明文として抽出
- 100文字を超える場合は`...`で省略

## Playbookテンプレート集

### 1. 基本的なWebサーバー（Nginx）

```yaml
- hosts: localhost
  connection: local
  become: false

  vars:
    container_name: "nginx-server"
    container_port: "8080"
    host_html_dir: "/opt/nginx/html"

  tasks:
    - name: Create HTML directory
      file:
        path: "{{ host_html_dir }}"
        state: directory
        mode: '0755'
      become: true

    - name: Create sample index.html
      copy:
        dest: "{{ host_html_dir }}/index.html"
        content: |
          <!DOCTYPE html>
          <html>
          <head><title>Nginx</title></head>
          <body><h1>Welcome to Nginx!</h1></body>
          </html>
        mode: '0644'
      become: true

    - name: Run Nginx container
      command: >
        docker run -d 
        --name {{ container_name }}
        -p {{ container_port }}:80
        -v {{ host_html_dir }}:/usr/share/nginx/html:ro
        --restart unless-stopped
        nginx:latest
```

### 2. データベース（MySQL）

```yaml
- hosts: localhost
  connection: local
  become: false

  vars:
    container_name: "mysql-server"
    container_port: "3306"
    mysql_root_password: "root_password_123"
    mysql_database: "myapp"
    mysql_user: "myuser"
    mysql_password: "mypass_456"
    data_dir: "/opt/mysql/data"

  tasks:
    - name: Create MySQL data directory
      file:
        path: "{{ data_dir }}"
        state: directory
        mode: '0755'
      become: true

    - name: Run MySQL container
      command: >
        docker run -d
        --name {{ container_name }}
        -p {{ container_port }}:3306
        -e MYSQL_ROOT_PASSWORD={{ mysql_root_password }}
        -e MYSQL_DATABASE={{ mysql_database }}
        -e MYSQL_USER={{ mysql_user }}
        -e MYSQL_PASSWORD={{ mysql_password }}
        -v {{ data_dir }}:/var/lib/mysql
        --restart unless-stopped
        mysql:8
      no_log: true  # パスワードのログ出力を抑制
```

### 3. Webアプリ（Dockerネットワーク使用）

```yaml
- hosts: localhost
  connection: local
  become: false

  vars:
    container_name: "webapp"
    container_port: "3000"
    docker_network: "app-network"
    db_host: "mysql-server"
    db_password: ""

  tasks:
    - name: Create Docker network
      command: docker network create {{ docker_network }}
      register: network_result
      failed_when: false
      changed_when: network_result.rc == 0

    - name: Run webapp container
      command: >
        docker run -d
        --name {{ container_name }}
        --network {{ docker_network }}
        -p {{ container_port }}:3000
        -e DB_HOST={{ db_host }}
        -e DB_PASSWORD={{ db_password }}
        --restart unless-stopped
        node:20-alpine
        node -e "console.log('App running')"
      no_log: "{{ db_password != '' }}"
```

### 4. Redis（キーバリューストア）

```yaml
- hosts: localhost
  connection: local
  become: false

  vars:
    container_name: "redis-server"
    container_port: "6379"
    data_dir: "/opt/redis/data"

  tasks:
    - name: Create Redis data directory
      file:
        path: "{{ data_dir }}"
        state: directory
        mode: '0755'
      become: true

    - name: Run Redis container
      command: >
        docker run -d
        --name {{ container_name }}
        -p {{ container_port }}:6379
        -v {{ data_dir }}:/data
        --restart unless-stopped
        redis:7 redis-server --appendonly yes
```

### 5. カスタムDockerイメージのビルドとデプロイ

```yaml
- hosts: localhost
  connection: local
  become: false

  vars:
    container_name: "custom-app"
    build_dir: "/tmp/custom-app-build"
    image_name: "custom-app:latest"
    container_port: "8080"

  tasks:
    - name: Create build directory
      file:
        path: "{{ build_dir }}"
        state: directory
        mode: '0755'

    - name: Copy Dockerfile
      copy:
        dest: "{{ build_dir }}/Dockerfile"
        content: |
          FROM node:20-alpine
          WORKDIR /app
          RUN echo 'console.log("Hello World")' > index.js
          CMD ["node", "index.js"]
        mode: '0644'

    - name: Build Docker image
      command: docker build -t {{ image_name }} {{ build_dir }}

    - name: Run custom app container
      command: >
        docker run -d
        --name {{ container_name }}
        -p {{ container_port }}:8080
        --restart unless-stopped
        {{ image_name }}

    - name: Clean up build directory
      file:
        path: "{{ build_dir }}"
        state: absent
```

## ベストプラクティス

### 1. 冪等性の確保

同じPlaybookを複数回実行しても安全になるようにします。

✅ **推奨**
```yaml
- name: Create network
  command: docker network create app-network
  register: result
  failed_when: false             # エラーを無視
  changed_when: result.rc == 0   # 成功時のみ変更とみなす
```

❌ **避けるべき**
```yaml
- name: Create network
  command: docker network create app-network
  # エラー時にPlaybook全体が失敗
```

### 2. コンテナの既存確認

```yaml
- name: Check if container exists
  command: docker ps -a --filter name={{ container_name }} --format '{{{{.Names}}}}'
  register: existing_container
  changed_when: false
  failed_when: false

- name: Remove existing container if needed
  command: docker rm -f {{ container_name }}
  when: existing_container.stdout == container_name

- name: Run new container
  command: docker run -d --name {{ container_name }} ...
```

### 3. データの永続化

```yaml
vars:
  data_dir: "/opt/{{ container_name }}/data"
  log_dir: "/opt/{{ container_name }}/logs"

tasks:
  - name: Create directories
    file:
      path: "{{ item }}"
      state: directory
      mode: '0755'
    loop:
      - "{{ data_dir }}"
      - "{{ log_dir }}"
    become: true

  - name: Run container with volumes
    command: >
      docker run -d
      --name {{ container_name }}
      -v {{ data_dir }}:/app/data
      -v {{ log_dir }}:/app/logs
      ...
```

### 4. 環境変数の安全な扱い

```yaml
# パスワード等の機密情報
vars:
  db_password: ""
  api_key: ""

tasks:
  - name: Run app
    command: >
      docker run -d
      -e DB_PASSWORD={{ db_password }}
      -e API_KEY={{ api_key }}
      ...
    no_log: "{{ db_password != '' or api_key != '' }}"
```

### 5. ネットワーク設定

```yaml
# 複数コンテナ間で通信させる場合
tasks:
  - name: Create app network
    command: docker network create app-network
    failed_when: false

  - name: Run database
    command: >
      docker run -d
      --name mysql-server
      --network app-network
      ...

  - name: Run webapp (DBに接続)
    command: >
      docker run -d
      --name webapp
      --network app-network
      -e DB_HOST=mysql-server  # コンテナ名で名前解決
      ...
```

### 6. ヘルスチェックと状態確認

```yaml
- name: Run container
  command: docker run -d --name {{ container_name }} ...
  register: docker_result

- name: Wait for container to be healthy
  command: docker inspect --format='{{{{.State.Health.Status}}}}' {{ container_name }}
  register: health_status
  until: health_status.stdout == "healthy"
  retries: 10
  delay: 3
  when: docker_result.rc == 0
  failed_when: false

- name: Show success message
  debug:
    msg: "Container {{ container_name }} is running and healthy"
  when: docker_result.rc == 0
```

## よくある問題と解決方法

### 問題1: コンテナ名が既に存在する

**エラー**:
```
Error response from daemon: Conflict. The container name "/my-app" is already in use
```

**解決方法**:
```yaml
- name: Remove existing container
  command: docker rm -f {{ container_name }}
  register: rm_result
  failed_when: false

- name: Run new container
  command: docker run -d --name {{ container_name }} ...
```

### 問題2: ポートが既に使用されている

**エラー**:
```
Error starting userland proxy: listen tcp4 0.0.0.0:8080: bind: address already in use
```

**解決方法**:
- variables.ymlで異なるポート番号を指定できるようにする
- または既存コンテナを停止

### 問題3: ディレクトリの権限エラー

**エラー**:
```
Permission denied: '/opt/app/data'
```

**解決方法**:
```yaml
- name: Create directory with proper permissions
  file:
    path: "{{ data_dir }}"
    state: directory
    mode: '0755'
    owner: root
    group: root
  become: true  # sudo権限で実行
```

### 問題4: Dockerネットワークが作成できない

**エラー**:
```
Error response from daemon: network with name app-network already exists
```

**解決方法**:
```yaml
- name: Create network (ignore if exists)
  command: docker network create {{ docker_network }}
  register: network_result
  failed_when: false
  changed_when: network_result.rc == 0
```

## テスト方法

### 1. ローカルで直接実行

```bash
cd playbooks/containers/your-app
ansible-playbook main.yml
```

### 2. 環境変数付きで実行

```bash
ansible-playbook main.yml \
  -e "container_name=test-app" \
  -e "container_port=9000"
```

### 3. 冪等性のテスト

```bash
# 1回目
ansible-playbook main.yml

# 2回目（エラーなく完了することを確認）
ansible-playbook main.yml
```

### 4. コンテナの動作確認

```bash
# コンテナが起動しているか確認
docker ps | grep your-app

# ログ確認
docker logs your-app

# 接続テスト
curl http://localhost:8080
```

## チェックリスト

Playbookを作成したら、以下を確認してください。

- [ ] main.ymlが正しいYAML形式で記述されている
- [ ] README.mdに最初の段落で説明が記載されている
- [ ] variables.yml（使用する場合）が正しい形式
- [ ] デフォルト値が設定されている
- [ ] 機密情報に`no_log`が設定されている
- [ ] 冪等性が確保されている（2回実行してもエラーにならない）
- [ ] データディレクトリが適切に作成・マウントされている
- [ ] ローカルで実行テスト済み
- [ ] コンテナが正常に起動することを確認済み

## 配置場所

作成したPlaybookは以下に配置します:

```bash
playbooks/
└── containers/
    ├── nginx/           # 既存Playbook
    ├── mysql/
    ├── redis/
    └── your-app/       # ← 新規Playbook
        ├── main.yml
        ├── README.md
        └── variables.yml
```

配置後、Webアプリケーションを再起動すると一覧に表示されます。

```bash
# 開発環境の場合
cd playbooks/app/webapp
make run

# 本番環境の場合
sudo systemctl restart kdinstall-webapp
```

## まとめ

Playbook作成の基本手順:

1. `playbooks/containers/`配下に新規ディレクトリ作成
2. `main.yml`でAnsibleタスクを記述
3. `README.md`で説明を記載
4. 必要に応じて`variables.yml`で環境変数定義
5. ローカルでテスト実行
6. Webアプリで動作確認

より詳しい情報は[VARIABLES.md](./VARIABLES.md)を参照してください。
