# HERMOS AI Marketplace

Interner Plugin-Marktplatz für Claude Cowork und Claude Code bei HERMOS.

Ein Claude-Marktplatz ist kein gehosteter Store, sondern schlicht ein Git-Repository mit
einem Katalog unter `.claude-plugin/marketplace.json`. Wer dieses Repository hinzufügt,
sieht die unten aufgeführten Plugins.

> English version: [README.md](README.md)

## Was drin ist

**`hermos-fusion` 0.2.0** – Edge-Geräteverwaltung über den Fusion-MCP-Server.

| Skill | Wofür |
|-------|-------|
| `fusion-fleet-report` | Flottenüberblick: online, still, offline. Gruppiert nach Org-Unit, Agents unter `minAgentVersion` als Handlungspunkt gekennzeichnet. |
| `fusion-device-triage` | Geordnete Fehlersuche an einem stummen Gerät – Diagnose, Rundlauf, dann Container, Transfers, Last. |
| `fusion-docs` | Beantwortet Fusion-Fragen aus der Dokumentation, die der MCP-Server mitliefert, und nennt die Quelldatei. |

| Command | Wofür |
|---------|-------|
| `/fusion-status [gerät]` | Kurzer Status zu einem Gerät. |
| `/fusion-fleet [org-unit oder tag]` | Flottenüberblick, wahlweise gefiltert. |

Alle Skills **lesen nur**. Fusion bietet auch zerstörende Werkzeuge – Geräte löschen,
Container-Aktionen, Transfers abbrechen. Die verlangen eine ausdrückliche Bestätigung.
`run_sql_query` sieht alle Mandanten und ist für Gerätefragen ausgeschlossen.

**`HERMOS-local-GPU` 0.2.0** – die eigene NVIDIA-GPU des Entwicklers als MCP-Server
(Projekt `gpu-mcp`: ein einzelnes, abhängigkeitsfreies Go-Binary, JSON-RPC über stdio,
kein Python/Node). Läuft **lokal** auf dem Rechner jedes Entwicklers mit ausreichender
GPU – NVIDIA mit **≥ 8 GB VRAM** (konfigurierbar über `GPU_MCP_MIN_VRAM_MB`); das
Plugin prüft das selbst.

| Tool | Wofür |
|------|-------|
| `gpu_get_status` | Vollständiger `nvidia-smi`-Bericht: Treiber, CUDA, Auslastung, Speicher, Temperatur, alle GPU-Prozesse. |
| `gpu_query_metrics` | Kompakte CSV-Metriken fürs Monitoring. |
| `gpu_list_processes` | CUDA-Compute-Prozesse (PID, Name, GPU-Speicher). |
| `gpu_check_requirements` | Prüft die GPU-Voraussetzung – Bericht endet mit `RESULT: MET` / `NOT MET`; auch als `gpu-mcp.exe --check`. |
| `gpu_run_command` | Führt einen Shell-Befehl aus, z. B. um GPU-Jobs zu starten – **führt Befehle aus**, Freigabe-Dialoge bleiben streng. |

Die ersten vier Tools lesen nur. `gpu_run_command` führt beliebige Befehle mit den
Rechten des angemeldeten Benutzers aus – Details und Sicherheitshinweise:
[`plugins/gpu-mcp/README_de.md`](plugins/gpu-mcp/README_de.md).

## Wo dieses Repo liegt — und wo die Quellen liegen

Arbeitskopie: `D:\DEV\HER\HER-MCP\hermos-ai-marketplace` — direkt neben den MCP-Server-Quellen,
die es veröffentlicht. Jeder MCP-Server hat sein eigenes privates Repository in der
Organisation `Hermos-AG`:

| Server | Quell-Repo | Arbeitskopie |
|---|---|---|
| `gpu-mcp` (Plugin `HERMOS-local-GPU`) | `Hermos-AG/HER-gpu-mcp` | `D:\DEV\HER\HER-MCP\gpu-mcp` |
| `unifi-network-mcp` | `Hermos-AG/HER-unifi-network-mcp` (upstream `sirkirby/unifi-network-mcp`) | `D:\DEV\HER\HER-MCP\unifi-network-mcp` |
| `windows-mcp` | `Hermos-AG/HER-windows-mcp` (upstream `CursorTouch/Windows-MCP`) | `D:\DEV\HER\HER-MCP\windows-mcp` |
| Fusion-MCP (Plugin `hermos-fusion`) | Teil der Fusion-Solution, `D:\DEV\HER\Fusion` | gehosteter Endpunkt, kein Checkout nötig |

