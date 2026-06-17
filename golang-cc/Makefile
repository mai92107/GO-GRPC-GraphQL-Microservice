APP_BIN := ./bin/server
COMPOSE_SERVICE := app
HEALTH_TOKEN := local-monitoring-token-change-me
HEALTH_URL := http://localhost:8080/actuator/health
export GOCACHE := $(CURDIR)/.gocache

.PHONY: test test-db build docker-build deploy logs status down health

test-db:
	docker compose up -d postgres
	docker compose exec -T postgres sh -c "psql -U credit_cards -d postgres -tc \"SELECT 1 FROM pg_database WHERE datname = 'credit_cards_test'\" | grep -q 1 || createdb -U credit_cards credit_cards_test"

test:
	go test -p 1 -count=1 ./...
	cd web && npm test -- --run

build:
	cd web && npm run build
	mkdir -p bin
	go build -o $(APP_BIN) .

docker-build:
	docker compose build $(COMPOSE_SERVICE)

deploy: test-db test build docker-build
	docker compose up -d
	$(MAKE) health

logs:
	docker compose logs -f $(COMPOSE_SERVICE)

status:
	docker compose ps

down:
	docker compose down

health:
	curl -fsS -H 'Authorization: Bearer $(HEALTH_TOKEN)' $(HEALTH_URL) >/dev/null
	docker compose ps
