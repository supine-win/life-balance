# 多阶段构建：Go 网关服务
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY gateway/go.mod gateway/go.sum ./
RUN go mod download

COPY gateway/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /gateway ./cmd/server

# 运行阶段
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata ffmpeg
WORKDIR /app

COPY --from=builder /gateway .
COPY gateway/configs/ ./configs/
COPY gateway/migrations/ ./migrations/

# 创建数据目录
RUN mkdir -p /app/data/uploads /app/data/logs

EXPOSE 8080

CMD ["./gateway"]
