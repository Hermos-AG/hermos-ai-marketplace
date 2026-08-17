# Release Notes

> English version: [RELEASE_NOTES.md](RELEASE_NOTES.md)

## 1.4.2 — 17. August 2026

Der Katalog baut und prüft sich jetzt selbst auf GitHub. Für ein neues gpu-mcp-Binary
braucht niemand mehr eine Go-Toolchain, und eine veraltete Release-Kopie kommt nicht mehr
nach `main`.

```mermaid
flowchart LR
    subgraph SRC["Hermos-AG/HER-gpu-mcp"]
        A["Push / PR"] --> B["build: vet, Cross-Compile, Smoke-Test"]
        T["Tag v*"] --> R["GitHub-Release<br/>gpu-mcp.exe + gpu-mcp-linux"]
    end
    subgraph CAT["Hermos-AG/hermos-ai-marketplace"]
        S["scripts/sync-gpu-mcp.ps1<br/>oder Workflow refresh-gpu-binaries"] --> P["Pull Request"]
        P --> V["validate: Manifeste, Versionen,<br/>Doku-Paare, Binary-Version"]
        V --> M["Merge → Sync nach Claude"]
    end
    R -.->|"Release-Kopie"| S
```

- **`validate`** läuft bei jedem Push und Pull Request: Katalog- und Plugin-Manifeste,
  Versionskonsistenz, zweisprachige Doku-Paare, JSON-Syntax und ob die mitgelieferten
  Binaries wirklich die Plugin-Version tragen.
- **`refresh-gpu-binaries`** (Actions → Run workflow) baut beide Binaries aus dem
  eingecheckten Go-Quellcode neu, testet den Linux-Build gegen das Fake-`nvidia-smi`, zieht
  die Versionen nach und öffnet den Pull Request.
- **Behoben:** Der Fusion-Eintrag in `marketplace.json` hieß `HERMOS-Fusion`, das Manifest
  aber `hermos-fusion` — jetzt stimmen beide überein.
- Plugins unverändert: `hermos-fusion` 0.3.1, `HERMOS-local-GPU` 0.2.0.
## 1.4.1 — 17. August 2026

Aufräum-Release: Der Katalog heißt jetzt **`hermos-ai-marketplace`** und liegt bei den
MCP-Server-Quellen statt für sich allein. Ein privates Repository je MCP-Server.

```mermaid
flowchart LR
    subgraph N["D:/DEV/HER/HER-MCP"]
        M["hermos-ai-marketplace/<br/>Katalog hermos 1.4.1"]
        G["gpu-mcp/ → HER-gpu-mcp"]
        U["unifi-network-mcp/ → HER-unifi-network-mcp"]
        W["windows-mcp/ → HER-windows-mcp"]
    end
    O["alt: D:/DEV/HER/Claude.Catalog<br/>HER-Claude-Catalog"] -.->|"umbenannt + verschoben"| M
    G -->|"Release-Kopie"| M
```

- **Für installierte Nutzer nichts zu tun** — GitHub leitet die alte Repository-URL weiter.
  Verweise trotzdem bei Gelegenheit auf `Hermos-AG/hermos-ai-marketplace` umstellen.
- **Plugins unverändert:** `hermos-fusion` 0.3.1, `HERMOS-local-GPU` 0.2.0.
- **Neue Doku:** Die Plugin-Checkliste `docs/ADDING-A-PLUGIN_de.md` (und die englische Fassung)
  liegt jetzt im Katalog selbst.
## 1.4.0 — 16. August 2026

Zweites Plugin im Katalog: `HERMOS-local-GPU` bringt die eigene NVIDIA-GPU des
Entwicklers in Claudes Reichweite – Monitoring, CUDA-Prozesse und kontrollierte
GPU-Jobs, alles lokal über ein einzelnes, abhängigkeitsfreies Binary (Projekt
`gpu-mcp`).

```mermaid
graph LR
    P["HERMOS-local-GPU 0.2.0"] --> A["gpu_get_status / gpu_query_metrics"]
    P --> B["gpu_list_processes"]
    P --> C["gpu_check_requirements"]
    P --> D["gpu_run_command"]
    C --> C1["NVIDIA ≥ 8 GB VRAM?"]
    C1 -->|"RESULT: MET"| E["GPU-Jobs erlaubt"]
    C1 -->|"RESULT: NOT MET"| F["nur Monitoring"]
```

- **Läuft auf dem Entwickler-Rechner** – das Plugin liefert `gpu-mcp.exe` mit und
  registriert sie automatisch; nichts zu konfigurieren. Cloud-Cowork-Sessions
  erreichen die GPU über die Geräte-Brücke der Desktop-App.
- **„Ausreichende GPU" wird geprüft, nicht angenommen** – NVIDIA mit ≥ 8 GB VRAM
  als Standard (`GPU_MCP_MIN_VRAM_MB`, `0` = nur Treiber-Check). Unter dem Minimum
  funktioniert das Monitoring weiter; der Check meldet `RESULT: NOT MET`, GPU-Jobs
  bleiben tabu.
- **Installation:** `/plugin install HERMOS-local-GPU@hermos` – danach Claude das
  Tool `gpu_check_requirements` ausführen lassen oder im Terminal
  `gpu-mcp.exe --check`.
- Katalog 1.3.1 → 1.4.0. `hermos-fusion` unverändert bei 0.3.1.

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