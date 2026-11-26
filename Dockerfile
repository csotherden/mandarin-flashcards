# syntax=docker/dockerfile:1.7

FROM golang:1.27 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY ent ./ent

RUN go generate ./ent

RUN CGO_ENABLED=0 go build -o /app/api-server ./cmd/api-server

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl ffmpeg \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd --gid 65532 app \
    && useradd --uid 65532 --gid 65532 --create-home --shell /usr/sbin/nologin app \
    && mkdir -p /var/lib/mandarin-files \
    && chown -R 65532:0 /var/lib/mandarin-files

WORKDIR /app
COPY --from=builder --chown=65532:0 /app/api-server ./api-server

USER 65532

ENV SERVER_HOST=0.0.0.0 \
    SERVER_PORT=8080 \
    FILE_STORE_DIR=/var/lib/mandarin-files

EXPOSE 8080

ENTRYPOINT ["/app/api-server"]
CMD ["-env", "/run/mandarin-secrets/secrets.env"]
