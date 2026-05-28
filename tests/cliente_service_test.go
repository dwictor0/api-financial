package tests

import (
	"os"
	"testing"
	"time"

	"api-financial/models"
	"api-financial/services"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func SetupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("Falha ao inicializar o banco de dados de testes em memória: " + err.Error())
	}

	err = db.AutoMigrate(&models.Cliente{}, &models.WebhookEvent{})
	if err != nil {
		panic("Falha ao rodar migrações de teste: " + err.Error())
	}

	return db
}

func TestMain(m *testing.M) {
	os.Setenv("PIPEFY_TOKEN", "token_fake_teste_ci")
	os.Setenv("PIPEFY_PIPE_ID", "999999")

	code := m.Run()

	os.Unsetenv("PIPEFY_TOKEN")
	os.Unsetenv("PIPEFY_PIPE_ID")

	os.Exit(code)
}

func TestCriarCliente_Sucesso(t *testing.T) {
	os.Setenv("PIPEFY_TOKEN", "token_fake_para_ci_cd")
	os.Setenv("PIPEFY_PIPE_ID", "123456")
	os.Setenv("PIPEFY_API_URL", "https://api.pipefy.com")

	defer os.Unsetenv("PIPEFY_TOKEN")
	defer os.Unsetenv("PIPEFY_PIPE_ID")
	defer os.Unsetenv("PIPEFY_API_URL")

	db := SetupTestDB()
	service := services.NewClienteService(db)

	novoCliente := models.Cliente{
		ClienteNome:     "Lucas Silva",
		ClienteEmail:    "lucas@test.com",
		ValorPatrimonio: 500000,
	}

	_, err := service.CriarCliente(novoCliente)
	assert.NoError(t, err)

	var clienteBanco models.Cliente
	errDB := db.First(&clienteBanco, "cliente_email = ?", "lucas@test.com").Error
	assert.NoError(t, errDB)
	assert.Equal(t, "Lucas Silva", clienteBanco.ClienteNome)

	time.Sleep(100 * time.Millisecond)
}

func TestCriarCliente_ErroDuplicado(t *testing.T) {
	db := SetupTestDB()
	service := services.NewClienteService(db)

	clienteExistente := models.Cliente{
		ClienteNome:     "Cliente Antigo",
		ClienteEmail:    "duplicado@test.com",
		ValorPatrimonio: 100000,
	}
	db.Create(&clienteExistente)

	novoClienteDuplicado := models.Cliente{
		ClienteNome:     "Novo Cliente",
		ClienteEmail:    "duplicado@test.com",
		ValorPatrimonio: 200000,
	}

	_, err := service.CriarCliente(novoClienteDuplicado)

	assert.Error(t, err)
}

func TestClienteProcessarCardUpdated_PrioridadeAlta(t *testing.T) {
	os.Setenv("PIPEFY_TOKEN", "token_fake_para_ci_cd")
	os.Setenv("PIPEFY_PIPE_ID", "123456")
	os.Setenv("API_URL", "https://api.pipefy.com")

	defer os.Unsetenv("PIPEFY_TOKEN")
	defer os.Unsetenv("PIPEFY_PIPE_ID")
	defer os.Unsetenv("API_URL")

	db := SetupTestDB()
	webhookService := services.NewWebhookService(db)

	clienteMock := models.Cliente{
		ClienteNome:     "Roberto VIP",
		ClienteEmail:    "rico@test.com",
		ValorPatrimonio: 1500000,
	}
	db.Create(&clienteMock)

	payload := models.WebhookCardUpdatedInput{
		EventID:      "evt_unique_111",
		CardID:       "card_777",
		ClienteEmail: "rico@test.com",
	}

	_, err := webhookService.ProcessarCardUpdated(payload)
	assert.NoError(t, err)

	var clienteAtualizado models.Cliente
	db.First(&clienteAtualizado, "cliente_email = ?", "rico@test.com")

	assert.Equal(t, "prioridade_alta", clienteAtualizado.Prioridade)

	time.Sleep(50 * time.Millisecond)
}

func TestClienteProcessarCardUpdated_Idempotencia(t *testing.T) {
	db := SetupTestDB()
	webhookService := services.NewWebhookService(db)

	eventoAntigo := models.WebhookEvent{
		EventID:   "evt_repetido_111",
		CardID:    "card_111",
		CreatedAt: time.Now(),
	}
	db.Create(&eventoAntigo)

	payloadDuplicado := models.WebhookCardUpdatedInput{
		EventID:      "evt_repetido_111",
		CardID:       "card_111",
		ClienteEmail: "idempotente@test.com",
	}

	_, err := webhookService.ProcessarCardUpdated(payloadDuplicado)

	assert.Error(t, err, "O sistema deveria bloquear o processamento de event_id duplicado")
}
