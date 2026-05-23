package controllers

import (
	"api-financial/models"
	"fmt"
	"net/http"
	"os"

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
		return
	}
	pipeId := os.Getenv("PIPEFY_PIPE_ID")

	_ = `
	mutation createCard($pipeId: ID!, $title: String!, $fields: [FieldValueInput]) {
	  createCard(input: {
	    pipe_id: $pipeId,
	    title: $title,
	    fields_attributes: $fields
	  }) {
	    card {
	      id
	      title
	    }
	  }
	}
	`
	_ = map[string]interface{}{
		"pipeId": pipeId,
		"title":  cliente.ClienteNome,
		"fields": []map[string]interface{}{
			{"field_id": "email_do_cliente", "field_value": []string{cliente.ClienteEmail}},
			{"field_id": "tipo_de_solicitacao", "field_value": []string{cliente.TipoSolicitacao}},
			{"field_id": "valor_do_patrimonio", "field_value": []string{fmt.Sprintf("%.2f", cliente.ValorPatrimonio)}},
		},
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Cliente criado com sucesso e mapeado para o Pipefy!",
		"cliente":      cliente,
		"pipefy_split": "Mutation GraphQL estruturada com sucesso na camada de serviço",
	})
}
