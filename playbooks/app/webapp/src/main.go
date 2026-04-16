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

	if cfg.EnableSSL {
		// HTTPS モード
		log.Printf("Server starting on https://localhost%s", addr)
		if err := router.RunTLS(addr, cfg.SSLCertPath, cfg.SSLKeyPath); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	} else {
		// HTTP モード (開発用)
		log.Printf("Server starting on http://localhost%s", addr)
		if err := router.Run(addr); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}
}
