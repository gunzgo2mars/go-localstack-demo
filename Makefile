# config
COMPOSE=docker-compose
LOCALSTACK=localstack

# cli
up:
	$(COMPOSE) up -d

down:
	$(COMPOSE) down

run-http-service:
	@echo "Starting HTTP service..."
	@export $$(grep -v '^#' .env | xargs) && \
	go run --race cmd/http/main.go

run-consume-service:
	@echo "Starting Consume service..."
	@export $$(grep -v '^#' .env | xargs) && \
	go run --race cmd/consume/main.go

logs-localstack:
	$(COMPOSE) logs -f $(LOCALSTACK)

upload-s3:
	curl -X POST http://localhost:7777/upload -F "file=@file.txt"

sync-s3:
	aws s3 sync s3://localstack-bucket ./s3-upload --endpoint-url=http://localhost:4566

request-encrypt:
	curl -X POST http://localhost:7777/encrypt -H "Content-Type: application/json" -d '{"text":"Testing encrypt"}'

request-decrypt:
	curl -X POST http://localhost:7777/decrypt -H "Content-Type: application/json" -d '{"ciphertext":"YTYxOTQ4ZjYtMTQ2MS00MzFhLTgxM2UtOTlhZDk5MTQwODFmJR6MQN9702NZvtiBqUBeu1OmNoUaz1VkpEy3TqvCg2V0QG7+CGaiDN/SWwrz6PwE"}'

clean:
	docker system prune -f
