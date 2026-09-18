<div align="center">
  <img src="webs/src/assets/images/logo.svg" width="280px" />
  
  **✨ Powerful proxy subscription management and conversion ✨**

  <p>
    <img src="https://img.shields.io/github/go-mod/go-version/linux-jin/sublinkPro?style=flat-square&logo=go&logoColor=white" alt="Go Version"/>
    <img src="https://img.shields.io/github/package-json/dependency-version/linux-jin/sublinkPro/react?filename=webs%2Fpackage.json&style=flat-square&logo=react&logoColor=white&color=61DAFB" alt="React Version"/>
    <img src="https://img.shields.io/github/package-json/dependency-version/linux-jin/sublinkPro/@mui/material?filename=webs%2Fpackage.json&style=flat-square&logo=mui&logoColor=white&label=MUI&color=007FFF" alt="MUI Version"/>
    <img src="https://img.shields.io/github/package-json/dependency-version/linux-jin/sublinkPro/vite?filename=webs%2Fpackage.json&style=flat-square&logo=vite&logoColor=white&color=646CFF" alt="Vite Version"/>
  </p>
  <p>
    <img src="https://img.shields.io/github/v/release/linux-jin/sublinkPro?style=flat-square&logo=github&label=Latest" alt="Latest Release"/>
    <img src="https://img.shields.io/github/release-date/linux-jin/sublinkPro?style=flat-square&logo=github&label=Release%20Date" alt="Release Date"/>
  </p>
  <p>
    <img src="https://img.shields.io/badge/GHCR-ghcr.io%2Flinux--jin%2Fsublink--pro-2088FF?style=flat-square&logo=github&logoColor=white" alt="GHCR Image"/>
    <img src="https://img.shields.io/github/actions/workflow/status/linux-jin/sublinkPro/build-release.yml?branch=main&style=flat-square&logo=githubactions&label=Container%20Build" alt="Container Build"/>
  </p>
  <p>
    <img src="https://img.shields.io/github/stars/linux-jin/sublinkPro?style=flat-square&logo=github&label=Stars" alt="GitHub Stars"/>
    <img src="https://img.shields.io/github/forks/linux-jin/sublinkPro?style=flat-square&logo=github&label=Forks" alt="GitHub Forks"/>
    <img src="https://img.shields.io/github/issues/linux-jin/sublinkPro?style=flat-square&logo=github&label=Issues" alt="GitHub Issues"/>
    <img src="https://img.shields.io/github/license/linux-jin/sublinkPro?style=flat-square&label=License" alt="License"/>
  </p>
  <p>
    <a href="https://github.com/linux-jin/sublinkPro/issues">
      <img src="https://img.shields.io/badge/Feedback-Issues-blue?style=flat-square&logo=github" alt="Issues"/>
    </a>
    <a href="https://github.com/linux-jin/sublinkPro/releases">
      <img src="https://img.shields.io/badge/Download-Releases-green?style=flat-square&logo=github" alt="Releases"/>
    </a>
  </p>
</div>

English | [简体中文](README.zh-CN.md)

