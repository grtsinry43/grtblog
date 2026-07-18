GOOSE ?= goose
GOOSE_DRIVER ?= postgres
MIGRATIONS_DIR ?= migrations
DB_DSN ?= postgres://postgres:postgres@localhost:5432/grtblog?sslmode=disable
SWAG ?= swag
SWAG_ENTRY ?= cmd/api/main.go
SWAG_OUT ?= docs

.PHONY: migrate-up migrate-down migrate-status migrate-version migrate-create docs test-federation-e2e test-backup-e2e

migrate-up:
	$(GOOSE) -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) "$(DB_DSN)" up

migrate-down:
	$(GOOSE) -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) "$(DB_DSN)" down

migrate-status:
	$(GOOSE) -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) "$(DB_DSN)" status

migrate-version:
	$(GOOSE) -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) "$(DB_DSN)" version

migrate-create:
ifndef NAME
	$(error NAME is required, e.g. make migrate-create NAME=add_posts_table)
endif
	$(GOOSE) -dir $(MIGRATIONS_DIR) create $(NAME) sql

docs:
	$(SWAG) init -g $(SWAG_ENTRY) -o $(SWAG_OUT) --outputTypes json --parseInternal

test-federation-e2e:
	go test -tags=federation_e2e -v ./tests/e2e -run '^TestFederationBetweenTwoIsolatedInstances$$' -count=1

test-backup-e2e:
	sh ./tests/e2e/backup_restore_container_e2e.sh
