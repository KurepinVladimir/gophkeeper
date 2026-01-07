# GophKeeper

GophKeeper — клиент-серверный менеджер секретов, реализованный на Go.
Проект предназначен для безопасного хранения и синхронизации приватных данных пользователя.

Система поддерживает хранение следующих типов данных:

- логины и пароли;
- произвольные текстовые данные;
- произвольные бинарные данные;
- данные банковских карт;
- произвольную текстовую метаинформацию.

---

## Стек технологий

### Server
- Go
- net/http + chi
- PostgreSQL
- pgx/v5/stdlib
- JWT (HMAC)
- bcrypt
- zap
- viper + pflag
- TLS (опционально)

### Client
- Go
- CLI-приложение
- HTTP API

---

## Архитектура

Проект построен с разделением ответственности по слоям:

- HTTP-слой — маршрутизация, middleware, обработка запросов (`internal/http`)
- Service-слой — бизнес-логика (`internal/service`)
- Repository-слой — работа с БД (`internal/repository`)
- Crypto / Security — шифрование данных, JWT, пароли
- Config / Logger — инфраструктурные компоненты

Клиент взаимодействует с сервером исключительно через HTTP API.

---

## Безопасность

- пароли пользователей хранятся в виде bcrypt-хэшей;
- авторизация реализована с использованием JWT;
- все пользовательские данные шифруются на сервере перед сохранением;
- используется envelope encryption:
  - данные шифруются data-key;
  - data-key шифруется master-key;
- поддерживается HTTPS (при наличии TLS-сертификата).

---

## Ограничения

- размер пользовательских данных ограничен на уровне базы данных;
- ограничение применяется к зашифрованным данным (`encrypted_data`);
- максимальный размер зашифрованных данных — **1 МБ**;
- ограничение реализовано через CHECK-constraint:
- при превышении лимита операция сохранения завершается ошибкой.

---

## Быстрый старт (Docker PostgreSQL)

### 1. Запуск PostgreSQL

```bash
docker compose up -d
```

### 2. Применение миграций

```bash
psql "postgres://gophkeeper:gophkeeper@localhost:5432/gophkeeper?sslmode=disable" \
  -f migrations/001_init.up.sql

psql "postgres://gophkeeper:gophkeeper@localhost:5432/gophkeeper?sslmode=disable" \
  -f migrations/002_add_encrypted_data_key.up.sql

psql "postgres://gophkeeper:gophkeeper@localhost:5432/gophkeeper?sslmode=disable" \
  -f migrations/003_limit_encrypted_data_size.up.sql
```

---

## Запуск сервера

```bash
go run ./cmd/server \
  -addr 127.0.0.1:8080 \
  -dsn "postgres://gophkeeper:gophkeeper@localhost:5432/gophkeeper?sslmode=disable" \
  -jwt dev-secret-change-me \
  -master_key BASE64_MASTER_KEY
```

---

## Сборка клиента

```bash
go build -o ./bin/gophkeeper \
  -ldflags "-X main.version=1.0.0 -X main.buildDate=2025-12-22" \
  ./cmd/client
```

Проверка версии:

```bash
./bin/gophkeeper version
```
---

## Сценарии взаимодействия (по ТЗ)

### Новый пользователь

#### 1. Регистрация

```bash
./bin/gophkeeper register \
  --server http://127.0.0.1:8080 \
  --login user1 \
  --password pass1
```

#### 2. Аутентификация

```bash
./bin/gophkeeper login \
  --server http://127.0.0.1:8080 \
  --login user1 \
  --password pass1
```

#### 3. Добавление данных

```bash
./bin/gophkeeper add login \
  --title "Github" \
  --username "u" \
  --password "p" \
  --meta "work"
```

#### 4. Синхронизация

```bash
./bin/gophkeeper sync --server http://127.0.0.1:8080
```

---

### Существующий пользователь

#### 1. Аутентификация

```bash
./bin/gophkeeper login \
  --server http://127.0.0.1:8080 \
  --login user1 \
  --password pass1
```

#### 2. Синхронизация

```bash
./bin/gophkeeper sync --server http://127.0.0.1:8080
```

#### 3. Получение списка данных

```bash
./bin/gophkeeper list
```

#### 4. Получение конкретного секрета

```bash
./bin/gophkeeper get --id 1
```

---

## Тестирование

Запуск всех тестов:

```bash
go test ./...
```

Покрытие серверной части:

```bash
go test ./internal/... -cover
```

---

## Документация (GoDoc)

```bash
go install golang.org/x/tools/cmd/godoc@latest
godoc -http=:6060
```

Открыть в браузере:

```
http://localhost:6060
```

Пакеты в `internal/` скрыты по правилам Go,
но доступны по прямым ссылкам, например:

```
http://localhost:6060/pkg/gophkeeper/internal/service/
```

---
