FROM golang:1.23.4-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/static-inline-editor ./cmd/server

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /out/static-inline-editor /usr/local/bin/static-inline-editor

ENV STATIC_EDITOR_ADDR=:8080
ENV STATIC_EDITOR_DATA_DIR=/data
ENV STATIC_EDITOR_TENANT_DIR=/config/tenants
ENV STATIC_EDITOR_SESSION_TTL=12h
ENV STATIC_EDITOR_MAGIC_LINK_TTL=15m
ENV STATIC_EDITOR_GIT_AUTHOR_NAME=Static Inline Editor

VOLUME ["/data", "/config"]

EXPOSE 8080

CMD ["static-inline-editor"]
