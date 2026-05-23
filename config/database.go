package config

import (
	"api-financial/models"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() *gorm.DB {
	err := godotenv.Load()
	if err != nil {
		log.Println("Erro ao inicializar as configuracoes de ambiente.")
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
		log.Fatal("Falha ao conectar ao banco de dado: ", err)
	}

	fmt.Printf("Banco de dados inicializado [%s!\n", dbName)

	DB = database

	RunMigrations()
	return DB
}

func RunMigrations() {
	fmt.Println("Executando AutoMigrate do GORM...")

	err := DB.AutoMigrate(
		&models.Cliente{},
		&models.WebhookEvent{},
	)
	if err != nil {
		log.Fatalf("Erro ao executar a migração do banco: %v", err)
	}

	fmt.Println("Migrações de entidades concluídas com sucesso!")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}
