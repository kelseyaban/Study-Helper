## Filename Makefile
include .envrc

.PHONY: fmt
fmt: 
	go fmt ./...

.PHONY: vet
vet: fmt
	go vet ./...

.PHONY: run
run: vet
	go run ./cmd/web -addr=${ADDRESS} -dsn=${FEEDBACK_DB_DSN}


.PHONY: db/psql
db/psql:
	psql ${FEEDBACK_DB_DSN}