`plugins/<name>/` in diesem Repo ist die **Release-Kopie** eines Servers: entwickelt wird im
Quell-Repo, der Release-Stand wird hierher kopiert, Version hochgezählt, Pull Request auf.
Übersicht aller Repositories: `D:\DEV\HER\HER-MCP\README.md`.

Dieses Repository hieß bis zum 17.08.2026 `HER-Claude-Catalog`; GitHub leitet die alte URL
weiter, bitte trotzdem `Hermos-AG/hermos-ai-marketplace` verwenden.
## Aufbau des Repositorys

```mermaid
graph TD
    R["hermos-ai-marketplace/"] --> M[".claude-plugin/marketplace.json"]
    R --> P["plugins/"]
    P --> F["hermos-fusion/"]
    F --> FM[".claude-plugin/plugin.json"]
    F --> FC[".mcp.json"]
    F --> FS["skills/"]
    F --> FK["commands/"]
    FS --> S1["fusion-fleet-report"]
    FS --> S2["fusion-device-triage"]
    FS --> S3["fusion-docs"]
    FK --> K1["fusion-status"]
    FK --> K2["fusion-fleet"]
    P --> G["gpu-mcp/"]
    G --> GM[".claude-plugin/plugin.json"]
    G --> GC[".mcp.json"]
    G --> GE["gpu-mcp.exe + Go-Quellcode"]
```

| Datei | Zweck |
|-------|-------|
| `.claude-plugin/marketplace.json` | Der Katalog. Listet jedes Plugin und dessen Ablageort. |
| `plugins/<name>/.claude-plugin/plugin.json` | Plugin-Manifest: Name, Version, Beschreibung. |
| `plugins/<name>/.mcp.json` | MCP-Server, die das Plugin mitbringt. |
| `plugins/<name>/skills/<skill>/SKILL.md` | Ein Skill. Das Frontmatter-Feld `description` entscheidet, wann Claude ihn zieht. |
| `plugins/<name>/commands/<cmd>.md` | Ein Slash-Command. |

## Authentifizierung

Der Fusion-MCP-Server liegt hinter Entra ID. Der Client öffnet einen Browser, du meldest
dich mit deinem gewöhnlichen Hermos-Konto an, und jeder Werkzeugaufruf läuft mit demselben
Mandanten und denselben Rollen wie im Fusion-Web-UI. In der `.mcp.json` steht deshalb
nichts ausser der URL.

```mermaid
sequenceDiagram
    participant C as Claude
    participant M as Fusion MCP
    participant E as Entra ID
    participant A as Fusion API
    C->>M: erster Aufruf, kein Token
    M-->>C: 401, verweist auf OAuth-Metadaten
    C->>E: Anmeldung im Browser
    E-->>C: Access Token
    C->>M: jeder Aufruf trägt das Token
    M->>A: Tausch gegen eine Fusion-Sitzung
    A-->>M: Identität, Mandant, Rollen
    M-->>C: Ergebnisse im Rahmen deiner Rechte
```

Verweigert der Client die automatische Registrierung, die Client-ID
`f44473d3-115d-4c76-ba23-71655a672c97` in den erweiterten Einstellungen des Konnektors
eintragen. Ein Client Secret gibt es nicht.

Für den Betrieb ohne Browser funktioniert stattdessen ein Personal Access Token
(`fpat_…`) als `Authorization`-Header. **Niemals in dieses Repository einchecken.**

Angemeldet, aber keine Daten sichtbar? Die erste Anmeldung legt nur ein Basiskonto an.
Org-Units und Rollen muss ein Fusion-Admin noch zuweisen.

## Installieren

```mermaid
flowchart LR
    A["Repo klonen oder pushen"] --> B{"Für wen?"}
    B -->|"Nur ich"| C["Cowork: Anpassen, Plugins durchsuchen, Persönlich, Plus, Marketplace hinzufügen"]
    B -->|"Ganze Firma"| D["Organisationseinstellungen, Plugins, Plugin hinzufügen, GitHub"]
    B -->|"Terminal"| E["claude, dann slash plugin marketplace add"]
    C --> F["hermos-fusion installieren"]
    D --> F
    E --> F
```

**Lokaler Testlauf** – ohne Push:

```bash
cd D:\DEV\HER\HER-MCP\hermos-ai-marketplace
claude
/plugin marketplace add .
/plugin install hermos-fusion@hermos
/plugin install HERMOS-local-GPU@hermos
```

`HERMOS-local-GPU` setzt eine NVIDIA-GPU mit ≥ 8 GB VRAM auf dem installierenden
Rechner voraus – nach der Installation prüfen mit dem Tool
`gpu_check_requirements` (oder `gpu-mcp.exe --check`).

