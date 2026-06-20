# Zentrale MFA vor WordPress-Admin

## Empfehlung

Für euren aktuellen Stack ist **Authelia** der passendste erste Kandidat für zentrales MFA vor allen WordPress-Admin-Routen.

Warum in eurem Fall:

- passt sehr gut zu **Traefik ForwardAuth**
- schützt **zentral vor WordPress**, statt Plugin-Logik in jede Instanz zu drücken
- deckt **TOTP** und **WebAuthn/Passkeys** ab
- ist leichtergewichtig als ein vollständiger IdP-Stack
- lässt sich schrittweise nur für `wp-login.php` und `/wp-admin` ausrollen

## Wann eher authentik?

`authentik` ist die stärkere Wahl, wenn ihr mittelfristig deutlich mehr als WordPress zentralisieren wollt:

- SSO für viele unterschiedliche Apps
- umfangreichere Flows und Self-Service
- zentrales IdP-/Directory-Gefühl
- stärkere Portal-/Benutzerverwaltungs-Anforderungen

Für **„WordPress-Admin zentral mit MFA absichern“** bleibt Authelia im Moment der pragmatischere Start.

## Zielbild

- öffentliche Website bleibt normal erreichbar
- nur `wp-login.php` und `/wp-admin` laufen zusätzlich durch ForwardAuth
- CrowdSec bleibt vorgeschaltet
- WordPress selbst behält seine eigenen Rollen und Passwörter

## Migrationsplan

### 1. Pilot nur für eine Instanz

- Authelia auf eigener Subdomain deployen, z. B. `auth.geller.men`
- ForwardAuth-Middleware in Traefik anlegen
- zunächst nur eine WordPress-Testinstanz an die Admin-Middleware hängen

### 2. Admin-Routen schützen

- `wp-login.php`
- `/wp-admin`

Nicht im ersten Schritt:

- komplettes Frontend
- `/wp-json`
- Webhooks oder externe Integrationen

## 3. Fallback sauber vorbereiten

- mindestens ein Break-Glass-Admin-Konto dokumentieren
- Recovery-Codes sicher ablegen
- für den Rollout-Zeitraum Passwortschutz/Maintenance nur bei Bedarf zusätzlich aktiv lassen

### 4. Rollout über alle WordPress-Instanzen

Das bestehende Modell mit `wp_traefik_middleware_admin` ist dafür bereits passend vorbereitet.

Empfohlene Reihenfolge:

- Pilot validieren
- zwei bis drei weitere Instanzen umstellen
- danach flächig ausrollen

### 5. Danach erst schärfer ziehen

Optional später:

- Geo/IP-Regeln für Admin-Zugänge
- zusätzliche Basic-Auth für Wartungsfenster
- feinere Policies pro Benutzergruppe

## Was das nicht ersetzt

Zentrales MFA ersetzt nicht:

- WordPress-Core-/Plugin-/Theme-Updates
- CrowdSec
- Wazuh/FIM
- starke individuelle Passwörter
- Salt-/Secret-Rotation nach Incidents
