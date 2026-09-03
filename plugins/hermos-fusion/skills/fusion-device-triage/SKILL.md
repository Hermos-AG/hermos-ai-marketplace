---
name: fusion-device-triage
description: Fehlersuche an einem einzelnen Fusion Edge-Gerät, wenn es nicht reagiert, Container hängen oder ein Transfer klemmt. Nutzen bei "Gerät reagiert nicht", "Device offline", "Container läuft nicht", "Transfer hängt", "warum antwortet X nicht", "device triage", "Störung an Anlage".
---

# Gerät eingrenzen

## Reihenfolge

Diese Reihenfolge einhalten – sie geht von der breiten Diagnose zur teuren Einzelabfrage.

1. **Gerät identifizieren.** Nennt die Person einen Namen statt einer ID, mit
   `list_devices` auflösen. Bei mehreren Treffern nachfragen statt raten.
2. **`get_device_diagnostics`** – der Einstieg. Liefert Zustand, `lastSeenAt`,
   Broker-Lage inklusive Queue-Tiefen, letzte Befehle und Transfers, SignalR.
   Hinweis: `totalMessagesReady` ist ohne Admin-Rolle `null`. Das heisst **nicht**,
   dass der Broker steht.
3. **`trace_device`** – Rundlauf API → Broker → Agent → zurück, mit Latenz.
   Timeout hier bei gesundem Eintrag aus Schritt 2 deutet auf die Agent-Strecke.
4. Erst danach vertiefen, je nach Befund:
   - Container: `list_live_containers` gegen `list_containers` (Soll gegen Ist),
     bei Auffälligkeit `get_container_logs`.
   - Transfers: `list_device_transfers`, auf `chunksAcked`/`chunksTotal` achten.
   - Last: `get_device_performance` für die letzten 48 Stunden roh,
     `get_device_telemetry` für den längeren Verlauf.
   - Bereits bekannt? `list_sentinel_findings`/`get_sentinel_finding` zeigen, ob
     für dieses Gerät schon eine Sentinel-Regel ausgelöst hat, bevor man selbst
     danach sucht; `get_device_uptime` zeigt die Neustart-Historie direkt und
     beantwortet damit, ob das Gerät in einer Boot-Schleife steckt. Ein 404 dort
     heisst nicht „Gerät gibt es nicht": es deckt auch „ausserhalb deiner
     Sichtbarkeit" und „sichtbar, aber Sentinel hat noch keine Uptime erfasst"
     ab. Also melden, dass keine Uptime-Daten vorliegen — nicht, dass das Gerät
     fehlt.

## Deutung

| Befund | Wahrscheinliche Ursache |
|--------|-------------------------|
| `lastSeenAt` alt, Queue wächst | Agent weg, Nachrichten stauen sich |
| Queue-Tiefe steigt, keine Consumer | Der konsumierende Dienst ist unten |
| `trace_device` Timeout, Diagnose sonst sauber | Agent-Strecke oder Gerät selbst |
| `updateRequired: true` | Agent unter `minAgentVersion` – erst aktualisieren, dann weitersuchen |

## Regeln

- **Nur lesen.** `container_action`, `send_device_command`, `delete_device`,
  `remove_device_image`, `abort_device_transfer`, `acknowledge_sentinel_finding`,
  `suppress_sentinel_finding` und alles Schreibende erst nach ausdrücklicher
  Zustimmung – und vorher benennen, was passieren wird.
- `run_sql_query` sieht alle Mandanten und braucht Admin. Für Fragen zu einem Gerät
  immer die dedizierten Werkzeuge nehmen, nie SQL.
- Befund und Vermutung trennen. "Queue bei 1.240, keine Consumer" ist ein Befund.
  "Der Worker ist abgestürzt" ist eine Vermutung – als solche kennzeichnen.
- Am Ende ein konkreter nächster Schritt, keine Liste von Möglichkeiten.