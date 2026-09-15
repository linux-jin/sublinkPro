English | [简体中文](socks5.zh-CN.md)

# SOCKS5 Gateway

SublinkPro can expose selected stored proxy nodes through a local SOCKS5 gateway. This is separate from the existing SOCKS5 node import/export support.

## What is included

- TCP `CONNECT` only
- IPv4, IPv6, and domain targets
- Optional username/password authentication (enabled by default)
- Best-node, random-node, round-robin, or specific-node selection
- Candidate node pools filtered by one or more groups, sources, protocols, and countries/regions; fields combine with AND and values inside one field combine with OR
- Mihomo outbound adapter pooling keyed by node ID and link hash, with an idle LRU cap
- Configurable per-request retry (up to 5 candidate nodes), dial timeout, and failed-node cooldown
- Optional fallback from a specific node; disabled by default to keep specific routing strict
- Global active-connection limit (`maxConnections`, default `256`, range `1-10000`)
- Per-client-IP active-connection limit (`maxConnectionsPerClient`, default `32`, range `1..maxConnections`)
- Idle timeout (`idleTimeoutSeconds`, default `600`, range `0-86400`; `0` disables it)
- Maximum connection duration (`maxConnectionDurationSeconds`, default `0`, range `0-604800`; `0` disables it)
- Administrator-only live monitoring with active/total/success/failure counters and upload/download byte counters
- Administrator controls to disconnect one active connection or all active connections
- Optional active node health checks with a fixed HTTPS probe target, bounded concurrency, latency reporting, exponential failure cooldown, and manual probe control
- Node health visualization for healthy, unhealthy, checking, and unknown nodes
- Health-aware routing: active-probe failures are removed from normal candidate sets, and `best` prioritizes fresh active-probe latency before applying the retry limit
- Runtime start/stop when settings are saved; no process restart is required

Retries happen before the SOCKS5 success reply is sent. Active-probe failures are skipped while another routable candidate exists; if every candidate is unhealthy, routing fails open and keeps candidates available for recovery. A successful real connection clears the routing exclusion without erasing the most recent active-probe latency. A failed adapter is discarded, and adapters are closed when the gateway is stopped or reapplied. If every candidate is cooling down, the node whose cooldown expires first is probed so the pool cannot remain permanently unavailable.

UDP `ASSOCIATE`, `BIND`, sticky sessions, multi-user routing, and multi-port listeners are not included yet. Phase 2.2 adds administrator-only live connection monitoring, aggregate counters, traffic byte counters, and connection termination controls.

## Configure

1. Sign in as an administrator.
2. Open **System Settings → SOCKS5 gateway** from the sidebar.
3. Keep **Listen address** as `127.0.0.1` for local-only access.
4. Choose a port (default `1080`) and node selection strategy.
5. Optionally restrict the candidate node pool by groups, sources, protocols, and countries/regions. Empty fields do not restrict the pool. Specific-node mode uses these filters only for fallback candidates.
6. Configure maximum attempts, per-node dial timeout, and failed-node cooldown.
7. Configure global and per-client connection limits.
8. Configure idle timeout and maximum connection duration; set either value to `0` to disable that limit.
9. For specific-node routing, enable fallback only if switching to another node is acceptable.
10. Keep authentication enabled and set a username/password.
11. Optionally enable active node health checks, then choose a 10-3600 second interval and 1-30 second per-node timeout.
12. Save the settings. Monitoring, health probing, and connection termination controls are administrator-only.

The gateway is disabled by default. Passwords are encrypted with the instance API encryption key. Settings responses expose `hasPassword` and, when present, `maskedPassword`; plaintext `password` is never returned. Omitting `password` preserves the saved password, while `clearPassword: true` clears it.

## Use

```bash
curl --proxy socks5h://127.0.0.1:1080 \
  -U "admin:your-password" \
  https://api.ipify.org
```

If you bind to `0.0.0.0` or another non-loopback address, authentication is mandatory. Expose the port only behind a firewall or private network and use a strong password. The backend rejects unauthenticated non-loopback listeners to prevent accidental open proxies.

## API

All endpoints below require an authenticated administrator. The destructive `POST`/`DELETE` operations are also restricted in demo mode.

- `GET /api/v1/settings/socks5` — read public gateway settings; the plaintext password is never returned.
- `POST /api/v1/settings/socks5` — save and apply settings. JSON fields include `enabled`, `listenAddress`, `port`, `username`, optional `password`, `clearPassword`, `nodeId`, `selection` (`best`, `random`, `round_robin`, or `specific`), `requireAuth`, `maxAttempts` (1-5), `dialTimeoutSeconds` (1-120), `failureCooldownSeconds` (0-3600), `specificFallback`, `maxConnections` (1-10000), `maxConnectionsPerClient` (1..maxConnections), `idleTimeoutSeconds` (0-86400), `maxConnectionDurationSeconds` (0-604800), `healthCheckEnabled`, `healthCheckIntervalSeconds` (10-3600), `healthCheckTimeoutSeconds` (1-30), `candidateGroups`, `candidateSources`, `candidateProtocols`, and `candidateCountries`. Candidate pool fields are string arrays; fields are combined with AND and values within each field use OR. Optional fields may be omitted to preserve saved values for legacy clients.
- `POST /api/v1/settings/socks5/stop` — stop the listener without changing saved settings.
- `GET /api/v1/settings/socks5/status` — return `data.config`, `data.stats`, and `data.health` with sweep state and per-node health details.
- `POST /api/v1/settings/socks5/health/probe` — start an immediate health sweep. An already-running sweep is reused instead of starting a duplicate.
- `GET /api/v1/settings/socks5/connections` — return the current active connection array. Each item includes `id`, `clientAddress`, `target`, `nodeName`, `phase`, `startedAt`, `lastActivity`, `uploadBytes`, and `downloadBytes`.
- `DELETE /api/v1/settings/socks5/connections/:id` — disconnect one active connection by ID.
- `DELETE /api/v1/settings/socks5/connections` — disconnect all active connections.

The status counters and active connection list belong to the current gateway server lifetime; stopping or reapplying settings creates a new registry and resets them.

Active health checks follow the current selection: strict specific-node mode probes only that node, specific-node fallback probes that node plus the configured fallback pool, and other strategies probe the configured candidate pool. Probes use a fixed Cloudflare HTTPS connectivity endpoint; administrators cannot configure an arbitrary probe URL. Sweeps use at most four workers and never overlap. Status responses include aggregate counts and at most the first 200 sorted node records to keep three-second monitoring polls bounded. Failed probes retire the failed adapter lease safely and apply exponential cooldown capped at one hour.
