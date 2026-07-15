# Local development against the DynamoDB Local emulator.
# `make dev` is THE single documented startup command (spec FR-004 / SC-001).

DYNAMODB_ENDPOINT ?= http://localhost:8000
AWS_REGION ?= us-east-1
AWS_ACCESS_KEY_ID ?= test
AWS_SECRET_ACCESS_KEY ?= test
JWT_SECRET ?= local-dev-secret-change-me
ADMIN_BOOTSTRAP_PASSWORD ?= admin123

# Env shared by the server and the test target.
LOCAL_ENV = \
	DYNAMODB_ENDPOINT=$(DYNAMODB_ENDPOINT) \
	AWS_REGION=$(AWS_REGION) \
	AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) \
	AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY)

.PHONY: dev db db-stop test

## dev: start the emulator (if needed) and run the API against it
dev: db
	@echo "Starting API on :8080 against $(DYNAMODB_ENDPOINT)"
	cd backend && $(LOCAL_ENV) JWT_SECRET=$(JWT_SECRET) \
		ADMIN_BOOTSTRAP_PASSWORD=$(ADMIN_BOOTSTRAP_PASSWORD) \
		go run ./cmd/server

## db: start DynamoDB Local and wait until it accepts connections
db:
	docker compose up -d dynamodb-local
	@echo "Waiting for DynamoDB Local on $(DYNAMODB_ENDPOINT)..."
	@for i in $$(seq 1 30); do \
		curl -s "$(DYNAMODB_ENDPOINT)" >/dev/null 2>&1 && exit 0; \
		sleep 1; \
	done; \
	echo "DynamoDB Local did not become ready" >&2; exit 1

## db-stop: stop and remove the emulator (discards local data)
db-stop:
	docker compose down

## test: run the backend test suite against the emulator
test: db
	cd backend && $(LOCAL_ENV) go test -p 1 ./...
