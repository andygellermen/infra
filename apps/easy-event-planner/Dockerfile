FROM golang:1.23.4-alpine AS builder

WORKDIR /src

ARG EEP_VERSION=dev

COPY go.mod ./
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -trimpath -ldflags="-s -w -X main.buildVersion=${EEP_VERSION}" \
  -o /out/easy-event-planner ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -trimpath -ldflags="-s -w -X main.buildVersion=${EEP_VERSION}" \
  -o /out/easy-event-planner-migrate ./cmd/migrate
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -trimpath -ldflags="-s -w -X main.buildVersion=${EEP_VERSION}" \
  -o /out/easy-event-planner-seed ./cmd/seed
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -trimpath -ldflags="-s -w -X main.buildVersion=${EEP_VERSION}" \
  -o /out/easy-event-planner-worker ./cmd/worker

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /out/easy-event-planner /usr/local/bin/easy-event-planner
COPY --from=builder /out/easy-event-planner-migrate /usr/local/bin/easy-event-planner-migrate
COPY --from=builder /out/easy-event-planner-seed /usr/local/bin/easy-event-planner-seed
COPY --from=builder /out/easy-event-planner-worker /usr/local/bin/easy-event-planner-worker

ENV EEP_ENV=production
ENV EEP_HTTP_ADDR=:8080
ENV EEP_BASE_URL=https://events.example.com
ENV EEP_DB_DRIVER=sqlite
ENV EEP_DB_PATH=/data/easy-event-planner.sqlite

VOLUME ["/data"]

EXPOSE 8080

CMD ["easy-event-planner"]
