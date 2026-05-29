# **API Financial - Client Management & Pipefy Integration**

> Um sistema interno para gerenciamento de clientes, cálculo de patrimônio investido e mapeamento de fluxos operacionais integrado com a API GraphQL do Pipefy e Webhooks.

--- 

## **Sumário**

- [Tecnologias Utilizadas](#tecnologias-utilizadas)
- [Instalação e Execução Local](#instalação-e-execução-local)
- [Execução dos Testes](#execução-dos-testes)
- [Exemplos de Requisição (cURL)](#exemplos-de-requisição-curl)
- [Guia Passo a Passo: Integração Completa com o Painel do Pipefy](#guia-passo-a-passo-integração-completa-com-o-painel-do-pipefy)
- [Visão de Produção na AWS](#visão-de-produção-na-aws)
- [Decisões Arquiteturais](#decisões-arquiteturais)
- [Fluxo de Processamento de Webhook](#fluxo-de-processamento-de-webhook)
- [Arquitetura da Aplicação](#arquitetura-da-aplicação)
- [Troubleshooting](#troubleshooting)

--- 

## **Tecnologias Utilizadas**

- [Go (Golang 1.26.3)](https://go.dev/)
- [PostgreSQL](https://www.postgresql.org/)
- [Docker & Docker Compose](https://www.docker.com/)
- [Swagger (swaggo)](https://github.com/swaggo/swag)
- [Makefile](https://www.gnu.org/software/make/manual/make.html)

---



## **Instalação e Execução Local**

Siga os passos abaixo para rodar o projeto localmente:

### 1. Clone o repositório

```bash
git clone https://github.com/seu-usuario/api-financial.git
cd api-financial
```

### 2. Configure as variáveis de ambiente

Copie o arquivo de exemplo e ajuste conforme necessário:

```bash
cp .env.example .env
```
> ⚠️ **Atenção:** 

### 3. Suba o ambiente completo (Banco + API)

Utilize o Makefile para inicializar o PostgreSQL, aguardar a prontidão da porta 5432 e compilar/iniciar a aplicação automaticamente:

```bash
make dev
```

### 4. Execução manual (alternativo)

Caso prefira gerenciar os containers manualmente:

```bash
docker compose up -d
```

### 5. Limpeza de ambiente (se necessário)

Em caso de inconsistências ou necessidade de limpar volumes:

```bash
make down
```

---

##  **Execução dos Testes**

Para executar a suíte de testes unitários, gerar o relatório de cobertura e abrir o resumo em HTML no navegador:

```bash
make test
```

### Métricas de Cobertura Atuais

| Camada | Cobertura |
|---|---|
| **Total Global** | 66% |
| **Model Cliente** | 100% |
| **Webhook Service** | 81.0% |
| **Client Service** | 59% |

> **Nota técnica:** A cobertura do Client Service foi mantida neste patamar intencionalmente para isolar a camada de regras de negócio sem introduzir mocks de chamadas de rede adicionais nesta iteração.

---

## **Exemplos de Requisição (cURL)**

### Endpoint 1 — Criar Cliente

```bash
curl -X POST http://localhost:8080/api/clientes \
  -H "Content-Type: application/json" \
  -d '{
    "cliente_nome": "João Silva",
    "cliente_email": "joao.silva@example.com",
    "tipo_solicitacao": "Atualização cadastral",
    "valor_patrimonio": 200000.00
  }'
```

**Resposta esperada (`201 Created`):**

```json
{
  "cliente": {
    "id": 1,
    "cliente_nome": "João Silva",
    "cliente_email": "joao.silva@example.com",
    "tipo_solicitacao": "Atualização cadastral",
    "valor_patrimonio": 200000.00,
    "status": "aguardando_analise",
    "prioridade": "nao_definida",
    "created_at": "2026-05-28T19:26:39.519513743-03:00",
    "updated_at": "2026-05-28T19:26:39.519513743-03:00"
  },
  "message": "Cliente criado com sucesso!",
  "status": "success"
}
```

---

### Endpoint 2 — Receber Webhook do Pipefy

```bash
curl -X POST http://localhost:8080/api/webhooks/pipefy/card-updated \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "evt_123",
    "card_id": "card_456",
    "cliente_email": "joao.silva@example.com",
    "timestamp": "2026-05-18T12:00:00Z"
  }'
```

**Resposta esperada (`200 OK`):**

```json
{
  "data": {
    "cliente_email": "joao.silva@example.com",
    "prioridade": "prioridade_alta",
    "status": "processado"
  },
  "message": "Webhook processado com sucesso e banco local atualizado!",
  "status": "success"
}
```

---

## **Guia Passo a Passo: Integração Completa com o Painel do Pipefy**

Para que a API em Go se comunique com o Pipefy com sucesso (tanto para criar cards quanto para receber os alertas de Webhook), siga atentamente o tutorial abaixo para configurar o seu painel.

### 1. Identificando o seu `Pipe ID`
O `Pipe ID` é o identificador único do seu fluxo de trabalho dentro do Pipefy, necessário para que a API saiba em qual quadro deve inserir os novos clientes.
1. Acesse o painel do **Pipefy** e entre no Pipe (quadro) que deseja utilizar.
2. Olhe para a barra de endereços (URL) do seu navegador.
3. O número que aparece logo após `/pipes/` é o seu ID.
   * *Exemplo:* Se a URL for `https://app.pipefy.com/pipes/301234567`, o seu **Pipe ID** é `301234567`.

### 2. Gerando o seu Personal Access Token (Bearer Token)
A API precisa de uma chave de autorização para ter permissão de enviar dados via GraphQL.
1. No canto inferior esquerdo do Pipefy, clique na sua **foto de perfil** e vá em **Account Preferences** (Configurações da Conta).
2. No menu esquerdo, clique em **Personal Access Tokens**.
3. Clique em **Generate new token**, dê um nome para identificar (ex: `API-Financial-Go`) e clique em gerar.
4. **Copie o token gerado imediatamente** e guarde-o com segurança. Ele deve ser adicionado no arquivo `.env` da sua aplicação na variável correspondente.

### 3. Criando a Automação e Configurando o Webhook (Gatilho HTTP)
Para fazer o Pipefy avisar a sua API em Go quando uma alteração acontecer, configure uma automação de disparo de requisição HTTP:
1. Dentro do seu Pipe, clique no botão **Automations** (Automatizar — ícone de raio no canto superior direito).
2. Clique em **Create a new automation** (Criar nova automação).
3. **Configure o Gatilho (Trigger):** * Escolha o evento: **"When a card field is updated"** (Quando um campo do card for atualizado).
   * Selecione o campo específico: Escolha o campo referente ao **nome da conta** (ou qualquer outro de sua preferência caso queira customizar o gatilho).
4. **Configure a Ação (Action):**
   * Escolha a opção **"Send a Webhook"** (Disparar um Webhook / Requisição HTTP).
   * **URL de Destino:** Insira o endpoint público da sua aplicação. Se estiver testando localmente, utilize uma ferramenta de tunelamento. 
   * **Método:** `POST`
   * **Payload da requisição:** Monte a estrutura no formato JSON esperado pela API para que os dados cheguem corretamente.

### 4. Configuração Obrigatória do Formulário de Início
Para que a integração via GraphQL consiga preencher e persistir os dados sem erros de validação de schema, as propriedades de identificação do cliente precisam estar exatamente iguais.
1. Dentro do seu Pipe, clique em **Form Settings** (Configurações do Formulário / Formulário de Início).
2. Clique em **Create new field** (Criar campo) e selecione o tipo **Short Text** (Texto Curto) ou **Email**.
3. No nome do campo (Label), digite **exatamente**: `email_cliente`
4. No campo identificador interno do Pipefy (ID do campo/API Key do campo), certifique-se de que ele foi gerado exatamente como `email_cliente`.
> ⚠️ **Atenção:** A nomenclatura `email_cliente` deve ser seguida de forma estrita e sem variações de maiúsculas ou espaços. Caso o ID do campo seja diferente, a camada de serviço da API em Go falhará ao tentar localizar a sintaxe do tipo da *mutation* na documentação oficial, impedindo a sincronização dos dados no banco PostgreSQL local.

### 5. Exemplo de Request Body
<img width="1400" height="760" alt="image" src="https://github.com/user-attachments/assets/dd7d1d6b-f73a-4545-9a9a-85408b559de1" />

>  ⚠️ **Atenção:** No exemplo visual e nos testes de integração locais, valores idênticos podem ser utilizados temporariamente para os campos event_id e card_id. Contudo, em um ambiente de produção real, é fundamental mapear as variáveis nativas do Pipefy de forma dinâmica: o event_id deve representar o identificador único do evento gerado (garantindo a idempotência e evitando processamento duplicado), enquanto o card_id deve apontar estritamente para o ID do card modificado dentro do fluxo.




---

## **Visão de Produção na AWS**

### Como essa arquitetura escalaria na AWS

A estrutura atual (API Go + PostgreSQL + Webhooks) mapeia naturalmente para os seguintes serviços gerenciados da AWS:

```
Pipefy Webhook
      │
      ▼
 API Gateway  ──────────────────────────────────────────┐
      │                                                  │
      ▼                                                  ▼
 Lambda (Go)                                    Lambda (Go)
 /clients                                    /webhooks/pipefy
      │                                                  │
      ▼                                                  ▼
  RDS (PostgreSQL)                            SQS Queue (buffer)
  Multi-AZ                                              │
                                                        ▼
                                               Lambda Worker
                                            (processa eventos
                                             de forma assíncrona)
                                                        │
                                                        ▼
                                             DynamoDB (log de
                                            eventos/idempotência)
```

### Justificativa por serviço

**API Gateway**
Atua como ponto de entrada único, com throttling nativo por rota e autenticação via API Key ou Cognito. Absorve picos de tráfego sem configuração manual de auto-scaling.

**AWS Lambda (Go)**
O binário Go compila para um executável estático extremamente leve, com cold starts abaixo de 100ms. Cada endpoint (`/clients`, `/webhooks/pipefy`) é uma função separada, permitindo escalar e versionar de forma independente.

**RDS PostgreSQL (Multi-AZ)**
Mantém o banco relacional já utilizado localmente, com failover automático, backups gerenciados e réplicas de leitura para consultas analíticas de patrimônio investido.

**SQS (Simple Queue Service)**
O webhook do Pipefy é enfileirado imediatamente, retornando `200 OK` ao Pipefy sem latência. Um Lambda Worker consome a fila de forma assíncrona, garantindo que falhas temporâneas não percam eventos (dead-letter queue configurada).

**DynamoDB**
Armazena o log de eventos processados com uma chave de idempotência (`event_id + action + timestamp`), evitando processamento duplicado de webhooks em caso de reenvio pelo Pipefy.

### Escalabilidade e resiliência

- **Concorrência de webhooks:** SQS + Lambda escalam automaticamente com o volume de eventos sem intervenção manual.
- **Picos de criação de clientes:** API Gateway + Lambda lidam com burst de até 3.000 req/s por padrão, configurável.
- **Zero downtime:** Deploys via Lambda aliases com traffic shifting gradual (ex: 10% → 50% → 100%).
- **Custo:** Modelo pay-per-use — sem custo em períodos ociosos, ideal para sistemas internos com tráfego irregular.

---

## **Decisões Arquiteturais**

### Go como linguagem principal
Go oferece performance próxima ao C com a simplicidade de uma linguagem moderna. Para uma API financeira com requisitos de latência baixa e processamento de webhooks assíncronos, a combinação de goroutines, tipagem estática e compilação para binário único é ideal.

### Docker para isolamento do ambiente
O ambiente de desenvolvimento é idêntico ao de produção via Docker Compose, eliminando o clássico problema de "funciona na minha máquina" e facilitando o onboarding de novos membros ao time.

### Makefile como automação central
Concentra todos os comandos do projeto (`make dev`, `make test`, `make down`) em uma interface única, reduzindo a curva de aprendizado e padronizando o fluxo de trabalho entre desenvolvedores.

### Swagger para documentação da API
A documentação é gerada automaticamente a partir das anotações no código via `swaggo`, mantendo-a sempre sincronizada com a implementação real.

---

## **Fluxo de Processamento de Webhook**

```
Pipefy emite evento
        │
        ▼
POST /webhooks/pipefy
        │
        ▼
Validação de assinatura (HMAC)
        │
        ▼
Verificação de duplicidade (idempotência por event_id)
        │
        ├── Duplicado → 409 Conflict (ignorado)
        │
        └── Novo evento
                │
                ▼
        Processamento assíncrono
                │
                ▼
        Atualização do estado do cliente no PostgreSQL
                │
                ▼
        200 OK
```

---

## **Arquitetura da Aplicação**

O sistema segue separação clara de responsabilidades em camadas:


- **Cmd:** Ponto de entrada (Bootstrap) responsável pela inicialização do servidor HTTP e injeção de dependências.
- **Config:** Gerenciamento centralizado de variáveis de ambiente e pooling de conexões com o PostgreSQL.
- **Controller:** Recebe a requisição HTTP, valida o payload e delega ao Service.
- **Service:** Contém toda a lógica de negócio (cálculo de patrimônio, regras de cliente).
- **Model:** Estruturas de dados relacionais mapeadas para as tabelas do PostgreSQL.
- **Routes:** Camada responsável pelo roteamento da API.
- **Docs:** Especificações OpenAPI geradas pelo Swagger.

---

## **Troubleshooting**

**Erro de conexão com o banco de dados:**
Verifique se o PostgreSQL está saudável no Docker:
```bash
docker compose logs db
```

**Porta 5432 ocupada:**
```bash
make down && make dev
```

**Containers em estado inconsistente:**
```bash
make down
docker system prune -f
make dev
```
