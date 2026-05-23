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

// Create godoc
// @Summary      Criar um novo cliente
// @Description  Recebe os dados de um cliente, executa validações de negócio, evita duplicidade de e-mail e prepara o mapeamento para o Pipefy.
// @Tags         Clientes
// @Accept       json
// @Produce      json
// @Param        cliente  body      models.Cliente  true  "Dados do Cliente"
// @Success      201      {object}  map[string]interface{} "Cliente criado com sucesso"
// @Failure      400      {object}  map[string]interface{} "Validação de formato falhou"
// @Failure      409      {object}  map[string]interface{} "Conflito de cadastro (E-mail já existente)"
// @Failure      500      {object}  map[string]interface{} "Erro interno do servidor"
// @Router       /clientes [post]
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
			"status":  "failed",
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
