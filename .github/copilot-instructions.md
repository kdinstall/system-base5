# Project Guidelines

## Project Overview

This project provides **one-line server environment setup** for Ubuntu 24 servers. Execute a single command to automatically install:

- **Docker** + CLI tools
- **Go Web Application** (Docker container management UI)
- **systemd service** (`kdinstall-webapp`) for continuous operation

**Installation command:**
```bash
curl -fsSL https://raw.githubusercontent.com/kdinstall/system-base5/master/script/start.sh | bash
```

**Target OS:** Ubuntu 24 only (root access required for installation)

## Architecture

Three-layer architecture:

1. **Installation Script** (`script/start.sh`)
   - Downloads latest release from GitHub
   - Executes Ansible playbooks in sequence

2. **Ansible Playbooks**
   - `playbooks/docker/main.yml` - Docker environment setup
   - `playbooks/app/main.yml` - Web application build and deployment
   - `playbooks/containers/*/main.yml` - Preset container templates

3. **Web Application** (`playbooks/app/webapp/`)
   - Go + Gin framework
   - Provides Docker container management UI
   - Executes Ansible playbooks for container installation

**Execution flow:**
```
script/start.sh → Docker setup → SSL cert gen → Web app build → systemd service
                                                                         ↓
                                                                  Web UI (:58080 HTTPS)
                                                                         ↓
                                                     Docker CLI / Ansible async execution
                                                                         ↓
                                                              Job Manager (in-memory)
                                                                         ↓
                                                              Real-time log streaming
```

## Technology Stack

**System:**
- Ansible (core latest)
- Docker 8.0.0 (geerlingguy.docker role)
- systemd (service management)

**Web Application:**
- Go 1.23.0+
- Gin v1.11.0 (web framework)
- Tailwind CSS v4 (frontend styling)
- Preline UI (UI components)
- Node.js (CSS build tool)
- html/template (Go templating)
- github.com/google/uuid v1.6.0 (job ID generation)

**Dependencies:**
- Docker CLI (via os/exec)
- Ansible CLI (ansible-playbook via os/exec)

**Concurrency:**
- sync.RWMutex (job map protection)
- goroutines (async playbook execution)
- context.Context (cancellation support)

## Directory Structure

```
script/
  start.sh                          # One-line installer
playbooks/
  docker/
    main.yml                        # Docker installation
    requirements.yml                # Ansible Galaxy dependencies
  app/
    main.yml                        # Web app deployment
    webapp/                         # Go source code
      src/
        main.go                     # Entry point
        router.go                   # Route definitions
        config/config.go            # Environment variables
        controllers/                # HTTP handlers
        lib/                        # Business logic
          ansible/ansible.go        # Ansible CLI wrapper (sync/async)
          docker/docker.go          # Docker CLI wrapper
          job/manager.go            # Job manager (singleton)
          playbook/manager.go       # Playbook management
          template/                 # Template utilities
        templates/                  # HTML templates
          containers.html           # Container list
          container_logs.html       # Container logs
          install.html              # Playbook installation
          install_config.html       # Variable configuration
          job_list.html             # Job history
          job_detail.html           # Job detail with real-time logs
        style/input.css             # Tailwind source
      docs/                         # Detailed documentation
      Makefile                      # Build commands
      go.mod                        # Go dependencies
      package.json                  # Node.js dependencies
  containers/
    nginx/main.yml                  # Nginx container setup
    mysql/main.yml                  # MySQL container setup
    postgresql/main.yml             # PostgreSQL container setup
    mongodb/main.yml                # MongoDB container setup
    redis/main.yml                  # Redis container setup
    nodejs-webapp/main.yml          # Node.js app container
```

## Web Application

**Deployed to:** `/opt/kdinstall/webapp`  
**Binary:** `/opt/kdinstall/bin/webapp`  
**User:** `kdi` (docker group member)  
**Default port:** 58080 (HTTPS)

### Controllers

**ContainerController** (`src/controllers/containerController.go`)
- `GET /containers` - List all containers
- `POST /containers/:id/start` - Start container
- `POST /containers/:id/stop` - Stop container
- `POST /containers/:id/restart` - Restart container
- `GET /containers/:id/logs` - View container logs

**InstallController** (`src/controllers/installController.go`)
- `GET /install` - Installation screen (playbook list)
- `GET /install/:name/config` - Configuration form (reads variables.yml)
- `POST /install/execute` - Execute playbook asynchronously with environment variables
- `GET /install/jobs` - Job list (installation history)
- `GET /install/jobs/:id` - Job detail with real-time logs
- `GET /install/jobs/:id/logs` - Get job logs incrementally (JSON)
- `GET /install/jobs/:id/status` - Get job status (JSON)
- `GET /install/jobs/running` - Get currently running job (JSON)

