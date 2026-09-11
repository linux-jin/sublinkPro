English | [简体中文](CHANGELOG.zh-CN.md)

# Changelog

All notable SublinkPro releases are documented here. The README only keeps a short summary of the latest release; feature-level details live under `docs/features/`.

## [Unreleased]

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
