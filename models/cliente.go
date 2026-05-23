package models

import (
	"time"

	"gorm.io/gorm"
)

type Cliente struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	ClienteNome     string         `gorm:"not null" json:"cliente_nome" binding:"required"`
	ClienteEmail    string         `gorm:"uniqueIndex;not null" json:"cliente_email" binding:"required,email"`
	TipoSolicitacao string         `gorm:"not null" json:"tipo_solicitacao" binding:"required"`
	ValorPatrimonio float64        `gorm:"not null" json:"valor_patrimonio" binding:"required"`
	Status          string         `gorm:"default:'Aguardando Análise'" json:"status"`
	Prioridade      string         `gorm:"default:'não definida'" json:"prioridade"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}
