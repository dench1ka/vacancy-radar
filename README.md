# vacancy-radar

[![CI](https://github.com/dench1ka/vacancy-radar/actions/workflows/ci.yml/badge.svg)](https://github.com/dench1ka/vacancy-radar/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dench1ka/vacancy-radar)](https://goreportcard.com/report/github.com/dench1ka/vacancy-radar)
[![Go Version](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go)](go.mod)

Небольшой сервис на Go, который сам ищет вакансии на hh.ru по заданным ключевым
словам и присылает уведомления в Telegram, как только появляется что-то новое.

Идея родилась из личной задачи: вместо того чтобы каждый день вручную
проверять hh.ru по запросу "golang junior", бот делает это за меня и шлёт
сообщение, если появилась новая подходящая вакансия.

## Как это работает

1. Пользователь регистрируется и логинится через REST API (JWT).
2. Создаёт подписку — ключевое слово + (опционально) регион hh.ru.
3. Привязывает свой Telegram chat ID.
4. Фоновый воркер каждые `POLL_INTERVAL` опрашивает hh.ru по всем активным
   подпискам (конкурентно, через worker pool на goroutines), сравнивает
   найденные вакансии с уже отправленными и шлёт новые в Telegram.

```
┌──────────┐      REST (JWT)       ┌──────────────┐
│  Client  │ ────────────────────▶ │  HTTP API    │
└──────────┘                       │ (net/http +  │
                                   │     chi)     │
                                   └──────┬───────┘
                                          │
                                   ┌──────▼───────┐         ┌───────────┐
                                   │  PostgreSQL  │◀──────▶ │  Worker   │
                                   └──────────────┘         │ (goroutines│
                                                            │ + ticker) │
                                          ┌─────────────┐   └─────┬─────┘
                                          │  api.hh.ru  │◀────────┘
                                          └─────────────┘         │
                                                                  ▼
                                                          ┌───────────────┐
                                                          │ Telegram Bot  │
                                                          └───────────────┘
```

## Стек

- Go 1.23, `net/http` + [chi](https://github.com/go-chi/chi) роутер
- PostgreSQL, `database/sql` + [pgx](https://github.com/jackc/pgx)
- JWT-авторизация ([golang-jwt](https://github.com/golang-jwt/jwt)) + bcrypt
- SQL-миграции ([golang-migrate](https://github.com/golang-migrate/migrate))
- Конкурентность: goroutines + worker pool для параллельного опроса подписок
- Docker / docker-compose
- GitHub Actions CI (vet, test -race, build)
- Юнит-тесты с моками внешних зависимостей (hh.ru клиент, Telegram, storage)

## Запуск локально

```bash
cp .env.example .env
# при желании впишите свой TELEGRAM_BOT_TOKEN

docker compose up --build
```

API будет доступен на `http://localhost:8080`.

## Запуск без Docker

```bash
# поднимите Postgres и примените миграции
make migrate-up

go run ./cmd/server
```

## API

| Метод  | Путь                         | Описание                              |
|--------|------------------------------|----------------------------------------|
| POST   | `/api/v1/auth/register`      | Регистрация (email + password)         |
| POST   | `/api/v1/auth/login`         | Логин, возвращает JWT                  |
| GET    | `/api/v1/me`                 | Текущий пользователь                   |
| PUT    | `/api/v1/me/telegram`        | Привязать Telegram chat ID             |
| POST   | `/api/v1/subscriptions`      | Создать подписку (keyword + area_id)   |
| GET    | `/api/v1/subscriptions`      | Список своих подписок                  |
| DELETE | `/api/v1/subscriptions/{id}` | Удалить подписку                       |
| GET    | `/api/v1/vacancies`          | Вакансии, которые уже были найдены     |

Все эндпоинты кроме `auth/*` и `/healthz` требуют заголовок
`Authorization: Bearer <token>`.

## Тесты

```bash
make test
```

Юнит-тесты покрывают: хэширование паролей и JWT (`internal/auth`), парсинг
ответа hh.ru (`internal/hhclient`), и логику дедупликации/конкурентной
обработки подписок в воркере (`internal/worker`) — всё через моки, без
реальной сети и базы.

## Что можно добавить дальше

- Rate limiting на регистрацию/логин
- Swagger/OpenAPI-описание API
- Интеграционные тесты с Testcontainers (поднятие реального Postgres)
- Пагинация и фильтры по зарплате в `/vacancies`
