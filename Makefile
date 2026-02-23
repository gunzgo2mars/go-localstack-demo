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

clean:
	docker system prune -f
