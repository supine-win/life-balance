# MCP Server
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY mcp-server/go.mod mcp-server/go.sum ./
RUN go mod download
COPY mcp-server/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /mcp-server .

FROM alpine:3.19
COPY --from=builder /mcp-server .
EXPOSE 8082
CMD ["./mcp-server"]
