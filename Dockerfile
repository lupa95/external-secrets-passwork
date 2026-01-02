ARG APP_VERSION=dev
ARG GO_VERSION=1.23.5
ARG ALPINE_VERSION=3.23.2

# Build Stage
FROM golang:${GO_VERSION}-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .

RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build -o external-secrets-passwork main.go

# Run Stage
FROM alpine:${ALPINE_VERSION} AS runtime

ARG APP_VERSION

# User
RUN addgroup -S app && adduser -S -G app -u 1000 app
USER app

# Workdir
WORKDIR /app
COPY --from=builder /app/external-secrets-passwork .

EXPOSE 8080

LABEL org.opencontainers.image.title="external-secrets-passwork"
LABEL org.opencontainers.image.description="Proxy to fetch Passwork passwords"
LABEL org.opencontainers.image.version="${APP_VERSION}"
LABEL org.opencontainers.image.source="https://github.com/lupa95/external-secrets-passwork"

CMD ["./external-secrets-passwork"]
