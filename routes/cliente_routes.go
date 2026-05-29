package routes

import (
	"api-financial/controllers"
	"api-financial/services"
	"log/slog"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupClienteRoutes(router *gin.Engine, db *gorm.DB, logger *slog.Logger) {
	clienteService := services.NewClienteService(db, logger)

	clienteController := controllers.NewClienteController(clienteService)

	v1 := router.Group("/api")
	{
		v1.POST("/clientes", clienteController.Create)
	}
}
