FROM golang:1.23.1-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build .

FROM alpine
WORKDIR /app
COPY --from=builder /app/payments /app/payments
COPY config.json /app/config.json
ENTRYPOINT ["./payments"]
EXPOSE $SERVER_PORT