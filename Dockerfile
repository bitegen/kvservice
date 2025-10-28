FROM golang:1.25 as builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build -ldflags="-s -w" -o ./bin/kvapp ./cmd/kvstore/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/kvapp ./bin/kvapp
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/migrations ./migrations

CMD ["./bin/kvapp"]
