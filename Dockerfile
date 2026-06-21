# syntax=docker/dockerfile:1

# ============================================================
# Stage 1 — build a static Go binary (no CGO: pgx/bcrypt/excelize are pure Go)
# ============================================================
FROM golang:1.26-alpine AS build

# git is needed by `go mod download` for some modules
RUN apk add --no-cache git

WORKDIR /src

# Cache dependencies first for faster rebuilds
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server .

# ============================================================
# Stage 2 — minimal runtime image
# ============================================================
FROM alpine:3.20

# ca-certificates: TLS to Supabase (sslmode=require) and payment gateways.
# tzdata: correct timestamps. Then create an unprivileged user.
RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 10001 appuser

WORKDIR /app
COPY --from=build /out/server /app/server

USER appuser

# Informational; Railway routes traffic to the PORT it injects at runtime.
EXPOSE 8080

ENTRYPOINT ["/app/server"]
