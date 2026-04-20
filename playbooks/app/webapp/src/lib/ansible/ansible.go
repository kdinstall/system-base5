package ansible

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"webapp/src/lib/job"
)

// PlaybookResult はPlaybook実行結果を保持する
type PlaybookResult struct {
	Success bool
	Output  string
	Error   string
}

// RunPlaybook はAnsible Playbookを実行する
// playbookPath: Playbookファイルのパス（例: "playbooks/containers/nginx/main.yml"）
// extraVars: 追加変数（キー=値形式、例: []string{"container_name=my-nginx", "port=8080"}）
func RunPlaybook(playbookPath string, extraVars []string) *PlaybookResult {
	args := []string{"-i", "localhost,", playbookPath}

	// 追加変数がある場合
	if len(extraVars) > 0 {
		args = append(args, "-e", strings.Join(extraVars, " "))
	}

	cmd := exec.Command("ansible-playbook", args...)

	// Ansible一時ディレクトリを/tmpに設定（ホームディレクトリが存在しないユーザー対応）
	env := os.Environ()
	env = append(env, "ANSIBLE_LOCAL_TEMP=/tmp/ansible")
	env = append(env, "ANSIBLE_REMOTE_TEMP=/tmp/ansible")
	env = append(env, "ANSIBLE_NOCOLOR=True")
	cmd.Env = env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &PlaybookResult{
		Success: err == nil,
		Output:  stdout.String(),
		Error:   stderr.String(),
	}

	if err != nil && result.Error == "" {
		result.Error = err.Error()
	}

	return result
}

// RunPlaybookWithConnection はAnsible Playbookを接続タイプ指定で実行する
// connection: "local" または "ssh" など
func RunPlaybookWithConnection(playbookPath string, connection string, extraVars []string) *PlaybookResult {
	args := []string{"-i", "localhost,", "--connection", connection, playbookPath}

	// 追加変数がある場合
	if len(extraVars) > 0 {
		args = append(args, "-e", strings.Join(extraVars, " "))
	}

	cmd := exec.Command("ansible-playbook", args...)

	// Ansible一時ディレクトリを/tmpに設定（ホームディレクトリが存在しないユーザー対応）
	env := os.Environ()
	env = append(env, "ANSIBLE_LOCAL_TEMP=/tmp/ansible")
	env = append(env, "ANSIBLE_REMOTE_TEMP=/tmp/ansible")
	env = append(env, "ANSIBLE_NOCOLOR=True")
	cmd.Env = env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &PlaybookResult{
		Success: err == nil,
		Output:  stdout.String(),
		Error:   stderr.String(),
	}

	if err != nil && result.Error == "" {
		result.Error = err.Error()
	}

	return result
}

// CheckAnsibleInstalled はAnsibleがインストールされているかチェックする
func CheckAnsibleInstalled() error {

	cmd := exec.Command("ansible-playbook", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ansible-playbook が見つかりません。Ansibleをインストールしてください")
	}
	return nil
}

// GetAnsibleVersion はAnsibleのバージョンを取得する
func GetAnsibleVersion() (string, error) {
	cmd := exec.Command("ansible-playbook", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	// 最初の行を返す
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		return lines[0], nil
	}

	return string(output), nil
}

// FormatPlaybookOutput はPlaybook出力を見やすく整形する
func FormatPlaybookOutput(output string) string {
	// ANSI カラーコードを削除（簡易版）
	// 本格的には regexp を使う必要がある
	output = strings.ReplaceAll(output, "\x1b[0m", "")
	output = strings.ReplaceAll(output, "\x1b[1m", "")
	output = strings.ReplaceAll(output, "\x1b[32m", "")
	output = strings.ReplaceAll(output, "\x1b[33m", "")
	output = strings.ReplaceAll(output, "\x1b[31m", "")
	output = strings.ReplaceAll(output, "\x1b[36m", "")

	return output
}

// RunPlaybookAsync runs an Ansible playbook asynchronously in a goroutine
// jobID: unique job identifier for logging
// playbookPath: path to the playbook file
// connection: connection type (e.g., "local")
// extraVars: extra variables to pass to ansible-playbook
func RunPlaybookAsync(jobID, playbookPath, connection string, extraVars []string) {
	go func() {
		jobManager := job.GetManager()

		// Update status to running
		jobManager.UpdateStatus(jobID, "running")
		jobManager.AppendLog(jobID, fmt.Sprintf("Starting playbook: %s", playbookPath))

		// Create context for cancellation support
		ctx, cancel := context.WithCancel(context.Background())
		jobManager.SetCancelFunc(jobID, cancel)

		// Build command arguments
		args := []string{"-i", "localhost,", "--connection", connection, playbookPath}
		if len(extraVars) > 0 {
			args = append(args, "-e", strings.Join(extraVars, " "))
		}

		// Create command with context
		cmd := exec.CommandContext(ctx, "ansible-playbook", args...)

		// Set environment variables
		env := os.Environ()
		env = append(env, "ANSIBLE_LOCAL_TEMP=/tmp/ansible")
		env = append(env, "ANSIBLE_REMOTE_TEMP=/tmp/ansible")
		env = append(env, "ANSIBLE_NOCOLOR=True")
		cmd.Env = env

		// Get stdout and stderr pipes
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			jobManager.SetError(jobID, fmt.Sprintf("Failed to create stdout pipe: %v", err))
			jobManager.UpdateStatus(jobID, "failed")
			return
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			jobManager.SetError(jobID, fmt.Sprintf("Failed to create stderr pipe: %v", err))
			jobManager.UpdateStatus(jobID, "failed")
			return
		}

		// Start the command
		if err := cmd.Start(); err != nil {
			jobManager.SetError(jobID, fmt.Sprintf("Failed to start ansible-playbook: %v", err))
			jobManager.UpdateStatus(jobID, "failed")
			return
		}

		// Read stdout in a goroutine
		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				line := scanner.Text()
				jobManager.AppendLog(jobID, line)
			}
		}()

		// Read stderr in a goroutine
		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				line := scanner.Text()
				jobManager.AppendLog(jobID, "[STDERR] "+line)
			}
		}()

		// Wait for command to complete
		err = cmd.Wait()

		// Check if context was cancelled
		if ctx.Err() == context.Canceled {
			jobManager.AppendLog(jobID, "Job was cancelled")
			jobManager.UpdateStatus(jobID, "failed")
			jobManager.SetError(jobID, "Job was cancelled by user")
			return
		}

		// Update final status
		if err != nil {
			jobManager.AppendLog(jobID, fmt.Sprintf("Playbook execution failed: %v", err))
			jobManager.SetError(jobID, err.Error())
			jobManager.UpdateStatus(jobID, "failed")
		} else {
			jobManager.AppendLog(jobID, "Playbook execution completed successfully")
			jobManager.UpdateStatus(jobID, "completed")
		}
	}()
}
