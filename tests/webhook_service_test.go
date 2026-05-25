package tests

import (
	"api-financial/models"
	"api-financial/services"
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
	db := setupTestDB(t)
	webhookService := services.NewWebhookService(db)
	clienteService := services.NewClienteService(db)

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

func TestProcessarCardUpdated_Idempotencia(t *testing.T) {
	db := setupTestDB(t)
	webhookService := services.NewWebhookService(db)
	clienteService := services.NewClienteService(db)

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
