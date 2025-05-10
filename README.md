# Распределённый вычислитель арифметических выражений

Проект представляет собой распределенный калькулятор для обработки арифметических выражений. Система состоит из трех микросервисов:
- **Сервис аутентификации (Auth)** - отвечает за регистрацию и авторизацию пользователей
- **Оркестратор** - управляет вычислениями и координирует работу агентов
- **API Gateway** - обрабатывает HTTP запросы и проксирует их к соответствующим микросервисам

## Стек технологий
- Go 1.24
- PostgreSQL
- gRPC
- Docker и Docker Compose
- Nginx
- JWT для аутентификации

## Установка и запуск

### Предварительные требования
- Docker и Docker Compose

### Запуск
```bash
# Клонирование репозитория
git clone https://github.com/flexer2006/y.lms-final-task-calc-go.git
cd y.lms-final-task-calc-go

# Копирование конфигурационного файла
cp .env.example deploy/.env

# Генерация самоподписанных SSL-сертификатов (опционально)
chmod +x ./scripts/gen-ssl.sh
./scripts/gen-ssl.sh

# Запуск сервисов
cd deploy
docker-compose up -d
```

После успешного запуска приложение будет доступно по адресам:
- HTTP: http://localhost/
- HTTPS: https://localhost/ (при использовании SSL)

## API Endpoints

### Аутентификация

#### Регистрация пользователя
```bash
curl --location 'http://localhost/api/v1/auth/register' \
  --header 'Content-Type: application/json' \
  --data '{
    "email": "user@example.com",
    "password": "password123",
    "name": "Тестовый Пользователь"
  }'
```

#### Вход пользователя
```bash
curl --location 'http://localhost/api/v1/auth/login' \
  --header 'Content-Type: application/json' \
  --data '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### Калькулятор

#### Создание вычисления
```bash
curl --location 'http://localhost/api/v1/calculations' \
  --header 'Content-Type: application/json' \
  --header 'Authorization: Bearer YOUR_TOKEN' \
  --data '{
    "expression": "2+2*2"
  }'
```

#### Получение списка вычислений
```bash
curl --location 'http://localhost/api/v1/calculations' \
  --header 'Authorization: Bearer YOUR_TOKEN'
```

#### Получение результата вычисления по ID
```bash
curl --location 'http://localhost/api/v1/calculations/{id}' \
  --header 'Authorization: Bearer YOUR_TOKEN'
```

### Проверка работоспособности сервисов

#### Проверка API Gateway
```bash
curl --location 'http://localhost/health'
```

#### Проверка сервиса авторизации
```bash
curl --location 'http://localhost/api/v1/auth/health'
```

#### Проверка сервиса оркестрации
```bash
curl --location 'http://localhost/api/v1/calculations/health'
```

## Тестирование

Для запуска всех тестов:
```bash
go test ./...
```

Для запуска тестов с покрытием:
```bash
go test -cover ./...
```

## Структура проекта

- cmd - точки входа для различных сервисов
- internal - внутренние пакеты приложения
- pkg - многоразовые библиотеки
- proto - определения Protocol Buffers
- migrations - миграции баз данных
- deploy - файлы для развертывания
- build - скрипты сборки и Dockerfile'ы

## Лицензия
MIT

ㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤㅤ
ㅤ
ㅤ
ㅤ

ㅤ
ㅤ
ㅤ
ㅤ
ㅤ
ㅤ
ㅤ

ㅤ
ㅤ

Послесловие: я очень сильно болею, сорян за качество проекта, хотел запотеть, но здоровье очень сильно подвело(((. Получилость как получилось, все свои ошибки сам знаю и вижу, как поправлюсь, полностью зарефакторю кор, вспомогательное и будет конфетка. *-*
