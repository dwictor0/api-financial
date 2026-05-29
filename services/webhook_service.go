package services

import (
	"api-financial/models"
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"gorm.io/gorm"
)

type WebhookService struct {
	DB     *gorm.DB
	Logger *slog.Logger
}

type graphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

func NewWebhookService(db *gorm.DB, logger *slog.Logger) *WebhookService {
	return &WebhookService{
		DB:     db,
		Logger: logger,
	}
}

func (s *WebhookService) ProcessarCardUpdated(input models.WebhookCardUpdatedInput) (*models.Cliente, error) {
	var eventoExistente models.WebhookEvent
	if err := s.DB.Where("event_id = ?", input.EventID).First(&eventoExistente).Error; err == nil {
		s.Logger.Error("AUDIT: Erro webhook ja esta cadastrado.",
			"acao", "PIPEFY_WEBHOOK_PROCESS",
			"card_id", input.CardID,
			"event_id", input.EventID,
			"status", "erro_conflito",
			"error", "evento duplicado",
		)
		return nil, fmt.Errorf("este evento já foi processado anteriormente")
	}

	var cliente models.Cliente
	if err := s.DB.Where("cliente_email = ?", input.ClienteEmail).First(&cliente).Error; err != nil {
		return nil, fmt.Errorf("cliente não encontrado para o e-mail fornecido")
	}

	if cliente.ValorPatrimonio >= 200000.00 {
		cliente.Prioridade = "prioridade_alta"
	} else {
		cliente.Prioridade = "prioridade_normal"
	}
	cliente.Status = "processado"

	if err := s.DB.Save(&cliente).Error; err != nil {
		return nil, err
	}

	novoEvento := models.WebhookEvent{
		EventID:   input.EventID,
		CardID:    input.CardID,
		CreatedAt: time.Now(),
	}
	if err := s.DB.Create(&novoEvento).Error; err != nil {
		s.Logger.Error("Erro ao salvar log de idempotencia do webhook", "error", err)
		return nil, err
	}

	mutation := `
    mutation UpdateFinancialCardFields($cardId: ID!, $containerFields: [UpdateCardFieldInput!]!) {
      updateCardFields(input: {
        card_id: $cardId,
        fields: $containerFields
      }) {
        card { id }
      }
    }`

	variables := map[string]interface{}{
		"cardId": input.CardID,
		"containerFields": []map[string]interface{}{
			{
				"field_id": "status_do_cliente",
				"value":    []string{cliente.Status},
			},
			{
				"field_id": "prioridade",
				"value":    []string{cliente.Prioridade},
			},
		},
	}

	bodyRequestBody, _ := json.Marshal(graphQLRequest{Query: mutation, Variables: variables})
	pipefyURL := os.Getenv("PIPEFY_API_URL")
	pipefyToken := os.Getenv("PIPEFY_TOKEN")

	if pipefyToken != "" && pipefyURL != "" {
		req, _ := http.NewRequest("POST", pipefyURL, bytes.NewBuffer(bodyRequestBody))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", pipefyToken))
		req.Header.Set("Content-Type", "application/json")
		httpClient := &http.Client{Timeout: 5 * time.Second}
		resp, err := httpClient.Do(req)
		if err == nil {
			defer resp.Body.Close()
		}
	}

	s.Logger.Info("AUDIT: Webhook do Pipefy processado com sucesso",
		"acao", "PIPEFY_WEBHOOK_SUCCESS",
		"card_id", input.CardID,
		"cliente_email", cliente.ClienteEmail,
		"novo_status", cliente.Status,
		"nova_prioridade", cliente.Prioridade,
		"status", "sucesso",
	)

	return &cliente, nil
}
