# Release Notes

> Deutsche Fassung: [RELEASE_NOTES_de.md](RELEASE_NOTES_de.md)

## 1.4.2 — 17 August 2026

The catalog now builds and checks itself on GitHub. Nobody needs a Go toolchain to ship a
new gpu-mcp binary, and a stale release copy can no longer reach `main`.

```mermaid
flowchart LR
    subgraph SRC["Hermos-AG/HER-gpu-mcp"]
        A["push / PR"] --> B["build: vet, cross-compile, smoke test"]
        T["tag v*"] --> R["GitHub release<br/>gpu-mcp.exe + gpu-mcp-linux"]
    end
    subgraph CAT["Hermos-AG/hermos-ai-marketplace"]
        S["scripts/sync-gpu-mcp.ps1<br/>or workflow refresh-gpu-binaries"] --> P["pull request"]
        P --> V["validate: manifests, versions,<br/>docs pairs, binary version"]
        V --> M["merge → sync to Claude"]
    end
    R -.->|"release copy"| S
```

- **`validate`** runs on every push and pull request: catalog and plugin manifests, version
  consistency, bilingual documentation pairs, JSON syntax, and whether the shipped binaries
  actually carry the plugin version.
- **`refresh-gpu-binaries`** (Actions → Run workflow) rebuilds both binaries from the
  committed Go source, smoke-tests the Linux build against the fake `nvidia-smi`, bumps the
  versions and opens the pull request for you.
- **Fixed:** the Fusion entry in `marketplace.json` was named `HERMOS-Fusion` while its
  manifest says `hermos-fusion`; both agree now.
- Plugins unchanged: `hermos-fusion` 0.3.1, `HERMOS-local-GPU` 0.2.0.
## 1.4.1 — 17 August 2026

Housekeeping release: the catalog is now called **`hermos-ai-marketplace`** and lives with
the MCP server sources instead of off on its own. One private repository per MCP server.

```mermaid
flowchart LR
    subgraph N["D:/DEV/HER/HER-MCP"]
        M["hermos-ai-marketplace/<br/>catalog hermos 1.4.1"]
        G["gpu-mcp/ → HER-gpu-mcp"]
        U["unifi-network-mcp/ → HER-unifi-network-mcp"]
        W["windows-mcp/ → HER-windows-mcp"]
    end
    O["old: D:/DEV/HER/Claude.Catalog<br/>HER-Claude-Catalog"] -.->|"renamed + moved"| M
    G -->|"release copy"| M
```

- **Nothing to do for installed users** — GitHub redirects the old repository URL. Still,
  switch references to `Hermos-AG/hermos-ai-marketplace` when you touch them.
- **Plugins unchanged:** `hermos-fusion` 0.3.1, `HERMOS-local-GPU` 0.2.0.
- **New docs:** the plugin checklist `docs/ADDING-A-PLUGIN.md` (and `_de`) now ships with the
  catalog itself.
## 1.4.0 — 16 August 2026

Second plugin in the catalog: `HERMOS-local-GPU` puts the developer's own NVIDIA
GPU at Claude's fingertips — monitoring, CUDA processes, and controlled GPU jobs,
all running locally through a single dependency-free binary (project `gpu-mcp`).

```mermaid
graph LR
    P["HERMOS-local-GPU 0.2.0"] --> A["gpu_get_status / gpu_query_metrics"]
    P --> B["gpu_list_processes"]
    P --> C["gpu_check_requirements"]
    P --> D["gpu_run_command"]
    C --> C1["NVIDIA ≥ 8 GB VRAM?"]
    C1 -->|"RESULT: MET"| E["GPU jobs allowed"]
    C1 -->|"RESULT: NOT MET"| F["monitoring only"]
```

- **Runs on the developer's machine** — the plugin ships `gpu-mcp.exe` and registers
  it automatically; nothing to configure. Cloud Cowork sessions reach the GPU through
  the desktop app's device bridge.
- **"Sufficient GPU" is checked, not assumed** — NVIDIA with ≥ 8 GB VRAM by default
  (`GPU_MCP_MIN_VRAM_MB`, `0` = driver check only). Below the minimum, monitoring
  still works; the check says `RESULT: NOT MET` and GPU jobs stay off.
