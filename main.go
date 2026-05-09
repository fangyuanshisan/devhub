package main

import (
	"log"
	"os"

	"devhub-gin-backend/internal/service"
	"devhub-gin-backend/internal/store"
	"devhub-gin-backend/internal/transport/httpapi"

	"github.com/gin-gonic/gin"
)

// main 组装仓储、业务服务和 HTTP 路由后启动 Gin 服务。
func main() {
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	var repo service.Repository
	if os.Getenv("CMS_STORE") == "memory" {
		repo = store.NewMemoryStore()
	} else {
		mysqlStore, err := store.NewMySQLStore(store.MySQLConfig{
			Host:     env("DB_HOST", "127.0.0.1"),
			Port:     env("DB_PORT", "3307"),
			User:     env("DB_USER", "devhub"),
			Password: env("DB_PASSWORD", "Devhub_123456"),
			Database: env("DB_NAME", "devhub"),
		})
		if err != nil {
			log.Fatalf("connect mysql: %v", err)
		}
		defer mysqlStore.Close()
		repo = mysqlStore
	}

	svc := service.New(repo)
	router := httpapi.NewRouter(svc)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("listen on port %s: %v", port, err)
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
