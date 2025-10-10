# Fichier docker de mon application Go

# Étape build
FROM golang:1.22 AS build
WORKDIR /app

# Copie uniquement go.mod (go.sum peut ne pas exister au début)
COPY go.mod ./
RUN go mod download

# Copie le reste du code
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app

# Étape runtime minimaliste
FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=build /app/app /app/app
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/app"]
