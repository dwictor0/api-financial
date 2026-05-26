#!/bin/bash

VERDE='\033[0;32m'
AZUL='\033[0;34m'
AMARELO='\033[1;33m'
SEM_COR='\033[0m'

echo -e "${AZUL} INICIANDO TESTES E COBERTURA GO ${SEM_COR}"

echo -e "${AMARELO}[1/3] Rodando os testes unitários (limpando cache)...${SEM_COR}"
go test ./tests -coverpkg=./... -coverprofile=coverage.out -count=1

if [ $? -ne 0 ]; then
    echo ""
    echo "Ops! Alguns testes falharam. Corrija os erros antes de gerar o HTML."
    exit 1
fi

echo ""
echo -e "${VERDE} Testes executados com sucesso! RESUMO DE COBERTURA:${SEM_COR}"
go tool cover -func=coverage.out

echo ""
echo -e "${AMARELO}[2/3] Traduzindo relatório para HTML...${SEM_COR}"
go tool cover -html=coverage.out -o coverage.html

echo -e "${AMARELO}[3/3] Abrindo relatório no navegador...${SEM_COR}"
if command -v xdg-open &> /dev/null; then
    xdg-open coverage.html 
elif command -v open &> /dev/null; then
    open coverage.html     
else
    echo -e "${VERDE}Relatório gerado em: $(pwd)/coverage.html (Abra manualmente no navegador)${SEM_COR}"
fi

echo -e "${VERDE} PROCESSO CONCLUÍDO COM SUCESSO! ${SEM_COR}"