package routes

import (
	"api-financial/controllers"
	"api-financial/services"
	"log/slog"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupWebhookRoutes(router *gin.Engine, db *gorm.DB, logger *slog.Logger) {
	webhookService := services.NewWebhookService(db, logger)
	webhookController := controllers.NewWebhookController(webhookService)
	v1 := router.Group("/api")
	{
		v1.POST("/webhooks/pipefy/card-updated", webhookController.HandleCardUpdated)
	}
}
