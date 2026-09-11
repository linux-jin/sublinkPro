[English](socks5.md) | 简体中文

# SOCKS5 网关（一期）

SublinkPro 可以将已保存的代理节点通过本地 SOCKS5 网关暴露出来。该功能与现有的 SOCKS5 节点导入/导出能力不同。

## 一期包含

- 仅支持 TCP `CONNECT`
- 支持 IPv4、IPv6 和域名目标
- 可选用户名/密码认证（默认启用）
- 最佳节点、随机节点或指定节点选择
- 复用现有 mihomo 出站适配器连接选定节点
- 保存设置后即时启动/停止，无需重启进程

本期不包含 UDP `ASSOCIATE`、`BIND`、粘性会话、多用户路由，以及 Resin 风格的健康节点池故障切换。

## 配置

1. 使用管理员账号登录。
2. 从侧边栏打开 **系统设置 → SOCKS5 网关**。
3. 本机使用请保持监听地址为 `127.0.0.1`。
4. 选择端口（默认 `1080`）和节点选择策略。
5. 保持认证开启，并设置用户名和密码。
6. 保存设置。

网关默认关闭。密码使用实例 API 加密密钥加密保存，设置 API 永远不会返回明文密码。

## 使用

```bash
curl --proxy socks5h://127.0.0.1:1080 \
  -U "admin:your-password" \
  https://api.ipify.org
```

如果绑定到 `0.0.0.0` 或其他非回环地址，请只在防火墙或私有网络后开放，并使用强密码。该网关是出站中继，未启用认证就暴露到公网会变成开放代理。

## API

管理员专用接口：

- `GET /api/v1/settings/socks5`
- `POST /api/v1/settings/socks5`
- `POST /api/v1/settings/socks5/stop`

响应会返回 `running` 和 `boundAddress`，不会返回明文密码。
