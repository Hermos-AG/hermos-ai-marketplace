# Release Notes

> English version: [RELEASE_NOTES.md](RELEASE_NOTES.md)

## 1.3.0 — 16. August 2026

Klein und gezielt: der Skill `fusion-docs` kennt jetzt die Fundstellen der
MCP-Server-eigenen Dokumentation – Tool-Katalog, OAuth-Stack, Release-Historie – sowie
die Seiten zu Telemetrie und Geräte-Logs. `hermos-fusion` 0.3.0.

## 1.2.0 — 16. August 2026

Alles aufgeschrieben, was nötig ist, damit der Katalog firmenweit der Standard wird.

### Wege zum Ausrollen

```mermaid
flowchart LR
    A["Wo arbeiten die Leute?"] --> B["Claude-App, Cowork"]
    A --> C["Terminal, Claude Code"]
    B --> D["Organisationseinstellungen"]
    C --> E["Managed Settings per MDM"]
    D --> F["Automatisch installiert"]
    E --> F
```

Neu: [`docs/DEPLOYMENT_de.md`](docs/DEPLOYMENT_de.md) und
[`docs/DEPLOYMENT.md`](docs/DEPLOYMENT.md), mit einsatzfertigen Nutzlasten für
`extraKnownMarketplaces`, `enabledPlugins` und `strictKnownMarketplaces` sowie dem
geprüften Ablageort für jede Plattform.

### Zwei Dinge, über die man stolpert

Projekt-Settings installieren ein Plugin bei anderen **nicht**. Sie registrieren nach dem
Workspace-Trust den Marktplatz, mehr nicht – installieren muss weiterhin jede Person
selbst. Nur Organisationseinstellungen und Managed Settings bringen das Plugin
tatsächlich auf fremde Rechner.

Der alte Windows-Pfad `C:\ProgramData\ClaudeCode\managed-settings.json` funktioniert seit
v2.1.75 nicht mehr. Was dort liegt, muss nach `C:\Program Files\ClaudeCode\` umziehen.

### Versionen

Katalog 1.2.0. `hermos-fusion` bleibt bei 0.2.0 – reine Dokumentation, keine Änderung am
Plugin, also bekommt niemand ein sinnloses Update.
## 1.1.0 — 16. August 2026

Das Gerüst ist weg. `hermos-fusion` arbeitet jetzt mit dem echten Werkzeugsatz von Fusion.

### Drei Skills

```mermaid
graph LR
    P["hermos-fusion 0.2.0"] --> A["fusion-fleet-report"]
    P --> B["fusion-device-triage"]
    P --> C["fusion-docs"]
    A --> A1["list_devices, seitenweise"]
    B --> B1["get_device_diagnostics"]
    B --> B2["trace_device"]
    C --> C1["search_docs, read_doc"]
```

- **fusion-fleet-report** blättert `list_devices` vollständig durch, gruppiert nach
  Org-Unit und unterscheidet drei Zustände: online, still, offline. Jeder Agent unter
  `minAgentVersion` wird als Handlungspunkt geführt, nicht als Randnotiz.
- **fusion-device-triage** legt die Reihenfolge fest. Erst die Diagnose, weil dieser eine
  Aufruf schon Gerätezustand, Broker-Queues, letzte Befehle und Transfers abdeckt. Dann
  der Rundlauf. Container, Transfers und Last erst danach.
- **fusion-docs** antwortet aus der Dokumentation, die der MCP-Server mitbringt, und
  nennt immer die Quelldatei.

### Weg durch die Fehlersuche

```mermaid
flowchart TD
    S["Gerät reagiert nicht"] --> D["get_device_diagnostics"]
    D -->|"lastSeenAt alt, Queue wächst"| Q["Agent weg, Nachrichten stauen sich"]
    D -->|"sieht gesund aus"| T["trace_device"]
    T -->|"Timeout"| AG["Agent-Strecke oder Gerät selbst"]
    T -->|"Latenz in Ordnung"| C["Container, Transfers, Last"]
```

### Authentifizierung, geklärt

Der Endpunkt nutzt Entra ID. Der Client öffnet den Browser, du meldest dich mit deinem
normalen Hermos-Konto an, und jeder Werkzeugaufruf läuft mit demselben Mandanten und
denselben Rollen wie im Fusion-Web-UI. In die `.mcp.json` gehört nichts ausser der URL.

Personal Access Tokens (`fpat_…`) gibt es für den Betrieb ohne Browser – CI, Maschinen
ohne Anzeige. Die gehören in eine lokale Konfiguration, niemals in dieses Repository.

### Standardmässig nur lesen

Fusion bietet auch zerstörende Werkzeuge: Geräte löschen, Container-Aktionen, Images
entfernen, Transfers abbrechen. Alle drei Skills lesen nur und verlangen eine
ausdrückliche Bestätigung, bevor sich etwas ändert. `run_sql_query` sieht alle Mandanten
und ist für Gerätefragen ganz ausgeschlossen.

### Umstieg von 1.0.0

Nichts zu migrieren. `/plugin uninstall`, dann `/plugin install`, oder am Marktplatz auf
„Nach Updates suchen" klicken.

## 1.0.0 — 13. August 2026

Erster Wurf: Marktplatz-Katalog, ein Plugin am Fusion-MCP-Endpunkt, Skill- und
Command-Gerüste mit TODO-Platzhaltern.