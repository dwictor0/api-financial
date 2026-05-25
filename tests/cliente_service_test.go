package tests

import (
	"api-financial/models"
	"api-financial/services"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
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

func TestCriarCliente_Sucesso(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewClienteService(db)

	cliente := models.Cliente{
		ClienteNome:     "Joao Teste",
		ClienteEmail:    "lucas@test.com",
		TipoSolicitacao: "abertura_conta",
		ValorPatrimonio: 150000.00,
	}

	resultado, err := service.CriarCliente(cliente)

	if err != nil {
		t.Errorf("nao esperava erro, mas obteve: %v", err)
	}

	if resultado.Status != "aguardando_analise" {
		t.Errorf("esperava status 'aguardando_analise', mas obteve: %s", resultado.Status)
	}

	if resultado.Prioridade != "nao_definida" {
		t.Errorf("esperava prioridade 'nao_definida', mas obteve: %s", resultado.Prioridade)
	}
}

func TestCriarCliente_ErroDuplicado(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewClienteService(db)

	cliente := models.Cliente{
		ClienteNome:     "Joao Duplicado",
		ClienteEmail:    "duplicado@test.com",
		TipoSolicitacao: "abertura_conta",
		ValorPatrimonio: 50000.00,
	}

	_, _ = service.CriarCliente(cliente)

	_, err := service.CriarCliente(cliente)

	if err == nil {
		t.Error("esperava um erro de conflito por email duplicado, mas o erro veio nulo")
	}
}
