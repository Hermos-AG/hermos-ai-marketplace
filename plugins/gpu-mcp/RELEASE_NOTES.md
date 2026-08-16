# Release Notes — gpu-mcp

[Deutsche Version → RELEASE_NOTES_de.md](RELEASE_NOTES_de.md)

## v0.2.0 — 2026-08-16 · Marketplace release

`gpu-mcp` is now a listed plugin in the **HERMOS Claude Catalog** (business
unit AI-DEV): every HERMOS developer installs it with two commands, and the
bundled binary registers itself — no config files. Because the server runs on
whatever machine the developer has, this release adds a built-in answer to
"is my GPU sufficient?": an 8 GB VRAM minimum (configurable), checkable by
Claude (`gpu_check_requirements`) and from the terminal (`gpu-mcp.exe --check`).

```mermaid
mindmap
  root((gpu-mcp v0.2.0))
    Marketplace
      plugin.json + .mcp.json
      /plugin install HERMOS-local-GPU@hermos
      category ai-dev
    Requirement check
      gpu_check_requirements tool
      --check CLI · exit 0/1
      GPU_MCP_MIN_VRAM_MB · default 8192
      startup preflight log
    Compatibility
      manual desktop-app setup still works
      monitoring works below minimum
      Linux binary for tests
```

**Update:** via marketplace — `/plugin install HERMOS-local-GPU@hermos`
(or reinstall); manual installs — replace `gpu-mcp.exe` while the Claude
desktop app is closed, then start it again.

## v0.1.1 — 2026-08-12 · Quoting fix

Commands containing double quotes (e.g. `"C:\path with spaces\tool.exe" -x`)
now reach `cmd.exe` verbatim instead of being mangled by Go's argument
escaping. Update by replacing `gpu-mcp.exe` while the Claude desktop app is
closed, then starting the app again.

## v0.1.0 — 2026-08-12 · Initial release

First release of the NVIDIA GPU bridge for the Claude desktop app: a single
dependency-free Windows binary (`gpu-mcp.exe`, Go, stdio MCP server) that makes
the local GPU visible and usable for Claude — including cloud Cowork sessions,
where the tools are proxied through the desktop app's device bridge.

### Highlights

```mermaid
mindmap
  root((gpu-mcp v0.1.0))
    Monitoring
      gpu_get_status
      gpu_query_metrics
      gpu_list_processes
    Jobs
      gpu_run_command
        timeout + exit code
        hidden console windows
    Robustness
      nvidia-smi auto-detection
      200 KB output cap
      WaitDelay against pipe hangs
    Zero dependencies
      single .exe
      no Python / Node
```

- **Three read-only monitoring tools** — full `nvidia-smi` report, compact CSV
  metrics, and CUDA process list, each with proper MCP annotations
  (`readOnlyHint`) so clients can treat them as safe.
- **One job tool** — `gpu_run_command` runs a command via `cmd /C` with a
  configurable timeout (default 60 s, max 1 h), returns combined output and
  exit code, and is annotated as destructive so approval prompts stay strict.
- **Windows polish** — `nvidia-smi` is auto-detected (PATH, System32, legacy
  driver folder, `GPU_MCP_NVIDIA_SMI` override); spawned processes use
  `CREATE_NO_WINDOW`, so no console windows flash.
- **Robust protocol loop** — newline-delimited JSON-RPC 2.0 with graceful
  handling of parse errors, unknown methods/tools, and `resources/prompts`
  probes; `WaitDelay` prevents hangs when timed-out shells leave children
  holding the output pipe.
- **Tested without hardware** — ships with a fake `nvidia-smi` and a 12-step
  protocol session covering all tools, error paths, and the timeout behavior.

### Known limitations

- Read-only tools cover NVIDIA GPUs only (no AMD/Intel).
- A timed-out command's child processes may keep running (the shell is killed,
  its children are not).
- Windows amd64 binary included; other platforms build from source.
