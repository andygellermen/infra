# Incident Response Runbook

## Severity

### SEV-1

möglicher Mailverlust, Accountübernahme in Breite, falscher Massensend,
großer Inbound-/Outbound-Ausfall.

### SEV-2

Teilfunktion stark beeinträchtigt, Queue wächst, Attachment Store
gestört.

### SEV-3

begrenzte Degradation ohne Datenverlust.

## Ablauf

``` text
Detect
 -> classify
 -> contain
 -> preserve evidence
 -> recover
 -> verify
 -> communicate
 -> postmortem
```

## Sicherheitsregel

Bei unklarem Send-Zustand niemals blind Queue replayen.

## Postmortem

Blameless, aber technisch konkret: - Timeline - Trigger - fehlende
Guardrails - Recovery - dauerhafte Maßnahmen - Test, der Wiederholung
verhindert
