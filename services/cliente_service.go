package services

import (
	"api-financial/models"
	"fmt"

	"gorm.io/gorm"
)

type ClienteService struct {
	DB *gorm.DB
}

func NewClienteService(db *gorm.DB) *ClienteService {
	return &ClienteService{DB: db}
}

func (s *ClienteService) CriarCliente(clienteInput models.Cliente) (*models.Cliente, error) {
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

	return &clienteInput, nil
}
