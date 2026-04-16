package docker

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Container はDockerコンテナの情報を保持する
type Container struct {
	ID      string
	Name    string
	Image   string
	Status  string
	State   string // running, exited, paused, etc.
	Created string
	Ports   string
}

// ListContainers は全てのコンテナ一覧を取得する
func ListContainers() ([]Container, error) {
	cmd := exec.Command("docker", "ps", "-a", "--format", "{{json .}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("docker ps failed: %v, output: %s", err, string(output))
	}

	var containers []Container
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		var rawContainer map[string]interface{}
		if err := json.Unmarshal([]byte(line), &rawContainer); err != nil {
			continue
		}

		container := Container{
			ID:      getString(rawContainer, "ID"),
			Name:    strings.TrimPrefix(getString(rawContainer, "Names"), "/"),
			Image:   getString(rawContainer, "Image"),
			Status:  getString(rawContainer, "Status"),
			State:   getString(rawContainer, "State"),
			Created: getString(rawContainer, "CreatedAt"),
			Ports:   getString(rawContainer, "Ports"),
		}

		containers = append(containers, container)
	}

	return containers, nil
}

// StartContainer は指定されたコンテナを起動する
func StartContainer(containerID string) error {
	cmd := exec.Command("docker", "start", containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker start failed: %v, output: %s", err, string(output))
	}
	return nil
}

// StopContainer は指定されたコンテナを停止する
func StopContainer(containerID string) error {
	cmd := exec.Command("docker", "stop", containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker stop failed: %v, output: %s", err, string(output))
	}
	return nil
}

// RestartContainer は指定されたコンテナを再起動する
func RestartContainer(containerID string) error {
	cmd := exec.Command("docker", "restart", containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker restart failed: %v, output: %s", err, string(output))
	}
	return nil
}

// GetLogs は指定されたコンテナのログを取得する（最新100行）
func GetLogs(containerID string) (string, error) {
	cmd := exec.Command("docker", "logs", "--tail", "100", containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("docker logs failed: %v", err)
	}
	return string(output), nil
}

// InspectContainer は指定されたコンテナの詳細情報を取得する
func InspectContainer(containerID string) (map[string]interface{}, error) {
	cmd := exec.Command("docker", "inspect", containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("docker inspect failed: %v, output: %s", err, string(output))
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse docker inspect output: %v", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("container not found")
	}

	return result[0], nil
}

// GetContainerByID は指定されたIDのコンテナを取得する
func GetContainerByID(containerID string) (*Container, error) {
	containers, err := ListContainers()
	if err != nil {
		return nil, err
	}

	for _, c := range containers {
		if c.ID == containerID || strings.HasPrefix(c.ID, containerID) {
			return &c, nil
		}
	}

	return nil, fmt.Errorf("container not found: %s", containerID)
}

// GetContainerByName は指定された名前のコンテナを取得する
func GetContainerByName(name string) (*Container, error) {
	containers, err := ListContainers()
	if err != nil {
		return nil, err
	}

	for _, c := range containers {
		if c.Name == name {
			return &c, nil
		}
	}

	return nil, fmt.Errorf("container not found: %s", name)
}

// IsRunning は指定されたコンテナが実行中かどうかを返す
func IsRunning(containerID string) (bool, error) {
	container, err := GetContainerByID(containerID)
	if err != nil {
		return false, err
	}
	return container.State == "running", nil
}

// FormatCreatedTime は作成日時を人間が読みやすい形式に変換する
func FormatCreatedTime(created string) string {
	// Docker の CreatedAt フォーマットをパース
	layouts := []string{
		"2006-01-02 15:04:05 -0700 MST",
		"2006-01-02 15:04:05 -0700",
		time.RFC3339,
	}

	var t time.Time
	var err error
	for _, layout := range layouts {
		t, err = time.Parse(layout, created)
		if err == nil {
			break
		}
	}

	if err != nil {
		return created
	}

	duration := time.Since(t)
	if duration < time.Minute {
		return fmt.Sprintf("%d秒前", int(duration.Seconds()))
	} else if duration < time.Hour {
		return fmt.Sprintf("%d分前", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%d時間前", int(duration.Hours()))
	} else {
		return fmt.Sprintf("%d日前", int(duration.Hours()/24))
	}
}

// getString は map から文字列を安全に取得する
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
