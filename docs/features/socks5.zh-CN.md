[English](socks5.md) | 简体中文

# SOCKS5 网关

SublinkPro 可以将已保存的代理节点通过本地 SOCKS5 网关暴露出来。该功能与现有的 SOCKS5 节点导入/导出能力不同。

## 当前包含

- 仅支持 TCP `CONNECT`
- 支持 IPv4、IPv6 和域名目标
- 可选用户名/密码认证（默认启用）
- 支持最佳节点、随机节点、轮询节点或指定节点选择
- 按节点 ID 和链接哈希缓存复用 mihomo 出站适配器，并限制空闲 LRU 缓存规模
- 支持每次请求最多 5 个候选节点重试、拨号超时和失败节点冷却
- 指定节点可选失败回退；默认关闭，保持指定节点路由严格性
- 全局活动连接上限（`maxConnections`，默认 `256`，范围 `1-10000`）
- 单客户端 IP 活动连接上限（`maxConnectionsPerClient`，默认 `32`，范围 `1..maxConnections`）
- 空闲超时（`idleTimeoutSeconds`，默认 `600`，范围 `0-86400`；`0` 表示禁用）
- 单连接最大持续时间（`maxConnectionDurationSeconds`，默认 `0`，范围 `0-604800`；`0` 表示禁用）
- 仅管理员可见的实时监控，展示活动/累计/成功/失败连接数及上下行字节数
- 仅管理员可执行断开单条活动连接或断开全部连接
- 保存设置后即时启动/停止，无需重启进程

所有自动切换均发生在发送 SOCKS5 成功响应之前。拨号失败的适配器会被丢弃，停止或重新应用网关配置时会关闭适配器池。如果全部候选节点都处于冷却状态，系统会探测最早到期的节点，避免节点池永久不可用。

目前仍不包含 UDP `ASSOCIATE`、`BIND`、粘滞会话、多用户路由和多端口监听。第二阶段新增了管理员实时连接监控、聚合计数、流量字节统计以及连接终止控制。

## 配置

1. 使用管理员账号登录。
2. 从侧边栏打开 **系统设置 → SOCKS5 网关**。
3. 本机使用请保持监听地址为 `127.0.0.1`。
4. 选择端口（默认 `1080`）和节点选择策略。
5. 配置最大尝试次数、单节点拨号超时和失败节点冷却。
6. 配置全局连接上限和单客户端连接上限。
7. 配置空闲超时和最大连接时长；设置为 `0` 可禁用对应限制。
8. 使用指定节点时，仅在允许切换其他节点的情况下开启失败回退。
9. 保持认证开启，并设置用户名和密码。
10. 保存设置。监控和断开连接操作仅管理员可用。

网关默认关闭。密码使用实例 API 加密密钥加密保存。设置响应仅返回 `hasPassword` 及（有密码时）`maskedPassword`，不会返回明文 `password`。省略 `password` 会保留已保存密码，`clearPassword: true` 会清除密码。

## 使用

```bash
curl --proxy socks5h://127.0.0.1:1080 \
  -U "admin:your-password" \
  https://api.ipify.org
```

如果绑定到 `0.0.0.0` 或其他非回环地址，系统会强制要求认证。请只在防火墙或私有网络后开放，并使用强密码；后端会拒绝未认证的非回环监听，避免意外形成开放代理。

## API

以下接口均要求登录并且仅管理员可用；有写入风险的 `POST`/`DELETE` 接口在演示模式下也会被限制。

- `GET /api/v1/settings/socks5` — 读取公开网关设置，不返回明文密码。
- `POST /api/v1/settings/socks5` — 保存并应用设置。JSON 字段包括 `enabled`、`listenAddress`、`port`、`username`、可选 `password`、`clearPassword`、`nodeId`、`selection`（`best`、`random`、`round_robin` 或 `specific`）、`requireAuth`、`maxAttempts`（1-5）、`dialTimeoutSeconds`（1-120）、`failureCooldownSeconds`（0-3600）、`specificFallback`、`maxConnections`（1-10000）、`maxConnectionsPerClient`（1..maxConnections）、`idleTimeoutSeconds`（0-86400）和 `maxConnectionDurationSeconds`（0-604800）。后四个第二阶段字段可省略，兼容旧客户端并保留已保存值。
- `POST /api/v1/settings/socks5/stop` — 停止监听，但不修改已保存设置。
- `GET /api/v1/settings/socks5/status` — 返回 `data.config` 和 `data.stats`；统计字段为 `activeConnections`、`totalConnections`、`successfulConnections`、`failedConnections`、`uploadBytes`、`downloadBytes`。
- `GET /api/v1/settings/socks5/connections` — 返回当前活动连接数组，每项包含 `id`、`clientAddress`、`target`、`nodeName`、`phase`、`startedAt`、`lastActivity`、`uploadBytes` 和 `downloadBytes`。
- `DELETE /api/v1/settings/socks5/connections/:id` — 按连接 ID 断开一条活动连接。
- `DELETE /api/v1/settings/socks5/connections` — 断开全部活动连接。

状态统计和活动连接列表仅属于当前网关进程生命周期；停止或重新应用设置后会创建新的注册中心，计数会重置。
