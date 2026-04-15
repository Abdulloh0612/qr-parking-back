# QR-Parking

Система идентификации владельцев автомобилей через QR-код. Веб-платформа, которая позволяет автовладельцам зарегистрировать свой автомобиль через уникальный QR-код для быстрой анонимной связи.

## Технический стек

| Компонент  | Технология                                                |
| ---------- | --------------------------------------------------------- |
| Backend    | Go 1.23+, Fiber v2, pgx v5, go-redis v9, golang-jwt      |
| Frontend   | SvelteKit, TypeScript, TailwindCSS, Lucide Icons          |
| БД         | PostgreSQL 16, Redis 7                                    |
| Bot        | Telegram Bot (telebot v3)                                 |
| Инфра      | Docker, Docker Compose, Nginx                             |

## Структура проекта

```
qr-parking/
├── main.go                   # Точка входа API (как в qr-parking)
├── server/                   # Сборка Fiber, маршруты, DI
├── cmd/
│   ├── bot/                  # Точка входа Telegram бота
│   └── hashpass/             # Утилита bcrypt для админ-пароля
├── handlers/                 # HTTP handlers (Fiber)
├── services/                 # Бизнес-логика
├── db/
│   ├── pool.go               # Подключение PostgreSQL (pgx pool)
│   └── repositories/         # Интерфейсы + реализации репозиториев
├── types/                    # Сущности и DTO
├── config/                   # Конфигурация из env
├── middleware/               # JWT, rate limit, device id
├── bot/                      # Telegram bot
├── migrations/               # SQL миграции
├── pkg/
│   ├── jwt/                  # JWT токены
│   ├── logger/               # Structured logging (zap)
│   ├── qrgen/                # Генерация QR-кодов
│   └── validator/            # Валидация данных
├── frontend/                 # SvelteKit приложение
│   └── src/routes/
│       ├── scan/[code]/      # Публичная страница сканирования
│       ├── register/         # Регистрация
│       ├── login/            # Вход
│       ├── profile/          # Личный кабинет
│       └── admin/            # Админ-панель
├── docker-compose.yml
├── Dockerfile
├── nginx.conf
└── Makefile
```

## Быстрый старт с Docker (рекомендуется)

### Требования

- Docker 20+ и Docker Compose v2

### Запуск

```bash
# 1. Клонируйте репозиторий
git clone <repo-url>
cd qr-parking

# 2. Создайте файл .env (уже создан с настройками по умолчанию)
# При необходимости отредактируйте .env:
#   - JWT_SECRET — секрет для JWT токенов (ОБЯЗАТЕЛЬНО сменить в продакшне)
#   - TG_BOT_TOKEN — токен Telegram бота (получить у @BotFather)

# 3. Запустите все сервисы
docker compose up -d

# 4. Проверьте что всё работает
docker compose ps
```

После запуска:
- **Frontend**: http://localhost (через Nginx)
- **API**: http://localhost/api/v1/
- **API напрямую**: http://localhost:8090

## Локальная разработка (без Docker для Go/SvelteKit)

### Требования

- Go 1.23+
- Node.js 18+
- PostgreSQL 16
- Redis 7

### Шаг 1 — Запустите PostgreSQL и Redis через Docker

```bash
docker compose up -d postgres redis
```

### Шаг 2 — Примените миграции

```bash
psql "postgres://qrparking:qrparking@localhost:5434/qrparking?sslmode=disable" -f migrations/001_init.up.sql
```

### Шаг 3 — Настройте переменные окружения

Файл `.env` уже создан с дефолтными значениями. При необходимости отредактируйте.

### Шаг 4 — Запустите API сервер

```bash
go run .
# или: make run-air   # Air, см. .air.toml
```

API будет доступен на http://localhost:8090

### Шаг 5 — Запустите Frontend (в отдельном терминале)

```bash
cd frontend
npm install
npm run dev
```

Frontend будет доступен на http://localhost:5173 (с проксированием API на порт API)

### Шаг 6 — Запустите Telegram бота (опционально, в отдельном терминале)

```bash
# Сначала установите TG_BOT_TOKEN в .env
go run ./cmd/bot/
```

## Makefile команды

```bash
make build              # Сборка Go бинарников
make run                # Запуск API (go run .)
make run-air            # API с Air (см. .air.toml)
make run-bot            # Запуск Telegram бота
make swagger            # Swagger (swag init, исключая ./qr-parking)
make migrate-up         # Применение миграций
make migrate-down       # Откат миграций
make docker-up          # Docker compose up
make docker-down        # Docker compose down
make docker-logs        # Логи Docker
make frontend-install   # npm install для frontend
make frontend-dev       # Запуск frontend dev server
make frontend-build     # Сборка frontend для продакшна
make test               # Запуск тестов
make lint               # Линтер (golangci-lint)
```

## API эндпоинты

### Публичные (без авторизации)

