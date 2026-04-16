# QR Parking — API Документация

**Base URL:** `http://localhost:8090/api/v1`

**Swagger UI:** `http://localhost:8090/swagger/`

---

## Формат ответов

Все эндпоинты возвращают JSON в одном из двух форматов:

**Успех:**
```json
{ "success": true, "data": { ... } }
```

**Ошибка:**
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Описание ошибки",
    "fields": {
      "phone": "Обязательное поле"
    }
  }
}
```

### Коды ошибок

| HTTP статус | code             | Когда возникает                         |
|-------------|------------------|-----------------------------------------|
| `400`       | `VALIDATION_ERROR` | Неверные или отсутствующие поля        |
| `401`       | —                | Нет токена или токен недействителен     |
| `403`       | —                | Нет прав (не администратор)             |
| `404`       | `NOT_FOUND`      | QR код / пользователь не найден        |
| `500`       | `INTERNAL_ERROR` | Внутренняя ошибка сервера              |

---

## 🚗 Публичные QR эндпоинты

> Авторизация не требуется.

---

### `GET /qr/:qr_id`

**Получить данные по QR коду**

Вызывается при сканировании QR наклейки. Возвращает профиль владельца и данные машины.

| Параметр | Тип    | Где      | Описание              |
|----------|--------|----------|-----------------------|
| `qr_id`  | string | path     | Код с QR наклейки     |

**Ответ — QR ещё не зарегистрирован:**
```json
{
  "success": true,
  "data": {
    "registered": false
  }
}
```

**Ответ — QR зарегистрирован:**
```json
{
  "success": true,
  "data": {
    "registered": true,
    "owner": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "phone": "+998901234567",
      "first_name": "Алибек",
      "last_name": "Каримов",
      "avatar_url": "https://cdn.example.com/photo.jpg",
      "vehicle_number": "01A123BC",
      "vehicle_brand": "Chevrolet Nexia 3",
      "vehicle_photo_url": "https://cdn.example.com/car.jpg",
      "whatsapp": "+998901234567",
      "instagram": "@alibek",
      "telegram": "@alibek_k",
      "vk": null,
      "facebook": null,
      "telegram_enabled": true,
      "scan_count": 42
    }
  }
}
```

---

### `POST /qr/:qr_id`

**Шаг 1 — Отправить OTP на телефон**

Первый шаг привязки QR к владельцу. Генерирует 6-значный код и отправляет его SMS на указанный номер. Код действителен **5 минут**.

| Параметр | Тип    | Где      | Описание          |
|----------|--------|----------|-------------------|
| `qr_id`  | string | path     | Код с QR наклейки |
| `phone`  | string | body     | Номер телефона    |

**Body:**
```json
{
  "phone": "+998901234567"
}
```

**Ответ `200`:**
```json
{
  "success": true,
  "data": {
    "message": "OTP отправлен на +998901234567"
  }
}
```

---

### `POST /qr-verify/:qr_id`

**Шаг 2 — Верифицировать OTP и получить токен**

Проверяет OTP из SMS. Если пользователь с таким телефоном уже есть в базе — привязывает QR к нему. Если нет — создаёт нового пользователя. Возвращает `access_token` (действует **30 дней**).

| Параметр | Тип    | Где  | Описание               |
|----------|--------|------|------------------------|
| `qr_id`  | string | path | Тот же QR код          |
| `phone`  | string | body | Номер телефона         |
| `otp`    | string | body | 6-значный код из SMS   |

**Body:**
```json
{
  "phone": "+998901234567",
  "otp": "394821"
}
```

**Ответ `201`:**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "is_new_user": true
  }
}
```

> Токен также устанавливается в cookie `owner_token` (httpOnly, SameSite=Lax).

---

### `POST /qr-message/:qr_id`

**Отправить сообщение владельцу машины**

Любой прохожий может написать владельцу — например, «у вас фары не выключены». Если у владельца привязан Telegram через бота — придёт уведомление в мессенджер.

| Параметр  | Тип    | Где  | Описание                      |
|-----------|--------|------|-------------------------------|
| `qr_id`   | string | path | Код с QR наклейки             |
| `message` | string | body | Текст сообщения (1–500 симв.) |

**Body:**
```json
{
  "message": "У вас спущено заднее колесо!"
}
```

**Ответ `200`:**
```json
{
  "success": true,
  "message": "Сообщение отправлено"
}
```

---

## 👤 Личный кабинет владельца

> Все эндпоинты требуют заголовок:
> ```
> Authorization: Bearer {access_token}
> ```

---

### `GET /me`

**Получить свой профиль**

Возвращает полный профиль: имя, телефон, аватар, соцсети и список машин.

