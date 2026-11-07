---
name: fusion-device-triage
description: Fehlersuche an einem einzelnen Fusion Edge-Gerät, wenn es nicht reagiert, Container hängen oder ein Transfer klemmt. Nutzen bei "Gerät reagiert nicht", "Device offline", "Container läuft nicht", "Transfer hängt", "warum antwortet X nicht", "device triage", "Störung an Anlage".
---

# Gerät eingrenzen

## Reihenfolge

Diese Reihenfolge einhalten – sie geht von der breiten Diagnose zur teuren Einzelabfrage.

1. **Gerät identifizieren.** Nennt die Person einen Namen statt einer ID, mit
   `list_devices` auflösen (`deviceId` ist der Bezeichner, den alle anderen
   Geräte-Werkzeuge nehmen). Bei mehreren Treffern nachfragen statt raten.
   `get_device` liefert `null`, wenn die ID unbekannt oder das Gerät für dich
   nicht sichtbar ist – das als „nicht gefunden" melden, nicht als Fehler.
   Gerät nicht in `list_devices`, aber in `list_discovered_devices` → siehe
   **Übernahme aus dem Discovery-Pool** unten.
2. **`get_device_diagnostics`** – der Einstieg. Liefert `device` (`status`
   `Active | Inactive`, `lastSeenUtc`), `broker` (Verbindung, Queues mit
   `messagesReady`/`messagesUnacked`/`consumers`), die letzten Befehle
   (`recentCommandTasks`) und Transfers (`recentTransfers` mit
   `chunksAcked`/`chunksTotal`) sowie `signalR`. Hinweis: Queue-Tiefen für den
   ganzen Broker und `totalMessagesReady` gibt es nur mit erweiterten
   Plattformrechten; alle anderen sehen nur die Queues des Geräts und
   `totalMessagesReady: null`. Das heisst **nicht**, dass der Broker steht.
3. **`trace_device`** – Rundlauf API → Broker → Agent → zurück. Liefert
   `{ latencyMs, response }`; in `response` stehen `success`, `timedOut` und
   `durationMs`. `success: false` oder `timedOut: true` bei gesundem Eintrag
   aus Schritt 2 deutet auf die Agent-Strecke.
4. Erst danach vertiefen, je nach Befund:
   - Container: `list_containers` ist der **Sollzustand** aus der Fusion-Datenbank,
     `list_live_containers` der **Istzustand** auf dem Gerät (`docker ps -a` über
     MQTT, das Gerät muss online sein; `success: false` mit `error` heisst „Gerät
     hat nicht geantwortet oder Docker hat abgelehnt" – der Aufruf selbst ist
     dann trotzdem gelungen, also `success` prüfen). `state` ist Dockers eigener
     Wert (`created | running | paused | restarting | removing | exited | dead`).
     Bei Auffälligkeit `get_container_logs` (`tail` 1–10000, Standard 100).
     **Zwei Bedeutungen von `containerId`:** `container_action` und
     `get_container_logs` nehmen die Docker-Container-ID, wie sie
     `list_live_containers` in `containerId` meldet; `get_container` nimmt das
     Feld `id` (GUID des Datenbanksatzes) einer `list_containers`-Zeile und
     liefert `null`, wenn die GUID unbekannt ist oder zu einem anderen Gerät
     gehört. `container_action` erwartet als `requestJson` genau
     `{"action":"Start"}`, `"Stop"`, `"Restart"` oder `"Remove"`; `Remove` ist
     `docker rm -f` auf dem Gerät (auch laufend), der Datenbanksatz bleibt.
   - Transfers: `list_device_transfers`, auf `state`, `chunksAcked`/`chunksTotal`
     und `resultSummary` achten.
   - Dateien auf dem Gerät (Agent ab 2.15.0): `list_device_directory` zeigt einen
     Ordner des Host-Dateisystems innerhalb der erlaubten Wurzeln (leerer Pfad =
     die Wurzeln; `/var/log` ist der übliche Einstieg). `DeviceTimeout` bei sonst
     gesundem Gerät heisst meist: Agent älter als 2.15.0. Eine Logdatei holen:
     `download_device_file`; mehrere Dateien oder einen Ordner als ZIP:
     `download_device_files` – beides wartet auf das Gerät und liefert kleine
     Ergebnisse (bis 8 MB) als `contentBase64`, grössere als `contentUrl`.
     Laufende Downloads: `get_device_downloads`.
   - Last: `get_device_performance` liefert die Rohproben eines Zeitfensters
     (`from`/`to` als ISO 8601 UTC, höchstens 31 Tage, längere Fenster werden auf
     die jüngsten 31 Tage gekürzt); bei mehr als etwa 500 Proben werden sie in
     Buckets zusammengefasst (`aggregated: true`), Lücken erscheinen als
     `"gap": true`. Für den längeren Verlauf einer Metrik `get_device_telemetry`
     (bis 365 Tage), die Metrik-Schlüssel kommen aus `list_device_metrics`.
     Installationen ohne Langzeit-Historie (`historyMode` RollingWindow in
     `api_health` → `capabilities`) kürzen `from` auf `rawRetentionHours` und
     antworten mit 413, wenn das Fenster mehr als 200000 Rohzeilen hält – vorher
     `api_health` lesen.
   - Fragen an den Agenten ohne Shell: `query_device` stellt genau eine der
     parameterlosen Fragen (`GET_AGENT_VERSION`, `GET_METRICS_INTERVAL`,
     `GET_CMD_TIMEOUT`, `GET_LOG_LEVEL`, `GET_PTY_STATUS`, `CHECK_INTERNET`,
     `CHECK_SCREEN`, `CHECK_VNC`, `HELP`) mit festen 15 s Timeout. Nur diese
     Namen übergeben – einen unbekannten Namen führt der Agent als
     Shell-Befehl aus. Der Aufruf wird als Command-Task aufgezeichnet.
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

