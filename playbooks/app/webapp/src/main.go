package main

import (
	"log"
	"webapp/src/config"
)

func main() {
	// ルータ起動
	router := initRouter()
	cfg := config.GetEnv()
	addr := ":" + cfg.ServerPort

	// HTTPS モードのみ（HTTPは無効）
	log.Printf("Server starting on https://localhost%s", addr)
	if err := router.RunTLS(addr, cfg.SSLCertPath, cfg.SSLKeyPath); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
