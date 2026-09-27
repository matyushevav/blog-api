# Blog API

REST API для блог-платформы на Go: регистрация и вход пользователей по JWT, посты, комментарии и асинхронный журнал событий.

## Содержание

- [Возможности](#возможности)
- [Технологии](#технологии)
- [Структура проекта](#структура-проекта)
- [Быстрый старт](#быстрый-старт)
- [Конфигурация](#конфигурация)
- [API](#api)
- [Журнал событий](#журнал-событий)
- [Проверка работы](#проверка-работы)

## Возможности

- регистрация и вход, пароли хранятся в виде bcrypt-хеша;
- авторизация по JWT (HS256), срок жизни токена настраивается;
- создание и чтение постов с пагинацией;
- комментарии к постам;
- валидация входных данных, ответы и ошибки в формате JSON;
- журнал событий: создание постов и комментариев пишется в `log.txt` отдельной горутиной через канал;
- логирование HTTP-запросов в консоль, перехват паник, CORS;
- graceful shutdown: по Ctrl+C или `docker stop` сервер дорабатывает текущие запросы, журнал дописывается, соединение с БД закрывается;
- хранение данных в PostgreSQL, миграции применяются автоматически при старте.

## Технологии

| Что | Чем |
|---|---|
| Язык | Go 1.24 |
| HTTP-роутер | [chi](https://github.com/go-chi/chi) |
| База данных | PostgreSQL 15, драйвер [lib/pq](https://github.com/lib/pq) |
| Аутентификация | [golang-jwt/jwt v5](https://github.com/golang-jwt/jwt) |
| Хеширование паролей | [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) |
| Валидация | [go-playground/validator](https://github.com/go-playground/validator) |
| Конфигурация | [godotenv](https://github.com/joho/godotenv) |
| Запуск | Docker, Docker Compose |

## Структура проекта

```
blog-api/
├── api/
│   └── main.go                  # точка входа: конфиг, сборка зависимостей, роутер, запуск и остановка
├── internal/
│   ├── errors/apperrors/        # ошибки приложения (not found, already exists и т.д.)
│   ├── handler/                 # HTTP-обработчики и перевод ошибок в HTTP-коды
│   ├── logger/                  # журнал событий: канал + горутина-воркер
│   ├── middleware/              # JWT-авторизация, логирование запросов, Recovery, CORS
│   ├── model/                   # модели, запросы, ответы и правила валидации
│   ├── repository/              # SQL-запросы к PostgreSQL
│   └── service/                 # бизнес-логика
├── pkg/
│   ├── auth/                    # JWT и bcrypt
│   └── database/                # подключение к PostgreSQL и миграции
├── migrations/                  # SQL-миграции, применяются по порядку имен
├── Dockerfile
├── docker-compose.yml
└── .env.example                 # пример конфигурации
```

Запрос проходит слои сверху вниз: `middleware` → `handler` → `service` → `repository` → PostgreSQL.

## Быстрый старт

Понадобятся Docker и Docker Compose. Для локального запуска без Docker ещё Go 1.24+.

### Вариант 1. Всё в Docker

```bash
docker compose up -d --build
```

Поднимутся два контейнера: PostgreSQL и приложение. API будет доступно на `http://localhost:8080`.

```bash
docker compose logs -f app      # логи приложения
docker compose down             # остановить
docker compose down -v          # остановить и удалить данные базы
```

Журнал событий внутри контейнера:

```bash
docker compose exec app cat log.txt
```

### Вариант 2. Приложение локально, база в Docker

```bash
cp .env.example .env            # один раз, затем задать свой JWT_SECRET
docker compose up -d db         # только база, наружу на порту 7432
go run ./api                    # из корня проекта
```

Запускать можно и из папки `api` (`go run .`): приложение само находит корень проекта по папке `migrations`, поэтому `.env`, миграции и `log.txt` всегда берутся из корня.

Остановка: Ctrl+C.

## Конфигурация

Настройки берутся из переменных окружения. При локальном запуске они загружаются из файла `.env` в корне проекта, пример лежит в `.env.example`. В Docker их задаёт `docker-compose.yml`.

| Переменная | По умолчанию | Описание |
|---|---|---|
| `DB_HOST` | `localhost` | хост PostgreSQL (в Docker: `db`) |
| `DB_PORT` | `5432` | порт PostgreSQL (локально с базой из compose: `7432`) |
| `DB_USER` | `postgres` | пользователь БД |
| `DB_PASSWORD` | `postgres` | пароль БД |
| `DB_NAME` | `blog_db` | имя базы |
| `DB_SSLMODE` | `disable` | режим SSL |
| `JWT_SECRET` | — | **обязательно**, секрет для подписи токенов; без него приложение не запустится |
| `JWT_EXPIRY_HOURS` | `24` | срок жизни токена в часах |
| `SERVER_HOST` | `0.0.0.0` | адрес HTTP-сервера |
| `SERVER_PORT` | `8080` | порт HTTP-сервера |
| `LOGS_FILE` | `./log.txt` | файл журнала событий |

В `docker-compose.yml` для `JWT_SECRET` задано запасное значение `change-me-in-production`, чтобы проект запускался без подготовки. Для реального использования задайте свой секрет в `.env`.

## API

Базовый адрес: `http://localhost:8080/api`. Все запросы и ответы в формате JSON.

### Эндпоинты

| Метод | Путь | Авторизация | Описание |
|---|---|---|---|
| GET | `/api/health` | — | проверка доступности |
| POST | `/api/register` | — | регистрация |
| POST | `/api/login` | — | вход |
| GET | `/api/profile` | JWT | профиль текущего пользователя |
| POST | `/api/posts` | JWT | создать пост |
| GET | `/api/posts` | — | список постов |
| GET | `/api/posts/{id}` | — | пост по ID |
| GET | `/api/posts/author/{authorID}` | — | посты автора |
| POST | `/api/posts/{postId}/comments` | JWT | добавить комментарий |
| GET | `/api/posts/{postId}/comments` | — | комментарии к посту |

Для эндпоинтов с авторизацией нужен заголовок:

```
Authorization: Bearer <token>
```

Токен выдаётся при регистрации и при входе.

### Формат ошибок

Ошибки обработчиков:

```json
{
  "error": "Not Found",
  "message": "post not found"
}
```

`error` — стандартное название HTTP-статуса, `message` — описание причины.

Ошибки авторизации (нет токена, токен неверный или истёк) возвращаются с кодом `401` в коротком виде:

```json
{
  "error": "invalid token"
}
```

| Код | Когда |
|---|---|
| `400` | некорректный JSON, не прошла валидация, неверный ID или параметр пагинации |
| `401` | нет токена, токен невалиден или истёк; неверный email или пароль |
| `404` | пост или пользователь не найден |
| `405` | неподдерживаемый HTTP-метод |
| `409` | пользователь с таким email или username уже существует |
| `500` | внутренняя ошибка сервера (подробности только в логах сервера) |

### Пагинация

Списки принимают параметры `limit` и `offset`:

- `limit` — сколько записей вернуть: по умолчанию 10, максимум 100 (большее значение будет урезано до 100);
- `offset` — сколько записей пропустить: по умолчанию 0.

Если параметр не число, ответ `400`. Ответ со списком содержит массив и общее количество записей:

```json
{
  "posts": [ ... ],
  "total": 42
}
```

---

### GET /api/health

Проверка, что сервис работает.

```bash
curl http://localhost:8080/api/health
```

**200 OK**

```json
{"status": "ok"}
```

---

### POST /api/register

Регистрация нового пользователя. Сразу возвращает токен.

| Поле | Правила |
|---|---|
| `username` | обязательно, 3–50 символов, уникальное |
| `email` | обязательно, корректный email, уникальный |
| `password` | обязательно, минимум 6 символов |

```bash
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "email": "test@example.com", "password": "password123"}'
```

**201 Created**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2026-09-28T19:51:49Z",
  "user": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "created_at": "2026-09-27T19:51:49Z"
  }
}
```

**Ошибки:** `400` — некорректный JSON или данные не прошли валидацию, `409` — email или username уже заняты.

---

### POST /api/login

Вход по email и паролю.

```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "password123"}'
```

**200 OK** — ответ такой же, как у регистрации: `token`, `expires_at`, `user`.

**Ошибки:** `400` — некорректный JSON или данные не прошли валидацию, `401` — неверный email или пароль.

Для несуществующего email и для неверного пароля ответ одинаковый (`invalid email or password`), чтобы по ответу нельзя было узнать, зарегистрирован ли адрес.

---

### GET /api/profile

Профиль владельца токена.

```bash
curl http://localhost:8080/api/profile \
  -H "Authorization: Bearer <token>"
```

**200 OK**

```json
{
  "id": 1,
  "username": "testuser",
  "email": "test@example.com",
  "created_at": "2026-09-27T19:51:49Z"
}
```

**Ошибки:** `401` — нет токена или он невалиден, `404` — пользователь удалён.

---

### POST /api/posts

Создание поста. Автором становится владелец токена.

| Поле | Правила |
|---|---|
| `title` | обязательно, 1–200 символов |
| `content` | обязательно, не пустое |

```bash
curl -X POST http://localhost:8080/api/posts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"title": "My First Post", "content": "This is the content of my first post"}'
```

**201 Created**

```json
{
  "id": 1,
  "title": "My First Post",
  "content": "This is the content of my first post",
  "author_id": 1,
  "created_at": "2026-09-27T19:51:50Z"
}
```

**Ошибки:** `400` — некорректный JSON или данные не прошли валидацию, `401` — нет токена или он невалиден.

Событие `user <id> created post <id>` записывается в журнал.

---

### GET /api/posts

Список постов, новые сверху.

```bash
curl "http://localhost:8080/api/posts?limit=10&offset=0"
```

**200 OK**

```json
{
  "posts": [
    {
      "id": 1,
      "title": "My First Post",
      "content": "This is the content of my first post",
      "author_id": 1,
      "created_at": "2026-09-27T19:51:50Z"
    }
  ],
  "total": 1
}
```

Если постов нет, `posts` — пустой массив `[]`.

**Ошибки:** `400` — `limit` или `offset` не число.

---

### GET /api/posts/{id}

Один пост по ID.

```bash
curl http://localhost:8080/api/posts/1
```

**200 OK** — объект поста, как в ответе на создание.

**Ошибки:** `400` — ID не положительное число, `404` — поста нет.

---

### GET /api/posts/author/{authorID}

Посты одного автора, новые сверху. Параметры пагинации и формат ответа такие же, как у `GET /api/posts`.

```bash
curl "http://localhost:8080/api/posts/author/1?limit=10&offset=0"
```

**200 OK** — `{"posts": [...], "total": N}`. Для автора без постов (или несуществующего) — пустой список и `total: 0`.

**Ошибки:** `400` — некорректный ID автора или параметр пагинации.

---

### POST /api/posts/{postId}/comments

Комментарий к посту. ID поста берётся из пути, автором становится владелец токена.

| Поле | Правила |
|---|---|
| `content` | обязательно, 1–1000 символов; пробелы по краям обрезаются |

```bash
curl -X POST http://localhost:8080/api/posts/1/comments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"content": "Great post!"}'
```

**201 Created**

```json
{
  "id": 1,
  "content": "Great post!",
  "post_id": 1,
  "author_id": 1,
  "created_at": "2026-09-27T19:51:50Z",
  "updated_at": "2026-09-27T19:51:50Z"
}
```

**Ошибки:** `400` — некорректный ID поста, JSON или текст, `401` — нет токена или он невалиден, `404` — поста нет.

Событие `user <id> created comment <id>` записывается в журнал.

---

### GET /api/posts/{postId}/comments

Комментарии к посту в порядке добавления, старые сверху.

```bash
curl "http://localhost:8080/api/posts/1/comments?limit=10&offset=0"
```

**200 OK**

```json
{
  "comments": [
    {
      "id": 1,
      "content": "Great post!",
      "post_id": 1,
      "author_id": 1,
      "created_at": "2026-09-27T19:51:50Z",
      "updated_at": "2026-09-27T19:51:50Z"
    }
  ],
  "total": 1
}
```

**Ошибки:** `400` — некорректный ID поста или параметр пагинации, `404` — поста нет.

## Журнал событий

При создании поста или комментария обработчик отправляет строку-событие в буферизованный канал и сразу отвечает клиенту, не дожидаясь записи на диск. Отдельная горутина-воркер читает канал и дописывает события в `log.txt` с задержкой в 1 секунду:

```
[2026-09-27 19:51:51] user 1 created post 1
[2026-09-27 19:51:52] user 1 created comment 1
```

При остановке сервера канал закрывается, воркер дописывает все оставшиеся события и только после этого закрывает файл. Если буфер канала (100 событий) переполнен, событие не записывается, а в консоль выводится предупреждение: запрос пользователя при этом не задерживается.

HTTP-запросы отдельно логируются в консоль: адрес клиента, метод, путь, код ответа и время обработки.

## Проверка работы

Для проверки API есть Postman-коллекция: [`postman/blog-api.postman_collection.json`](postman/blog-api.postman_collection.json). В ней 23 запроса по всем эндпоинтам: успешные сценарии и основные ошибки (400, 401, 404, 409). У каждого запроса есть тест на ожидаемый код ответа.

1. Запустите приложение на чистой базе (см. [Быстрый старт](#быстрый-старт)).
2. В Postman: **Import** → выберите файл коллекции.
3. Выполните **Auth → Register**, затем **Auth → Login**.
4. Скопируйте `token` из ответа Login и вставьте его в переменную коллекции `token` (вкладка **Variables** у коллекции), сохраните.
5. Выполняйте остальные запросы по порядку или запустите всю коллекцию через **Run collection**.

У коллекции две переменные:

| Переменная | Значение |
|---|---|
| `baseUrl` | адрес API, по умолчанию `http://localhost:8080/api` |
| `token` | JWT для запросов с авторизацией, заполняется вручную после Login |

В защищённых запросах токен задан на вкладке **Authorization** (тип **Bearer Token**, значение `{{token}}`)
