# Этап 1: сборка
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Git нужен для загрузки некоторых Go-модулей
RUN apk add --no-cache git

# Копируем зависимости отдельно для кэширования
COPY go.mod go.sum ./
RUN go mod download

# Исходный код
COPY . .

# Сборка статического бинарника
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /books-service ./cmd/server

# Этап 2: рантайм
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata wget \
    && addgroup -S app \
    && adduser -S -G app app

WORKDIR /app

COPY --from=builder --chown=app:app /books-service .

USER app

EXPOSE 8082

CMD ["./books-service"]