package tests

import (
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"api-financial/models"
	"api-financial/services"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var globalDB *gorm.DB

func SetupTestDB() *gorm.DB {
	if globalDB != nil {
		globalDB.Exec("DELETE FROM clientes")
		globalDB.Exec("DELETE FROM webhook_events")
		return globalDB
	}

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("Falha ao inicializar o banco de dados de testes: " + err.Error())
	}

	err = db.AutoMigrate(&models.Cliente{}, &models.WebhookEvent{})
	if err != nil {
		panic("Falha ao rodar migrações de teste: " + err.Error())
	}

	globalDB = db
	return globalDB
}

func TestMain(m *testing.M) {
	os.Setenv("PIPEFY_TOKEN", "token_fake_teste_ci")
	os.Setenv("PIPEFY_PIPE_ID", "999999")
	os.Setenv("PIPEFY_API_URL", "https://api.pipefy.com")

	code := m.Run()

	os.Unsetenv("PIPEFY_TOKEN")
	os.Unsetenv("PIPEFY_PIPE_ID")
	os.Unsetenv("PIPEFY_API_URL")

	os.Exit(code)
}

func TestCriarCliente_Sucesso(t *testing.T) {
	db := SetupTestDB()
	testLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := services.NewClienteService(db, testLogger)

	novoCliente := models.Cliente{
		ClienteNome:     "Lucas Silva",
		ClienteEmail:    "lucas@test.com",
		ValorPatrimonio: 500000,
		TipoSolicitacao: "Abertura de Conta",
	}

	res, err := service.CriarCliente(novoCliente)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	time.Sleep(150 * time.Millisecond)

	var clienteBanco models.Cliente
	errDB := db.First(&clienteBanco, "cliente_email = ?", "lucas@test.com").Error
	assert.NoError(t, errDB)
	assert.Equal(t, "Lucas Silva", clienteBanco.ClienteNome)
}

func TestCriarCliente_ErroEmailVazio(t *testing.T) {
	db := SetupTestDB()
	testLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := services.NewClienteService(db, testLogger)

	novoCliente := models.Cliente{
		ClienteNome:     "Sem Email",
		ClienteEmail:    "",
		ValorPatrimonio: 100000,
		TipoSolicitacao: "Teste",
	}

	res, err := service.CriarCliente(novoCliente)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestCriarCliente_ErroDuplicado(t *testing.T) {
	db := SetupTestDB()
	testLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := services.NewClienteService(db, testLogger)

	clienteExistente := models.Cliente{
		ClienteNome:     "Cliente Antigo",
		ClienteEmail:    "duplicado@test.com",
		ValorPatrimonio: 100000,
		TipoSolicitacao: "Atualização",
	}
	db.Create(&clienteExistente)

	novoClienteDuplicado := models.Cliente{
		ClienteNome:     "Novo Cliente",
		ClienteEmail:    "duplicado@test.com",
		ValorPatrimonio: 200000,
		TipoSolicitacao: "Atualização",
	}

	_, err := service.CriarCliente(novoClienteDuplicado)
	assert.Error(t, err)
}

func TestCriarCliente_EnvNaoConfigurado(t *testing.T) {
	os.Unsetenv("PIPEFY_TOKEN")
	defer os.Setenv("PIPEFY_TOKEN", "token_fake_teste_ci")

	db := SetupTestDB()
	testLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := services.NewClienteService(db, testLogger)

	novoCliente := models.Cliente{
		ClienteNome:     "Teste Env Faltante",
		ClienteEmail:    "env_em_falta@test.com",
		ValorPatrimonio: 150000,
		TipoSolicitacao: "Inclusão",
	}

	_, err := service.CriarCliente(novoCliente)
	assert.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
}

func TestCriarCliente_ErroUrlInvalida(t *testing.T) {
	os.Setenv("PIPEFY_API_URL", "%%invalid-url-cache%%")
	defer os.Setenv("PIPEFY_API_URL", "https://api.pipefy.com")

	db := SetupTestDB()
	testLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := services.NewClienteService(db, testLogger)

	novoCliente := models.Cliente{
		ClienteNome:     "Teste Url Invalida",
		ClienteEmail:    "url_errada@test.com",
		ValorPatrimonio: 120000,
		TipoSolicitacao: "Inclusão",
	}

	_, err := service.CriarCliente(novoCliente)
	assert.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
}

func TestClienteProcessarCardUpdated_PrioridadeAlta(t *testing.T) {
	db := SetupTestDB()
	testLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	webhookService := services.NewWebhookService(db, testLogger)

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
	assert.Equal(t, "processado", clienteAtualizado.Status)
}

func TestClienteProcessarCardUpdated_PrioridadeNormal(t *testing.T) {
	db := SetupTestDB()
	testLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	webhookService := services.NewWebhookService(db, testLogger)

	clienteMock := models.Cliente{
		ClienteNome:     "Lucas Varejo",
		ClienteEmail:    "normal@test.com",
		ValorPatrimonio: 150000,
	}
	db.Create(&clienteMock)

	payload := models.WebhookCardUpdatedInput{
		EventID:      "evt_normal_999",
		CardID:       "card_555",
		ClienteEmail: "normal@test.com",
	}

	clienteAtualizado, err := webhookService.ProcessarCardUpdated(payload)
	assert.NoError(t, err)
	assert.Equal(t, "prioridade_normal", clienteAtualizado.Prioridade)
}

func TestClienteProcessarCardUpdated_Idempotencia(t *testing.T) {
	db := SetupTestDB()
	testLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	webhookService := services.NewWebhookService(db, testLogger)

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
	assert.Error(t, err)
}

func TestClienteProcessarCardUpdated_ClienteNaoEncontrado(t *testing.T) {
	db := SetupTestDB()
	testLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	webhookService := services.NewWebhookService(db, testLogger)

	payload := models.WebhookCardUpdatedInput{
		EventID:      "evt_fantasma_777",
		CardID:       "card_000",
		ClienteEmail: "nao_existe@test.com",
	}

	_, err := webhookService.ProcessarCardUpdated(payload)
	assert.Error(t, err)
}

func TestClienteStatusString_Cobertura(t *testing.T) {
	statusLista := []models.ClienteStatus{
		models.StatusAguardandoAnalise,
		models.StatusEmAnalise,
		models.StatusAprovado,
		models.StatusRecusado,
		models.StatusProcessado,
	}

	esperado := []string{
		"aguardando_analise",
		"em_analise",
		"aprovado",
		"recusado",
		"processado",
	}

	for i, status := range statusLista {
		resultado := status.String()
		assert.Equal(t, esperado[i], resultado)
	}
}
