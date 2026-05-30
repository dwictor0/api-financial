package tests

import (
	"api-financial/models"
	"api-financial/services"
	"io"
	"log/slog"
	"os"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestWebhookDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("falha ao abrir banco em memoria: %v", err)
	}

	err = db.AutoMigrate(&models.Cliente{}, &models.WebhookEvent{})
	if err != nil {
		t.Fatalf("falha ao rodar automigrate de teste: %v", err)
	}

	return db
}

func TestProcessarCardUpdated_PrioridadeAlta(t *testing.T) {
	db := setupTestWebhookDB(t)
	testLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	webhookService := services.NewWebhookService(db, testLogger)
	clienteService := services.NewClienteService(db, testLogger)

	email := "rico@test.com"
	_, _ = clienteService.CriarCliente(models.Cliente{
		ClienteNome:     "Client1",
		ClienteEmail:    email,
		ValorPatrimonio: 300000.00,
	})

	input := models.WebhookCardUpdatedInput{
		EventID:      "evt_alta_123",
		CardID:       "card_999",
		ClienteEmail: email,
		Timestamp:    "2026-05-24T12:00:00Z",
	}

	clienteAtualizado, err := webhookService.ProcessarCardUpdated(input)
	if err != nil {
		t.Fatalf("erro inesperado ao processar webhook: %v", err)
	}

	if clienteAtualizado.Prioridade != "prioridade_alta" {
		t.Errorf("esperava 'prioridade_alta', mas obteve: %s", clienteAtualizado.Prioridade)
	}

	if clienteAtualizado.Status != "processado" {
		t.Errorf("esperava status 'processado', mas obteve: %s", clienteAtualizado.Status)
	}
}

func TestProcessarCardUpdated_PrioridadeNormal(t *testing.T) {
	db := setupTestWebhookDB(t)
	testLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	webhookService := services.NewWebhookService(db, testLogger)
	clienteService := services.NewClienteService(db, testLogger)

	email := "normal@test.com"
	_, _ = clienteService.CriarCliente(models.Cliente{
		ClienteNome:     "Client2",
		ClienteEmail:    email,
		ValorPatrimonio: 150000.00,
	})

	input := models.WebhookCardUpdatedInput{
		EventID:      "evt_normal_123",
		CardID:       "card_555",
		ClienteEmail: email,
		Timestamp:    "2026-05-24T12:00:00Z",
	}

	clienteAtualizado, err := webhookService.ProcessarCardUpdated(input)
	if err != nil {
		t.Fatalf("erro inesperado ao processar webhook: %v", err)
	}

	if clienteAtualizado.Prioridade != "prioridade_normal" {
		t.Errorf("esperava 'prioridade_normal', mas obteve: %s", clienteAtualizado.Prioridade)
	}
}

func TestProcessarCardUpdated_Idempotencia(t *testing.T) {
	db := setupTestWebhookDB(t)
	testLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	webhookService := services.NewWebhookService(db, testLogger)
	clienteService := services.NewClienteService(db, testLogger)

	email := "idempotente@test.com"
	_, _ = clienteService.CriarCliente(models.Cliente{
		ClienteNome:     "cliente2",
		ClienteEmail:    email,
		ValorPatrimonio: 50000.00,
	})

	input := models.WebhookCardUpdatedInput{
		EventID:      "evt_repetido_111",
		CardID:       "card_111",
		ClienteEmail: email,
		Timestamp:    "2026-05-24T12:00:00Z",
	}

	_, errFirst := webhookService.ProcessarCardUpdated(input)
	if errFirst != nil {
		t.Fatalf("primeiro processamento falhou: %v", errFirst)
	}

	_, errSecond := webhookService.ProcessarCardUpdated(input)
	if errSecond == nil {
		t.Error("a idempotencia falhou: o sistema permitiu processar o mesmo event_id duas vezes")
	}
}

func TestProcessarCardUpdated_ClienteNaoEncontrado(t *testing.T) {
	db := setupTestWebhookDB(t)
	testLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	webhookService := services.NewWebhookService(db, testLogger)

	input := models.WebhookCardUpdatedInput{
		EventID:      "evt_inexistente_123",
		CardID:       "card_000",
		ClienteEmail: "nao_existe_no_banco@test.com",
		Timestamp:    "2026-05-24T12:00:00Z",
	}

	_, err := webhookService.ProcessarCardUpdated(input)
	if err == nil {
		t.Error("esperava erro de cliente não encontrado, mas a execução retornou sucesso")
	}
}

func TestProcessarCardUpdated_BloqueioAntiSSRF(t *testing.T) {
	db := setupTestWebhookDB(t)
	testLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	webhookService := services.NewWebhookService(db, testLogger)
	clienteService := services.NewClienteService(db, testLogger)

	os.Setenv("PIPEFY_API_URL", "http://host-perigoso-hack.com")
	defer os.Setenv("PIPEFY_API_URL", "https://api.pipefy.com")

	email := "ssrf@test.com"
	_, _ = clienteService.CriarCliente(models.Cliente{
		ClienteNome:     "Client SSRF",
		ClienteEmail:    email,
		ValorPatrimonio: 250000.00,
	})

	input := models.WebhookCardUpdatedInput{
		EventID:      "evt_ssrf_999",
		CardID:       "card_777",
		ClienteEmail: email,
		Timestamp:    "2026-05-24T12:00:00Z",
	}

	_, err := webhookService.ProcessarCardUpdated(input)
	if err != nil {
		t.Fatalf("erro inesperado: o bloqueio deve logar o ocorrido e retornar o cliente: %v", err)
	}
}

func TestProcessarCardUpdated_UrlCorrompida(t *testing.T) {
	db := setupTestWebhookDB(t)
	testLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	webhookService := services.NewWebhookService(db, testLogger)
	clienteService := services.NewClienteService(db, testLogger)

	os.Setenv("PIPEFY_API_URL", "%%invalid-url-cache-format%%")
	defer os.Setenv("PIPEFY_API_URL", "https://api.pipefy.com")

	email := "corrompido@test.com"
	_, _ = clienteService.CriarCliente(models.Cliente{
		ClienteNome:     "Client Erro URL",
		ClienteEmail:    email,
		ValorPatrimonio: 250000.00,
	})

	input := models.WebhookCardUpdatedInput{
		EventID:      "evt_url_errada_888",
		CardID:       "card_777",
		ClienteEmail: email,
		Timestamp:    "2026-05-24T12:00:00Z",
	}

	_, err := webhookService.ProcessarCardUpdated(input)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
}
