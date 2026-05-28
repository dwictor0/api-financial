package controllers

import (
	"api-financial/models"
	"api-financial/services"
	"fmt"
	"log/slog"
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
// @Description  Recebe a notificação do Pipefy, valida a idempotência do evento e updates a prioridade/status do cliente.
// @Tags         Webhooks
// @Accept       json
// @Produce      json
// @Param        webhook  body      models.WebhookCardUpdatedInput  true  "Payload do Webhook"
// @Success      200      {object}  map[string]interface{} "Webhook processado com sucesso"
// @Failure      400      {object}  map[string]interface{} "JSON inválido ou campos obrigatórios ausentes"
// @Failure      409      {object}  map[string]interface{} "Conflito: Evento duplicado ou erro de negócio"
// @Router       /api/webhooks/pipefy/card-updated [post]
func (wc *WebhookController) HandleCardUpdated(c *gin.Context) {
	var input models.WebhookCardUpdatedInput

	if err := c.ShouldBindJSON(&input); err != nil {
		fmt.Println("ERRO DE VALIDAÇÃO DO GIN:", err.Error())
		if errs, ok := err.(validator.ValidationErrors); ok {
			camposIncompletos := make(map[string]string)
			for _, e := range errs {
				camposIncompletos[e.Field()] = "Este campo é obrigatório ou está no formato incorreto."
			}
			slog.Warn("AUDIT: Campos obrigatorios ausentes no webhook",
				"acao", "PIPEFY_WEBHOOK_RECEIVE",
				"status", "erro_validacao",
				"detalhes", camposIncompletos,
			)
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  StatusValidationFailed,
				"details": camposIncompletos,
			})
			return
		}

		slog.Warn("AUDIT: Payload do webhook nao e um JSON valido",
			"acao", "PIPEFY_WEBHOOK_RECEIVE",
			"status", "erro_json_invalido",
			"error", err.Error(),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  StatusInvalidJSON,
			"details": "O corpo da requisição não é um JSON válido.",
		})
		return
	}

	clienteAtualizado, err := wc.Service.ProcessarCardUpdated(input)
	if err != nil {
		slog.Error("AUDIT: Erro webhook ja esta cadastrado.",
			"acao", "PIPEFY_WEBHOOK_PROCESS",
			"card_id", input.CardID,
			"status", "erro_conflito",
			"error", err.Error(),
		)
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
	slog.Info("AUDIT: Webhook do Pipefy processado com sucesso",
		"acao", "PIPEFY_WEBHOOK_SUCCESS",
		"card_id", input.CardID,
		"cliente_email", clienteAtualizado.ClienteEmail,
		"novo_status", clienteAtualizado.Status,
		"nova_prioridade", clienteAtualizado.Prioridade,
		"status", "sucesso",
	)
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
