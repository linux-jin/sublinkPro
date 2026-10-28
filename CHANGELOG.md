English | [简体中文](CHANGELOG.zh-CN.md)

# Changelog

All notable SublinkPro releases are documented here. The README only keeps a short summary of the latest release; feature-level details live under `docs/features/`.

## [Unreleased]

### Node classification

- Extract a shared node-condition editor and evaluator for tag automation and smart groups, preserving legacy tag JSON, event triggers, exclusivity and tag add/remove behavior.
- Add optional candidate and membership conditions to smart groups, keeping check scope separate from passing membership and preserving legacy country/name matching, scheduling and subscription references.
- Strictly validate new fields/operators/values and reject check-dependent candidate fields. Report additional membership exclusions, preserve omitted fields on legacy PUT, support explicit clearing and restore backups without the new columns.

- Consolidate tag management and smart groups under **Node Classification**, with separate Tags and Smart Groups tabs. Keep tag lists and automation rules inside Tags.
- Persist the selected top-level tab in the URL for refresh, deep links and browser history. Redirect both old management URLs while preserving unrelated query parameters and fragments.
- Keep tag sub-tab labels visible on mobile and add accessible tab/panel associations. Only mount the active top-level page to avoid polling by an inactive smart-group page.
- Preserve all existing APIs, tag/group settings, automatic checks and subscription references; this change adds neither automatic member tagging nor tag-based candidate filtering.

### Updates page

- Remove the promotional/support card (Star, sponsorship links and affiliate recommendations) from the fork's Updates page. Keep the release log and its refresh control unchanged.

## [1.16.0] - 2026-10-27

### HTTP and Mixed proxy gateway listeners

- Add per-listener SOCKS5, HTTP and Mixed inbound modes while retaining existing API paths and default SOCKS5 behavior. Old listener updates preserve omitted protocol values.
- Share independent-account/legacy profile authentication, routing, health, retries and sticky sessions; support plain HTTP forwarding and HTTPS CONNECT without decryption or an extra proxy process.
- Authenticate and route each keep-alive request separately, strip proxy credentials/hop headers, preserve pre-read tunnel bytes, and restrict retries to dialing before application data is sent. Stream uploads/responses concurrently and release incomplete uploads when an origin rejects, disconnects, or returns invalid headers.
- Include HTTP/CONNECT in connection limits, timeout/disconnect controls, payload statistics and protocol-aware monitoring. Reject unauthenticated non-loopback listeners and unsupported upgrades.
- Rename the UI to Proxy gateway, add protocol selectors and username compatibility hints, and update bilingual guides/API examples and regression tests.

### Detection-center and smart-group UI cleanup

- Align the center with the upstream PR: keep Plans and Automatic Jobs only, removing its added history/results layer, dedicated endpoint, component, test and unused translations. Existing node/task pages are unchanged.
- Preserve smart-group candidate job scopes, single/selected-node checks, schedules, plan-reference protection and subscription membership rules.
- Remove internal-mechanism explanations, repeated banners and verbose copy from smart-group pages/forms. Use Included / Excluded for group membership; retain source airports, exclusion reasons, field constraints, timezone and real errors.
- Unify bilingual detection-plan labels, fix nested candidate-dialog headings and the subscription smart-group loader's Hook dependency warning. Subscribe the smart-group page to task completions only rather than every progress update.

### Detection-center fixes

- Do not treat context cancellation after a successful subscription completion as failure; preserve usage, automatic tags and subsequent checks. Real loader failures still stop checks.
- Validate scope for paused original schedules with empty Cron so an empty filter cannot expand to all nodes.
- Automatic-job five-field Cron is validated by the server; named months/weekdays and single-value steps are not blocked by the browser's numeric preview.

## [1.15.0] - 2026-10-22

### Detection center

- Split detection plans, independent automatic jobs, run history and latest node results into separate tabs. Plans expose latency, optional download speed, exit IP/location, IP quality and metadata-driven AI/streaming modules, while retaining detailed request and traffic options.
- Add reusable-plan scheduled/airport-update jobs with explicit airport/group/tag or smart-group candidate scopes, server-timezone Cron, optional synchronous airport pre-updates, per-job overlap protection and coalesced airport follow-ups. Empty scopes never expand to all nodes; failed pre-updates stop detection.
- Map existing profile schedules as compatibility jobs without copying schedules or changing profile IDs. Preserve single/selected-node detection, airport and smart-group references, legacy APIs and Telegram scheduling. Protect referenced plans from deletion and preserve schedules during plan-only edits.
- Add paginated current results with source airports, provider evidence/times and dynamic membership reasons; distinguish these from historical run summaries and from final subscription inclusion. AI probe results do not automatically become smart-group eligibility rules or account/model verification.
- Include independent jobs and referenced candidate groups in database restore, including legacy-backup cleanup. Validate new job/configuration paths and add scope, queue, pre-update, compatibility, paging, reference and restore regression coverage.

## [1.14.1] - 2026-10-21

### Large-node performance

