FROM golang:1.23.4-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/sheet-helper ./cmd/server

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /out/sheet-helper /usr/local/bin/sheet-helper
COPY testdata ./testdata

ENV SHEET_HELPER_ADDR=:8080
ENV SHEET_HELPER_DB_PATH=/data/sheet-helper.db
ENV SHEET_HELPER_SEED_FILE=/app/testdata/seed.json
ENV SHEET_HELPER_COOKIE_SECRET=change-me

VOLUME ["/data"]

EXPOSE 8080

CMD ["sheet-helper"]
