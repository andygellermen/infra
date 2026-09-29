# Runtime / Docker Topology v1

## Container

``` text
traefik
easypeasy-api
easypeasy-worker
easypeasy-web
easypeasy-db
easypeasy-object-adapter
```

Inbound kann je nach Edge als Worker/Adapter laufen; eigener
SMTP-Receiver wäre separater Container.

## Netzwerk

``` text
Internet
  |
Traefik :443
  |
  +--> web
  +--> api

SES/Event ingress
  |
  +--> worker/api ingress endpoint
          |
          +--> DB
          +--> Object Store
          +--> SES outbound
```

Traefik terminiert HTTP(S), nicht SMTP.

## Worker Jobs

-   inbound parse
-   attachment verify/scan
-   newsletter classify
-   scheduled send
-   send retry
-   retention/purge
-   snapshot generation
-   integrity sweep

## PostgreSQL

Serverkanon. SQLite bleibt Clientcache.

## Object Store

Adapter abstrahiert S3-kompatiblen bzw. anderen Store.

## Deployment

Ansible: - env/secrets provisioning - Docker network - migrations -
health checks - backup hooks

Portainer bleibt Betriebsoberfläche, nicht Konfigurationswahrheit.
