# Сервис сокращения ссылок с аналитикой

Аналог Bit.ly на Go. Backend-сервис для генерации коротких кодов, HTTP-редиректов, сбора статистики переходов (гео, рефереры, user-agent) и предоставления дашборда владельцам ссылок.

## Архитектура

```
┌─────────────┐     POST /api/shorten    ┌────────────┐     ┌────────────┐
│ Пользователь│ ───────────────────────→ │            │ ──→ │ PostgreSQL │
│ (владелец)  │ ←─────────────────────── │  Go Gin    │ ←── │  (links,   │
└─────────────┘     201 + short_url      │  Backend   │     │   clicks)  │
                                         │            │     └────────────┘
┌─────────────┐     GET /:code           │            │     ┌────────────┐
│ Пользователь│ ───────────────────────→ │            │ ──→ │   Redis    │
│ (посетитель)│ ←─────────────────────── │            │ ←── │   (кэш)    │
└─────────────┘     301 Redirect         └────────────┘     └────────────┘
```

## Технологический стек

| Компонент | Технология |
|-----------|-----------|
| Язык | Go 1.25 |
| HTTP фреймворк | Gin |
| База данных | PostgreSQL 16 |
| Кэш | Redis 7 |
| Контейнеризация | Docker + Docker Compose |
| Драйвер БД | pgx v5 |
| SAST | staticcheck, go vet |

## Быстрый старт

### Локальный запуск (без Docker)

```bash
# Требуется: Go 1.25+, PostgreSQL, Redis

# Клонировать и перейти в директорию
cd url-shortener

# Скопировать конфигурацию
cp .env.example .env
# Отредактировать .env под своё окружение

# Запустить миграции (выполнить 001_init.sql в PostgreSQL)

# Запустить сервер
go run ./cmd/server
```

### Запуск через Docker Compose

```bash
docker compose up --build
```

Сервер будет доступен на `http://localhost:8080`.

## API

### Создать короткую ссылку

```
POST /api/shorten
Content-Type: application/json

{
    "original_url": "https://example.com/very/long/url"
}
```

Ответ:
```json
{
    "short_url": "http://localhost:8080/abc123",
    "short_code": "abc123",
    "original_url": "https://example.com/very/long/url"
}
```

### Редирект по короткому коду

```
GET /:code
```
→ HTTP 301 Redirect на оригинальный URL

### Получить статистику по ссылке

```
GET /api/link/:id/stats
```

Ответ:
```json
{
    "total_clicks": 1542,
    "unique_clicks": 892,
    "clicks_by_day": [
        {"date": "2026-05-01", "count": 145},
        {"date": "2026-05-02", "count": 230}
    ],
    "top_referrers": [
        {"referrer": "twitter.com", "count": 412},
        {"referrer": "direct", "count": 380}
    ],
    "top_countries": [
        {"country": "RU", "count": 890},
        {"country": "US", "count": 312}
    ]
}
```

### Список ссылок владельца

```
GET /api/links?owner_id=user-123
```

### Удалить ссылку

```
DELETE /api/link
Content-Type: application/json

{
    "link_id": 1
}
```

### Health-check

```
GET /health
→ {"status": "ok"}
```

## Тестирование

```bash
# Запустить все тесты
go test -v -count=1 ./...

# Статический анализ
go vet ./...
staticcheck ./...
```

Проект содержит **67 модульных и интеграционных тестов**, покрывающих все ключевые компоненты.

## Переменные окружения

| Переменная | По умолчанию | Описание |
|-----------|-------------|----------|
| SERVER_PORT | 8080 | Порт сервера |
| BASE_URL | http://localhost:8080 | Базовый URL для коротких ссылок |
| POSTGRES_HOST | localhost | Хост PostgreSQL |
| POSTGRES_PORT | 5432 | Порт PostgreSQL |
| POSTGRES_USER | urlshortener | Пользователь БД |
| POSTGRES_PASSWORD | urlshortener_secret | Пароль БД |
| POSTGRES_DB | urlshortener | Имя БД |
| REDIS_HOST | localhost | Хост Redis |
| REDIS_PORT | 6379 | Порт Redis |
| REDIS_DB | 0 | Номер БД Redis |
| CACHE_TTL | 3600 | TTL кэша (секунды) |
