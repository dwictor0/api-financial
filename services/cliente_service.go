package services

import (
	"api-financial/models"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"gorm.io/gorm"
)

type ClienteService struct {
	DB *gorm.DB
}

func NewClienteService(db *gorm.DB) *ClienteService {
	return &ClienteService{DB: db}
}

func (s *ClienteService) CriarCliente(clienteInput models.Cliente) (*models.Cliente, error) {
	if clienteInput.ClienteEmail == "" {
		clienteInput.ClienteEmail = fmt.Sprintf("sem-email-%d@sistema.com", time.Now().UnixNano())
	}

	var clienteExistente models.Cliente
	err := s.DB.Unscoped().Where("cliente_email = ?", clienteInput.ClienteEmail).First(&clienteExistente).Error

	if err == nil {
		return nil, fmt.Errorf("Não foi possível processar a requisição com os dados informados")
	}

	clienteInput.Status = "aguardando_analise"
	clienteInput.Prioridade = "nao_definida"

	if err := s.DB.Create(&clienteInput).Error; err != nil {
		return nil, err
	}

	go func(c models.Cliente) {
		apiURL := os.Getenv("PIPEFY_API_URL")
		token := os.Getenv("PIPEFY_TOKEN")
		pipeId := os.Getenv("PIPEFY_PIPE_ID")

		if apiURL == "" || token == "" || pipeId == "" {
			fmt.Println("[Pipefy] Variáveis de ambiente (.env) não configuradas. Pulando envio.")
			return
		}

		parsedURL, err := url.ParseRequestURI(apiURL)
		if err != nil {
			fmt.Println("[Pipefy] Formato da API_URL configurada no .env é inválido:", err)
			return
		}

		if parsedURL.Scheme != "https" || parsedURL.Host != "api.pipefy.com" {
			fmt.Printf("[Pipefy] Bloqueio de segurança: Domínio '%s' não autorizado para requisições.\n", parsedURL.Host)
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
			fmt.Println("[Pipefy] Erro ao serializar JSON da mutation:", err)
			return
		}

		req, err := http.NewRequest("POST", parsedURL.String(), bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println("[Pipefy] Erro ao construir request HTTP:", err)
			return
		}

		req.Header.Set("Authorization", token)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("[Pipefy] Falha física na conexão de rede:", err)
			return
		}
		defer resp.Body.Close()

		var corpoResposta map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&corpoResposta); err == nil {
			if errosGraphQL, existe := corpoResposta["errors"]; existe {
				fmt.Printf("[Pipefy] Erro interno de validação do GraphQL: %v\n", errosGraphQL)
				return
			}
			fmt.Printf("[Pipefy] Resposta bruta recebida: %v\n", corpoResposta)
		}

		if resp.StatusCode == http.StatusOK {
			fmt.Printf("[Pipefy] Card criado com sucesso no Kanban para o e-mail: %s\n", c.ClienteEmail)
		} else {
			fmt.Printf("[Pipefy] Servidor respondeu com código de erro HTTP: %d\n", resp.StatusCode)
		}
	}(clienteInput)

	return &clienteInput, nil
}