| Метод | URL                          | Описание                           |
| ----- | ---------------------------- | ---------------------------------- |
| GET   | `/api/v1/scan/:code`         | Получить данные QR-кода            |
| POST  | `/api/v1/scan/:code/message` | Отправить сообщение владельцу      |
| GET   | `/api/v1/qr/:code/status`    | Проверить статус QR                |
| GET   | `/api/v1/qr/:code/image`     | Получить QR-код как PNG            |

### Авторизованные (JWT)

| Метод  | URL                               | Описание                    |
| ------ | --------------------------------- | --------------------------- |
| GET    | `/api/v1/me/`                     | Профиль                     |
| PUT    | `/api/v1/me/`                     | Обновить профиль            |
| GET    | `/api/v1/me/vehicles`             | Мои автомобили              |
| POST   | `/api/v1/me/vehicles`             | Добавить автомобиль         |
| PUT    | `/api/v1/me/vehicles/:id`         | Обновить авто               |
| PATCH  | `/api/v1/me/vehicles/:id/privacy` | Переключить приватность     |
| GET    | `/api/v1/me/socials`              | Мои соц. сети               |
| POST   | `/api/v1/me/socials`              | Добавить соц. сеть          |
| PUT    | `/api/v1/me/socials/:id`          | Обновить соц. сеть          |
| DELETE | `/api/v1/me/socials/:id`          | Удалить соц. сеть           |
| GET    | `/api/v1/me/qrcodes`              | Мои QR-коды                 |
| GET    | `/api/v1/me/messages`             | Мои сообщения               |
| GET    | `/api/v1/me/scans`                | Мои сканирования            |
| POST   | `/api/v1/telegram/link`           | Привязать Telegram          |
| DELETE | `/api/v1/telegram/unlink`         | Отвязать Telegram           |

### Админ (JWT с `admin_session: true`, только после `POST /admin/login`)

| Метод | URL                                | Описание                    |
| ----- | ---------------------------------- | --------------------------- |
| GET   | `/api/v1/admin/stats`              | Статистика платформы        |
| GET   | `/api/v1/admin/users`              | Список пользователей        |
| GET   | `/api/v1/admin/users/:id`          | Детали пользователя         |
| PATCH | `/api/v1/admin/users/:id/block`    | Заблокировать пользователя  |
| GET   | `/api/v1/admin/qrcodes`            | Все QR-коды                 |
| POST  | `/api/v1/admin/qrcodes/generate`   | Генерировать QR-коды        |
| PATCH | `/api/v1/admin/qrcodes/:id/block`  | Заблокировать QR            |
| GET   | `/api/v1/admin/scans`              | Все сканирования            |
| GET   | `/api/v1/admin/messages`           | Все сообщения               |

## Переменные окружения

| Переменная    | Описание                       | По умолчанию                |
| ------------- | ------------------------------ | --------------------------- |
| `SERVER_PORT` | Порт API сервера               | `8090`                      |
| `DATABASE_URL`| URL подключения к PostgreSQL   | —                           |
| `REDIS_URL`   | Адрес Redis                    | `localhost:6379`            |
| `JWT_SECRET`  | Секретный ключ для JWT         | —                           |
| `TG_BOT_TOKEN`| Токен Telegram бота            | —                           |
| `APP_BASE_URL`| Базовый URL приложения         | `http://localhost:8090`     |
| `APP_ENV`     | Окружение (development/production) | `development`           |

## Админ-панель

Вход **только по логину и паролю** (не через телефон и OTP).

1. Откройте страницу **`/admin/login`** (например `http://localhost:3000/admin/login` в Docker или `http://localhost:5173/admin/login` в dev).
2. После успешного входа откроется **`/admin`** (дашборд). API `/api/v1/admin/*` принимает только токены, выданные через `POST /api/v1/admin/login` (в JWT есть `admin_session: true`). Войти в админку по обычному OTP нельзя.

### Миграция БД (колонки логина/пароля)

Если база уже создана без второй миграции:

```bash
psql "$DATABASE_URL" -f migrations/002_admin_credentials.up.sql
```

### Создать учётную запись администратора

1. У пользователя в таблице `users` должны быть **`is_admin = true`**, **`admin_username`** (уникальный логин) и **`password_hash`** (bcrypt).

2. Сгенерируйте хеш пароля:

```bash
go run ./cmd/hashpass/ 'ВашНадёжныйПароль'
```

3. Запишите в БД (пример: логин `admin`, подставьте свой UUID пользователя или номер телефона):

```bash
docker compose exec postgres psql -U qrparking -d qrparking -c \
  "UPDATE users SET is_admin = TRUE, admin_username = 'admin', password_hash = 'СЮДА_ВЫВОД_cmd_hashpass' WHERE phone = '+998901234567';"
```

Либо создайте отдельного пользователя с заведомым телефоном и назначьте ему поля выше.

---

**QR-Parking** — *Меньше конфликтов. Больше уважения на дороге.*
