---
name: fusion-fleet-report
description: Erstellt einen Flottenüberblick über alle Fusion Edge-Geräte – wer ist online, wer hängt bei der Agent-Version hinterher, wie verteilt sich der Bestand über die Org-Units. Nutzen bei Fragen wie "Wie steht die Flotte da", "Welche Geräte sind offline", "Wo müssen Agents aktualisiert werden", "Geräteübersicht", "fleet report", "welche Geräte hat TNT".
---

# Flottenbericht

## Vorgehen

1. `list_devices` mit `pageSize: 200` aufrufen. Bei `hasMore: true` weiterblättern
   (`page: 2`, `3`, …), bis alle Geräte geladen sind. `totalCount` gegen die Anzahl
   eingesammelter Einträge prüfen und die Zahl im Bericht nennen.
2. Je Gerät auswerten:
   - **Erreichbarkeit** über `lastSeenAt`. Schwellen: unter 15 Minuten = online,
     unter 24 Stunden = still, älter = offline. Immer das absolute Datum mitgeben,
     nicht nur "vor 3 Tagen".
   - **Agent-Stand** über `agentVersion`, `updateAvailable`, `updateRequired`.
     `updateRequired: true` bedeutet, der Agent liegt unter `minAgentVersion` –
     das ist ein Handlungspunkt, kein Hinweis.
   - **Zuordnung** über `orgUnitName` und `tags`.
3. Gruppieren nach Org-Unit. Innerhalb der Gruppe die Problemfälle nach oben.
4. Wird zusätzlich nach Sentinel-Befunden gefragt: `list_sentinel_findings`
   liefert offene Findings fleet-weit, ergänzend zu den drei Zahlen oben – nicht
   automatisch, nur wenn die Frage danach verlangt.
5. Wird zusätzlich nach unregistrierten Geräten gefragt: `list_discovered_devices`
   liefert den Bestand, der aktuell sendet, aber noch nicht in `list_devices`
   auftaucht – ebenfalls nur auf Nachfrage, nicht automatisch. Eine Übernahme
   läuft über den Skill `fusion-device-triage` (`adopt_discovered_device`),
   nicht hier.
6. Wird gefragt, ob Alarme auch ankommen: `list_notification_channels` zeigt die
   Benachrichtigungskanäle des Mandanten (E-Mail, Teams, Webhook) mit letztem
   Erfolg und letztem Fehler, `list_notification_outbox` die Zustellhistorie
   (`Pending`/`Sent`/`Failed`/`Dropped`; `Dropped` = absichtlich entprellt).
   Regeln anlegen (`create_notification_rule`) oder eine Ad-hoc-Nachricht
   senden (`send_notification`) nur auf ausdrückliche Anweisung: echte
   Menschen erhalten sie, Wortlaut vorher bestätigen lassen.

## Ausgabeformat

Kurzer Vorspann mit drei Zahlen: Gesamtbestand, davon offline, davon Agent-Update
erforderlich. Danach eine Tabelle je Org-Unit:

| Gerät | Zustand | Agent | Zuletzt gesehen | Tags |

Am Schluss ein Abschnitt "Handlungsbedarf" – nur Geräte mit `updateRequired: true`
oder mehr als 24 Stunden ohne Lebenszeichen. Ist die Liste leer, das ausdrücklich
sagen statt den Abschnitt wegzulassen.

## Regeln

- Zeitstempel aus Fusion sind UTC. Für Leser in DE/CH nach lokaler Zeit umrechnen
  und die Zone dazuschreiben.
- Keine Zahlen schätzen. Fehlende Felder (`deviceType`, `location`, `hostname` sind
  oft leer) als "–" ausweisen, nicht erfinden.
- `hasInternet: null` heisst "nie geprüft", nicht "kein Internet". Unterscheiden.
- Nur lesen. Dieser Skill ruft nie ein Werkzeug auf, das etwas verändert.