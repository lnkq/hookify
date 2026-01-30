FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o hookify ./cmd/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/hookify .
COPY --from=builder /app/migrations ./migrations

EXPOSE 50051

CMD ["./hookify"]
