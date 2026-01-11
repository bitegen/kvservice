# kvservice

## Описание
HTTP‑сервис реализует key‑value хранилище с поддержкой операций PUT, GET и DELETE.
Для сохранения состояния доступны два варианта: запись в обычный текстовый файл или использование базы данных PostgreSQL.

## Переменные окружения 
Необходимо определить следующие переменные окружения:
```
POSTGRES_DB=
POSTGRES_HOST=
POSTGRES_USER=
POSTGRES_PASSWORD=
POSTGRES_PORT=
CONFIG_PATH= // yaml файлики в /internal/configs
```

## Запуск приложения через docker-compose
```
docker compose up -d
```