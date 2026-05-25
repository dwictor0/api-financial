package models

import (
	"time"

	"gorm.io/gorm"
)

type ClienteStatus uint

const (
	StatusAguardandoAnalise ClienteStatus = iota
	StatusEmAnalise
	StatusAprovado
	StatusRecusado
	StatusProcessado
)

func (s ClienteStatus) String() string {
	return [...]string{"aguardando_analise", "em_analise", "aprovado", "recusado", "processado"}[s]
}

type Cliente struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	ClienteNome     string         `gorm:"not null" json:"cliente_nome" binding:"required"`
	ClienteEmail    string         `gorm:"uniqueIndex;not null" json:"cliente_email" binding:"required,email"`
	TipoSolicitacao string         `gorm:"not null" json:"tipo_solicitacao" binding:"required"`
	ValorPatrimonio float64        `gorm:"not null" json:"valor_patrimonio" binding:"required"`
	Status          string         `gorm:"default:'aguardando_analise'" json:"status"`
	Prioridade      string         `gorm:"default:'nao_definida'" json:"prioridade"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

type CriarClienteInput struct {
	ClienteNome     string  `json:"cliente_nome" binding:"required"`
	ClienteEmail    string  `json:"cliente_email" binding:"required,email"`
	TipoSolicitacao string  `json:"tipo_solicitacao" binding:"required"`
	ValorPatrimonio float64 `json:"valor_patrimonio" binding:"required"`
}