> [!IMPORTANT]
> This repository, [`linux-jin/sublinkPro`](https://github.com/linux-jin/sublinkPro), is a feature-enhanced community fork of the upstream project [`ZeroDeng01/sublinkPro`](https://github.com/ZeroDeng01/sublinkPro). It regularly incorporates upstream fixes while giving first-class priority to native Loon profiles, the SOCKS5 gateway, WebDAV automation, OpenVPN round-trip support, large-node management, and multi-architecture container delivery.

---

## 🚀 Why This Fork

The additions maintained by this fork are treated as first-class product features, not secondary patches:

1. **Native Loon profile generation** — select a `.lcf` template to generate a complete Loon profile locally, preserving policy groups, rules, plugins, scripts, and MITM sections. If no Loon template is selected, the existing Sub-Store conversion remains available as a compatibility fallback.
2. **SOCKS5 gateway and routing profiles** — expose managed nodes through a health-aware TCP CONNECT gateway with independent profile-bound accounts, multiple listeners, P2C/round-robin selection, retries, cooldown, sticky sessions, and live load visibility.
3. **WebDAV backup automation** — upload backup ZIP files manually or on a five-field cron schedule, browse remote backups, and restore them from the web UI.
4. **OpenVPN YAML round-trip** — import, edit, and export Mihomo/Clash `type: openvpn` entries while retaining supported certificates, keys, and transport fields.
5. **Large-node and container operations** — server-side pagination/search/filtering, reduced rendering overhead, improved grouped selection, and multi-architecture GHCR images.

> [!TIP]
> Start with `ghcr.io/linux-jin/sublink-pro:latest` for stable deployments or `ghcr.io/linux-jin/sublink-pro:dev` to test the newest fork features.

---

## 📖 Project Overview

This repository is maintained by [`linux-jin`](https://github.com/linux-jin) as an enhanced fork of [`ZeroDeng01/sublinkPro`](https://github.com/ZeroDeng01/sublinkPro). The upstream project is itself a deeply refactored continuation of [sublinkX](https://github.com/gooaclok819/sublinkX) and [sublinkE](https://github.com/eun1e/sublinkE). Thanks to the upstream maintainer and all original contributors.

The fork follows this maintenance model:

- 🔄 **Track upstream**: periodically merge upstream fixes, protocol improvements, dependency updates, and general features
- 🧩 **Maintain fork extensions**: native Loon profile generation, the SOCKS5 gateway, WebDAV backup automation, OpenVPN YAML round-trip support, large-node-list optimizations, and GHCR multi-architecture images
- 🐛 **Fork-specific feedback**: report issues related to these extensions in the [linux-jin/sublinkPro issue tracker](https://github.com/linux-jin/sublinkPro/issues)
- 🎨 **Frontend framework**: Based on [Berry Free React Material UI Admin Template](https://github.com/codedthemes/berry-free-react-admin-template)
- ⚡ **Backend stack**: Go + Gin + Gorm
- 🔐 **Default account**: `admin` / `123456`, change it immediately after installation
- 💻 **Demo**: [https://demo.sublinkpro.dpdns.org/](https://demo.sublinkpro.dpdns.org/), username: admin, password: 123456

> [!WARNING]
> ⚠️ This project is not database compatible with the original projects. Don't mix their databases.
>
> ⚠️ Don't use this project, or any derivative of it, for activities that violate the laws and regulations of your location or the location of the users you serve. This project is for personal development, learning, and exchange only.

---

## ✨ Highlights

### Additions maintained by this fork

| Feature | Description | Details |
|:---|:---|:---:|
| 🌙 **Native Loon full-profile output** | Store and select `.lcf` templates, inject locally generated nodes, expand Loon Remote Filters, preserve rules/plugins/scripts/MITM, and fall back to Sub-Store only when no native template is configured | [📖](#-multi-protocol-support) |
| 🧦 **SOCKS5 gateway and routing profiles** | Health-aware TCP CONNECT gateway with independent accounts, multiple listeners, P2C and round-robin routing, sticky sessions, routing profiles, node filters, retries, cooldown, connection governance, and live load monitoring | [📖](docs/features/socks5.md) |
| 💾 **WebDAV backup and scheduling** | Encrypt WebDAV credentials, upload backups manually or by cron, browse remote ZIP files, restore from the page, and handle collection redirects used by services such as TeraCLOUD | [📖](docs/features/backup.md) |
| 🔐 **OpenVPN YAML round-trip** | Import and export Mihomo/Clash `type: openvpn` proxy entries while preserving certificates, keys, transport options, and other supported fields | [📖](#-multi-protocol-support) |
| 🚀 **Large node-list optimization** | Compact server projections, reduced rendering overhead, server-side pagination, search and filtering, plus improved grouped node selection for large installations | — |
| 📦 **Multi-architecture GHCR images** | Automated `amd64`, `arm64`, `arm/v7`, and `386` images; use `latest` for stable releases and `dev` for development builds | [📖](docs/installation.md) |

### Upstream and shared capabilities

| Feature | Description | Details |
|:---|:---|:---:|
| 🏷️ **Smart tag system** | Automatic rule based tagging, no code filtering, IP quality conditions | [📖](docs/features/tags.md) |
| ⚡ **Professional speed test system** | Two stage tests, smart latency measurement, IP quality and unlock checks | [📖](docs/features/speedtest.md) |
| 🔗 **Chain proxy** | Native Dialer-Proxy support, visual configuration, IP quality based node selection | [📖](docs/features/chain-proxy.md) |
| 🤖 **AI template editing** | Generate operation based previews from natural language, review read-only diffs, accept into the editor, then save normally | [📖](docs/features/template-ai.md) |
| ✈️ **Airport management** | Multi format import, scheduled updates, traffic monitoring, one click full refresh | [📖](docs/features/airport.md) |
| 🗂️ **Group ordering** | Drag airport priority within a group to control node order in subscription output | [📖](docs/development.md) |
| 📋 **Subscription sharing** | Multiple links, expiration policies, access statistics | [📖](docs/features/subscription-share.md) |
| 🌐 **Host management** | Domain mappings, DNS configuration, CDN preferred IPs | [📖](docs/features/host.md) |
| ☁️ **Cloudflare Tunnel** | Expose the admin UI without a public IP, with cloudflared managed from the page | [📖](docs/features/cloudflare-tunnel.md) |
| 🤖 **Telegram Bot** | Remote speed tests, subscription management, system monitoring | [📖](docs/features/telegram-bot.md) |
| 📜 **Script system** | Node filtering, content post processing, chained scripts | [📖](docs/script_support.md) |
| 🔔 **Webhooks** | Supports PushDeer, Bark, DingTalk, ServerChan, and other notification platforms | [📖](docs/configuration.md) |
| 🔐 **Security features** | Token authorization, API Key, IP allow and block lists, access logs | [📖](docs/configuration.md) |
| 🦾 **AI agent skill** | Drive the whole system in natural language via the REST API — add nodes, build and share subscriptions, manage airports and templates. Portable `SKILL.md` format, works with any compatible AI agent | [📖](skill-sublinkpro/README.md) |

---

## 🆕 Release Status

**Latest stable release:** `v1.11.0`

**Current `main` / `dev` additions after v1.11.0:**

- Native Loon `.lcf` template management and full-profile generation, with Sub-Store retained as the no-template compatibility fallback.
- A sanitized public `template/loon.lcf`; private certificates, credentials, subscriptions, host mappings, and SSIDs are not included.
- Legacy `.lcf` assignments are automatically repaired, and node-import result dialogs now include complete Chinese and English translations.

The v1.11.0 stable release introduced independent SOCKS5 routing profiles, username-based profile selection, extended Clash/Mihomo field preservation, and front-proxy-aware node speed tests.

See [CHANGELOG.md](CHANGELOG.md) for the complete version history, or browse [GitHub Releases](https://github.com/linux-jin/sublinkPro/releases).

---

## 🚀 Quick Start

### Docker Compose, recommended

> [!IMPORTANT]
> Runtime data is stored in these directories by default. Keep them during upgrades and migrations:
>
> - `./db`: database, configuration files, GeoIP, and other local data
> - `./template`: template files
> - `./logs`: runtime logs

Create `docker-compose.yml`:

```yaml
services:
  sublinkpro:
    image: ghcr.io/linux-jin/sublink-pro
    container_name: sublinkpro
    ports:
      - "8000:8000"
    volumes:
      - "./db:/app/db"
      - "./template:/app/template"
      - "./logs:/app/logs"
    restart: unless-stopped
```

Start the service:

```bash
docker-compose up -d
```

Open `http://localhost:8000` and sign in with `admin` / `123456`.

SQLite is used by default. To switch to MySQL or PostgreSQL, set the database connection through `SUBLINK_DSN`, `dsn:` in the config file, or the `--dsn` command line flag. See [⚙️ Configuration](docs/configuration.md) for examples.

> [!NOTE]
> Even when `SUBLINK_WEB_BASE_PATH` is configured to hide the admin UI entry, API paths (`/api/*`) and subscription or share paths (`/c/*`) stay at the root path. This is a project specific frontend and backend integration rule.

> [!TIP]
> For more install methods, including Docker, one line scripts, updates, and upgrades, see the [📦 Installation Guide](docs/installation.md).

> [!TIP]
> The Docker image includes `cloudflared`. After signing in, open `User Center -> Cloudflare Tunnel`, enter the token, and start it. When auto connect is enabled, the Tunnel connects when the service starts.

### Migrate from SQLite to MySQL / PostgreSQL

If your earlier instance used SQLite and you now want to migrate to MySQL or PostgreSQL, use this flow:

1. Sign in to the old SQLite instance, open **System Backup** from the avatar menu in the upper right, and export `backup.zip`.
2. Configure the `DSN` for MySQL or PostgreSQL in the new instance, and make sure the target database is a fresh empty database.
3. Start the new instance and open `Settings -> Data Migration`.
4. Upload the `backup.zip` exported from the old instance.
5. Choose whether to migrate `AccessKey` and subscription access logs, then start the migration.
6. After migration completes, **manually restart the project instance**, then sign in again and check the data.

> [!IMPORTANT]
> Using `backup.zip` is recommended. Uploading a `.db` file directly migrates database records only and won't restore the template directory.

> [!NOTE]
> If you migrate `AccessKey`, make sure both old and new instances use the same `API encryption key`; otherwise old API Keys may no longer work.

> [!TIP]
> If migration finishes with “N warnings”, open the corresponding “Database Migration” task in `Task Center` to view details.

---

## 📖 Documentation

### 🔧 Installation and Configuration

| Document | Description |
|:---|:---|
| [📦 Installation](docs/installation.md) | Docker, one line scripts, updates, Watchtower automatic updates |
| [⚙️ Configuration](docs/configuration.md) | Environment variables, command line flags, CAPTCHA configuration |
| [📝 Changelog](CHANGELOG.md) | Complete version history and release notes |

### ✨ Feature Guides

| Document | Description |
|:---|:---|
| [🌙 Native Loon output](#-multi-protocol-support) | Complete `.lcf` profile generation with local node injection and Sub-Store compatibility fallback |
| [🧦 SOCKS5 gateway](docs/features/socks5.md) | Local TCP CONNECT gateway with independent accounts, multiple listeners, profiles, P2C/round-robin routing, retries, cooldown, sticky sessions, limits, and monitoring |
| [💾 System backup and WebDAV](docs/features/backup.md) | Manual/scheduled backups, remote listing, and restore |
| [🔐 OpenVPN round-trip](#-multi-protocol-support) | Mihomo/Clash OpenVPN YAML import, editing, and export |
| [🏷️ Smart tag system](docs/features/tags.md) | Automatic rule based tagging, no code filtering, IP quality rules |
| [⚡ Speed test system](docs/features/speedtest.md) | Test design, IP quality checks, unlock checks, parameter tuning |
| [🌍 Unlock checks](docs/features/unlock-check.md) | Streaming and AI availability checks, Provider architecture, extensions |
| [🔗 Chain proxy](docs/features/chain-proxy.md) | Dialer-Proxy, condition based node selection, configuration flow |
| [🤖 AI template editing](docs/features/template-ai.md) | Operation based previews, read-only diff review, accept into editor, normal save |
| [✈️ Airport management](docs/features/airport.md) | Subscription import, scheduled updates, traffic monitoring |
| [📋 Subscription sharing](docs/features/subscription-share.md) | Multiple links, expiration policies, access statistics |
| [🌐 Host management](docs/features/host.md) | Domain mappings, DNS configuration, speed test persistence |
| [☁️ Cloudflare Tunnel](docs/features/cloudflare-tunnel.md) | Create a Tunnel, get a token, configure public access |
| [🤖 Telegram Bot](docs/features/telegram-bot.md) | Command list and setup guide |
| [📜 Script support](docs/script_support.md) | Node filtering, content post processing, function reference |
| [🔐 Multi factor authentication, MFA](docs/features/mfa.md) | TOTP setup, recovery codes, emergency reset flow |

### 👨‍💻 Developers

| Document | Description |
|:---|:---|
| [🛠️ Development Guide](docs/development.md) | Project structure, local development, scheduled task development |
| [🔌 Protocol Extension Guide](docs/development.md#-protocol-extension-guide) | Add a protocol, register capabilities, field metadata, ProtocolDemo example |

---

## 📡 Multi Protocol Support

| Client | Supported protocols |
|:---|:---|
| **v2ray** | base64 common format, without Clash/mihomo specific protocols such as Mieru, Snell, and OpenVPN |
| **clash / mihomo** | ss, ssr, trojan, vmess, vless, hy, hy2, tuic, AnyTLS, Socks5, HTTP, HTTPS, Mieru, Snell, OpenVPN |
| **surge** | ss, trojan, vmess, hy2, tuic, AnyTLS, Snell |
| **loon (native template)** | ss, ssr, trojan, vmess, vless, hy2, AnyTLS, Socks5, HTTP, HTTPS, WireGuard; unsupported protocols are skipped instead of downgraded |

> [!NOTE]
> Loon has two output paths. When a subscription selects a Loon `.lcf` template, SublinkPro generates the complete profile locally and does not call Sub-Store. Without a Loon template, the existing Sub-Store node-conversion path remains available for backward compatibility. The repository ships a sanitized public `template/loon.lcf`; personal certificates and credentials must be managed privately.

> [!NOTE]
> Mieru currently supports Clash/mihomo YAML import and export only. Official Mieru has `mieru://` and `mierus://` share links, but does not define a general URL schema suitable for field by field editing. For raw editing and Clash/mihomo import write back, SublinkPro uses an internal editable form: `mieru://username:password@server:port?...#name`, with port ranges written as `portRange=2090-2099`. v2ray and Surge don't support Mieru in SublinkPro. Subscription output skips that protocol instead of converting it to a downgraded form.

> [!NOTE]
> Snell supports Clash/mihomo and Surge output only. Snell has no official share link schema, so SublinkPro uses an internal editable form for raw editing and Clash/mihomo/Surge import write back: `snell://server:port?psk=xxx&version=3&obfs=http&obfs-host=xxx#name`. v2ray does not support Snell in SublinkPro; subscription output skips it instead of converting it to a downgraded form.

> [!NOTE]
> OpenVPN supports importing and exporting Mihomo/Clash `proxies:` YAML entries. SublinkPro stores them in an internal round-trip form, `openvpn://server:port?...#name`; this is not an official OpenVPN share-link standard and does not add direct `.ovpn` file import. Certificate and private-key material is URL-encoded, not encrypted, in the stored node link, so protect database access and do not share the internal link.

---

## 🖼️ Preview

<details open>
<summary><b>Show or hide screenshots</b></summary>

| | |
|:---:|:---:|
| ![Preview 1](docs/images/1.jpg) | ![Preview 2](docs/images/2.jpg) |
| ![Preview 3](docs/images/3.jpg) | ![Preview 4](docs/images/4.jpg) |
| ![Preview 5](docs/images/5.jpg) | ![Preview 6](docs/images/6.jpg) |
| ![Preview 7](docs/images/7.jpg) | ![Preview 8](docs/images/8.jpg) |
| ![Preview 9](docs/images/9.jpg) | ![Preview 10](docs/images/10.jpg) |
| ![Preview 11](docs/images/11.jpg) | ![Preview 12](docs/images/12.jpg) |

</details>

---

## 📊 Project Stats

<div align="center">

[//]: # (  <img src="https://repobeez.abhijithganesh.com/api/insert/linux-jin/sublinkPro" alt="Repobeez" height="0" width="0" style="display: none"/>)
  
  ![Star History Chart](https://star-history.dera.page/svg?repos=linux-jin/sublinkPro&type=Date)
</div>

---

## 🤝 Contributing and Support

If this project helps you, you are welcome to:

- ⭐ Star the project
- 🐛 Open an [Issue](https://github.com/linux-jin/sublinkPro/issues) for bugs or suggestions
- 🔧 Submit a Pull Request
- 📖 Improve the docs and tutorials

### 🌟 Recommended services

If you need to buy a server, you can support the maintainer through the links below. Purchases through these links may provide commission rewards to the maintainer. Check official pages for exact pricing, promotion eligibility, network performance, and renewal rules.

- **[BandwagonHost](https://bandwagonhost.com/aff.php?aff=19245)**: premium routes, multiple data centers, and CN2 GIA options. Good for high quality route servers, or as a stable landing server with other nodes. Highlights: many VPS locations, CN2 GIA optimized routes, quality route machines.
- **[Vultr](https://www.vultr.com/?ref=8055869)**: many regions, hourly billing, IP and region changes, starting at $2.5 per month. Good for websites, AI service hosting, route servers, and landing servers. Highlights: many regions, hourly billing, low price, stable service, IPv6 only machines.
- **[Aliyun](https://www.aliyun.com/minisite/goods?userCode=brje0cbs)**: 99 RMB per year server offer with the same price for new purchase and renewal. Good for domestic deployment and testing, including newapi and SublinkPro. Registration through the link may provide discounts, including AI related discounts. Highlights: suitable for China based deployment and development testing, 99 RMB per year new purchase and renewal, discount coupons.

### 🙏 Acknowledgements

Thanks to these open source projects and maintainers:

- [ZeroDeng01/sublinkPro](https://github.com/ZeroDeng01/sublinkPro), the upstream project this fork follows and extends
- [sublinkX](https://github.com/gooaclok819/sublinkX) / [sublinkE](https://github.com/eun1e/sublinkE), original projects
- [Berry Free React Admin Template](https://github.com/codedthemes/berry-free-react-admin-template), frontend template
- [Mihomo](https://github.com/MetaCubeX/mihomo), proxy core

---

<div align="center">
  <sub>Fork maintained by <a href="https://github.com/linux-jin">linux-jin</a> · Based on <a href="https://github.com/ZeroDeng01/sublinkPro">ZeroDeng01/sublinkPro</a></sub>
</div>