### Key Features

- **Preset playbooks**: Nginx, MySQL, PostgreSQL, MongoDB, Redis, Node.js webapp
- **Dynamic download**: Fetch playbooks from Git URLs or direct YAML files
- **Variable injection**: POST form fields (`env_*` prefix) → Ansible extra vars
- **Async execution**: Playbooks run in background via goroutines
- **Job management**: Track installation history and status in memory
- **Real-time logs**: JavaScript polling updates logs every second
- **Single job limit**: Only one installation can run at a time (prevents resource conflicts)
- **Docker operations**: Execute `docker` commands via `os/exec.Command`
- **Ansible execution**: Execute `ansible-playbook` via `lib/ansible/ansible.go`

### Directory Structure

```
src/
  main.go                           # Gin server initialization
  router.go                         # Route registration + template loading
  config/config.go                  # SERVER_PORT, PLAYBOOKS_DIR
  controllers/
    containerController.go          # Docker management
    installController.go            # Playbook execution
  lib/
    docker/docker.go                # Docker CLI wrapper
    ansible/ansible.go              # Ansible CLI wrapper (sync/async)
    job/manager.go                  # Job manager (singleton, in-memory)
    playbook/manager.go             # Playbook download/management
    template/                       # Template utilities
  templates/                        # HTML files
    containers.html                 # Container list
    container_logs.html             # Container logs
    install.html                    # Playbook installation
    install_config.html             # Variable configuration
    job_list.html                   # Job history
    job_detail.html                 # Job detail with real-time logs
  style/
    input.css                       # Tailwind source
    output.css                      # Generated CSS (git-ignored)
```

## Development Guide

### Web Application Development

```bash
cd playbooks/app/webapp

# Full development cycle (install + build + run)
make dev

# Build CSS only
make build

# Watch mode (auto-rebuild CSS)
make watch

# Run app (requires pre-built CSS)
make run
```

**Detailed documentation:** See `playbooks/app/webapp/docs/DEVELOPMENT.md`

### Deployment

```bash
# Full system deployment
curl -fsSL https://raw.githubusercontent.com/kdinstall/system-base5/master/script/start.sh | bash

# Update web app only
cd playbooks/app
ansible-galaxy install --role-file=requirements.yml
ansible-playbook -i localhost, main.yml
```

### Testing

```bash
# Test mode (uses master branch instead of latest release)
bash script/start.sh -test
```

## Implementation Conventions

### Job Management Pattern

**Creating and executing async jobs:**
```go
import (
    "github.com/google/uuid"
    "webapp/src/lib/job"
    "webapp/src/lib/ansible"
)

// Generate job ID
jobID := uuid.New().String()

// Create job in manager
jm := job.GetManager()
jm.CreateJob(jobID, playbookName)

// Execute asynchronously
err := ansible.RunPlaybookAsync(jobID, playbookPath, "local", extraVars)
if err != nil {
    return err
}

// Redirect to job detail page
c.Redirect(http.StatusSeeOther, fmt.Sprintf("/install/jobs/%s", jobID))
```

**Job status checking (single job limit):**
```go
jm := job.GetManager()
if runningJob := jm.GetRunningJob(); runningJob != nil {
    return c.HTML(http.StatusUnprocessableEntity, "install.html", gin.H{
        "error": fmt.Sprintf("ジョブ '%s' が実行中です", runningJob.Name),
    })
}
```

### Ansible Variable Handling

**Reading variables.yml:**
```go
// lib/playbook/manager.go
variables := parseVariablesFile("variables.yml")
// Returns []Variable with name, label, type, default, required
```

**Passing to Ansible:**
```go
// Form fields: env_container_name, env_db_host, etc.
extraVars := []string{}
for key, value := range form {
    if strings.HasPrefix(key, "env_") {
        varName := strings.TrimPrefix(key, "env_")
        extraVars = append(extraVars, varName+"="+value)
    }
}
// Execute async: ansible-playbook --extra-vars "container_name=value db_host=value"
```

### Docker/Ansible Command Execution

**Pattern:**
```go
import "os/exec"

// Docker example
cmd := exec.Command("docker", "ps", "-a", "--format", "{{.ID}}|{{.Names}}|{{.Status}}")
output, err := cmd.CombinedOutput()

// Ansible example
cmd := exec.Command("ansible-playbook", "-i", "localhost,", "main.yml", "--extra-vars", vars)
```

