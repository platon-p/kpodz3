FROM golang:1.23.8-alpine AS builder

WORKDIR /app/proto
COPY proto/ .

WORKDIR /app/orders
COPY orders/go.mod orders/go.sum ./
RUN go mod download
COPY orders/ .
RUN go build .

FROM alpine
WORKDIR /app
COPY --from=builder /app/orders/orders .
ENTRYPOINT [ "./orders" ]
