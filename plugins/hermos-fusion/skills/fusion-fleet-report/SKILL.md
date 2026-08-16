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