package main

import (
	"api-financial/config"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	config.ConnectDB()
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
			"status":  "running",
		})
	})
	log.Println("Servidor HTTP iniciado na porta :8080")

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Falha ao iniciar o servidor: %v", err)
	}
}
