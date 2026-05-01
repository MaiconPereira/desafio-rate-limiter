# Desafio 4 RATE LIMIT

## Objetivo

Desenvolver um Rate Limiter em Go que funcione como middleware para controlar o fluxo de requisições de um serviço web.

O sistema limita tráfego por IP ou por Token de acesso usando Redis para persistência e orquestração.

## Regras de negócio

- Limitação por IP: restringe o número máximo de requisições por segundo de um mesmo endereço IP.
- Limitação por Token: restringe o número máximo de requisições por segundo de um token enviado no header `API_KEY: <TOKEN>`.
- Precedência: quando o header `API_KEY` estiver presente, a regra do Token prevalece sobre a regra de IP.

Quando o limite é excedido, a API responde com HTTP `429` e o corpo:

```text
you have reached the maximum number of requests or actions allowed within a certain time frame
```

O IP ou Token fica bloqueado pelo tempo configurado em `RATE_LIMIT_BLOCK_DURATION`.

## Configuração

Todas as configurações podem ser passadas por variáveis de ambiente ou pelo `docker-compose.yaml`.

```env
APP_PORT=8080
REDIS_ADDR=redis:6379
REDIS_PASSWORD=
REDIS_DB=0
RATE_LIMIT_IP_RPS=10
RATE_LIMIT_TOKEN_RPS=100
RATE_LIMIT_TOKEN_OVERRIDES=basic-token=5,premium-token=100
RATE_LIMIT_BLOCK_DURATION=5m
```

`RATE_LIMIT_TOKEN_RPS` é o limite padrão para qualquer token. `RATE_LIMIT_TOKEN_OVERRIDES` permite limites específicos por token no formato `token=limite,token2=limite`.

## Arquitetura

A regra de negócio fica em `internal/limiter`.

A persistência usa o padrão Strategy através da interface:

```go
type Store interface {
    Increment(ctx context.Context, key string, expiration time.Duration) (int64, error)
    Block(ctx context.Context, key string, expiration time.Duration) error
    IsBlocked(ctx context.Context, key string) (bool, error)
}
```

A implementação obrigatória com Redis está em `internal/infra/redisstore`. Para trocar a persistência no futuro, crie uma nova implementação dessa interface e injete no `limiter.New(...)`.

O middleware HTTP está separado em `internal/middleware`.

## Como executar

Suba a aplicação e o Redis:

```bash
docker compose up --build
```

A API estará em:

```text
http://localhost:8080
```

Exemplos:

```bash
curl http://localhost:8080/
curl -H 'API_KEY: basic-token' http://localhost:8080/
curl -H 'API_KEY: premium-token' http://localhost:8080/
```

## Como executar os testes

Pelo Docker Compose:

```bash
docker compose --profile test run --rm tests
```

Ou localmente:

```bash
go test ./...
```

Os testes demonstram bloqueio por IP, bloqueio por Token, precedência Token > IP e resposta HTTP `429` com a mensagem exigida.

## Testes manuais

Para testar manualmente, suba a aplicação:

```bash
docker compose up --build
```

A aplicação ficará disponível em `http://localhost:8080`.

### 1. Teste de limite por IP

No `docker-compose.yaml`, o limite padrão por IP está configurado assim:

```yaml
RATE_LIMIT_IP_RPS: 10
```

Isso significa que o mesmo IP pode fazer até 10 requisições por segundo sem token.

Execute várias chamadas rapidamente:

```bash
for i in $(seq 1 12); do curl -i http://localhost:8080/; echo; done
```

As primeiras requisições devem retornar:

```http
HTTP/1.1 200 OK
```

Depois que o limite for excedido, a API deve retornar:

```http
HTTP/1.1 429 Too Many Requests
```

Com exatamente este corpo:

```text
you have reached the maximum number of requests or actions allowed within a certain time frame
```

### 2. Teste de limite por Token

O token deve ser enviado no header `API_KEY`.

No `docker-compose.yaml`, o token `basic-token` está configurado com limite específico:

```yaml
RATE_LIMIT_TOKEN_OVERRIDES: "basic-token=5,premium-token=100"
```

Teste o token `basic-token`:

```bash
for i in $(seq 1 7); do curl -i -H 'API_KEY: basic-token' http://localhost:8080/; echo; done
```

As primeiras 5 requisições dentro do mesmo segundo devem passar com `200 OK`. As próximas devem retornar `429 Too Many Requests`.

### 3. Teste de precedência Token > IP

A regra principal do desafio é: quando o header `API_KEY` existe, a configuração do Token prevalece sobre a configuração de IP.

No projeto, o IP tem limite padrão de 10 requisições por segundo:

```yaml
RATE_LIMIT_IP_RPS: 10
```

Mas o token `premium-token` tem limite de 100 requisições por segundo:

```yaml
RATE_LIMIT_TOKEN_OVERRIDES: "basic-token=5,premium-token=100"
```

Execute:

```bash
for i in $(seq 1 20); do curl -i -H 'API_KEY: premium-token' http://localhost:8080/; echo; done
```

Mesmo passando de 10 requisições, as chamadas com `premium-token` devem continuar retornando `200 OK`, porque o limite do token é maior que o limite do IP.

### 4. Teste do tempo de bloqueio

O tempo de bloqueio está configurado por:

```yaml
RATE_LIMIT_BLOCK_DURATION: 5m
```

Depois que um IP ou Token excede o limite, novas requisições para a mesma chave devem continuar retornando `429` durante esse período.

Para testar mais rápido, altere temporariamente no `docker-compose.yaml`:

```yaml
RATE_LIMIT_BLOCK_DURATION: 10s
```

Recrie a aplicação:

```bash
docker compose down
docker compose up --build
```

Estoure o limite:

```bash
for i in $(seq 1 12); do curl -i http://localhost:8080/; echo; done
```

Faça uma nova chamada imediatamente:

```bash
curl -i http://localhost:8080/
```

Ela deve retornar `429`.

Aguarde 10 segundos e tente novamente:

```bash
sleep 10
curl -i http://localhost:8080/
```

A resposta deve voltar para `200 OK`.

### 5. Teste usando o arquivo api.http

O arquivo `api.http` contém exemplos prontos:

```http
GET http://localhost:8080/

###

GET http://localhost:8080/
API_KEY: basic-token

###

GET http://localhost:8080/
API_KEY: premium-token
```

Use a extensão REST Client do VS Code ou a ferramenta HTTP do seu editor para executar as chamadas.

### 6. Limpar Redis entre testes

Como o Redis guarda as contagens e bloqueios, pode ser útil limpar os dados entre testes manuais:

```bash
docker compose exec redis redis-cli FLUSHALL
```

Ou derrube e suba tudo novamente:

```bash
docker compose down
docker compose up --build
```
