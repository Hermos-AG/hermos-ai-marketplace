# HERMOS Claude Catalog

Internal plugin marketplace for Claude Cowork and Claude Code at HERMOS.

A Claude marketplace is not a hosted store — it is a plain Git repository containing a
`.claude-plugin/marketplace.json` catalog. Anyone who adds this repository sees the
plugins listed below.

> Deutsche Fassung: [README_de.md](README_de.md)

## What is in it

**`hermos-fusion` 0.2.0** — edge device management through the Fusion MCP server.

| Skill | What it does |
|-------|--------------|
| `fusion-fleet-report` | Fleet overview: online, quiet, offline. Grouped by org unit, agents below `minAgentVersion` flagged as action items. |
| `fusion-device-triage` | Ordered investigation of an unresponsive device — diagnostics, round-trip trace, then containers, transfers, load. |
| `fusion-docs` | Answers Fusion questions from the documentation the MCP server ships with, naming the source file. |

| Command | What it does |
|---------|--------------|
| `/fusion-status [device]` | Short status for one device. |
| `/fusion-fleet [org-unit or tag]` | Fleet overview, optionally filtered. |

All skills are **read-only**. Fusion also exposes destructive tools — deleting devices,
container actions, aborting transfers. Those require explicit confirmation.
`run_sql_query` spans all tenants and is excluded from device questions.

## Repository layout

```mermaid
graph TD
    R["Claude.Catalog/"] --> M[".claude-plugin/marketplace.json"]
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
```

| File | Purpose |
|------|---------|
| `.claude-plugin/marketplace.json` | The catalog. Lists every plugin and where it lives. |
| `plugins/<name>/.claude-plugin/plugin.json` | Plugin manifest: name, version, description. |
| `plugins/<name>/.mcp.json` | MCP servers the plugin brings along. |
| `plugins/<name>/skills/<skill>/SKILL.md` | A skill. Frontmatter `description` decides when Claude reaches for it. |
| `plugins/<name>/commands/<cmd>.md` | A slash command. |

## Authentication

The Fusion MCP server sits behind Entra ID. Your client opens a browser, you sign in with
your ordinary Hermos account, and every tool call runs with the same tenant and roles you
have in the Fusion web UI. `.mcp.json` therefore carries nothing but the URL.

```mermaid
sequenceDiagram
    participant C as Claude
    participant M as Fusion MCP
    participant E as Entra ID
    participant A as Fusion API
    C->>M: first call, no token
    M-->>C: 401, points at OAuth metadata
    C->>E: browser sign-in
    E-->>C: access token
    C->>M: every call carries the token
    M->>A: exchange for a Fusion session
    A-->>M: identity, tenant, roles
    M-->>C: results scoped to your permissions
```

If the client refuses automatic registration, enter the client ID
`f44473d3-115d-4c76-ba23-71655a672c97` in the connector's advanced settings. There is no
client secret.

For headless use a personal access token (`fpat_…`) works instead, passed as an
`Authorization` header. **Never commit one to this repository.**

Signed in but no data visible? The first login only creates a basic account. A Fusion
admin still has to assign org units and roles.

## Install

```mermaid
flowchart LR
    A["Clone or push repo"] --> B{"Who is it for?"}
    B -->|"Just me"| C["Cowork: Customize, Browse plugins, Personal, plus, Add marketplace"]
    B -->|"Whole company"| D["Organization settings, Plugins, Add plugin, GitHub"]
    B -->|"Terminal"| E["claude then slash plugin marketplace add"]
    C --> F["Install hermos-fusion"]
    D --> F
    E --> F
```

**Local test run** — no push needed:

```bash
cd D:\DEV\HER\Claude.Catalog
claude
/plugin marketplace add .
/plugin install hermos-fusion@hermos
```

**Personal, from GitHub:** Cowork tab, "Customize" in the sidebar, "Browse plugins",
"Personal", the "+" button, "Add marketplace from GitHub", then `Hermos-AG/HER-Claude-Catalog`.

**Organization-wide:** Organization settings, "Plugins", "Add plugin", source GitHub,
repository as `Hermos-AG/HER-Claude-Catalog`. Requires a Team or Enterprise plan with Owner rights, and
both Cowork and Skills enabled.

## Constraints worth knowing

- An organization marketplace needs a **private or internal** repository on github.com.
  Public repos are rejected there, and self-hosted GitHub Enterprise Server is not
  supported.
- Relative sources such as `./plugins/hermos-fusion` work everywhere. The `github`,
  `url` and `git-subdir` source types only resolve when the target repository is public.
  `npm` and `pip` are not supported at all.
- Automatic sync fires when a pull request containing a version bump is merged into the
  default branch. A direct push does not trigger it — use "Check for updates" instead.
- Skills from a plugin work in Chat, Desktop and Cowork. Hooks and sub-agents run in
  Cowork only.

## Releasing a change

```mermaid
sequenceDiagram
    participant Dev as Developer
    participant Repo as GitHub
    participant Claude as Claude
    Dev->>Dev: Edit skill or command
    Dev->>Dev: Bump version in plugin.json and marketplace.json
    Dev->>Dev: Update both changelogs and release notes
    Dev->>Repo: Open pull request
    Repo->>Repo: Merge into default branch
    Repo->>Claude: Sync
    Claude->>Dev: New version visible to the team
```

Users only receive an update when `version` in `plugin.json` changes, so bump it on every
release. Leave `version` out entirely and every commit counts as a new version instead.

## Adding another plugin

1. `plugins/<new-name>/.claude-plugin/plugin.json`
2. Skills, commands, agents, hooks or `.mcp.json` next to it
3. A new entry in the `plugins` array of `.claude-plugin/marketplace.json`
4. Bump versions, update both changelogs, open a pull request
5. `claude plugin validate .` before pushing

## Docs

- **[Deployment](docs/DEPLOYMENT.md)** — roll this out for everyone, by default
- Marketplaces: https://code.claude.com/docs/en/plugin-marketplaces
- Plugins in the Claude app: https://support.claude.com/en/articles/13837440-use-plugins-in-claude
- Organization administration: https://support.claude.com/en/articles/13837433-manage-plugins-for-your-organization
- Fusion MCP client setup: `Fusion.McpServer/docs/MCP_CLIENT_SETUP.md` (via the `fusion-docs` skill)