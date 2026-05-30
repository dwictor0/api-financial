package main

import (
	"api-financial/config"
	"api-financial/routes"
	"log"
	"log/slog"

	"github.com/gin-gonic/gin"
)

// @title           API Financial - MVP
// @version         1.0
// @description     API de gerenciamento financeiro e integração com CRM.
// @host            localhost:8080
// @BasePath        /ap1
func main() {
	config.InitLogger()
	logger := slog.Default()
	db := config.ConnectDB()
	r := gin.Default()
	routes.SetupClienteRoutes(r, db, logger)
	routes.SetupSwaggerRoute(r)
	routes.SetupWebhookRoutes(r, db, logger)
	log.Println("Servidor HTTP iniciado na porta :8080")

	if err := r.Run(":8080"); err != nil {
		slog.Error("Falha ao iniciar o servidor", "error", err.Error())
	}
}
