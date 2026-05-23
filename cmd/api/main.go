package main

import (
	"api-financial/config"
	"api-financial/routes"
	"log"

	"github.com/gin-gonic/gin"
)

// @title           API Financial - MVP
// @version         1.0
// @description     API de gerenciamento financeiro e integração com CRM.
// @host            localhost:8080
// @BasePath        /ap1
func main() {
	db := config.ConnectDB()
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
			"status":  "running",
		})
	})
	routes.SetupClienteRoutes(r, db)
	routes.SetupSwaggerRoute(r)
	log.Println("Servidor HTTP iniciado na porta :8080")

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Falha ao iniciar o servidor: %v", err)
	}
}
