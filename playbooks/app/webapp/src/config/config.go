package config

import (
	"log"
	"os"
	"path/filepath"
)

// Env はアプリケーション設定を保持する構造体
type Env struct {
	AppName      string
	ServerPort   string
	PlaybooksDir string
}

// GetEnv は設定を返す（環境変数でオーバーライド可能）
func GetEnv() Env {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "58080"
	}

	playbooksDir := os.Getenv("PLAYBOOKS_DIR")
	if playbooksDir == "" {
		// デフォルト: /opt/kdinstall/containers
		// WorkingDirectory=/opt/kdinstall/webapp から相対パスで取得
		wd, _ := os.Getwd()
		playbooksDir = filepath.Join(wd, "..", "containers")
	}

	// playbooksDir が空でないかチェック
	if playbooksDir == "" {
		log.Println("エラー: PLAYBOOKS_DIR が設定されていません")
	} else {
		// ディレクトリの存在確認
		if info, err := os.Stat(playbooksDir); err != nil {
			if os.IsNotExist(err) {
				log.Printf("エラー: playbooksディレクトリが存在しません: %s\n", playbooksDir)
			} else {
				log.Printf("エラー: playbooksディレクトリの確認に失敗しました: %s (%v)\n", playbooksDir, err)
			}
		} else if !info.IsDir() {
			log.Printf("エラー: playbooksパスがディレクトリではありません: %s\n", playbooksDir)
		}
	}

	return Env{
		AppName:      "Docker管理",
		ServerPort:   port,
		PlaybooksDir: playbooksDir,
	}
}
