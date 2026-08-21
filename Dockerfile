# =========================================================
#  Production Dockerfile — multi-stage build
# =========================================================

# ---------- Stage 1: build ----------
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Кэшируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Собираем статический бинарник
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /app/server .

# ---------- Stage 2: runtime ----------
FROM alpine:3.20

# CA-сертификаты для HTTPS-запросов (если понадобятся)
RUN apk add --no-cache ca-certificates tzdata

# Создаём непривилегированного пользователя
RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

# Копируем бинарник
COPY --from=builder /app/server /app/server

# Копируем UI-шаблоны и статику
COPY --from=builder /app/ui /app/ui

# Директории для данных
RUN mkdir -p /app/data /app/firmwares /app/flashers \
    && chown -R app:app /app

# Переключаемся на непривилегированного пользователя
USER app

# Порт по умолчанию
EXPOSE 8200

# Точка входа — сервер
ENTRYPOINT ["/app/server"]