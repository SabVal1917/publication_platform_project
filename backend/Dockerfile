FROM golang:1.27.0-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -mod=readonly -o /out/api ./cmd/api
RUN go build -mod=readonly -o /out/migrate ./cmd/migrate

FROM alpine:latest

WORKDIR /app

COPY --from=builder /out/api ./api
COPY --from=builder /out/migrate ./migrate
COPY --from=builder /app/migrations ./migrations

CMD ["./api"]
