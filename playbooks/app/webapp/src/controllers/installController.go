package controllers

import (
	"net/http"
	"strings"
	"webapp/src/config"
	"webapp/src/lib/ansible"
	"webapp/src/lib/playbook"
	tmpl "webapp/src/lib/template"

	"github.com/gin-gonic/gin"
)

// InstallController はコンテナインストールのアクションを提供する
type InstallController struct{}

// Index はインストール画面を表示する (GET /install)
func (ic *InstallController) Index(c *gin.Context) {
	basePath := config.GetEnv().PlaybooksDir

	playbooks, err := playbook.ListLocalPlaybooks(basePath)

	var errorMsg string
	if err != nil {
		errorMsg = "Playbook一覧の取得に失敗しました: " + err.Error()
	}

	c.HTML(http.StatusOK, "install.html", tmpl.MergeData(gin.H{
		"page_title":  "コンテナインストール",
		"active_page": "install",
		"playbooks":   playbooks,
		"error":       errorMsg,
		"flash":       c.Query("flash"),
	}))
}

// Config は環境変数設定画面を表示する (GET /install/:name/config)
func (ic *InstallController) Config(c *gin.Context) {
	playbookName := c.Param("name")
	basePath := config.GetEnv().PlaybooksDir

	// Playbookの存在確認
	if err := playbook.ValidatePlaybookExists(basePath, playbookName); err != nil {
		c.HTML(http.StatusNotFound, "404.html", tmpl.MergeData(gin.H{"page_title": "Not Found"}))
		return
	}

	// variables.ymlを読み込む
	playbookDir := basePath + "/" + playbookName
	variables, err := playbook.ReadVariables(playbookDir)
	if err != nil {
		playbooks, _ := playbook.ListLocalPlaybooks(basePath)
		c.HTML(http.StatusInternalServerError, "install.html", tmpl.MergeData(gin.H{
			"page_title":  "コンテナインストール",
			"active_page": "install",
			"playbooks":   playbooks,
			"error":       "環境変数定義の読み込みに失敗しました: " + err.Error(),
		}))
		return
	}

	// 環境変数設定画面を表示
	c.HTML(http.StatusOK, "install_config.html", tmpl.MergeData(gin.H{
		"page_title":    "環境変数設定",
		"active_page":   "install",
		"playbook_name": playbookName,
		"variables":     variables,
	}))
}

// Execute はPlaybookを実行してコンテナをインストールする (POST /install/execute)
func (ic *InstallController) Execute(c *gin.Context) {
	playbookName := c.PostForm("playbook")
	downloadURL := c.PostForm("download_url")
	downloadType := c.PostForm("download_type") // "git" or "url"

	basePath := config.GetEnv().PlaybooksDir

	// URL指定でのダウンロード
	if downloadURL != "" {
		// 名前を生成（URLの最後の部分から）
		parts := strings.Split(strings.TrimSuffix(downloadURL, "/"), "/")
		name := parts[len(parts)-1]
		name = strings.TrimSuffix(name, ".git")
		name = strings.TrimSuffix(name, ".yml")
		name = strings.TrimSuffix(name, ".yaml")

		var err error
		if downloadType == "git" {
			err = playbook.DownloadFromGit(downloadURL, basePath, name)
		} else {
			err = playbook.DownloadFromURL(downloadURL, basePath, name)
		}

		if err != nil {
			playbooks, _ := playbook.ListLocalPlaybooks(basePath)
			c.HTML(http.StatusUnprocessableEntity, "install.html", tmpl.MergeData(gin.H{
				"page_title":  "コンテナインストール",
				"active_page": "install",
				"playbooks":   playbooks,
				"error":       "ダウンロードに失敗しました: " + err.Error(),
				"input":       gin.H{"download_url": downloadURL},
			}))
			return
		}

		playbookName = name
	}

	// Playbookの存在確認
	if err := playbook.ValidatePlaybookExists(basePath, playbookName); err != nil {
		playbooks, _ := playbook.ListLocalPlaybooks(basePath)
		c.HTML(http.StatusUnprocessableEntity, "install.html", tmpl.MergeData(gin.H{
			"page_title":  "コンテナインストール",
			"active_page": "install",
			"playbooks":   playbooks,
			"error":       "Playbookが見つかりません: " + playbookName,
		}))
		return
	}

	// 環境変数を受け取る（env_で始まるフィールド）
	var extraVars []string
	for key, values := range c.Request.PostForm {
		if strings.HasPrefix(key, "env_") && len(values) > 0 {
			// env_をプレフィックスから削除
			varName := strings.TrimPrefix(key, "env_")
			varValue := values[0]
			// 空文字列でない場合のみ追加
			if varValue != "" {
				extraVars = append(extraVars, varName+"="+varValue)
			}
		}
	}

	// Playbook実行
	playbookPath := playbook.GetPlaybookPath(basePath, playbookName)
	result := ansible.RunPlaybookWithConnection(playbookPath, "local", extraVars)

	// 結果を表示
	c.HTML(http.StatusOK, "install.html", tmpl.MergeData(gin.H{
		"page_title":    "コンテナインストール",
		"active_page":   "install",
		"playbooks":     nil, // インストール結果表示時は一覧表示しない
		"result":        result,
		"playbook_name": playbookName,
		"show_result":   true,
	}))
}