### Error Handling

**Controller pattern:**
```go
if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{
        "error": err.Error(),
    })
    return
}
```

### Tailwind CSS Build

**Build process:**
```bash
npm install --include=optional    # Install Tailwind v4
npm run build                     # Generates src/style/output.css
```

**package.json script:**
```json
{
  "scripts": {
    "build": "tailwindcss -i ./src/style/input.css -o ./src/style/output.css",
    "watch": "tailwindcss -i ./src/style/input.css -o ./src/style/output.css --watch"
  }
}
```

### Template Loading

**Pattern:**
```go
// router.go
router.LoadHTMLGlob("src/templates/*.html")

// Controller
c.HTML(http.StatusOK, "containers.html", gin.H{
    "containers": containers,
})
```

**Log rendering (preserve newlines):**
```html
<!-- CORRECT: Preserve whitespace -->
<pre id="log-content">{{ range .job.Logs }}{{ . }}
{{ end }}</pre>

<!-- WRONG: Strips newlines with {{- -->
<pre id="log-content">{{- range .job.Logs }}{{ . }}
{{- end }}</pre>
```

## Build and Deploy

### Production Build (via playbooks/app/main.yml)

1. Install system dependencies (Go, Node.js, Ansible, Git)
2. Create user `kdi` and add to docker group
3. Create directory structure under `/opt/kdinstall/`
4. Copy source code to `/opt/kdinstall/webapp`
5. Build CSS: `npm install && npm run build`
6. Build binary: `go build -o /opt/kdinstall/bin/webapp ./src`
7. Configure systemd service `kdinstall-webapp`

### systemd Service Configuration

**Service file:** `/etc/systemd/system/kdinstall-webapp.service`

```ini
[Unit]
Description=kdinstall-webapp HTTPS web application
After=network-online.target docker.service

[Service]
Type=simple
User=kdi
Group=kdi
ExecStart=/opt/kdinstall/bin/webapp
Environment=SERVER_PORT=58080
Environment=PLAYBOOKS_DIR=/opt/kdinstall/containers
Environment=SSL_CERT_PATH=/opt/kdinstall/certs/server.crt
Environment=SSL_KEY_PATH=/opt/kdinstall/certs/server.key
Restart=always
```

### Environment Variables

- `SERVER_PORT` - Web server listen port (default: 58080)
- `PLAYBOOKS_DIR` - Container playbooks directory (default: `/opt/kdinstall/containers`)
- `SSL_CERT_PATH` - SSL certificate file path (default: `/opt/kdinstall/certs/server.crt`)
- `SSL_KEY_PATH` - SSL private key file path (default: `/opt/kdinstall/certs/server.key`)

**Configuration:** Set in `playbooks/app/main.yml` variables or systemd service file

### SSL/TLS Configuration

The application serves **HTTPS only** (HTTP is disabled) using a self-signed certificate:

- Certificate is auto-generated on first deployment (10 years validity)
- Stored at `/opt/kdinstall/certs/server.{crt,key}`
- For development, generate test certificates with `openssl` command
- For production with Let's Encrypt, update `SSL_CERT_PATH` and `SSL_KEY_PATH` in systemd service

### Update Strategy

Re-run the one-line installer to fetch latest code and rebuild:
```bash
curl -fsSL https://raw.githubusercontent.com/kdinstall/system-base5/master/script/start.sh | bash
```

The playbook detects changes and rebuilds only when necessary.

## Code Style

- **Go**: Follow standard Go conventions, use `gofmt`
- **Templates**: Keep logic minimal, pass data from controllers
- **CSS**: Use Tailwind utility classes, avoid custom CSS when possible
- **Ansible**: Use YAML best practices, document variables clearly

## Key Files

- `script/start.sh` - Entry point for one-line installation
- `playbooks/app/webapp/src/main.go` - Web application entry point
- `playbooks/app/webapp/src/router.go` - Route definitions and template loading
- `playbooks/app/webapp/src/lib/job/manager.go` - Job management (singleton pattern)
- `playbooks/app/webapp/src/lib/ansible/ansible.go` - Ansible execution (sync/async)
- `playbooks/app/webapp/src/controllers/installController.go` - Installation and job endpoints
- `playbooks/app/webapp/src/templates/job_detail.html` - Real-time log viewer
- `playbooks/app/webapp/docs/` - Detailed architecture and API documentation
- `playbooks/app/main.yml` - Web app deployment playbook
- `playbooks/docker/main.yml` - Docker setup playbook
