package controllers

import (
	"api-financial/models"
	"api-financial/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type WebhookController struct {
	Service *services.WebhookService
}

func NewWebhookController(svc *services.WebhookService) *WebhookController {
	return &WebhookController{Service: svc}
}

// HandleCardUpdated godoc
// @Summary      Processar Webhook de Atualização do Pipefy
// @Description  Recebe a notificação do Pipefy, valida a idempotência do evento e atualiza a prioridade/status do cliente.
// @Tags         Webhooks
// @Accept       json
// @Produce      json
// @Param        webhook  body      models.WebhookCardUpdatedInput  true  "Payload do Webhook"
// @Success      200      {object}  map[string]interface{} "Webhook processado com sucesso"
// @Failure      400      {object}  map[string]interface{} "JSON inválido ou campos obrigatórios ausentes"
// @Failure      409      {object}  map[string]interface{} "Conflito: Evento duplicado ou erro de negócio"
// @Router       /webhooks/pipefy/card-updated [post]
func (wc *WebhookController) HandleCardUpdated(c *gin.Context) {
	var input models.WebhookCardUpdatedInput

	if err := c.ShouldBindJSON(&input); err != nil {
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

	clienteAtualizado, err := wc.Service.ProcessarCardUpdated(input)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"status":  StatusConflictError,
			"details": err.Error(),
		})
		return
	}

	_ = `
	mutation updateCardField($cardId: ID!, $fieldId: ID!, $value: [String]!) {
	  updateCardFieldValue(input: {
		card_id: $cardId,
		field_id: $fieldId,
		value: $value
	  }) {
		card { id }
	  }
	}
	`
	_ = map[string]interface{}{
		"cardId": input.CardID,
		"updates": []map[string]interface{}{
			{"field_id": "status_do_cliente", "value": []string{clienteAtualizado.Status}},
			{"field_id": "prioridade", "value": []string{clienteAtualizado.Prioridade}},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  StatusSuccess,
		"message": "Webhook processado com sucesso e banco local atualizado!",
		"data": gin.H{
			"cliente_email": clienteAtualizado.ClienteEmail,
			"status":        clienteAtualizado.Status,
			"prioridade":    clienteAtualizado.Prioridade,
		},
	})
}
