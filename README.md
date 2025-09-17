# CATA Take Home Test

## Project Description

A robust backend service that integrates Pokemon data from PokeAPI, featuring:
- External API data synchronization
- MySQL data storage
- Redis caching
- Automatic data refresh every 15 minutes

## Prerequisites

- Docker (latest version)
- Docker Compose
- Go (version 1.20+)
- Git

## Preparation if you don't have mysql and redis on your local machine

- `docker pull mysql:8.0`
- `docker pull redis:7.0`

- ```docker run -d \
  --name pokemon-mysql \
  --network poke-network \
  -e MYSQL_ROOT_PASSWORD=your_secret_password \
  -e MYSQL_DATABASE=pokemondb \
  -p 3306:3306 \
  mysql:8.0

- ```docker run -d \
  --name pokemon-redis \
  --network poke-network \
  -p 6379:6379 \
  redis:7.0


## Run on local

- `go run ./cmd/server/main.go`

## Test

Note: before running test, you need to run the server first
- `go test ./...`

## Run with docker-compose

- `docker-compose up -d`


## Endpoints

- POST http://localhost:8080/sync
- GET http://localhost:8080/items


### Note: I'm using port mysql 3307 on docker-compose because I already installed mysql on my local machine, but you can change it to 3306 if you want to use default port.
### Note: change it on .env and docker-compose.yaml