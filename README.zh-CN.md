<div align="center">
  <img src="webs/src/assets/images/logo.svg" width="280px" />
  
  **✨ 强大的代理订阅管理与转换工具 ✨**

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
      <img src="https://img.shields.io/badge/问题反馈-Issues-blue?style=flat-square&logo=github" alt="Issues"/>
    </a>
    <a href="https://github.com/linux-jin/sublinkPro/releases">
      <img src="https://img.shields.io/badge/版本下载-Releases-green?style=flat-square&logo=github" alt="Releases"/>
    </a>
  </p>
</div>

[English](README.md) | 简体中文

> [!IMPORTANT]
> 本仓库 [`linux-jin/sublinkPro`](https://github.com/linux-jin/sublinkPro) 是基于上游项目 [`ZeroDeng01/sublinkPro`](https://github.com/ZeroDeng01/sublinkPro) 维护的功能增强分支。项目会持续合并上游修复，同时将 Loon 原生完整配置、SOCKS5 网关、WebDAV 自动备份、OpenVPN 往返转换、大规模节点管理和多架构容器发布作为优先维护能力。

---

## 🚀 为什么选择本增强分支

本分支新增的能力会作为一等功能优先维护，而不是放在上游共有功能之后：

1. **Loon 原生完整配置输出**：为订阅选择 `.lcf` 模板后，由 SublinkPro 本地生成完整 Loon 配置，保留策略组、规则、插件、脚本和 MITM 等 section；未选择模板时仍保留 Sub-Store 兼容转换。
2. **SOCKS5 网关与路由 Profile**：将已管理节点作为健康感知的 TCP CONNECT 出口，支持独立账号绑定 Profile、多监听入口、P2C/轮询、重试、冷却、粘性会话和实时负载观察。
3. **WebDAV 自动备份**：支持手动上传或使用五段 Cron 定时上传备份 ZIP，可浏览远程备份并直接从页面恢复。
4. **OpenVPN YAML 往返转换**：导入、编辑、导出 Mihomo/Clash `type: openvpn` 节点，并保留支持的证书、密钥与传输字段。
5. **大规模节点与容器运维**：服务端分页、搜索和筛选，减少前端渲染开销，改进分组选择，并提供 GHCR 多架构镜像。

> [!TIP]
> 稳定环境使用 `ghcr.io/linux-jin/sublink-pro:latest`；需要测试本分支最新功能时使用 `ghcr.io/linux-jin/sublink-pro:dev`。

---

## 📖 项目简介

本仓库由 [`linux-jin`](https://github.com/linux-jin) 基于上游项目 [`ZeroDeng01/sublinkPro`](https://github.com/ZeroDeng01/sublinkPro) 继续维护和增强。上游项目本身是在 [sublinkX](https://github.com/gooaclok819/sublinkX) / [sublinkE](https://github.com/eun1e/sublinkE) 基础上进行深度重构和优化的项目。感谢上游维护者与所有原始贡献者的付出。

本分支采用以下维护方式：

- 🔄 **持续同步上游**：定期合并上游的问题修复、协议改进、依赖升级和通用功能
- 🧩 **维护分支增强功能**：Loon 原生完整配置、SOCKS5 网关、WebDAV 自动备份、OpenVPN YAML 往返转换、大规模节点列表优化以及 GHCR 多架构镜像
- 🐛 **分支问题反馈**：与这些增强功能相关的问题，请提交到 [linux-jin/sublinkPro Issues](https://github.com/linux-jin/sublinkPro/issues)
- 🎨 **前端框架**：基于 [Berry Free React Material UI Admin Template](https://github.com/codedthemes/berry-free-react-admin-template)
- ⚡ **后端技术**：Go + Gin + Gorm
- 🔐 **默认账号**：`admin` / `123456`（请安装后务必修改）
- 💻 **演示系统**: [https://demo.sublinkpro.dpdns.org/](https://demo.sublinkpro.dpdns.org/) 用户名：admin 密码：123456

> [!WARNING]
> ⚠️ 本项目和原项目数据库不兼容，请不要混用。
>
> ⚠️ 请不要使用本项目以及任何本项目的衍生项目进行违反您以及您所服务用户的所在地法律法规的活动。本项目仅供个人开发和学习交流使用。

---

## ✨ 功能亮点

### 本分支维护的增强功能

| 功能 | 说明 | 详情 |
|:---|:---|:---:|
| 🌙 **Loon 原生完整配置输出** | 管理并选择 `.lcf` 模板，注入本地生成节点、展开 Loon Remote Filter，保留规则/插件/脚本/MITM；仅在未配置原生模板时使用 Sub-Store 兼容转换 | [📖](#-多协议支持) |
| 🧦 **SOCKS5 网关与路由配置** | 健康感知 TCP CONNECT 网关，支持独立账号、多监听入口、P2C/轮询选路、粘性会话、路由 Profile、节点过滤、重试、冷却、连接治理和实时负载监控 | [📖](docs/features/socks5.zh-CN.md) |
| 💾 **WebDAV 备份与定时任务** | 加密保存 WebDAV 凭据，支持手动或 Cron 定时上传、浏览远程 ZIP、页面恢复，并兼容 TeraCLOUD 等服务的集合重定向 | [📖](docs/features/backup.zh-CN.md) |
| 🔐 **OpenVPN YAML 往返转换** | 支持导入和导出 Mihomo/Clash `type: openvpn` 节点，并保留证书、密钥、传输选项及其他已支持字段 | [📖](#-多协议支持) |
| 🚀 **大规模节点管理优化** | 精简服务端列表数据、降低前端渲染开销，支持服务端分页、搜索与筛选，并改进大量节点下的分组选择体验 | — |
| 📦 **GHCR 多架构镜像** | 自动构建 `amd64`、`arm64`、`arm/v7` 和 `386` 镜像；稳定版使用 `latest`，开发测试版使用 `dev` | [📖](docs/installation.zh-CN.md) |

### 上游及共有能力

| 功能 | 说明 | 详情 |
|:---|:---|:---:|
| 🏷️ **智能标签系统** | 自动规则打标签、零代码筛选、支持 IP 质量条件 | [📖](docs/features/tags.zh-CN.md) |
| ⚡ **专业测速系统** | 双阶段测试、智能延迟测量、支持 IP 质量检测与解锁检测 | [📖](docs/features/speedtest.zh-CN.md) |
| 🔗 **链式代理** | Dialer-Proxy 原生支持、可视化配置、支持按 IP 质量选节点 | [📖](docs/features/chain-proxy.zh-CN.md) |
| 🤖 **AI 模板编辑** | 用自然语言生成操作式预览，审阅只读对比，接受到编辑器后再正常保存 | [📖](docs/features/template-ai.zh-CN.md) |
| ✈️ **机场管理** | 多格式导入、定时更新、流量监控、一键全量拉取 | [📖](docs/features/airport.zh-CN.md) |
| 🗂️ **分组排序** | 分组内机场优先级拖拽排序，控制订阅输出中的节点顺序 | [📖](docs/development.zh-CN.md) |
| 📋 **订阅分享** | 多链接管理、过期策略、访问统计 | [📖](docs/features/subscription-share.zh-CN.md) |
| 🌐 **Host 管理** | 域名映射、DNS 配置、CDN 优选 | [📖](docs/features/host.zh-CN.md) |
| ☁️ **Cloudflare Tunnel** | 无公网 IP 暴露管理界面、页面托管 cloudflared | [📖](docs/features/cloudflare-tunnel.zh-CN.md) |
| 🤖 **Telegram Bot** | 远程测速、订阅管理、系统监控 | [📖](docs/features/telegram-bot.zh-CN.md) |
| 📜 **脚本系统** | 节点过滤、内容后处理、多脚本链式执行 | [📖](docs/script_support.zh-CN.md) |
| 🔔 **Webhooks** | 支持 PushDeer、Bark、钉钉、方糖等多平台通知 | [📖](docs/configuration.zh-CN.md) |
| 🔐 **安全特性** | Token 授权、API Key、IP 黑/白名单、访问日志 | [📖](docs/configuration.zh-CN.md) |
| 🦾 **AI agent 技能** | 通过 REST API 用自然语言驱动整个系统——添加节点、创建与分享订阅、管理机场与模板。采用可移植的 `SKILL.md` 格式，任何兼容的 AI agent 均可使用 | [📖](skill-sublinkpro/README.zh-CN.md) |

---

## 🆕 版本状态

**最新稳定版：** `v1.11.0`

**v1.11.0 之后已进入 `main` / `dev` 的功能：**

- 新增 Loon `.lcf` 模板管理和原生完整配置输出；未配置模板时保留 Sub-Store 兼容转换。
- 仓库内置脱敏公开版 `template/loon.lcf`，不包含私人证书、凭据、订阅、Host 映射和 Wi-Fi SSID。
- 自动修复历史 `.lcf` 模板归类与订阅绑定，并补齐节点导入结果弹窗的中英文翻译。

v1.11.0 稳定版主要新增 SOCKS5 独立路由 Profile、用户名选择 Profile、Clash/Mihomo 扩展字段保留，以及支持前置代理的节点测速。

完整版本历史请查看 [CHANGELOG.zh-CN.md](CHANGELOG.zh-CN.md)，正式发布包参见 [GitHub Releases](https://github.com/linux-jin/sublinkPro/releases)。

---

## 🚀 快速开始

### Docker Compose（推荐）

> [!IMPORTANT]
> 运行时数据默认保存在以下目录中，请在升级和迁移时保留：
>
> - `./db`：数据库、配置文件、GeoIP 等本地数据
> - `./template`：模板文件
> - `./logs`：运行日志

创建 `docker-compose.yml`：

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

启动服务：

```bash
docker-compose up -d
```

访问 `http://localhost:8000`，使用默认账号 `admin` / `123456` 登录。

默认使用 SQLite；如需切换到 MySQL 或 PostgreSQL，可通过 `SUBLINK_DSN`、配置文件 `dsn:` 或命令行 `--dsn` 指定数据库连接，示例见 [⚙️ 配置说明](docs/configuration.zh-CN.md)。

> [!NOTE]
> 即使配置了 `SUBLINK_WEB_BASE_PATH` 隐藏管理界面入口，API (`/api/*`) 与订阅/分享访问路径 (`/c/*`) 仍保持在根路径下，这是本项目特有的前后端集成行为。

> [!TIP]
> 更多安装方式（Docker、一键脚本、更新升级等）请参阅 [📦 安装部署指南](docs/installation.zh-CN.md)

> [!TIP]
> Docker 镜像已内置 `cloudflared`。登录后可在 `用户中心 -> Cloudflare Tunnel` 填写 token 并启动；启用自动连接后会随服务启动连接 Tunnel。

### 从 SQLite 迁移到 MySQL / PostgreSQL

如果您早期使用的是 SQLite，现在希望迁移到 MySQL 或 PostgreSQL，建议按以下流程操作：

1. 在旧的 SQLite 实例中登录后台，点击右上角头像菜单中的 **系统备份**，导出 `backup.zip`
2. 在新实例中配置好 MySQL 或 PostgreSQL 的 `DSN`，并确保目标库是一个全新的空库
3. 启动新实例后，进入 `设置 -> 数据迁移`
4. 上传旧实例导出的 `backup.zip`
5. 根据需要选择是否迁移 `AccessKey`、订阅访问日志，然后开始迁移
6. 迁移完成后，**请手动重启项目实例**，再重新登录检查数据

> [!IMPORTANT]
> 推荐使用 `backup.zip` 迁移。直接上传 `.db` 只会迁移数据库记录，不会恢复模板目录。

> [!NOTE]
> 如果迁移了 `AccessKey`，请确保新旧实例使用相同的 `API 加密密钥`；否则旧 API Key 可能无法继续使用。

> [!TIP]
> 如果迁移完成后提示“有 N 条警告”，可以到 `任务中心` 打开对应的“数据库迁移”任务查看详细警告内容。

---

## 📖 文档导航

### 🔧 安装与配置

| 文档 | 说明 |
|:---|:---|
| [📦 安装部署](docs/installation.zh-CN.md) | Docker、一键脚本、更新升级、Watchtower 自动更新 |
| [⚙️ 配置说明](docs/configuration.zh-CN.md) | 环境变量、命令行参数、验证码配置 |
| [📝 更新日志](CHANGELOG.zh-CN.md) | 完整版本历史和发布说明 |

### ✨ 功能详解

| 文档 | 说明 |
|:---|:---|
| [🌙 Loon 原生输出](#-多协议支持) | 完整 `.lcf` 配置生成、本地节点注入与 Sub-Store 兼容回退 |
| [🧦 SOCKS5 网关](docs/features/socks5.zh-CN.md) | 支持独立账号、多监听入口、Profile、P2C/轮询、重试、冷却、粘性会话、连接限制和实时监控的 TCP CONNECT 网关 |
| [💾 系统备份与 WebDAV](docs/features/backup.zh-CN.md) | 手动/定时备份、远程列表和恢复 |
| [🔐 OpenVPN 往返转换](#-多协议支持) | Mihomo/Clash OpenVPN YAML 导入、编辑与导出 |
| [🏷️ 智能标签系统](docs/features/tags.zh-CN.md) | 自动规则打标签、零代码筛选、IP 质量规则 |
| [⚡ 测速系统](docs/features/speedtest.zh-CN.md) | 测速原理、IP 质量检测、解锁检测、参数配置 |
| [🌍 解锁检测](docs/features/unlock-check.zh-CN.md) | 流媒体 / AI 可用区检测、Provider 架构、扩展方式 |
| [🔗 链式代理](docs/features/chain-proxy.zh-CN.md) | Dialer-Proxy、条件选节点、配置流程 |
| [🤖 AI 模板编辑](docs/features/template-ai.zh-CN.md) | 操作式预览、只读对比审阅、接受到编辑器、正常保存 |
| [✈️ 机场管理](docs/features/airport.zh-CN.md) | 订阅导入、定时更新、流量监控 |
| [📋 订阅分享](docs/features/subscription-share.zh-CN.md) | 多链接管理、过期策略、访问统计 |
| [🌐 Host 管理](docs/features/host.zh-CN.md) | 域名映射、DNS 配置、测速持久化 |
| [☁️ Cloudflare Tunnel](docs/features/cloudflare-tunnel.zh-CN.md) | 创建 Tunnel、获取 token、配置公网访问 |
| [🤖 Telegram 机器人](docs/features/telegram-bot.zh-CN.md) | 命令列表、配置指南 |
| [📜 脚本功能](docs/script_support.zh-CN.md) | 节点过滤、内容后处理、函数参考 |
| [🔐 双重验证（MFA）](docs/features/mfa.zh-CN.md) | TOTP 设置、恢复码、应急重置流程 |

### 👨‍💻 开发者

| 文档 | 说明 |
|:---|:---|
| [🛠️ 开发指南](docs/development.zh-CN.md) | 项目结构、本地开发、定时任务开发 |
| [🔌 协议扩展指南](docs/development.zh-CN.md#-新增协议接入指南) | 如何新增协议、注册能力、字段元数据、ProtocolDemo 示例 |

---

## 📡 多协议支持

| 客户端 | 支持协议 |
|:---|:---|
| **v2ray** | base64 通用格式（不输出 Clash/mihomo 专属协议，如 Mieru、Snell、OpenVPN） |
| **clash / mihomo** | ss, ssr, trojan, vmess, vless, hy, hy2, tuic, AnyTLS, Socks5, HTTP, HTTPS, Mieru, Snell, OpenVPN |
| **surge** | ss, trojan, vmess, hy2, tuic, AnyTLS, Snell |
| **loon（原生模板）** | ss, ssr, trojan, vmess, vless, hy2, AnyTLS, Socks5, HTTP, HTTPS, WireGuard；不支持的协议会跳过而不是降级转换 |

> [!NOTE]
> Loon 提供两条输出路径：订阅选择 Loon `.lcf` 模板后，由 SublinkPro 本地生成完整配置，不调用 Sub-Store；未选择模板时，继续保留原有 Sub-Store 节点转换作为向后兼容路径。仓库提供脱敏公开版 `template/loon.lcf`，私人证书和凭据需要自行安全管理。

> [!NOTE]
> Mieru 当前仅支持 Clash/mihomo YAML 导入与导出。Mieru 官方存在 `mieru://` / `mierus://` 分享链接，但未定义适合逐字段编辑的通用 URL schema；SublinkPro 为原始编辑与 Clash/mihomo 导入回写使用内部可编辑形态：`mieru://username:password@server:port?...#name`，端口范围使用 `portRange=2090-2099`。v2ray 与 Surge 当前不支持 Mieru，订阅输出会跳过该协议而不是降级转换。

> [!NOTE]
> Snell 当前仅支持 Clash/mihomo 与 Surge 输出。Snell 没有官方分享链接方案，SublinkPro 为原始编辑与 Clash/mihomo、Surge 导入回写使用内部可编辑形态：`snell://server:port?psk=xxx&version=3&obfs=http&obfs-host=xxx#name`。v2ray 当前不支持 Snell，订阅输出会跳过该协议而不是降级转换。

> [!NOTE]
> OpenVPN 支持导入和导出 Mihomo/Clash `proxies:` YAML 节点。SublinkPro 会将其保存为内部往返格式 `openvpn://server:port?...#name`；这不是 OpenVPN 官方分享链接规范，也不代表已支持直接导入 `.ovpn` 文件。证书与私钥会经过 URL 编码（并非加密）后保存在节点链接中，请保护数据库访问权限且不要分享该内部链接。

---

## 🖼️ 项目预览

<details open>
<summary><b>点击展开/收起预览图</b></summary>

| | |
|:---:|:---:|
| ![预览1](docs/images/1.jpg) | ![预览2](docs/images/2.jpg) |
| ![预览3](docs/images/3.jpg) | ![预览4](docs/images/4.jpg) |
| ![预览5](docs/images/5.jpg) | ![预览6](docs/images/6.jpg) |
| ![预览7](docs/images/7.jpg) | ![预览8](docs/images/8.jpg) |
| ![预览9](docs/images/9.jpg) | ![预览10](docs/images/10.jpg) |
| ![预览11](docs/images/11.jpg) | ![预览12](docs/images/12.jpg) |

</details>

---

## 📊 项目统计

<div align="center">

[//]: # (  <img src="https://repobeez.abhijithganesh.com/api/insert/linux-jin/sublinkPro" alt="Repobeez" height="0" width="0" style="display: none"/>)
  
  ![Star History Chart](https://star-history.dera.page/svg?repos=linux-jin/sublinkPro&type=Date)
</div>

---

## 🤝 贡献与支持

如果这个项目对您有帮助，欢迎：

- ⭐ **Star** 这个项目表示支持
- 🐛 提交 [Issue](https://github.com/linux-jin/sublinkPro/issues) 反馈问题或建议
- 🔧 提交 Pull Request 贡献代码
- 📖 完善文档和使用教程

### 🌟 优质推荐

如果需要购买服务器，可以通过以下链接支持维护者。请注意，点击购买可能会为维护者带来佣金奖励；具体价格、活动资格、线路表现与续费规则请以官方页面为准。

- **[BandwagonHost (搬瓦工)](https://bandwagonhost.com/aff.php?aff=19245)**：精品线路，提供多机房与 CN2 GIA 等线路方案，适合优质线路机。也可以当作稳定落地机结合打野节点使用。亮点：多机房 VPS 方案、可关注 CN2 GIA 优化线路、高质量线路机。
- **[Vultr](https://www.vultr.com/?ref=8055869)**：海量机房可选，按小时结算收费，IP随时更换、地区可随时更换，最低$2.5/月，适合建站和 AI 服务托管等，也可作为线路机和落地机使用。亮点：多机房地区可选、按小时计费、低价稳定、有纯 V6 机器。
- **[阿里云 (Aliyun)](https://www.aliyun.com/minisite/goods?userCode=brje0cbs)**：99 元/年服务器新购续费同价，适合日常建站和测试，可部署 newapi 和 sublinkPro 等服务，通过链接注册可享受优惠（AI 等均可使用折扣）。亮点：适合国内部署与开发测试、99 元/年服务器新购续费同价、可领折扣券。

### 🙏 致谢

感谢以下项目与维护者的开源贡献：

- [ZeroDeng01/sublinkPro](https://github.com/ZeroDeng01/sublinkPro) - 本分支持续跟随并扩展的上游项目
- [sublinkX](https://github.com/gooaclok819/sublinkX) / [sublinkE](https://github.com/eun1e/sublinkE) - 原始项目
- [Berry Free React Admin Template](https://github.com/codedthemes/berry-free-react-admin-template) - 前端模板
- [Mihomo](https://github.com/MetaCubeX/mihomo) - 代理核心

---

<div align="center">
  <sub>增强分支由 <a href="https://github.com/linux-jin">linux-jin</a> 维护 · 基于上游 <a href="https://github.com/ZeroDeng01/sublinkPro">ZeroDeng01/sublinkPro</a></sub>
</div>
