package routes

import (
	"api-financial/controllers"
	"api-financial/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupWebhookRoutes(router *gin.Engine, db *gorm.DB) {
	webhookService := services.NewWebhookService(db)
	webhookController := controllers.NewWebhookController(webhookService)
	v1 := router.Group("/api")
	{
		v1.POST("/webhooks/pipefy/card-updated", webhookController.HandleCardUpdated)
	}
}
