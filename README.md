# Real Estate Backend

![Go](https://img.shields.io/badge/Go-1.23%2B-blue)

Бэкенд для платформы недвижимости: управление объектами, коллекциями, страницами,
менеджерами, медиа и отправкой писем. Реализован на Go с использованием Gin,
PostgreSQL, Redis и MinIO.

## Основные возможности
- Аутентификация, refresh-токены, сессии
- Управление объектами недвижимости и коллекциями
- Загрузка файлов в MinIO
- Кэширование в Redis
- Email-уведомления
- Публичные и защищенные API

## Стек
- Go, Gin
- PostgreSQL
- Redis
- MinIO
- Docker / Docker Compose

## Быстрый старт (локально)
1. Скопируйте переменные окружения:
   - `cp .env.example config/app.env`
2. Запустите зависимости через Docker:
   - `docker compose up -d`
3. Запустите сервис:
   - `make run`

## API
Базовый путь: `/api/v1`

Примеры:
- `POST /api/v1/user/sign-in`
- `POST /api/v1/user/auth/refresh`
- `GET /api/v1/real-estate`

## Makefile
- `make run` — запуск приложения
- `make build` — сборка бинарника
- `make test` — запуск тестов
- `make lint` — запуск линтера (если установлен `golangci-lint`)
- `make docker-up` / `make docker-down` — управление зависимостями

## Структура проекта
```
cmd/                 // точка входа
internal/            // бизнес-логика и слои приложения
pkg/                 // общие пакеты
db-setup/            // SQL миграции
templates/           // email-шаблоны
```

## Конфигурация
Все переменные — в `config/app.env` (см. `.env.example`).

