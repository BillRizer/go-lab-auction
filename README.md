# API de Consulta de Temperatura por CEP

Esta aplicação em Go recebe um CEP, identifica a cidade correspondente e retorna a temperatura atual em graus Celsius, Fahrenheit e Kelvin.

## Aplicacao online no Gcloud

https://go-lab-auction-jfc5xq2lta-uc.a.run.app/



## Requisitos

- Chave de API da [WeatherAPI](https://www.weatherapi.com/)

## Funcionalidades

- Validação de CEP (formato de 8 dígitos)
- Consulta de localização por CEP usando a API ViaCEP
- Consulta de temperatura atual usando a API WeatherAPI
- Conversão de temperatura para Celsius, Fahrenheit e Kelvin
- Respostas padronizadas conforme os requisitos do sistema

## Endpoints

**GET /weather/{cep}**

Retorna a temperatura atual para a localidade associada ao CEP.

**Respostas**

- **200 OK**: CEP válido e temperatura encontrada
  ```json
  { "temp_C": 28.5, "temp_F": 83.3, "temp_K": 301.65 }
  ```

- **422 Unprocessable Entity**: CEP com formato inválido
  ```json
  { "message": "invalid zipcode" }
  ```

- **404 Not Found**: CEP não encontrado
  ```json
  { "message": "can not find zipcode" }
  ```

**GET /health**

Verifica se o serviço está operacional.

### Como Executar
```
git clone git@github.com:BillRizer/go-lab-auction.git
cd go-lab-auction

```
crie o .env
```
PORT=8080
WEATHER_API_KEY=XXXXXXXX
```

A API estará disponível em `http://localhost:8080`

```
docker compose up --build
```


### Testes

Para executar os testes automatizados:

```bash
go test -v
```
