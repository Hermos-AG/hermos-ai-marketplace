# Changelog — unifi-network

[English version: CHANGELOG.md](CHANGELOG.md)

Änderungen am **Eintrag im HERMOS AI Marketplace**. Die Historie des Servers
selbst liegt upstream: [`sirkirby/unifi-mcp` Releases](https://github.com/sirkirby/unifi-mcp/releases).

## [0.25.1] - 2026-08-26

### Hinzugefügt

- Erster Eintrag als `unifi-network`, Kategorie `networking`.
- Release-Kopie des Upstream-Plugins: Manifeste, die fünf Skills
  (`unifi-network`, `network-health-check`, `firewall-auditor`,
  `firewall-manager`, `unifi-network-setup`) sowie die Skripte für
  Voraussetzungen und Umgebungsvariablen.
- Serverstart über `uvx unifi-network-mcp==0.25.1`, gepinnt auf das
  veröffentlichte PyPI-Release.
- Zweisprachige Dokumentation mit Architekturdiagramm, Zugangsdaten-Konzept und
  Sicherheitshinweisen.

### Hinweis

- Zugangsdaten werden ausschließlich als Umgebungsvariablen referenziert
  (`UNIFI_NETWORK_*`); im Repository liegt nichts Geheimes.
