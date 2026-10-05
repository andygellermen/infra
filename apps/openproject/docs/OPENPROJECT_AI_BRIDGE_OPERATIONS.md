# OpenProject AI Bridge im Infra-Stack betreiben

Die Lifecycle-Skripte verwalten die separate private Bridge-Installation auf dem
Infra-Host. Sie speichern keine Bridge-Konfiguration oder Secrets im Infra-Repo.
Standardmäßig verwenden sie:

- Bridge-Checkout: `/home/andy/openproject-ai-bridge`
- private Laufzeitdaten: `/srv/openproject-ai-bridge`
- Compose-Datei: `deploy/compose/compose.yml` im Bridge-Checkout
- Compose-Umgebung: `/srv/openproject-ai-bridge/compose.env`

Alle mutierenden Befehle und die Statusprüfung verwenden dieselbe lokale Sperre.
Parallel gestartete Lifecycle-Aktionen werden abgewiesen.

## Bedienung

```bash
./scripts/opai-bridge-start.sh
./scripts/opai-bridge-status.sh
./scripts/opai-bridge-stop.sh
./scripts/opai-bridge-redeploy.sh
./scripts/opai-bridge-update.sh
```

`start` validiert zuerst das gerenderte Compose-Modell und die erforderlichen
Secret-Dateien. Danach startet es Bridge, Cody-Tunnel und Chad-Tunnel und wartet
auf den Healthcheck aller drei Dienste.

`stop` stoppt alle drei Dienste, entfernt aber weder Container noch Netzwerke
oder das persistente Audit-Volume. Ein anschließendes `start` nimmt denselben
Laufzeitstand wieder auf.

`status` zeigt den Compose-Status und den Healthzustand jedes Dienstes. Der
Befehl endet mit einem Fehlercode, sobald ein Dienst fehlt oder nicht `healthy`
ist. Für eine gezielte Fehleranalyse bleiben die redigierten Compose-Logs auf
dem Host verfügbar.

`redeploy` prüft die Laufzeitkonfiguration, baut das Bridge-Image mit aktualisierter
Builder-Basis neu, lädt die beiden fest gepinnten Tunnel-Images und ersetzt die
Dienste erst nach einem erfolgreichen Build. Das Audit-Volume bleibt erhalten.

`update` verweigert die Ausführung bei lokalen Änderungen im Bridge-Checkout,
führt dort `git pull --ff-only` aus und startet anschließend denselben geprüften
Redeploy-Ablauf. Dadurch werden weder lokale Änderungen zusammengeführt noch
History umgeschrieben. OpenProject-Core-Updates werden weiterhin ausschließlich
mit `op-update.sh` durchgeführt; anschließend prüft `opai-bridge-redeploy.sh`,
ob die Bridge mit der aktualisierten OpenProject API v3 betriebsbereit ist.

## Abweichende Pfade und Timeout

Für Testsysteme können die Standardwerte pro Aufruf überschrieben werden:

```bash
OPAI_BRIDGE_REPO_DIR=/opt/openproject-ai-bridge \
OPAI_BRIDGE_RUNTIME_DIR=/secure/openproject-ai-bridge \
OPAI_BRIDGE_HEALTH_TIMEOUT=240 \
./scripts/opai-bridge-redeploy.sh
```

Zusätzlich kann `OPAI_BRIDGE_LOCK_FILE` einen abweichenden absoluten Pfad für die
lokale Prozesssperre setzen. Auf dem regulären Infra-Host sollten die Standardpfade
beibehalten werden.

## Update-Reihenfolge mit OpenProject

Bei einem regulären OpenProject-Core-Update:

1. OpenProject mit `sudo ./scripts/op-update.sh --domain=<domain> --version=<version>`
   aktualisieren und dessen Bereitschaftsprüfung abwarten.
2. Mit `./scripts/opai-bridge-redeploy.sh` die vorhandene Bridge-Version gegen
   die aktualisierte API starten und alle Healthchecks prüfen.
3. Wenn das Bridge-Repository eine dazugehörige Anpassung enthält, stattdessen
   `./scripts/opai-bridge-update.sh` verwenden.
4. Mit `./scripts/opai-bridge-status.sh` den Zustand von Bridge und Tunneln
   abschließend kontrollieren.

Die Skripte geben keine Secret-Inhalte und keine Containerlogs automatisch aus.
Fehlschläge liefern einen Fehlercode und nennen nur den betroffenen Dienst oder
die fehlgeschlagene Werkzeugklasse.
