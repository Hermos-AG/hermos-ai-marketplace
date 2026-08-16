---
name: fusion-docs
description: Beantwortet Fragen zu Fusion aus der mitgelieferten Projektdokumentation statt aus dem Gedächtnis – Architektur, MQTT, Auth und Entra ID, Org-Units, Container-Registry, Dateitransfer, Deploy-Runbook, Gerätebefehle. Nutzen bei "wie funktioniert X in Fusion", "was steht in der Doku zu", "wie ist Y konfiguriert", "Fusion Architektur", "MQTT Topics", "wie melde ich einen Agent an".
---

# Fusion-Doku befragen

Der MCP-Server liefert die Markdown-Dokumentation aller Fusion-Projekte mit, immer
zum Stand des jeweiligen Server-Builds. Diese Quelle schlägt jede Erinnerung.

## Vorgehen

1. `search_docs` mit ein bis drei prägnanten Begriffen. Die Suche ist ein
   Teilstring-Treffer ohne Gross-/Kleinschreibung – also `MQTT Topic`, nicht
   "Wie sind die MQTT-Topics aufgebaut".
2. Passenden Treffer mit `read_doc` öffnen. Der Pfad kommt aus dem Suchergebnis.
3. Findet die Suche nichts, `list_docs` aufrufen und über den Dateinamen gehen.

## Sprachpaare

Fast jedes Dokument existiert doppelt: `AUTH.md` und `AUTH_de.md`. Bei einer
deutschen Frage die `_de`-Fassung lesen, sonst das Original. Nicht beide laden.

## Wo was steht

| Thema | Datei |
|-------|-------|
| Gesamtarchitektur | `Fusion.API/docs/ARCHITECTURE.md` |
| Anmeldung, Token, Rollen | `Fusion.API/docs/AUTH.md` |
| Entra-ID-Einrichtung | `Fusion.API/docs/ENTRA_ID_SETUP.md` |
| MQTT-Einstieg und Topics | `Fusion.API/docs/MQTT_QUICKSTART.md`, `MQTT_GUIDE.md` |
| Geräte anlegen | `Fusion.API/docs/DEVICE_ONBOARDING.md` |
| Befehle an Geräte | `Fusion.API/docs/DEVICE_COMMANDS.md` |
| Container auf Geräten | `Fusion.API/docs/DEVICE_CONTAINERS.md` |
| Org-Unit-Baum und Rechte | `Fusion.API/docs/ORG_UNITS.md` |
| Dateitransfer | `Fusion.API/docs/FILE_TRANSFER.md` |
| Deployment | `Fusion.API/docs/DEPLOY_RUNBOOK.md` |
| MCP-Client anbinden | `Fusion.McpServer/docs/MCP_CLIENT_SETUP.md` |
| MCP-Server: Überblick, Tool-Katalog, Betrieb | `Fusion.McpServer/README.md` |
| MCP-Server: OAuth-Stack, Token-Tausch, PATs | `Fusion.McpServer/docs/OAUTH.md` |
| MCP-Server: was sich geändert hat | `Fusion.McpServer/CHANGELOG.md`, `RELEASE_NOTES.md` |
| Telemetrie-API und Metrik-Schlüssel | `Fusion.API/docs/DEVICE_TELEMETRY.md` |
| Geräte-Logs | `Fusion.API/docs/DEVICE_LOGS.md` |

## Regeln

- Immer die Quelldatei nennen, aus der die Antwort stammt.
- Steht es nicht in der Doku, das sagen – und nicht aus allgemeinem Wissen über
  MQTT oder Docker auffüllen, ohne es zu kennzeichnen.
- Einige Dateien sind sehr gross (`CHANGELOG.md` über 180 kB). Erst `search_docs`,
  dann gezielt lesen, statt eine grosse Datei komplett zu ziehen.