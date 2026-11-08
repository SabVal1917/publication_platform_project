FROM golang:1.27.0-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG GO_BUILD_TAGS=""
RUN go build -mod=readonly -tags="${GO_BUILD_TAGS}" -o /out/api ./cmd/api
RUN go build -mod=readonly -o /out/migrate ./cmd/migrate
RUN go build -mod=readonly -o /out/prober ./cmd/prober

FROM alpine:latest

WORKDIR /app

COPY --from=builder /out/api ./api
COPY --from=builder /out/migrate ./migrate
COPY --from=builder /out/prober ./prober
COPY --from=builder /app/migrations ./migrations

CMD ["./api"]
