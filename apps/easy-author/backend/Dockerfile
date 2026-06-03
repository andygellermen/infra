FROM golang:1.23.4 AS builder

WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /easy-author-backend ./cmd/server

FROM debian:bookworm-slim
WORKDIR /srv/easy-author
COPY --from=builder /easy-author-backend /usr/local/bin/easy-author-backend
EXPOSE 8086
ENV EASY_AUTHOR_ADDR=0.0.0.0:8086
CMD ["easy-author-backend"]
