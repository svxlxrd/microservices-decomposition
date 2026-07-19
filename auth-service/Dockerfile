# Этап 1: сборка
FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /auth-service ./cmd/server

# Этап 2: рантайм
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata wget \
    && addgroup -S app \
    && adduser -S -G app app

WORKDIR /app

COPY --from=builder /auth-service .

RUN chown app:app /app/auth-service

USER app

EXPOSE 8081

CMD ["./auth-service"]