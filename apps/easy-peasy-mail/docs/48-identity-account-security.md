# Identity & Account Security

## MVP Baseline

-   Passkey/WebAuthn bevorzugt
-   MFA-Fallback
-   sichere Recovery Codes
-   Device Registry
-   Session Registry
-   Device Revocation
-   Login Rate Limits
-   verdächtige Session-Erkennung
-   Re-Authentication für sensible Aktionen

## Device Trust

Geräte sind explizite Account-Entitäten.

``` text
ACTIVE
REVOKED
```

Revocation beendet Sessions/Tokens und verhindert neue
Sync-/Vault-Capabilities.

## Recovery

Recovery darf nicht allein vom Zugriff auf dieselbe
EasyPeasyMail-Adresse abhängen.

Mögliche Faktoren: - Recovery Code - zweiter Passkey - bestätigtes
Gerät - separate Recovery-Adresse - administrativer Prozess bei
Business-Tenants

## Kritische Aktionen

Erneute Authentisierung bei: - Passwort/Passkey ändern - Recovery
ändern - Gerät entfernen - Weiterleitung aktivieren - Alias/Domain
ändern - Mass Export - Account Closure
