package services

import (
	"api-financial/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
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

func (s *WebhookService) enviarMutacaoPipefy(httpClient *http.Client, pipefyURL, pipefyToken, cardID, status, prioridade string) error {
	query := fmt.Sprintf(`mutation UpdateCardFields($cardId: ID!) {
		status: updateCardField(input: {card_id: $cardId, field_id: "status_do_cliente", new_value: "%s"}) {
			card { title }
		}
		prioridade: updateCardField(input: {card_id: $cardId, field_id: "prioridade", new_value: "%s"}) {
			card { title }
		}
	}`, status, prioridade)

	variables := map[string]interface{}{
		"cardId": cardID,
	}

	bodyRequestBody, err := json.Marshal(graphQLRequest{Query: query, Variables: variables})
	if err != nil {
		return fmt.Errorf("erro ao serializar payload GraphQL: %w", err)
	}

	// #nosec G704 - URL validada de forma estrita contra ataques SSRF antes desta chamada
	req, err := http.NewRequest("POST", pipefyURL, bytes.NewBuffer(bodyRequestBody))
	if err != nil {
		return fmt.Errorf("erro ao construir request HTTP para o Pipefy: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", pipefyToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("falha de rede ao conectar com o Pipefy: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	slog.Info("[PIPEFY-RESPONSE-DEBUG]", "payload", string(bodyBytes))

	return nil
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

	pipefyURL := os.Getenv("PIPEFY_API_URL")
	pipefyToken := os.Getenv("PIPEFY_TOKEN")

	if pipefyToken != "" && pipefyURL != "" {
		parsedURL, err := url.ParseRequestURI(pipefyURL)
		if err != nil {
			s.Logger.Error("[pipefy-webhook] Formato da PIPEFY_API_URL é inválido",
				"error", err,
				"url_fornecida", pipefyURL)
			return &cliente, nil
		}

		if parsedURL.Scheme != "https" || parsedURL.Host != "api.pipefy.com" {
			s.Logger.Error("[pipefy-webhook] Bloqueio Anti-SSRF: Domínio ou protocolo externo não autorizado",
				"host_bloqueado", parsedURL.Host)
			return &cliente, nil
		}

		httpClient := &http.Client{Timeout: 5 * time.Second}

		if err := s.enviarMutacaoPipefy(httpClient, parsedURL.String(), pipefyToken, input.CardID, cliente.Status, cliente.Prioridade); err != nil {
			s.Logger.Error("Erro ao atualizar campos no Pipefy", "error", err)
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
