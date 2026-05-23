package controllers

import (
	"api-financial/models"
	"api-financial/services"
	"bytes"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ResponseStatus uint

const (
	StatusSuccess ResponseStatus = iota
	StatusValidationFailed
	StatusInvalidJSON
	StatusConflictError
)

func (r ResponseStatus) String() string {
	return [...]string{"success", "validation_failed", "invalid_json", "conflict_error"}[r]
}

func (r ResponseStatus) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(r.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

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
		if errs, ok := err.(validator.ValidationErrors); ok {
			camposIncompletos := make(map[string]string)

			for _, e := range errs {
				camposIncompletos[e.Field()] = "Este campo é obrigatório ou está no formato incorreto."
			}

			c.JSON(http.StatusBadRequest, gin.H{
				"status":  StatusValidationFailed,
				"details": camposIncompletos,
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"status":  StatusInvalidJSON,
			"details": "O corpo da requisição não é um JSON válido.",
		})
		return
	}

	clienteCriado, err := cc.Service.CriarCliente(cliente)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"status":  StatusConflictError,
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
		"status":  StatusSuccess,
		"message": "Cliente criado com sucesso!",
		"cliente": clienteCriado,
	})
}
