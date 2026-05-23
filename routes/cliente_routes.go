package routes

import (
	"api-financial/controllers"
	"api-financial/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupClienteRoutes(router *gin.Engine, db *gorm.DB) {
	clienteService := services.NewClienteService(db)

	clienteController := controllers.NewClienteController(clienteService)

	v1 := router.Group("/api")
	{
		v1.POST("/clientes", clienteController.Create)
	}
}