**Persönlich, von GitHub:** Reiter Cowork, in der Seitenleiste „Anpassen", dann
„Plugins durchsuchen", „Persönlich", die Schaltfläche „+", „Marketplace von GitHub
hinzufügen", anschliessend `Hermos-AG/hermos-ai-marketplace` eintragen.

**Firmenweit:** Organisationseinstellungen, „Plugins", „Plugin hinzufügen", Quelle
GitHub, Repository als `Hermos-AG/hermos-ai-marketplace`. Setzt einen Team- oder Enterprise-Plan mit
Owner-Rechten voraus, ausserdem müssen Cowork und Skills aktiviert sein.

## Randbedingungen, die man kennen sollte

- Ein Organisations-Marktplatz braucht ein **privates oder internes** Repository auf
  github.com. Öffentliche Repos werden dort abgelehnt, eigene GitHub-Enterprise-Server
  werden nicht unterstützt.
- Relative Quellen wie `./plugins/hermos-fusion` funktionieren überall. Die Quelltypen
  `github`, `url` und `git-subdir` lösen nur auf, wenn das Ziel-Repository öffentlich ist.
  `npm` und `pip` gehen gar nicht.
- Die automatische Synchronisierung startet, wenn ein Pull Request mit Versionssprung in
  den Default-Branch gemerged wird. Ein direkter Push löst sie nicht aus – dann von Hand
  „Nach Updates suchen".
- Skills eines Plugins wirken in Chat, Desktop und Cowork. Hooks und Sub-Agenten laufen
  nur in Cowork.

## Continuous Integration

| Workflow | Auslöser | Was er tut |
|---|---|---|
| `validate` | Push auf `main`, jeder Pull Request | `scripts/validate_catalog.py`: Katalog- und Plugin-Manifeste, Versionskonsistenz, zweisprachige Doku-Paare, JSON-Syntax und ob die mitgelieferten Binaries die Plugin-Version tragen. `claude plugin validate .` läuft informativ mit. |
| `refresh-gpu-binaries` | manuell (Actions → Run workflow) | Baut die Binaries in `plugins/gpu-mcp` (windows/amd64 + linux/amd64) aus dem eingecheckten Go-Quellcode neu, testet den Linux-Build gegen das Fake-`nvidia-smi`, zieht die Versionen nach, öffnet einen Pull Request. |

Lokal kopiert `pwsh scripts/sync-gpu-mcp.ps1` den Release-Stand aus dem Quell-Repo
`Hermos-AG/HER-gpu-mcp`, validiert ihn und öffnet den Pull Request. Das Quell-Repo hat seinen
eigenen `build`-Workflow (vet, Cross-Compile, Smoke-Test) und veröffentlicht die Binaries bei
einem `v*`-Tag als GitHub-Release.

Actions-Minuten: Die Organisation liegt im GitHub-Free-Plan, der 2.000 Minuten pro Monat für
private Repositories enthält. Eine Katalogvalidierung braucht deutlich unter einer Minute.
## Eine Änderung ausliefern

```mermaid
sequenceDiagram
    participant Dev as Entwickler
    participant Repo as GitHub
    participant Claude as Claude
    Dev->>Dev: Skill oder Command ändern
    Dev->>Dev: Version in plugin.json und marketplace.json erhöhen
    Dev->>Dev: Beide Changelogs und Release Notes nachziehen
    Dev->>Repo: Pull Request öffnen
    Repo->>Repo: In Default-Branch mergen
    Repo->>Claude: Synchronisierung
    Claude->>Dev: Neue Version für das Team sichtbar
```

Nutzer bekommen ein Update erst, wenn sich `version` in der `plugin.json` ändert – also
bei jedem Release erhöhen. Lässt man `version` ganz weg, zählt stattdessen jeder Commit
als neue Version.

## Ein weiteres Plugin aufnehmen

1. `plugins/<neuer-name>/.claude-plugin/plugin.json`
2. Skills, Commands, Agents, Hooks oder `.mcp.json` daneben legen
3. Neuer Eintrag im `plugins`-Array von `.claude-plugin/marketplace.json`
4. Versionen erhöhen, beide Changelogs nachziehen, Pull Request öffnen
5. Vor dem Push `claude plugin validate .`

## Dokumentation

- **[Deployment](docs/DEPLOYMENT_de.md)** — für alle ausrollen, standardmässig installiert
- Marktplätze: https://code.claude.com/docs/de/plugin-marketplaces
- Plugins in der Claude-App: https://support.claude.com/de/articles/13837440-plugins-in-claude-verwenden
- Organisationsverwaltung: https://support.claude.com/de/articles/13837433-verwalten-sie-cowork-plugins-fur-ihre-organisation
- Fusion-MCP-Client einrichten: `Fusion.McpServer/docs/MCP_CLIENT_SETUP_de.md` (über den Skill `fusion-docs`)