FROM golang:1.25 as builder

WORKDIR /app

COPY . .

RUN go build -o ./bin/kvapp ./cmd/kvstore/main.go

FROM alpine:latest

COPY --from=builder /bin/kvapp /bin/kvapp 

CMD ["./bin/kvapp"]