- Use bounded top-K selection for shallow node pages instead of copying and sorting every matching node. Totals remain exact; equal sort values now use ascending ID for deterministic pagination. Unpaginated consumers remain supported.
- Defer unlock-summary JSON parsing until cheap filters and name/link search need it; reuse normalized filter rules. Country/group/source counts read maintained index counts instead of copying all nodes.
- Isolate node-list rendering from task progress, stabilize list callbacks, cancel superseded list requests, and coalesce bursts of completion-triggered refreshes. Column-resize previews no longer rerender the entire table on every mouse frame.
- Smart-group member previews reuse one candidate snapshot and freshness cutoff rather than scanning candidates twice; response fields and eligibility rules remain unchanged.

### Upstream sync

- Merged upstream through `f361753` (2026-10-21): preserve VMess/VLESS WebSocket headers beyond Host during Clash/Mihomo round trips, and recognize Gemini availability markers in both escaped inline JavaScript and unescaped JSON.
- Kept VMess links with object-shaped `headers` compatible alongside the new JSON-string extension; unknown extension types do not reject the entire node.
- Updated regexp2 to `2.9.0`, GeoIP2 to `2.5.0`, React/React DOM to `19.3.1`, and related Go, i18n and development dependencies. Fork smart groups, native Loon behavior and GHCR workflows remain intact.

## [1.14.0] - 2026-10-21

### Country-matching fixes

- Fixed smart-group country abbreviations matching protocol names (`TR` → `trojan`, `ES` → `vmess`/`vless`, `RO` → `trojan`), including existing default rules and name-inferred country display.
- Recognize exact selected flag emoji without requiring saved country rules; keep country/keyword OR behavior and do not alter protocol or persisted node country fields.

### Upstream sync

- Merged upstream through `fc37921` (2026-10-12), including the previously contributed OpenVPN, WebDAV and SOCKS5 features, without replacing fork-specific smart groups, Loon output, large-node-list optimizations or GHCR publishing.
- Preserved complete Hysteria2 userpass credentials and standardized TUIC exports on `allow_insecure`, while accepting legacy `insecure` links.
- Hardened WebDAV collection redirects to accept only same-origin locations without following them. Database restore now preserves target SOCKS5 settings/accounts/listeners/profiles and stops/reloads the gateway against restored nodes.
- Updated mihomo to `1.19.32`, regexp2 to `2.8.4`, Vite to `8.3.4`, and related Go/frontend dependencies. Corrected SOCKS5 limit descriptions to per-listener limits and retained a single gateway menu entry.

### Added

- Added a manual dev build workflow trigger to retry image publication when a matching push build is missing.

- Added opt-in smart-group five-field Cron candidate checks, saved default profiles, server-timezone hints, running status and last-trigger/next-run display.
- Added airport-refresh linkage limited to matching candidates from the updated airport, including unchanged/failed nodes, with same-profile batch deduplication and queued follow-up updates.

### Changed

- Protected groups against overlapping manual/Cron/update checks and referenced profiles against deletion or incompatible TCP edits. Results dynamically decide membership without OpenClash control or moving nodes.

### Fixed

- Reject zero-step and invalid numeric Cron before preview evaluation to prevent browser freezes with `*/0`.


## [1.13.2] - 2026-09-30

### Changed

- Synced upstream dependency updates: mihomo, Go networking and database libraries, React, React Router, Monaco Editor, Vite, and frontend tooling.

## [1.13.1] - 2026-09-28

### Added

- Smart-group candidates now display their original source airport separately from their original group, and distinguish nodes currently passing this smart group's checks from excluded candidates. Final subscriptions may still filter or deduplicate eligible nodes.

### Changed

- Redesigned the smart-group candidate dialog as compact responsive rows with one content scroller, readable status/airport details, clear summary counts and test actions, and mobile-friendly pagination.

## [1.13.0] - 2026-09-23

### Added

- Added independent smart groups that select nodes by landing country across all original groups without moving or duplicating nodes. A node can belong to multiple smart groups.
- Added configurable latency, download-speed, and test-result freshness requirements; nodes with successful latency checks are included dynamically (and a successful speed check when minimum speed is positive).
- Added a candidate-check action using existing node-check profiles to test all nodes in the selected countries, including previously failed or untested candidates.
- Added smart-group management, member preview, and Dynamic/Mixed subscription selection with matching subscription previews.

### Changed

- Made smart groups selectable directly from subscription add/edit forms, with automatic Dynamic/Mixed mode switching and a one-click create-subscription entry on each smart group. Subscription lists now count linked smart groups separately, and smart-group choices load independently of unrelated options.
- Expanded the smart-group country chooser beyond countries already present on nodes (including Philippines), added optional node-name keyword and source-group filters, and showed candidate counts for diagnosing empty groups.
- A zero minimum speed now accepts successful TCP latency checks; positive speed thresholds still require download-speed results. Batch country fill now updates the node cache immediately.
- Moved smart-group candidate inspection and testing into a paginated dialog showing untested, failed, and healthy nodes with per-node exclusion reasons; added whole-group and single-node checks.

### Fixed

- Fixed a smart-group page render crash while candidate data is still loading by guarding pagination against an empty response.
- Smart groups match selected countries by stored landing country or country names/codes in node names, even when the two disagree; an optional name keyword is a further OR alternative. Source-group and health requirements remain AND restrictions. Name clues never overwrite landing-country metadata.
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
