package controllers

import (
	"api-financial/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ClienteController struct {
	DB *gorm.DB
}

func NewClienteController(db *gorm.DB) *ClienteController {
	return &ClienteController{DB: db}
}

func (cc *ClienteController) Create(c *gin.Context) {
	var cliente models.Cliente

	if err := c.ShouldBindJSON(&cliente); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validacao falhou",
			"details": err.Error(),
		})
		return
	}
	cliente.Status = "aguardando_analise"
	cliente.Prioridade = "nao_definida"

	if err := cc.DB.Create(&cliente).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar o cliente no banco"})
	}
}
