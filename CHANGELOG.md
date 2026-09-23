English | [简体中文](CHANGELOG.zh-CN.md)

# Changelog

All notable SublinkPro releases are documented here. The README only keeps a short summary of the latest release; feature-level details live under `docs/features/`.

## [Unreleased]

### Added

- Added independent smart groups that select nodes by landing country across all original groups without moving or duplicating nodes. A node can belong to multiple smart groups.
- Added configurable latency, download-speed, and test-result freshness requirements; nodes with successful latency checks are included dynamically (and a successful speed check when minimum speed is positive).
- Added a candidate-check action using existing node-check profiles to test all nodes in the selected countries, including previously failed or untested candidates.
- Added smart-group management, member preview, and Dynamic/Mixed subscription selection with matching subscription previews.

### Changed

- Expanded the smart-group country chooser beyond countries already present on nodes (including Philippines), added optional node-name keyword and source-group filters, and showed candidate counts for diagnosing empty groups.
- A zero minimum speed now accepts successful TCP latency checks; positive speed thresholds still require download-speed results. Batch country fill now updates the node cache immediately.

- Moved smart-group candidate inspection and testing into a paginated dialog showing untested, failed, and healthy nodes with per-node exclusion reasons; added whole-group and single-node checks.

### Fixed

- Smart groups now use configured country-name rules when a node has no stored landing country, so country candidates can be tested rather than remaining permanently empty. Stored landing country still takes precedence; no node metadata is overwritten.
- Added exclusion-reason counts and explicit name-inferred labels to distinguish no matching candidates from failed checks, thresholds, and stale results.

## [1.12.0] - 2026-09-18

### Added

- Added independent SOCKS5 accounts with encrypted per-account passwords and direct routing-profile binding.
- Added multiple SOCKS5 listeners with stable IDs, per-listener default profiles, authentication overrides, aggregate monitoring, and listener-scoped connection IDs.
- Added first-class Loon `.lcf` template management and native full-profile generation, including local node injection, Remote Filter expansion, and preservation of Loon rules, plugins, scripts, and MITM sections.
- Added a sanitized public `template/loon.lcf` without private certificates, credentials, subscriptions, host mappings, or SSIDs.

### Changed

- Prioritized fork-maintained features in the README and clarified that native Loon output is preferred when a Loon template is configured.
- Kept Sub-Store as the compatibility fallback for Loon subscriptions that do not select a native template.
- Automatically repair legacy `.lcf` template categories and subscriptions that previously stored a Loon template in the Clash template slot.

### Fixed

- Added the missing Chinese and English translations for the node-import result dialog.

## [1.11.0] - 2026-09-17

### Added

- Added SOCKS5 routing profiles with independent candidate pools, selection strategies, retries, cooldown, and sticky-session settings.
- Added backward-compatible username routing: `username` uses the default profile, `username@profile` selects a profile, and `username@profile.account` adds account-scoped affinity.
- Added routing-profile CRUD, profile-aware connection monitoring, profile-filtered node load views, and union health probing across enabled profiles.

### Fixed

- Preserved unmodeled Clash/Mihomo proxy fields, such as extended `smux` options, across YAML import, metadata-only edits, subscription refreshes, and Clash/Mihomo export.
- Made latency and speed tests honor a node's configured `DialerProxyName` front proxy.

### Changed

- Synchronized the latest upstream dependency updates and clarified the relationship between this enhanced fork and `ZeroDeng01/sublinkPro` in the README.

## [1.10.0] - 2026-09-17

### Added

- Added an administrator-only, server-paginated SOCKS5 node load table with search, health filtering, sorting, latency-source details, and live P2C scores.
- Added an API and UI action to reset per-node success, failure, and last-selection counters without interrupting active connections.

## [1.9.0] - 2026-09-17

### Added

- Added a SOCKS5 `smart` selection strategy using power-of-two choices (P2C), active connection load, health latency, and consecutive failure penalties.
- Added bounded in-memory per-node runtime routing counters with exact active-connection lifecycle tracking.

## [1.8.0] - 2026-09-16

### Added

- Added in-memory SOCKS5 sticky sessions keyed by client IP or authenticated SOCKS5 username, with a configurable sliding TTL.
- Sticky leases automatically fall back when their node becomes unhealthy, enters cooldown, changes link, or leaves the configured candidate pool.

## [1.7.0] - 2026-09-16

### Added

