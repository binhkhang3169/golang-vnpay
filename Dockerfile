# Stage 1: Build binary
FROM golang:1.24-bookworm AS builder

WORKDIR /app

RUN apt-get update && apt-get install -y librdkafka-dev pkg-config build-essential

COPY go.mod go.sum ./ 
RUN go mod tidy

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o payment_service ./cmd/server

# Stage 2: Runtime
FROM debian:bookworm-slim

# Thiết lập múi giờ của Việt Nam
RUN ln -sf /usr/share/zoneinfo/Asia/Ho_Chi_Minh /etc/localtime && \
    echo "Asia/Ho_Chi_Minh" > /etc/timezone

WORKDIR /app

# Copy binary từ builder
COPY --from=builder /app/payment_service .

EXPOSE 8080

ENTRYPOINT ["./payment_service"]
