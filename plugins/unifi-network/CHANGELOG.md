# Changelog — unifi-network

[Deutsche Fassung: CHANGELOG_de.md](CHANGELOG_de.md)

Changes to the **listing in the HERMOS AI Marketplace**. The server's own history
is upstream: [`sirkirby/unifi-mcp` releases](https://github.com/sirkirby/unifi-mcp/releases).

## [0.25.1] - 2026-08-26

### Added

- First listing as `unifi-network`, category `networking`.
- Release copy of the upstream plugin: manifests, the five skills
  (`unifi-network`, `network-health-check`, `firewall-auditor`,
  `firewall-manager`, `unifi-network-setup`) and the prerequisite/env scripts.
- Server started as `uvx unifi-network-mcp==0.25.1`, pinned to the published
  PyPI release.
- Bilingual documentation with architecture diagram, credential handling and
  security notes.

### Note

- Credentials are only referenced as environment variables (`UNIFI_NETWORK_*`);
  nothing secret is stored in this repository.