- Added SOCKS5 candidate node pool filters for groups, sources, protocols, and countries/regions.
- Added searchable, grouped, server-paginated selection for the SOCKS5 specific outbound node.

### Changed

- SOCKS5 active health results now participate in routing; `best` prioritizes measured probe latency, and probe-failed nodes are skipped while alternatives exist.

## [1.6.0] - 2026-09-15

### Added

- Added optional active SOCKS5 node health sweeps, per-node status and latency reporting, exponential cooldown, and manual probe controls.
- Added upload/download totals, per-connection traffic, duration, and idle time to the SOCKS5 monitor.

### Changed

- SOCKS5 shutdown now cancels connections that are still in the handshake phase.

## [1.5.0] - 2026-09-15

### Added

- Added configurable global and per-client SOCKS5 connection limits.
- Added idle and maximum-duration connection timeouts.
- Added administrator-only live SOCKS5 monitoring with active connection details, aggregate counters, and traffic byte statistics.
- Added administrator controls to disconnect one active SOCKS5 connection or all active connections.
- Added status, active-connections, and connection-control endpoints for SOCKS5 operations.

### Security

- SOCKS5 connection and disconnect management APIs require administrator authentication.
- Connection identifiers and traffic details are not exposed to non-administrator users.

See [SOCKS5 Gateway](docs/features/socks5.md).

## [1.4.0] - 2026-09-14

### Added

- Added SOCKS5 round-robin routing, configurable retry attempts, per-node dial timeout, failed-node cooldown, and optional specific-node fallback.
- Added reusable mihomo adapter pooling with link-change invalidation and lifecycle cleanup.

### Performance

- Reduced node-management response size with an optional compact list projection.
- Reduced repeated node-list requests and unnecessary rendering work for large node collections.

### Security

- Reject unauthenticated SOCKS5 listeners bound to non-loopback addresses.

## [1.3.0] - 2026-09-11

### Added

- Added an administrator-configurable local SOCKS5 gateway.
- Added TCP `CONNECT` support for IPv4, IPv6, and domain targets.
- Added username/password authentication, best-node selection, random-node selection, and specific-node selection.
- Added runtime start/stop controls and a dedicated **System Settings → SOCKS5 gateway** sidebar page.
- Added encrypted storage for the SOCKS5 password and administrator-only settings APIs.

### Security

- The gateway is disabled by default.
- The default listener is `127.0.0.1:1080`.
- UDP `ASSOCIATE` and `BIND` are intentionally not included in this phase.

See [SOCKS5 Gateway](docs/features/socks5.md).

## [1.2.21] - 2026-09-10

### Added

- Added scheduled WebDAV backup with a five-field cron expression.
- Scheduled runs reuse the manual ZIP upload flow, show progress as `webdav_backup`, and skip overlapping uploads.
- Restores preserve the current WebDAV connection and schedule settings.

### Fixed

- WebDAV directory creation now sends `MKCOL` with a trailing slash and accepts same-host collection redirects, fixing TeraCLOUD-style uploads.

See [System Backup and WebDAV](docs/features/backup.md).

## [1.2.20] - 2026-09-10

### Added

- Added the administrator-only System Backup settings page.
- Added WebDAV backup ZIP upload, remote listing, download, and restore through the existing database migration workflow.
- Added encrypted WebDAV passwords, HTTPS/private-network safeguards, archive limits, ZIP bomb/path traversal/symlink defenses, redirect blocking, cleanup, and restore task tracking.

### Fixed

- Fixed administrator detection for `/v1/users/me` responses containing `roles: ["ADMIN"]`.

See [System Backup and WebDAV](docs/features/backup.md).

## [1.2.19] - 2026-09-09

### Added

- Added Clash/mihomo `type: openvpn` node import and export support.
- Preserved OpenVPN transport, cipher/auth, credentials, PEM blocks, TLS static keys, keepalive, DNS, IP stack, and `dialer-proxy` fields.
- Added the internal lossless `openvpn://server:port?...#name` representation for editing and round trips.

### Compatibility

- OpenVPN nodes are emitted for Clash/mihomo. Unsupported v2ray and Surge outputs skip them.
- Direct `.ovpn` file import is not included.

See the OpenVPN notes in [Airport Management](docs/features/airport.md) and [Subscription Sharing](docs/features/subscription-share.md).

> Versions before `v1.2.19` remain available through the repository history and Git tags.