## Übernahme aus dem Discovery-Pool

Ein Gerät, das in `list_discovered_devices` steht, sendet einen gültigen
`openMetaData`-Feed, ist aber noch in keinem Mandanten registriert – kein Ausfall.
Jede Pool-Zeile trägt `hasEnrolment`: `true` heisst, das Gerät hat seinen
Befehlsschlüssel bereits verschlüsselt mitgeliefert.

1. `adopt_discovered_device` (nur nach Zustimmung) legt das Gerät im eigenen
   Mandanten an, wahlweise in einer Org-Unit, und liefert den DeviceDto.
   Telemetrie läuft ab sofort. Ein `409 DeviceIdExists` heisst: inzwischen
   anderswo registriert.
2. Auf `cmdAuthKeyEnrolment` in der Antwort achten:
   - `pending` – die API bestätigt den mitgelieferten Schlüssel gerade per
     Probe-Befehl. Das wird innerhalb von Sekunden zu `enrolled`
     (`hasCmdAuthKey: true`, Befehle möglich) oder `probe-failed`; das Gerät mit
     `get_device` erneut lesen, um den Ausgang zu sehen.
   - `none` – es läuft keine Probe und es startet auch keine (der Pool-Eintrag
     hatte kein `hasEnrolment`, oder die API kennt die Schlüssel-ID des Chiffrats
     nicht mehr). Nicht auf `enrolled` warten.
   - `probe-failed` – das Gerät hat die Probe nicht beantwortet.
3. Bei `none` und `probe-failed` bleibt das Gerät ohne Befehlsschlüssel, bis die
   QR-Registrierung für dieselbe DeviceId ihn ablegt (oder `update_device` mit
   `cmdAuthKey`). Telemetrie ist davon nicht betroffen; Befehle, Container-Aktionen
   und Transfers scheitern bis dahin mit „no command key".

## Deutung

| Befund | Wahrscheinliche Ursache |
|--------|-------------------------|
| `lastSeenUtc` alt, Queue wächst | Agent weg, Nachrichten stauen sich |
| Queue-Tiefe steigt, keine Consumer | Der konsumierende Dienst ist unten |
| `trace_device` Timeout, Diagnose sonst sauber | Agent-Strecke oder Gerät selbst |
| `list_live_containers` mit `success: false`, Gerät sonst erreichbar | Kein Befehlsschlüssel (`hasCmdAuthKey: false`) oder Docker auf dem Gerät gestört – `error` lesen |
| `hasCmdAuthKey: false`, `cmdAuthKeyEnrolment: none` oder `probe-failed` | Gerät nie registriert oder Probe gescheitert – Telemetrie läuft, Befehle nicht; QR-Registrierung nachholen |
| `updateRequired: true` | Agent unter `minAgentVersion` – erst aktualisieren, dann weitersuchen |
| Nicht in `list_devices`, aber in `list_discovered_devices` | Gerät sendet gültige Telemetrie, ist nur noch nicht registriert – kein Ausfall |

## Null, 404 und Fehlertexte

- `get_device`, `get_container` und `get_device_command_task` antworten mit
  `null`, wenn die ID unbekannt oder das Objekt nicht sichtbar ist. Das ist
  „nicht gefunden", kein Werkzeugfehler.
- `get_device_history`, `get_device_tags`, `get_device_performance`,
  `list_containers`, `list_live_containers`, `get_container_logs`,
  `list_device_command_tasks` und die Befehls-Werkzeuge antworten auf eine
  unbekannte oder unsichtbare `deviceId` mit einem 404-Fehler.
- Ein fehlgeschlagener Aufruf kommt als Werkzeugfehler
  `Fusion.API answered {status} {code}: {message}`. Status und Code so
  weitergeben, wie sie stehen (`404 DeviceNotFound`, `409 DeviceIdExists`,
  `504 DeviceTimeout`) – nicht umdeuten.

## Regeln

- **Nur lesen.** `container_action`, `send_device_command`, `delete_device`,
  `remove_device_image`, `abort_device_transfer`, `adopt_discovered_device`,
  `acknowledge_sentinel_finding`, `suppress_sentinel_finding`,
  `start_device_download`, `download_device_file`, `download_device_files`,
  `abort_device_download` und alles Schreibende erst nach ausdrücklicher
  Zustimmung – und vorher benennen, was passieren wird. Ein Download liest zwar
  nur auf dem Gerät, belegt aber dort Platte und Bandbreite und legt eine Kopie
  der Datei in der Plattform ab; `list_device_directory` ist reines Lesen.
  `send_device_command` führt jeden Text, der kein Agent-Befehl ist, als
  Shell-Befehl auf dem Gerät aus – für Fragen `query_device` nehmen.
- Für Fragen zu einem Gerät immer die dedizierten Geräte-Werkzeuge nehmen.
- Befund und Vermutung trennen. "Queue bei 1.240, keine Consumer" ist ein Befund.
  "Der Worker ist abgestürzt" ist eine Vermutung – als solche kennzeichnen.
- Am Ende ein konkreter nächster Schritt, keine Liste von Möglichkeiten.
