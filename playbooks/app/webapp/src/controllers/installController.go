package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"
	"webapp/src/config"
	"webapp/src/lib/ansible"
	"webapp/src/lib/job"
	"webapp/src/lib/playbook"
	tmpl "webapp/src/lib/template"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	jobManager := job.GetManager()

	// Check if there's already a running job
	if runningJob := jobManager.GetRunningJob(); runningJob != nil {
		c.HTML(http.StatusConflict, "install.html", tmpl.MergeData(gin.H{
			"page_title":       "コンテナインストール",
			"active_page":      "install",
			"error":            "既にインストールが実行中です。完了するまでお待ちください。",
			"running_job_id":   runningJob.ID,
			"running_job_name": runningJob.Name,
		}))
		return
	}

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

	// Generate job ID
	jobID := uuid.New().String()

	// Create job
	jobManager.CreateJob(jobID, playbookName)

	// Run playbook asynchronously
	playbookPath := playbook.GetPlaybookPath(basePath, playbookName)
	ansible.RunPlaybookAsync(jobID, playbookPath, "local", extraVars)

	// Redirect to job detail page
	c.Redirect(http.StatusFound, "/install/jobs/"+jobID)
}

// JobList displays a list of all jobs (GET /install/jobs)
func (ic *InstallController) JobList(c *gin.Context) {
	jobManager := job.GetManager()
	jobs := jobManager.ListJobs()

	c.HTML(http.StatusOK, "job_list.html", tmpl.MergeData(gin.H{
		"page_title":  "ジョブ一覧",
		"active_page": "install",
		"jobs":        jobs,
	}))
}

// JobDetail displays details of a specific job (GET /install/jobs/:id)
func (ic *InstallController) JobDetail(c *gin.Context) {
	jobID := c.Param("id")
	jobManager := job.GetManager()
	j := jobManager.GetJob(jobID)

	if j == nil {
		c.HTML(http.StatusNotFound, "404.html", tmpl.MergeData(gin.H{
			"page_title": "Not Found",
		}))
		return
	}

	c.HTML(http.StatusOK, "job_detail.html", tmpl.MergeData(gin.H{
		"page_title":  "ジョブ詳細",
		"active_page": "install",
		"job":         j,
	}))
}

// JobLogs returns job logs as JSON (GET /install/jobs/:id/logs?offset=N)
func (ic *InstallController) JobLogs(c *gin.Context) {
	jobID := c.Param("id")
	offsetStr := c.DefaultQuery("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	jobManager := job.GetManager()
	j := jobManager.GetJob(jobID)

	if j == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found",
		})
		return
	}

	logs := jobManager.GetLogs(jobID, offset)

	c.JSON(http.StatusOK, gin.H{
		"logs":   logs,
		"offset": offset + len(logs),
	})
}

// JobStatus returns job status as JSON (GET /install/jobs/:id/status)
func (ic *InstallController) JobStatus(c *gin.Context) {
	jobID := c.Param("id")
	jobManager := job.GetManager()
	j := jobManager.GetJob(jobID)

	if j == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         j.ID,
		"name":       j.Name,
		"status":     j.Status,
		"start_time": j.StartTime.Format(time.RFC3339),
		"end_time":   j.EndTime.Format(time.RFC3339),
		"error":      j.Error,
	})
}

// GetRunningJob returns the currently running job as JSON (GET /install/jobs/running)
func (ic *InstallController) GetRunningJob(c *gin.Context) {
	jobManager := job.GetManager()
	runningJob := jobManager.GetRunningJob()

	if runningJob == nil {
		c.JSON(http.StatusOK, gin.H{
			"running": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"running": true,
		"job": gin.H{
			"id":     runningJob.ID,
			"name":   runningJob.Name,
			"status": runningJob.Status,
		},
	})
}
