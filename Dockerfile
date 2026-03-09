# Go Quiz — Dockerfile
# Multi-stage build для минимального размера образа

# === Stage 1: Build ===
FROM golang:1.21-alpine AS builder

# Установка зависимостей
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Рабочая директория
WORKDIR /app

# Копирование go.mod и go.sum
COPY go.mod go.sum ./

# Загрузка зависимостей
RUN go mod download

# Копирование исходного кода
COPY . .

# Компиляция приложения
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o qwen_test .

# === Stage 2: Run ===
FROM alpine:latest

# Установка зависимостей для runtime
RUN apk add --no-cache ca-certificates sqlite

# Создание пользователя для безопасности
RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -s /bin/sh -D appuser

# Рабочая директория
WORKDIR /app

# Копирование бинарника из builder stage
COPY --from=builder /app/qwen_test .

# Копирование статики и шаблонов
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/questions.json .

# Создание директорий для данных
RUN mkdir -p /app/data /app/backups && \
    chown -R appuser:appgroup /app

# Переключение на не-root пользователя
USER appuser

# Порт приложения
EXPOSE 8080

# Переменные окружения
ENV PORT=:8080
ENV DB_PATH=/app/data/qwen_test.db
ENV BACKUP_DIR=/app/backups

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/stats || exit 1

# Запуск приложения
CMD ["./qwen_test"]
