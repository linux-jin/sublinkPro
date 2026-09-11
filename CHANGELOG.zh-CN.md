[English](CHANGELOG.md) | 简体中文

# 更新日志

这里记录 SublinkPro 的版本更新。README 只保留最新版本摘要，功能级细节放在 `docs/features/` 下。

## [未发布]

暂无未发布改动。

## [1.3.0] - 2026 年 9 月 11 日

### 新增

- 新增管理员可配置的本地 SOCKS5 网关。
- 支持 IPv4、IPv6 和域名目标的 TCP `CONNECT`。
- 支持用户名/密码认证、最佳节点、随机节点和指定节点选择。
- 新增运行时启动/停止控制，以及 **系统设置 → SOCKS5 网关** 侧边栏页面。
- SOCKS5 密码加密保存，并新增管理员专用设置 API。

### 安全

- 网关默认关闭。
- 默认监听地址为 `127.0.0.1:1080`。
- 一期明确不包含 UDP `ASSOCIATE` 和 `BIND`。

参见 [SOCKS5 网关](docs/features/socks5.zh-CN.md)。

## [1.2.21] - 2026 年 9 月 10 日

### 新增

- 新增 WebDAV 定时备份和 5 段 Cron 表达式配置。
- 定时任务复用手动 ZIP 上传流程，在任务中心以 `webdav_backup` 展示进度，并跳过重叠上传。
- 恢复备份时保留当前 WebDAV 连接和定时计划设置。

### 修复

- 创建 WebDAV 目录时使用带尾斜杠的 `MKCOL`，并接受同主机集合重定向，修复 TeraCLOUD 一类服务的上传问题。

参见 [系统备份与 WebDAV](docs/features/backup.zh-CN.md)。

## [1.2.20] - 2026 年 9 月 10 日

### 新增

- 新增仅管理员可用的系统备份设置页。
- 新增 WebDAV 备份 ZIP 上传、远程列表、下载和基于现有数据库迁移流程的恢复。
- 新增加密保存 WebDAV 密码、HTTPS/私有网络安全限制、归档大小限制、ZIP bomb/路径穿越/符号链接防护、重定向阻止、临时文件清理和恢复任务跟踪。

### 修复

- 修复 `/v1/users/me` 返回 `roles: ["ADMIN"]` 时管理员识别失败的问题。

参见 [系统备份与 WebDAV](docs/features/backup.zh-CN.md)。

## [1.2.19] - 2026 年 9 月 9 日

### 新增

- 新增 Clash/mihomo `type: openvpn` 节点导入和导出支持。
- 保留 OpenVPN 传输、加密/认证、凭据、PEM 块、TLS 静态密钥、保活、DNS、IP 栈和 `dialer-proxy` 等字段。
- 使用内部无损格式 `openvpn://server:port?...#name` 支持编辑和往返转换。

### 兼容性

- OpenVPN 节点会输出到 Clash/mihomo；v2ray 和 Surge 等不支持的输出会跳过。
- 当前不包含直接导入 `.ovpn` 文件的能力。

参见 [机场管理](docs/features/airport.zh-CN.md) 和 [订阅分享](docs/features/subscription-share.zh-CN.md) 中的 OpenVPN 说明。

> `v1.2.19` 之前的版本可通过仓库历史和 Git tag 查看。
