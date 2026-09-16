English | [简体中文](CHANGELOG.zh-CN.md)

# Changelog

All notable SublinkPro releases are documented here. The README only keeps a short summary of the latest release; feature-level details live under `docs/features/`.

## [Unreleased]

No unreleased changes yet.

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
