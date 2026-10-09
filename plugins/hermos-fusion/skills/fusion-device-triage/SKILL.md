---
name: fusion-device-triage
description: Fehlersuche an einem einzelnen Fusion Edge-Gerät, wenn es nicht reagiert, Container hängen oder ein Transfer klemmt. Nutzen bei "Gerät reagiert nicht", "Device offline", "Container läuft nicht", "Transfer hängt", "warum antwortet X nicht", "device triage", "Störung an Anlage".
---

# Gerät eingrenzen

## Reihenfolge

Diese Reihenfolge einhalten – sie geht von der breiten Diagnose zur teuren Einzelabfrage.

1. **Gerät identifizieren.** Nennt die Person einen Namen statt einer ID, mit
   `list_devices` auflösen. Bei mehreren Treffern nachfragen statt raten. Gerät
   nicht in `list_devices`, aber in `list_discovered_devices` → per
   `adopt_discovered_device` übernehmen (liefert `409`, wenn es zwischenzeitlich
   anderswo registriert wurde), danach QR-Registrierung für den Befehlsschlüssel.
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
   - Dateien auf dem Gerät (Agent ab 2.15.0): `list_device_directory` zeigt einen
     Ordner des Host-Dateisystems innerhalb der erlaubten Wurzeln (leerer Pfad =
     die Wurzeln; `/var/log` ist der übliche Einstieg). `DeviceTimeout` bei sonst
     gesundem Gerät heisst meist: Agent älter als 2.15.0. Eine Logdatei holen:
     `download_device_file`; mehrere Dateien oder einen Ordner als ZIP:
     `download_device_files` – beides wartet auf das Gerät und liefert kleine
     Ergebnisse inline, grosse als `contentUrl`. Laufende Downloads:
     `get_device_downloads`.
   - Last: `get_device_performance` für die letzten 48 Stunden roh,
     `get_device_telemetry` für den längeren Verlauf.
   - Ging ein Alarm raus? `list_notification_outbox` (Filter `state`, ggf.
     `channelId`) zeigt, ob und wohin Fusion.Sentinel-Findings zu diesem Gerät
     zugestellt wurden; `Failed` mit `lastError` erklärt, warum niemand eine Mail
     bekam, `Bundled` heisst: kam oder kommt mit der Tagesübersicht (`digestRowId`).
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
| Nicht in `list_devices`, aber in `list_discovered_devices` | Gerät sendet gültige Telemetrie, ist nur noch nicht registriert – kein Ausfall |

## Regeln

- **Nur lesen.** `container_action`, `send_device_command`, `delete_device`,
  `remove_device_image`, `abort_device_transfer`, `adopt_discovered_device`,
  `acknowledge_sentinel_finding`, `suppress_sentinel_finding`,
  `start_device_download`, `download_device_file`, `download_device_files`,
  `abort_device_download` und alles Schreibende erst nach ausdrücklicher
  Zustimmung – und vorher benennen, was passieren wird. Ein Download liest zwar
  nur auf dem Gerät, belegt aber dort Platte und Bandbreite und legt eine Kopie
  der Datei in der Plattform ab; `list_device_directory` ist reines Lesen.
- `run_sql_query` braucht Admin-Rechte. Für Fragen zu einem Gerät immer die
  dedizierten Werkzeuge nehmen, nie SQL.
- Befund und Vermutung trennen. "Queue bei 1.240, keine Consumer" ist ein Befund.
  "Der Worker ist abgestürzt" ist eine Vermutung – als solche kennzeichnen.
- Am Ende ein konkreter nächster Schritt, keine Liste von Möglichkeiten.