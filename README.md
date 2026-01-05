# GophKeeper (PostgreSQL + CLI)

GophKeeper — учебный менеджер паролей (client-server).
Данные **шифруются на клиенте**, сервер хранит только ciphertext.

## Стек
- Server: Go, chi, zap, viper, JWT, bcrypt, PostgreSQL (`pgx/v5/stdlib`)
- Client: Go CLI (cobra), resty, AES-GCM, scrypt (KDF)
- Tests: testify + sqlmock

## Быстрый старт (Docker Postgres)

```bash
docker compose up -d
```

Создать таблицы:

```bash
psql "postgres://gophkeeper:gophkeeper@localhost:5432/gophkeeper?sslmode=disable" -f ./migrations/001_init.up.sql
```

Запуск сервера:

```bash
go run ./cmd/server -a 127.0.0.1:8080 -dsn "postgres://gophkeeper:gophkeeper@localhost:5432/gophkeeper?sslmode=disable" -jwt dev-secret-change-me
```

Сборка клиента (версия/дата через ldflags):

```bash
go build -o ./bin/gophkeeper -ldflags "-X main.version=1.0.0 -X main.buildDate=2025-12-22" ./cmd/client
./bin/gophkeeper version
```

## CLI сценарий

Регистрация/логин:

```bash
./bin/gophkeeper register --server http://127.0.0.1:8080 --login user1 --password pass1
./bin/gophkeeper login --server http://127.0.0.1:8080 --login user1 --password pass1
```

Добавление данных и синхронизация:

```bash
./bin/gophkeeper add login --title "Github" --username "u" --password "p" --meta "work"
./bin/gophkeeper sync --server http://127.0.0.1:8080
./bin/gophkeeper list
./bin/gophkeeper get --id 1
```

## Тесты и покрытие

```bash
go test ./... -cover
```

> В проекте сделан упор на unit-тесты: crypto, service, repository (sqlmock), middleware/auth.
