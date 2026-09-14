English | [简体中文](socks5.zh-CN.md)

# SOCKS5 Gateway

SublinkPro can expose selected stored proxy nodes through a local SOCKS5 gateway. This is separate from the existing SOCKS5 node import/export support.

## What is included

- TCP `CONNECT` only
- IPv4, IPv6, and domain targets
- Optional username/password authentication (enabled by default)
- Best-node, random-node, round-robin, or specific-node selection
- Mihomo outbound adapter pooling keyed by node ID and link hash, with an idle LRU cap
- Configurable per-request retry (up to 5 candidate nodes), dial timeout, and failed-node cooldown
- Optional fallback from a specific node; disabled by default to keep specific routing strict
- Runtime start/stop when settings are saved; no process restart is required

Retries happen before the SOCKS5 success reply is sent. A failed adapter is discarded, and adapters are closed when the gateway is stopped or reapplied. If every candidate is cooling down, the node whose cooldown expires first is probed so the pool cannot remain permanently unavailable.

UDP `ASSOCIATE`, `BIND`, sticky sessions, multi-user routing, multi-port listeners, and a full traffic statistics panel are not included yet.

## Configure

1. Sign in as an administrator.
2. Open **System Settings → SOCKS5 gateway** from the sidebar.
3. Keep **Listen address** as `127.0.0.1` for local-only access.
4. Choose a port (default `1080`) and node selection strategy.
5. Configure maximum attempts, per-node dial timeout, and failed-node cooldown.
6. For specific-node routing, enable fallback only if switching to another node is acceptable.
7. Keep authentication enabled and set a username/password.
8. Save the settings.

The gateway is disabled by default. Passwords are encrypted with the instance API encryption key and are never returned by the settings API.

## Use

```bash
curl --proxy socks5h://127.0.0.1:1080 \
  -U "admin:your-password" \
  https://api.ipify.org
```

If you bind to `0.0.0.0` or another non-loopback address, authentication is mandatory. Expose the port only behind a firewall or private network and use a strong password. The backend rejects unauthenticated non-loopback listeners to prevent accidental open proxies.

## API

The administrator-only endpoints are:

- `GET /api/v1/settings/socks5`
- `POST /api/v1/settings/socks5`
- `POST /api/v1/settings/socks5/stop`

The response includes `running` and `boundAddress`; it never includes the plaintext password.
