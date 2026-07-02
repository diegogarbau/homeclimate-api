# ── Build stage ───────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /app

# copiar dependencias primero (aprovecha cache de capas)
# go.su[m] es un glob opcional: el módulo es solo-stdlib y aún no tiene go.sum
COPY go.mod go.su[m] ./
RUN go mod download

# copiar código fuente
COPY . .

# compilar binario estático
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o homeclimate-api ./cmd/server

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM alpine:3.19

WORKDIR /app

# certificados SSL para llamadas HTTPS (Open-Meteo, Nominatim)
RUN apk --no-cache add ca-certificates tzdata

# copiar solo el binario del build stage
COPY --from=builder /app/homeclimate-api .

# usuario no-root por seguridad
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["./homeclimate-api"]