**Ответ `200`:**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "phone": "+998901234567",
    "first_name": "Алибек",
    "last_name": "Каримов",
    "avatar_url": null,
    "vehicles": [
      {
        "id": "660e8400-e29b-41d4-a716-446655440000",
        "plate_number": "01A123BC",
        "car_model": "Chevrolet Nexia 3",
        "photo_url": null,
        "telegram_enabled": true,
        "qr_code": "abc123xyz"
      }
    ],
    "whatsapp": "+998901234567",
    "instagram": null,
    "telegram": "@alibek_k",
    "vk": null,
    "facebook": null
  }
}
```

---

### `PATCH /me`

**Обновить профиль**

Передавай только те поля, которые нужно изменить. Поле отсутствует или `null` — не трогается. Пустая строка `""` у соцсетей — **удалить** ссылку.

| Поле        | Тип    | Описание                                    |
|-------------|--------|---------------------------------------------|
| `first_name`| string | Имя                                         |
| `last_name` | string | Фамилия                                     |
| `avatar_url`| string | Ссылка на фото профиля                      |
| `whatsapp`  | string | Номер WhatsApp (пустая строка — удалить)    |
| `instagram` | string | Никнейм Instagram (пустая строка — удалить) |
| `telegram`  | string | Никнейм Telegram (пустая строка — удалить)  |
| `vk`        | string | Ссылка VK (пустая строка — удалить)         |
| `facebook`  | string | Ссылка Facebook (пустая строка — удалить)   |

**Body:**
```json
{
  "first_name": "Алибек",
  "last_name": "Каримов",
  "avatar_url": "https://cdn.example.com/photo.jpg",
  "whatsapp": "+998901234567",
  "instagram": "@alibek",
  "telegram": "@alibek_k",
  "vk": "",
  "facebook": null
}
```

**Ответ `200`:** такой же формат как у `GET /me`

---

### `GET /me/vehicles`

**Получить список своих машин**

**Ответ `200`:**
```json
{
  "success": true,
  "data": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440000",
      "plate_number": "01A123BC",
      "car_model": "Chevrolet Nexia 3",
      "photo_url": "https://cdn.example.com/car.jpg",
      "telegram_enabled": true,
      "qr_code": "abc123xyz"
    }
  ]
}
```

---

### `PATCH /me/vehicles/:id`

**Обновить данные машины**

Передавай только поля для изменения.

| Параметр | Тип    | Где  | Описание        |
|----------|--------|------|-----------------|
| `id`     | UUID   | path | UUID машины     |

| Поле               | Тип    | Описание                              |
|--------------------|--------|---------------------------------------|
| `plate_number`     | string | Госномер (например, `01A123BC`)       |
| `car_model`        | string | Марка и модель                        |
| `photo_url`        | string | Ссылка на фото машины                 |
| `telegram_enabled` | bool   | Получать ли уведомления в Telegram    |

**Body:**
```json
{
  "plate_number": "01A123BC",
  "car_model": "Chevrolet Nexia 3",
  "photo_url": "https://cdn.example.com/car.jpg",
  "telegram_enabled": true
}
```

**Ответ `200`:**
```json
{
  "success": true,
  "data": {
    "id": "660e8400-e29b-41d4-a716-446655440000",
    "plate_number": "01A123BC",
    "car_model": "Chevrolet Nexia 3",
    "photo_url": "https://cdn.example.com/car.jpg",
    "telegram_enabled": true,
    "qr_code": "abc123xyz"
  }
}
```

---

## 🔐 Админ панель

> Все эндпоинты (кроме `/admin/login`) требуют заголовок:
> ```
> Authorization: Bearer {access_token}
> ```

---

### `POST /admin/login`

**Войти в админку**

**Body:**
```json
{
  "username": "admin",
  "password": "secret123"
}
```

**Ответ `200`:**
```json
{
  "tokens": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
  },
  "admin_session": true
}
```

---

### `GET /admin/admins`

**Список всех администраторов**

| Параметр | Тип | Где   | По умолчанию |
|----------|-----|-------|--------------|
| `page`   | int | query | `1`          |
| `limit`  | int | query | `20`         |

**Ответ `200`:**
```json
{
  "data": [
    {
      "id": 1,
      "display_id": "uuid",
      "username": "admin",
      "created_at": "2026-01-15T10:00:00Z",
      "updated_at": "2026-01-15T10:00:00Z"
    }
  ],
  "total": 3,
  "page": 1,
  "limit": 20
}
```

---

### `GET /admin/admins/:id`

**Получить администратора по UUID**

| Параметр | Тип  | Где  | Описание            |
|----------|------|------|---------------------|
| `id`     | UUID | path | UUID администратора |

**Ответ `200`:**
```json
{
  "data": {
    "id": 1,
    "display_id": "uuid",
    "username": "admin",
    "created_at": "2026-01-15T10:00:00Z",
    "updated_at": "2026-01-15T10:00:00Z"
  }
}
```

---

### `PATCH /admin/admins/:id/block`

**Удалить администратора**

> Нельзя удалить самого себя.

| Параметр | Тип  | Где  | Описание            |
|----------|------|------|---------------------|
| `id`     | UUID | path | UUID администратора |

**Ответ `200`:**
```json
{ "message": "admin removed" }
```

---

### `GET /admin/users`

**Список всех пользователей (владельцев)**

| Параметр | Тип | Где   | По умолчанию |
|----------|-----|-------|--------------|
| `page`   | int | query | `1`          |
| `limit`  | int | query | `20`         |

**Ответ `200`:**
```json
{
  "data": [
    {
      "id": 1,
      "display_id": "550e8400-e29b-41d4-a716-446655440000",
      "phone": "+998901234567",
      "first_name": "Алибек",
      "last_name": "Каримов",
      "is_admin": false,
      "avatar_url": null,
      "created_at": "2026-01-15T10:00:00Z",
      "updated_at": "2026-01-15T10:00:00Z"
    }
  ],
  "total": 150,
  "page": 1,
  "limit": 20
}
```

---

### `GET /admin/users/:id`

**Получить пользователя по UUID**

| Параметр | Тип  | Где  | Описание           |
|----------|------|------|--------------------|
| `id`     | UUID | path | UUID пользователя  |

**Ответ `200`:** тот же формат что в списке пользователей.

---

### `PATCH /admin/users/:id/block`

**Заблокировать пользователя**

Удаляет пользователя и все его данные (CASCADE).

| Параметр | Тип  | Где  | Описание           |
|----------|------|------|--------------------|
| `id`     | UUID | path | UUID пользователя  |

**Ответ `200`:**
```json
{ "message": "user blocked" }
```

---

### `POST /admin/qrcodes/generate`

**Сгенерировать новые QR коды**

Создаёт N уникальных QR кодов со статусом `unregistered`.

| Поле    | Тип | Описание                    |
|---------|-----|-----------------------------|
| `count` | int | Количество кодов (1–1000)   |

**Body:**
```json
{ "count": 50 }
```

**Ответ `201`:**
```json
{
  "count": 50,
  "data": [
    {
      "id": 101,
      "code": "abc123xyz456",
      "status": "unregistered",
      "created_at": "2026-04-17T10:00:00Z"
    }
  ]
}
```

---

### `PATCH /admin/qrcodes/:id/block`

**Заблокировать QR код**

| Параметр | Тип    | Где  | Описание                              |
|----------|--------|------|---------------------------------------|
| `id`     | string | path | Код QR строкой (`abc123xyz`) или числовой id |

**Ответ `200`:**
```json
{ "message": "QR code blocked" }
```

---

### `GET /admin/messages`

**Список всех сообщений (по всем машинам)**

| Параметр | Тип | Где   | По умолчанию |
|----------|-----|-------|--------------|
| `page`   | int | query | `1`          |
| `limit`  | int | query | `20`         |

**Ответ `200`:**
```json
{
  "data": [
    {
      "id": "uuid",
      "qr_code_id": "abc123xyz",
      "vehicle_id": "uuid",
      "content": "У вас фары включены",
      "sender_name": null,
      "sender_phone": null,
      "is_delivered": true,
      "delivered_at": "2026-04-17T11:00:00Z",
      "created_at": "2026-04-17T10:55:00Z"
    }
  ],
  "total": 320,
  "page": 1,
  "limit": 20
}
```

---

### `GET /admin/messages/:user_id`

**Сообщения конкретного пользователя**

Возвращает все сообщения, отправленные на машины данного пользователя.

| Параметр  | Тип  | Где   | Описание                        |
|-----------|------|-------|---------------------------------|
| `user_id` | UUID | path  | UUID пользователя               |
| `page`    | int  | query | Страница (по умолчанию `1`)     |
| `limit`   | int  | query | Лимит (по умолчанию `20`)       |

**Ответ `200`:** тот же формат что у `GET /admin/messages`

---

## 🔄 Типичный сценарий использования

```
1. Пользователь сканирует QR наклейку на машине
   → GET /qr/{qr_id}
   
   Если registered: false — начинаем регистрацию:
   
2. Вводит номер телефона
   → POST /qr/{qr_id}   { "phone": "+998901234567" }
   
3. Получает SMS с кодом, вводит его
   → POST /qr-verify/{qr_id}   { "phone": "...", "otp": "394821" }
   ← Получает access_token
   
4. Заполняет данные своей машины
   → PATCH /me/vehicles/{id}   { "plate_number": "01A123BC", ... }
   
5. Заполняет свой профиль
   → PATCH /me   { "first_name": "Алибек", "whatsapp": "+998...", ... }
   
6. Кто-то сканирует его QR — видит профиль и пишет сообщение
   → GET /qr/{qr_id}
   → POST /qr-message/{qr_id}   { "message": "Спущено колесо!" }
```
