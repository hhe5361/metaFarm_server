FROM golang:1.24-alpine3.21 AS builder

WORKDIR /app

ENV CGO_ENABLED=1

RUN apk add --no-cache \
    gcc \
    musl-dev

COPY . .
RUN go build -o metafarm cmd/metafarm/main.go

FROM alpine:3.21

EXPOSE 8080

WORKDIR /app
COPY --from=builder /app/metafarm .

CMD ["./metafarm"]
