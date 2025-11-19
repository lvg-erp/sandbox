# === Переменные ===
MIGRATIONS_DIR := migrations
DB_URL := $(DATABASE_URL)
ifeq ($(DB_URL),)
    DB_URL := postgres://user:pass@localhost:5464/cinema?sslmode=disable
endif

# === Миграции ===
.PHONY: migrate up down reset status create

# Применить миграции (основная команда)
migrate: up

# UP — применить все
up:
	@echo "Применение миграций..."
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" up
	@echo "Миграции применены!"

# DOWN — откатить последнюю
down:
	@echo "Откат последней миграции..."
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" down
	@echo "Откат выполнен!"

# RESET — удалить всё
reset:
	@echo "Сброс всей схемы..."
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" reset
	@echo "Схема очищена!"

# STATUS
status:
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" status

# Создать новую миграцию
create:
	@test -n "$(name)" || (echo "Использование: make create name=название_миграции" && exit 1)
	@goose -dir $(MIGRATIONS_DIR) create $(name) sql

# === Локальный запуск ===
run: up
	@go run cmd/app/main.go

# === Docker ===
docker-up:
	@docker compose up --build

docker-migrate:
	@docker compose run --rm cinema make up