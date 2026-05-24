package services

import (
	"api-financial/models"
	"errors"
	"time"

	"gorm.io/gorm"
)

type WebhookService struct {
	DB *gorm.DB
}

func NewWebhookService(db *gorm.DB) *WebhookService {
	return &WebhookService{DB: db}
}

func (ws *WebhookService) ProcessarCardUpdated(input models.WebhookCardUpdatedInput) (*models.Cliente, error) {
	parsedTime, err := time.Parse(time.RFC3339, input.Timestamp)
	if err != nil {
		parsedTime = time.Now()
	}

	webhookLog := models.WebhookEvent{
		EventID:      input.EventID,
		CardID:       input.CardID,
		ClienteEmail: input.ClienteEmail,
		Timestamp:    parsedTime,
	}

	tx := ws.DB.Begin()

	if err := tx.Create(&webhookLog).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("evento ja processado ou duplicado")
	}

	var cliente models.Cliente
	if err := tx.First(&cliente, "cliente_email = ?", input.ClienteEmail).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("cliente nao encontrado no sistema")
	}

	if cliente.ValorPatrimonio >= 200000.00 {
		cliente.Prioridade = "prioridade_alta"
	} else {
		cliente.Prioridade = "prioridade_normal"
	}

	cliente.Status = models.StatusProcessado.String()

	if err := tx.Save(&cliente).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return &cliente, nil
}
