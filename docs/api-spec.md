# API Спецификация

## Base URL
`{VITE_API_URL}/api/v1`

---

## Endpoints

### 1. Проверить статус QR кода

**GET** `/qr/:qr_id/status`

Проверить зарегистрирован ли QR код.

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "registered": "boolean"
  }
}
```

**Использование:**
- `registered: false` - QR код не зарегистрирован, редирект на `/register?qr={id}`
- `registered: true` - QR код зарегистрирован, редирект на `/contact/{id}`

---

### 2. Регистрация QR кода

**POST** `/qr/register`

Регистрация нового автомобиля с QR кодом.

**Request Body:**
```json
{
  "qr_id": "string (required)",
  "phone": "string (required, +996...)",
  "first_name": "string (required)",
  "last_name": "string (required)",
  "vehicle_number": "string (required)",
  "vehicle_brand": "string (required)",
  "whatsapp": "string (optional)",
  "instagram": "string (optional)",
  "photo_url": "string (optional)",
  "reviews_enabled": "boolean (default: true)",
  "telegram_enabled": "boolean (default: false)",
  "device_id": "string (required)"
}
```

**Response:** `201 Created`
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "qr_id": "string",
    "created_at": "timestamp"
  }
}
```

---

### 3. Получить информацию о QR

**GET** `/qr/:qr_id`

Получить контактные данные владельца по QR коду.

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "phone": "string",
    "first_name": "string",
    "last_name": "string",
    "vehicle_number": "string",
    "vehicle_brand": "string",
    "whatsapp": "string | null",
    "instagram": "string | null",
    "photo_url": "string | null",
    "reviews_enabled": "boolean",
    "average_rating": "number (0-5)",
    "total_reviews": "number",
    "scan_count": "number"
  }
}
```

---

### 4. Отправить сообщение в Telegram

**POST** `/qr/:qr_id/message`

Отправить сообщение владельцу через Telegram бота.

**Request Body:**
```json
{
  "message": "string (required, max 500 chars)"
}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Сообщение отправлено"
}
```

**Rate Limit:** 1 сообщение в минуту с одного IP

---

### 5. Оставить рейтинг

**POST** `/qr/:qr_id/rating`

Оставить оценку владельцу.

**Request Body:**
```json
{
  "rating": "number (1-5, required)",
  "comment": "string (optional, max 200 chars)"
}
```

**Response:** `201 Created`
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "rating": "number",
    "created_at": "timestamp"
  }
}
```

---

### 6. Получить данные аккаунта

**GET** `/account`

Получить данные аккаунта по device_id.

**Headers:**
```
X-Device-ID: string (required)
```

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "qr_id": "string",
    "phone": "string",
    "first_name": "string",
    "last_name": "string",
    "vehicle_number": "string",
    "vehicle_brand": "string",
    "whatsapp": "string | null",
    "instagram": "string | null",
    "photo_url": "string | null",
    "reviews_enabled": "boolean",
    "telegram_enabled": "boolean",
    "scan_count": "number",
    "created_at": "timestamp"
  }
}
```

---

### 7. Получить сообщения аккаунта

**GET** `/account/messages`

Получить все сообщения из Telegram.

**Headers:**
```
X-Device-ID: string (required)
```

**Response:** `200 OK`
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "message": "string",
      "created_at": "timestamp"
    }
  ]
}
```

---

### 7. Получить отзывы аккаунта

**GET** `/account/reviews`

Получить все отзывы.

**Headers:**
```
X-Device-ID: string (required)
```

**Response:** `200 OK`
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "rating": "number",
      "comment": "string | null",
      "created_at": "timestamp"
    }
  ]
}
```

---

### 9. Обновить профиль

**PATCH** `/account`

Обновить данные профиля.

**Headers:**
```
X-Device-ID: string (required)
```

**Request Body:**
```json
{
  "phone": "string (optional)",
  "whatsapp": "string (optional)",
  "instagram": "string (optional)",
  "photo_url": "string (optional)",
  "reviews_enabled": "boolean (optional)",
  "telegram_enabled": "boolean (optional)"
}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "updated_at": "timestamp"
  }
}
```

---

## Error Responses

**400 Bad Request**
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Описание ошибки",
    "fields": {
      "phone": "Неверный формат номера"
    }
  }
}
```

**404 Not Found**
```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "QR код не найден"
  }
}
```

**429 Too Many Requests**
```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT",
    "message": "Слишком много запросов. Попробуйте через 1 минуту",
    "retry_after": 60
  }
}
```

**500 Internal Server Error**
```json
{
  "success": false,
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "Что-то пошло не так. Попробуйте позже"
  }
}
```
