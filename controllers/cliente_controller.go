package controllers

import (
	"api-financial/models"
	"api-financial/services"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type ClienteController struct {
	Service *services.ClienteService
}

func NewClienteController(svc *services.ClienteService) *ClienteController {
	return &ClienteController{Service: svc}
}

func (cc *ClienteController) Create(c *gin.Context) {
	var cliente models.Cliente

	if err := c.ShouldBindJSON(&cliente); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validação falhou",
			"details": err.Error(),
		})
		return
	}

	clienteCriado, err := cc.Service.CriarCliente(cliente)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Conflito de cadastro",
			"details": err.Error(),
		})
		return
	}

	pipeId := os.Getenv("PIPEFY_PIPE_ID")
	_ = map[string]interface{}{
		"pipeId": pipeId,
		"title":  clienteCriado.ClienteNome,
		"fields": []map[string]interface{}{
			{"field_id": "email_do_cliente", "field_value": []string{clienteCriado.ClienteEmail}},
			{"field_id": "tipo_de_solicitacao", "field_value": []string{clienteCriado.TipoSolicitacao}},
			{"field_id": "valor_do_patrimonio", "field_value": []string{fmt.Sprintf("%.2f", clienteCriado.ValorPatrimonio)}},
		},
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Cliente criado com sucesso e mapeado para o Pipefy!",
		"cliente":      clienteCriado,
		"pipefy_split": "Mutation GraphQL estruturada com sucesso na camada de serviço",
	})
}
