FROM golang:1.23.8-alpine AS builder

WORKDIR /app/proto
COPY proto/ .

WORKDIR /app/payments
COPY payments/go.mod payments/go.sum ./
RUN go mod download
COPY payments/ .
RUN go build .

FROM alpine
WORKDIR /app
COPY --from=builder /app/payments/payments .
ENTRYPOINT [ "./payments" ]