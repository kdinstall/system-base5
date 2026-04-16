package main

import (
	"log"
	"webapp/src/config"
)

func main() {
	// ルータ起動
	router := initRouter()
	addr := ":" + config.GetEnv().ServerPort
	log.Printf("Server starting on http://localhost%s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
