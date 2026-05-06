# Stage 1: Builder & Development
FROM golang:alpine AS builder

WORKDIR /app

# Cài đặt các công cụ cần thiết cho development
RUN apk add --no-cache git

# Cài đặt Air để hot reload
RUN go install github.com/air-verse/air@latest

# Copy go.mod và go.sum
COPY go.mod go.sum* ./
RUN go mod download

# Copy toàn bộ mã nguồn
COPY . .

# Chạy migrate và build (dành cho chế độ production)
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# --- Chế độ Development (Dùng Air) ---
FROM builder AS dev
WORKDIR /app
# Trong chế độ dev, Air sẽ quản lý việc chạy app và migrate
CMD ["air", "-c", ".air.toml"]


# --- Chế độ Production (Dùng Alpine nhẹ) ---
FROM alpine:latest AS prod
WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["sh", "-c", "./main -migrate && ./main"]
