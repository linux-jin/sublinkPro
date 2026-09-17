[English](socks5.md) | 简体中文

# SOCKS5 网关

SublinkPro 可以将已保存的代理节点通过本地 SOCKS5 网关暴露出来。该功能与现有的 SOCKS5 节点导入/导出能力不同。

## 当前包含

- 仅支持 TCP `CONNECT`
- 支持 IPv4、IPv6 和域名目标
- 可选用户名/密码认证（默认启用）
- 支持最佳节点、随机节点、轮询节点、P2C 智能负载均衡或指定节点选择
- 支持按多个分组、来源、协议和国家/地区组合过滤候选节点池；不同条件之间为“且”，同一条件内多个值为“或”
- 支持可选的内存粘性会话，可按客户端 IP 或已认证的 SOCKS5 用户名绑定，滑动 TTL 范围为 60-604800 秒
- 按节点 ID 和链接哈希缓存复用 mihomo 出站适配器，并限制空闲 LRU 缓存规模
- 支持每次请求最多 5 个候选节点重试、拨号超时和失败节点冷却
- 指定节点可选失败回退；默认关闭，保持指定节点路由严格性
- 全局活动连接上限（`maxConnections`，默认 `256`，范围 `1-10000`）
- 单客户端 IP 活动连接上限（`maxConnectionsPerClient`，默认 `32`，范围 `1..maxConnections`）
- 空闲超时（`idleTimeoutSeconds`，默认 `600`，范围 `0-86400`；`0` 表示禁用）
- 单连接最大持续时间（`maxConnectionDurationSeconds`，默认 `0`，范围 `0-604800`；`0` 表示禁用）
- 仅管理员可见的实时监控，展示活动/累计/成功/失败连接数及上下行字节数
- 仅管理员可执行断开单条活动连接或断开全部连接
- 可选节点主动健康探测，提供固定 HTTPS 探测目标、受控并发、延迟统计、指数失败冷却和手动立即探测
- 展示健康、异常、探测中和未知节点的健康状态
- 健康感知选路：主动探测失败的节点会退出正常候选集，`best` 会在截取重试节点前优先使用最新主动探测延迟排序
- 保存设置后即时启动/停止，无需重启进程

智能选路会用 P2C 选择首次尝试节点，综合主动探测延迟（或节点已存延迟）、活动连接负载和连续失败惩罚；其余重试节点按评分排序。粘性命中仍保持第一优先级，智能评分用于回退节点。所有自动切换均发生在发送 SOCKS5 成功响应之前。主动探测失败的节点会在仍有其他可路由节点时被跳过；若全部候选都异常，则采用 fail-open 方式保留候选以便恢复。真实连接成功会解除节点的选路排除状态，同时保留最近一次主动探测延迟。拨号失败的适配器会被丢弃，停止或重新应用网关配置时会关闭适配器池。如果全部候选节点都处于冷却状态，系统会探测最早到期的节点，避免节点池永久不可用。

目前仍不包含 UDP `ASSOCIATE`、`BIND`、多用户凭据路由和多端口监听。粘性租约仅保存在内存中，停止网关或重新应用设置后会清空。第二阶段新增了管理员实时连接监控、聚合计数、流量字节统计以及连接终止控制。

## 配置

1. 使用管理员账号登录。
2. 从侧边栏打开 **系统设置 → SOCKS5 网关**。
3. 本机使用请保持监听地址为 `127.0.0.1`。
4. 选择端口（默认 `1080`）和节点选择策略。
5. 可选按分组、来源、协议和国家/地区限制候选节点池；留空表示不限制。指定节点模式仅将这些条件用于失败回退候选。
6. 配置最大尝试次数、单节点拨号超时和失败节点冷却。
7. 配置全局连接上限和单客户端连接上限。
8. 配置空闲超时和最大连接时长；设置为 `0` 可禁用对应限制。
9. 可选启用粘性会话，并选择客户端 IP 或已认证的 SOCKS5 用户名作为粘性键。有效期范围为 60-604800 秒，每次连接成功都会刷新。
10. 使用指定节点时，仅在允许切换其他节点的情况下开启失败回退。
11. 保持认证开启，并设置用户名和密码。
12. 可选启用节点主动健康探测，并配置 10-3600 秒的探测间隔和 1-30 秒的单节点超时。
13. 保存设置。监控、健康探测和断开连接操作仅管理员可用。

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
- `POST /api/v1/settings/socks5` — 保存并应用设置。JSON 字段包括 `enabled`、`listenAddress`、`port`、`username`、可选 `password`、`clearPassword`、`nodeId`、`selection`（`best`、`random`、`round_robin`、`smart` 或 `specific`）、`requireAuth`、`maxAttempts`（1-5）、`dialTimeoutSeconds`（1-120）、`failureCooldownSeconds`（0-3600）、`specificFallback`、`maxConnections`（1-10000）、`maxConnectionsPerClient`（1..maxConnections）、`idleTimeoutSeconds`（0-86400）、`maxConnectionDurationSeconds`（0-604800）、`healthCheckEnabled`、`healthCheckIntervalSeconds`（10-3600）、`healthCheckTimeoutSeconds`（1-30）、`candidateGroups`、`candidateSources`、`candidateProtocols`、`candidateCountries`、`stickySessionEnabled`、`stickySessionMode`（`client_ip` 或 `username`）和 `stickySessionTtlSeconds`（60-604800）。候选池字段均为字符串数组；不同字段之间为“且”，字段内多个值为“或”。可选字段均可省略，以兼容旧客户端并保留已保存值。
- `POST /api/v1/settings/socks5/stop` — 停止监听，但不修改已保存设置。
- `GET /api/v1/settings/socks5/status` — 返回 `data.config`、`data.stats` 和 `data.health`；后者包含探测运行状态和逐节点健康详情。
- `POST /api/v1/settings/socks5/health/probe` — 立即启动一轮节点健康探测；已有探测运行时不会重复启动。
- `GET /api/v1/settings/socks5/connections` — 返回当前活动连接数组，每项包含 `id`、`clientAddress`、`target`、`nodeName`、`phase`、`startedAt`、`lastActivity`、`uploadBytes` 和 `downloadBytes`。
- `DELETE /api/v1/settings/socks5/connections/:id` — 按连接 ID 断开一条活动连接。
- `DELETE /api/v1/settings/socks5/connections` — 断开全部活动连接。

状态统计、活动连接列表和粘性租约仅属于当前网关进程生命周期；停止或重新应用设置后会创建新的注册中心并清空这些内存状态。

主动健康探测遵循当前选择策略：严格指定节点模式只探测该节点，开启指定节点回退时探测指定节点及配置后的回退池，其他策略探测配置后的候选节点池。探测固定使用 Cloudflare HTTPS 连通性检测地址，不允许管理员配置任意探测 URL。每轮最多使用 4 个 worker 且不会重叠；状态响应包含完整聚合计数，但最多返回排序后的前 200 条节点记录，避免三秒轮询产生过大响应；失败探测会安全淘汰对应适配器租约，并使用最长一小时的指数冷却。
