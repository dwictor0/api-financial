export DB_HOST ?= localhost
export DB_PORT ?= 5432
export DB_USER ?= postgres
export DB_PASSWORD ?= root
export DB_NAME ?= api_financial

export PATH := $(PATH):/usr/local/go/bin

.PHONY: help up down test run dev deps docs

help:
	@echo "================================================================="
	@echo "                   Comandos Disponíveis - API Financial          "
	@echo "================================================================="
	@echo "  make dev      - Liga o banco, aguarda ele subir e inicia a API"
	@echo "  make up       - Apenas inicializa o banco de dados (Docker)"
	@echo "  make down     - Para o banco de dados e limpa os containers"
	@echo "  make run      - Compila e executa a API localmente"
	@echo "  make test     - Executa os testes e abre o relatório no navegador"
	@echo "  make docs     - Atualiza a documentação automática do Swagger"
	@echo "  make deps     - Instala as ferramentas Go locais do projeto"
	@echo "================================================================="

up:
	@echo "Inicializando container do banco de dados..."
	@docker compose up -d

down:
	@echo "Parando e limpando ambiente de banco de dados..."
	@docker compose down -v

deps:
	@echo "Instalando ferramentas e dependências do ecossistema Go..."
	@if ! command -v go > /dev/null; then \
		echo "Erro: O compilador do Go não foi encontrado no PATH."; exit 1; \
	fi
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go mod tidy
	@echo "Dependências atualizadas com sucesso!"

docs: deps
	@echo "Atualizando a documentação do Swagger..."
	@swag init -g cmd/api/main.go -d ./ --parseDependency
	@echo "Documentação atualizada com sucesso!"

test:
	@echo "================================================================="
	@echo "  [1/3] Rodando os testes unitários (limpando cache)..."
	@echo "================================================================="
	@if ! command -v go > /dev/null; then \
		echo "Erro: O compilador do Go não foi encontrado no PATH. Verifique a instalação."; exit 1; \
	fi
	@go test ./tests -coverpkg=./... -coverprofile=coverage.out -count=1 || (echo "\nOps! Alguns testes falharam. Corrija os erros antes de gerar o HTML." && exit 1)
	
	@echo ""
	@echo "================================================================="
	@echo "  [2/3] Testes concluídos! RESUMO DE COBERTURA NO TERMINAL:"
	@echo "================================================================="
	@go tool cover -func=coverage.out
	
	@echo ""
	@echo "================================================================="
	@echo "  [3/3] Traduzindo para HTML e abrindo no navegador..."
	@echo "================================================================="
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
	@if ! command -v go > /dev/null; then \
		echo "Erro Crítico: O comando 'go' não está acessível neste ambiente (mesmo tentando /usr/local/go/bin)."; \
		echo "Se usou 'sudo make dev', tente rodar apenas 'make dev' sem o sudo."; \
		exit 1; \
	fi
	@go run cmd/api/main.go

dev: up
	@echo "Aguardando o banco de dados responder na porta $(DB_PORT)..."
	@if command -v nc > /dev/null; then \
		while ! nc -z $(DB_HOST) $(DB_PORT); do sleep 1; done; \
	else \
		sleep 4; \
	fi
	@echo "Banco de dados está online e aceitando conexões!"
	@make run