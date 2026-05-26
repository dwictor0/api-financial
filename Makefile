export DB_HOST ?= localhost
export DB_PORT ?= 3306
export DB_USER ?= root
export DB_PASSWORD ?= root
export DB_NAME ?= api_financial_db

.PHONY: help up down test run dev

help:
	@echo "  make dev      - Liga o banco, aguarda ele subir e inicia a API"
	@echo "  make up       - Apenas inicializa o banco de dados (Docker)"
	@echo "  make down     - Para o banco de dados e limpa os containers"
	@echo "  make run      - Compila e executa a API localmente"
	@echo "  make test     - Executa os testes e abre o relatório no navegador"
	@echo "================================================================="

up:
	@echo "Inicializando container do banco de dados..."
	@docker-compose up -d

down:
	@echo "Parando e limpando ambiente de banco de dados..."
	@docker-compose down -v

test:
	@go test ./tests -coverpkg=./... -coverprofile=coverage.out -count=1 || (echo "\n❌ Ops! Alguns testes falharam. Corrija os erros antes de gerar o HTML." && exit 1)
	
	@echo "Testes concluídos! RESUMO:"
	@go tool cover -func=coverage.out
	
	@go tool cover -html=coverage.out -o coverage.html
	@if command -v xdg-open > /dev/null; then \
		xdg-open coverage.html; \
	elif command -v open > /dev/null; then \
		open coverage.html; \
	else \
		echo "Relatório gerado com sucesso em: $(shell pwd)/coverage.html (Abra manualmente)"; \
	fi
	@echo ""
	@echo "PROCESSO CONCLUÍDO COM SUCESSO!"

run:
	@echo "Iniciando servidor da API Financial..."
	@go run main.go

dev: up
	@echo "Aguardando o banco de dados responder na porta $(DB_PORT)..."
	@if command -v nc > /dev/null; then \
		while ! nc -z $(DB_HOST) $(DB_PORT); do sleep 1; done; \
	elif command -v timeout > /dev/null; then \
		until docker exec $$(docker ps -q -f name=db) mysqladmin ping -u$(DB_USER) -p$(DB_PASSWORD) --silent &> /dev/null; do sleep 1; done; \
	else \