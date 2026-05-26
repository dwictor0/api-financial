package config

import (
	"api-financial/models"
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() *gorm.DB {
	err := godotenv.Load()
	if err != nil {
		slog.Warn("Configurações de ambiente (.env) não encontradas, usando fallbacks.")
	}

	host := getEnv("DB_HOST", "localhost")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "root")
	dbName := getEnv("DB_NAME", "api_financial")
	port := getEnv("DB_PORT", "5432")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=America/Sao_Paulo",
		host, user, password, dbName, port)

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		slog.Error("Falha critica ao conectar ao banco de dados", "error", err.Error())
	}

	slog.Info("Banco de dados inicializado com sucesso", "database", dbName, "host", host)

	DB = database

	RunMigrations()
	return DB
}

func RunMigrations() {
	slog.Info("Executando AutoMigrate do GORM...")

	err := DB.AutoMigrate(
		&models.Cliente{},
		&models.WebhookEvent{},
	)
	if err != nil {
		slog.Error("Erro ao executar a migracao do banco", "error", err.Error())

	}

	slog.Info("Migrações de entidades concluídas com sucesso!")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}
