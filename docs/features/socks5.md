English | [简体中文](socks5.zh-CN.md)

# SOCKS5 Gateway (Phase One)

SublinkPro can expose selected stored proxy nodes through a local SOCKS5 gateway. This is separate from the existing SOCKS5 node import/export support.

## What is included

- TCP `CONNECT` only
- IPv4, IPv6, and domain targets
- Optional username/password authentication (enabled by default)
- Best-node, random-node, or specific-node selection
- Existing mihomo outbound adapters are reused for the selected node
- Runtime start/stop when settings are saved; no process restart is required

UDP `ASSOCIATE`, `BIND`, sticky sessions, multi-user routing, and Resin-style health-pool failover are not included in this phase.

## Configure

1. Sign in as an administrator.
2. Open **System Settings → SOCKS5 gateway** from the sidebar.
3. Keep **Listen address** as `127.0.0.1` for local-only access.
4. Choose a port (default `1080`) and node selection strategy.
5. Keep authentication enabled and set a username/password.
6. Save the settings.

The gateway is disabled by default. Passwords are encrypted with the instance API encryption key and are never returned by the settings API.

## Use

```bash
curl --proxy socks5h://127.0.0.1:1080 \
  -U "admin:your-password" \
  https://api.ipify.org
```

If you bind to `0.0.0.0` or another non-loopback address, expose the port only behind a firewall or private network and use a strong password. The gateway is an outbound relay and can become an open proxy if it is exposed without authentication.

## API

The administrator-only endpoints are:

- `GET /api/v1/settings/socks5`
- `POST /api/v1/settings/socks5`
- `POST /api/v1/settings/socks5/stop`

The response includes `running` and `boundAddress`; it never includes the plaintext password.
