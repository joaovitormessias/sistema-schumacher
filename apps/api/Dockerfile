FROM golang:1.23-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/schumacher-api ./cmd/api

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY entrypoint.sh /app/entrypoint.sh
COPY --from=builder /out/schumacher-api /app/schumacher-api

RUN chmod +x /app/entrypoint.sh /app/schumacher-api

ENV PORT=8080

EXPOSE 8080

ENTRYPOINT ["/app/entrypoint.sh"]
