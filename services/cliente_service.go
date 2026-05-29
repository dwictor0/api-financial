package services

import (
	"api-financial/models"
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"gorm.io/gorm"
)

type ClienteService struct {
	DB     *gorm.DB
	Logger *slog.Logger
}

func NewClienteService(db *gorm.DB, logger *slog.Logger) *ClienteService {
	return &ClienteService{
		DB:     db,
		Logger: logger,
	}
}

func (s *ClienteService) CriarCliente(clienteInput models.Cliente) (*models.Cliente, error) {
	if clienteInput.ClienteEmail == "" {
		clienteInput.ClienteEmail = fmt.Sprintf("sem-email-%d@sistema.com", time.Now().UnixNano())
		s.Logger.Info("Cliente criado sem e-mail fornecido. Gerado e-mail temporário.",
			slog.String("email_gerado", clienteInput.ClienteEmail))
	}

	var clienteExistente models.Cliente
	err := s.DB.Unscoped().Where("cliente_email = ?", clienteInput.ClienteEmail).First(&clienteExistente).Error

	if err == nil {
		s.Logger.Warn("Tentativa de cadastro com e-mail duplicado",
			slog.String("email", clienteInput.ClienteEmail))
		return nil, fmt.Errorf("Não foi possível processar a requisição com os dados informados")
	}

	clienteInput.Status = "aguardando_analise"
	clienteInput.Prioridade = "nao_definida"

	if err := s.DB.Create(&clienteInput).Error; err != nil {
		s.Logger.Error("Erro ao persistir cliente no banco de dados local",
			slog.Any("error", err),
			slog.String("email", clienteInput.ClienteEmail))
		return nil, err
	}

	go func(c models.Cliente) {
		apiURL := os.Getenv("PIPEFY_API_URL")
		token := os.Getenv("PIPEFY_TOKEN")
		pipeId := os.Getenv("PIPEFY_PIPE_ID")

		if apiURL == "" || token == "" || pipeId == "" {
			s.Logger.Warn("[pipefy] Variáveis de ambiente (.env) não configuradas. Pulando envio do card.")
			return
		}

		parsedURL, err := url.ParseRequestURI(apiURL)
		if err != nil {
			s.Logger.Error("[pipefy] Formato da API_URL configurada no .env é inválido",
				slog.Any("error", err),
				slog.String("url_fornecida", apiURL))
			return
		}

		if parsedURL.Scheme != "https" || parsedURL.Host != "api.pipefy.com" {
			s.Logger.Error("[pipefy] Bloqueio de segurança: Domínio não autorizado para requisições externa (Anti-SSRF)",
				slog.String("host_bloqueado", parsedURL.Host))
			return
		}

		mutation := `
		mutation createCard($pipeId: ID!, $title: String!, $fieldValues: [FieldValueInput]!) {
		  createCard(input: {
		    pipe_id: $pipeId,
		    title: $title,
		    fields_attributes: $fieldValues
		  }) {
		    card { id title }
		  }
		}
		`

		variables := map[string]interface{}{
			"pipeId": pipeId,
			"title":  c.ClienteNome,
			"fieldValues": []map[string]interface{}{
				{
					"field_id":    "email_cliente",
					"field_value": c.ClienteEmail,
				},
			},
		}

		payload := map[string]interface{}{
			"query":     mutation,
			"variables": variables,
		}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			s.Logger.Error("[pipefy] Erro ao serializar JSON da mutation GraphQL",
				slog.Any("error", err))
			return
		}

		// #nosec G704 - URL validada estruturalmente contra SSRF acima
		req, err := http.NewRequest("POST", parsedURL.String(), bytes.NewBuffer(jsonData))
		if err != nil {
			s.Logger.Error("[pipefy] Erro ao construir request HTTP para o Pipefy",
				slog.Any("error", err))
			return
		}

		req.Header.Set("Authorization", token)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}

		// #nosec G704 - Request seguro utilizando apenas o host api.pipefy.com
		resp, err := client.Do(req)
		if err != nil {
			s.Logger.Error("[pipefy] Falha física na conexão de rede com o Pipefy",
				slog.Any("error", err))
			return
		}
		defer resp.Body.Close()

		var corpoResposta map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&corpoResposta); err == nil {
			if errosGraphQL, existe := corpoResposta["errors"]; existe {
				s.Logger.Error("[pipefy] Erro interno de validação retornado pelo schema GraphQL",
					slog.Any("graphql_errors", errosGraphQL))
				return
			}
			s.Logger.Info("[pipefy] Resposta do servidor recebida com sucesso",
				slog.Any("raw_response", corpoResposta))
		}

		if resp.StatusCode == http.StatusOK {
			s.Logger.Info("[pipefy] Card criado com sucesso no Kanban",
				slog.String("cliente_email", c.ClienteEmail))
		} else {
			s.Logger.Warn("[pipefy] Servidor do Pipefy respondeu com código de falha HTTP",
				slog.Int("status_code", resp.StatusCode))
		}
	}(clienteInput)

	return &clienteInput, nil
}
