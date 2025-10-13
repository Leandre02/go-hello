# -------- Étape 1 : build
FROM golang:1.23-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git ca-certificates

# Cache deps
COPY go.mod go.sum ./
RUN go mod download

# Code (uniquement src/ pour rester propre)
COPY src ./src

# Build binaire statique
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" \
    -o /out/monitoring ./src/cmd/server/main.go

# -------- Étape 2 : image finale minuscule
FROM scratch
# Certifs HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Binaire
COPY --from=builder /out/monitoring /monitoring
# Front & SQL (si utiles)
COPY src/web/ /web/
COPY src/database/init.sql /migrations/10_init.sql
COPY src/database/dbtrigger.sql /migrations/20_trigger.sql

EXPOSE 8080
ENTRYPOINT ["/monitoring"]
