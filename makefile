# === Shared Tools ===

fmt:
	goimports -w .
	gofumpt -w .

lint:
	golangci-lint run --fix ./...

test:
	gotestsum -- -count=1 -tags=integration ./...

test-unit:
	gotestsum -- -count=1 ./...

test-integration:
	gotestsum -- -count=1 -tags=integration ./internal/integration/...

# === Microservices ===

lint-sub:
	cd microservices/subscription && golangci-lint run --fix ./...

test-sub:
	cd microservices/subscription && gotestsum -- -count=1 ./...

lint-email:
	cd microservices/email && golangci-lint run --fix ./...

test-email:
	cd microservices/email && gotestsum -- -count=1 ./...

lint-weather:
	cd microservices/weather && golangci-lint run --fix ./...

test-weather:
	cd microservices/weather && gotestsum -- -count=1 ./...

lint-micro: lint-sub lint-email lint-weather
test-micro: test-sub test-email test-weather

# === Monolith ===

lint-monolith:
	golangci-lint run --fix ./monolith/...

test-monolith:
	cd monolith && gotestsum -- -count=1 ./...

# === End-to-End ===

e2e-up:
	docker compose up -d --quiet-pull

e2e-test:
	go test -tags=e2e ./test/e2e -v

e2e-down:
	docker compose down -v

e2e: e2e-up
	sleep 5
	make e2e-test
	make e2e-down

# === Global Check ===

check: fmt lint test

# === RUN ===

COMPOSE=docker compose -f microservices/docker-compose.yml

.PHONY: up down restart logs build

up:
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down

restart:
	$(COMPOSE) down
	$(COMPOSE) up -d --build

logs:
	$(COMPOSE) logs -f --tail=100

build:
	$(COMPOSE) build


# === Generate PROTO ===

PROTO_DIR=microservices/proto
WEATHER_OUT_DIR=microservices/weather/internal/proto
SUBSCRIPTION_OUT_DIR=microservices/subscription/internal/adapter/weatherpb

generate-proto:
	mkdir -p $(WEATHER_OUT_DIR)
	mkdir -p $(SUBSCRIPTION_OUT_DIR)

	protoc -I=$(PROTO_DIR) \
	       --go_out=$(WEATHER_OUT_DIR) --go_opt=paths=source_relative \
	       --go-grpc_out=$(WEATHER_OUT_DIR) --go-grpc_opt=paths=source_relative \
	       $(PROTO_DIR)/weather.proto

	protoc -I=$(PROTO_DIR) \
	       --go_out=$(SUBSCRIPTION_OUT_DIR) --go_opt=paths=source_relative \
	       --go-grpc_out=$(SUBSCRIPTION_OUT_DIR) --go-grpc_opt=paths=source_relative \
	       $(PROTO_DIR)/weather.proto