- **Install:** `/plugin install HERMOS-local-GPU@hermos` — then ask Claude to run
  `gpu_check_requirements`, or run `gpu-mcp.exe --check` in a terminal.
- Catalog 1.3.1 → 1.4.0. `hermos-fusion` unchanged at 0.3.1.

## 1.3.0 — 16 August 2026

Small, targeted: the `fusion-docs` skill now knows where the MCP server's own
documentation lives — tool catalogue, OAuth stack, release history — plus the telemetry
and device-log pages. `hermos-fusion` 0.3.0.

## 1.2.0 — 16 August 2026

Everything needed to make the catalog the default for the whole company, written down.

### Rollout routes

```mermaid
flowchart LR
    A["Where do people work?"] --> B["Claude app, Cowork"]
    A --> C["Terminal, Claude Code"]
    B --> D["Organization settings"]
    C --> E["Managed settings via MDM"]
    D --> F["Installed automatically"]
    E --> F
```

New: [`docs/DEPLOYMENT.md`](docs/DEPLOYMENT.md) and
[`docs/DEPLOYMENT_de.md`](docs/DEPLOYMENT_de.md), with paste-ready payloads for
`extraKnownMarketplaces`, `enabledPlugins` and `strictKnownMarketplaces`, and the verified
deployment location for every platform.

### Two things that catch people out

Project settings do **not** install a plugin for others. They register the marketplace
after workspace trust, nothing more — each user still installs. Only organization settings
and managed settings actually put the plugin on someone else's machine.

The old Windows path `C:\ProgramData\ClaudeCode\managed-settings.json` stopped working in
v2.1.75. Anything deployed there has to move to `C:\Program Files\ClaudeCode\`.

### Versions

Catalog 1.2.0. `hermos-fusion` stays at 0.2.0 — documentation only, no plugin change, so
nobody gets a pointless update.
## 1.1.0 — 16 August 2026

The scaffold is gone. `hermos-fusion` now works against the real Fusion tool set.

### Three skills

```mermaid
graph LR
    P["hermos-fusion 0.2.0"] --> A["fusion-fleet-report"]
    P --> B["fusion-device-triage"]
    P --> C["fusion-docs"]
    A --> A1["list_devices, paged"]
    B --> B1["get_device_diagnostics"]
    B --> B2["trace_device"]
    C --> C1["search_docs, read_doc"]
```

- **fusion-fleet-report** walks every page of `list_devices`, groups by org unit and
  separates three states: online, quiet, offline. It flags every agent below
  `minAgentVersion` as an action item rather than a note.
- **fusion-device-triage** fixes the order of investigation. Diagnostics first, because
  that one call already covers device state, broker queues, recent commands and
  transfers. Round-trip trace second. Containers, transfers and load only after that.
- **fusion-docs** answers from the documentation the MCP server ships with, and always
  names the source file.

### Triage path

```mermaid
flowchart TD
    S["Device not responding"] --> D["get_device_diagnostics"]
    D -->|"lastSeenAt old, queue growing"| Q["Agent gone, messages piling up"]
    D -->|"looks healthy"| T["trace_device"]
    T -->|"timeout"| AG["Agent leg or device itself"]
    T -->|"latency fine"| C["Containers, transfers, load"]
```

### Authentication, settled

The endpoint uses Entra ID. The client opens a browser, you sign in with your normal
Hermos account, and every tool call runs with the same tenant and roles you have in the
Fusion web UI. Nothing goes into `.mcp.json` beyond the URL.

Personal access tokens (`fpat_…`) exist for headless use — CI, machines without a
browser. Those belong in a local configuration, never in this repository.

### Read-only by default

Fusion exposes destructive tools: deleting devices, container actions, removing images,
aborting transfers. All three skills are read-only and require explicit confirmation
before anything changes. `run_sql_query` spans all tenants and is excluded from device
questions entirely.

### Upgrade from 1.0.0

Nothing to migrate. `/plugin uninstall` then `/plugin install`, or click "Check for
updates" on the marketplace.

## 1.0.0 — 13 August 2026

First cut: marketplace catalog, one plugin wired to the Fusion MCP endpoint, skill and
command scaffolds carrying TODO placeholders.