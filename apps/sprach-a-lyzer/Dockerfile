FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/sprach-a-lyzer ./cmd/server
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/sprach-a-lyzer-migrate ./cmd/migrate
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/sprach-a-lyzer-seed ./cmd/seed
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/sprach-a-lyzer-analyze ./cmd/analyze
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/sprach-a-lyzer-normalize-dimensions ./cmd/normalize-dimensions

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /out/ /usr/local/bin/
COPY data/ ./data/

ENV SAL_ENV=production
ENV SAL_HTTP_ADDR=:8080

EXPOSE 8080

USER 65532:65532

CMD ["sprach-a-lyzer"]
