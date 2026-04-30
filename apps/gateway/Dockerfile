FROM golang:1.26.2-alpine AS builder

WORKDIR /workspace

COPY apps/gateway/go.mod apps/gateway/go.sum ./apps/gateway/
WORKDIR /workspace/apps/gateway
RUN go mod download

WORKDIR /workspace
COPY apps/gateway ./apps/gateway
WORKDIR /workspace/apps/gateway
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/bifrost-gateway ./cmd/bifrost-gateway
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/bifrost-migrate ./cmd/bifrost-migrate

FROM alpine:3.22

RUN apk add --no-cache ca-certificates wget

WORKDIR /app
COPY --from=builder /out/bifrost-gateway ./bifrost-gateway
COPY --from=builder /out/bifrost-migrate ./bifrost-migrate

EXPOSE 8080

ENTRYPOINT ["./bifrost-gateway"]
