package api

import (
	"net/http"
	"strings"
	"sublink/models"
	socks5service "sublink/services/socks5"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

func requireSocks5Admin(c *gin.Context) bool {
	username, ok := currentUsernameFromContext(c)
	if !ok {
		return false
	}
	user := &models.User{Username: username}
	if err := user.Find(); err != nil || !strings.EqualFold(user.Role, "admin") {
		utils.ForbiddenI18n(c, "仅管理员可管理 SOCKS5 服务", "settings.socks5.api.adminRequired", nil)
		return false
	}
	return true
}

func GetSocks5Settings(c *gin.Context) {
	if !requireSocks5Admin(c) {
		return
	}
	cfg, err := socks5service.LoadConfig()
	if err != nil {
		utils.FailWithI18n(c, "读取 SOCKS5 设置失败: "+err.Error(), "settings.socks5.api.loadFailed", map[string]any{"message": err.Error()})
		return
	}
	running, bound := socks5service.DefaultManager().Status()
	utils.OkDetailedI18n(c, "SOCKS5 设置已加载", socks5service.ToPublicConfig(cfg, running, bound), "settings.socks5.api.loaded", nil)
}

func UpdateSocks5Settings(c *gin.Context) {
	if !requireSocks5Admin(c) {
		return
	}
	var req struct {
		Enabled                      bool   `json:"enabled"`
		ListenAddress                string `json:"listenAddress"`
		Port                         int    `json:"port"`
		Username                     string `json:"username"`
		Password                     string `json:"password"`
		ClearPassword                bool   `json:"clearPassword"`
		NodeID                       int    `json:"nodeId"`
		Selection                    string `json:"selection"`
		RequireAuth                  *bool  `json:"requireAuth"`
		MaxAttempts                  *int   `json:"maxAttempts"`
		DialTimeoutSeconds           *int   `json:"dialTimeoutSeconds"`
		FailureCooldownSeconds       *int   `json:"failureCooldownSeconds"`
		SpecificFallback             *bool  `json:"specificFallback"`
		MaxConnections               *int   `json:"maxConnections"`
		MaxConnectionsPerClient      *int   `json:"maxConnectionsPerClient"`
		IdleTimeoutSeconds           *int   `json:"idleTimeoutSeconds"`
		MaxConnectionDurationSeconds *int   `json:"maxConnectionDurationSeconds"`
		HealthCheckEnabled           *bool  `json:"healthCheckEnabled"`
		HealthCheckIntervalSeconds   *int   `json:"healthCheckIntervalSeconds"`
		HealthCheckTimeoutSeconds    *int   `json:"healthCheckTimeoutSeconds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithI18n(c, "参数错误", "settings.socks5.api.invalidRequest", nil)
		return
	}
	current, err := socks5service.LoadConfig()
	if err != nil {
		utils.FailWithI18n(c, "读取 SOCKS5 设置失败: "+err.Error(), "settings.socks5.api.loadFailed", map[string]any{"message": err.Error()})
		return
	}
	password := current.Password
	if req.ClearPassword {
		password = ""
	} else if req.Password != "" {
		password = req.Password
	}
	requireAuth := current.RequireAuth
	if req.RequireAuth != nil {
		requireAuth = *req.RequireAuth
	}
	maxAttempts := current.MaxAttempts
	if req.MaxAttempts != nil {
		maxAttempts = *req.MaxAttempts
	}
	dialTimeoutSeconds := current.DialTimeoutSeconds
	if req.DialTimeoutSeconds != nil {
		dialTimeoutSeconds = *req.DialTimeoutSeconds
	}
	failureCooldownSeconds := current.FailureCooldownSeconds
	if req.FailureCooldownSeconds != nil {
		failureCooldownSeconds = *req.FailureCooldownSeconds
	}
	specificFallback := current.SpecificFallback
	if req.SpecificFallback != nil {
		specificFallback = *req.SpecificFallback
	}
	maxConnections := current.MaxConnections
	if req.MaxConnections != nil {
		maxConnections = *req.MaxConnections
	}
	maxConnectionsPerClient := current.MaxConnectionsPerClient
	if req.MaxConnectionsPerClient != nil {
		maxConnectionsPerClient = *req.MaxConnectionsPerClient
	}
	idleTimeoutSeconds := current.IdleTimeoutSeconds
	if req.IdleTimeoutSeconds != nil {
		idleTimeoutSeconds = *req.IdleTimeoutSeconds
	}
	maxConnectionDurationSeconds := current.MaxConnectionDurationSeconds
	if req.MaxConnectionDurationSeconds != nil {
		maxConnectionDurationSeconds = *req.MaxConnectionDurationSeconds
	}
	healthCheckEnabled := current.HealthCheckEnabled
	if req.HealthCheckEnabled != nil {
		healthCheckEnabled = *req.HealthCheckEnabled
	}
	healthCheckIntervalSeconds := current.HealthCheckIntervalSeconds
	if req.HealthCheckIntervalSeconds != nil {
		healthCheckIntervalSeconds = *req.HealthCheckIntervalSeconds
	}
	healthCheckTimeoutSeconds := current.HealthCheckTimeoutSeconds
	if req.HealthCheckTimeoutSeconds != nil {
		healthCheckTimeoutSeconds = *req.HealthCheckTimeoutSeconds
	}
	cfg, err := socks5service.SaveConfig(socks5service.Config{
		Enabled:                      req.Enabled,
		ListenAddress:                req.ListenAddress,
		Port:                         req.Port,
		Username:                     req.Username,
		Password:                     password,
		NodeID:                       req.NodeID,
		Selection:                    req.Selection,
		RequireAuth:                  requireAuth,
		MaxAttempts:                  maxAttempts,
		DialTimeoutSeconds:           dialTimeoutSeconds,
		FailureCooldownSeconds:       failureCooldownSeconds,
		SpecificFallback:             specificFallback,
		MaxConnections:               maxConnections,
		MaxConnectionsPerClient:      maxConnectionsPerClient,
		IdleTimeoutSeconds:           idleTimeoutSeconds,
		MaxConnectionDurationSeconds: maxConnectionDurationSeconds,
		HealthCheckEnabled:           healthCheckEnabled,
		HealthCheckIntervalSeconds:   healthCheckIntervalSeconds,
		HealthCheckTimeoutSeconds:    healthCheckTimeoutSeconds,
		ClearPassword:                req.ClearPassword,
	})
	if err != nil {
		utils.FailWithI18n(c, "SOCKS5 设置无效: "+err.Error(), "settings.socks5.api.invalidSettings", map[string]any{"message": err.Error()})
		return
	}
	if err := socks5service.DefaultManager().Apply(cfg); err != nil {
		utils.FailWithI18n(c, "启动 SOCKS5 服务失败: "+err.Error(), "settings.socks5.api.startFailed", map[string]any{"message": err.Error()})
		return
	}
	running, bound := socks5service.DefaultManager().Status()
	utils.OkDetailedI18n(c, "SOCKS5 设置已保存", socks5service.ToPublicConfig(cfg, running, bound), "settings.socks5.api.saved", nil)
}

func StopSocks5(c *gin.Context) {
	if !requireSocks5Admin(c) {
		return
	}
	socks5service.DefaultManager().Stop()
	cfg, err := socks5service.LoadConfig()
	if err != nil {
		utils.FailWithI18n(c, "读取 SOCKS5 设置失败: "+err.Error(), "settings.socks5.api.loadFailed", map[string]any{"message": err.Error()})
		return
	}
	running, bound := socks5service.DefaultManager().Status()
	utils.OkDetailedI18n(c, "SOCKS5 服务已停止", socks5service.ToPublicConfig(cfg, running, bound), "settings.socks5.api.stopped", nil)
}

func GetSocks5Status(c *gin.Context) {
	if !requireSocks5Admin(c) {
		return
	}
	cfg, err := socks5service.LoadConfig()
	if err != nil {
		utils.FailWithI18n(c, "读取 SOCKS5 设置失败: "+err.Error(), "settings.socks5.api.loadFailed", map[string]any{"message": err.Error()})
		return
	}
	running, bound := socks5service.DefaultManager().Status()
	utils.OkDetailedI18n(c, "SOCKS5 状态已加载", gin.H{"config": socks5service.ToPublicConfig(cfg, running, bound), "stats": socks5service.DefaultManager().GatewaySnapshot(), "health": socks5service.DefaultManager().HealthSnapshot()}, "settings.socks5.api.statusLoaded", nil)
}

func GetSocks5Connections(c *gin.Context) {
	if !requireSocks5Admin(c) {
		return
	}
	connections := socks5service.DefaultManager().Connections()
	utils.OkDetailedI18n(c, "SOCKS5 连接已加载", connections, "settings.socks5.api.connectionsLoaded", nil)
}

func DeleteSocks5Connection(c *gin.Context) {
	if !requireSocks5Admin(c) {
		return
	}
	id := strings.TrimSpace(c.Param("id"))
	if id == "" || len(id) > 64 {
		utils.FailWithI18n(c, "连接 ID 无效", "settings.socks5.api.invalidConnection", nil)
		return
	}
	if !socks5service.DefaultManager().CloseConnection(id) {
		utils.FailWithCodeI18n(c, http.StatusNotFound, "连接不存在", "settings.socks5.api.connectionNotFound", nil)
		return
	}
	utils.OkDetailedI18n(c, "SOCKS5 连接已断开", nil, "settings.socks5.api.connectionClosed", nil)
}

func DeleteSocks5Connections(c *gin.Context) {
	if !requireSocks5Admin(c) {
		return
	}
	socks5service.DefaultManager().CloseAllConnections()
	utils.OkDetailedI18n(c, "SOCKS5 连接已全部断开", nil, "settings.socks5.api.connectionsClosed", nil)
}

func ProbeSocks5Health(c *gin.Context) {
	if !requireSocks5Admin(c) {
		return
	}
	started, available := socks5service.DefaultManager().TriggerHealthProbe()
	if !available {
		utils.FailWithI18n(c, "SOCKS5 健康探测未启动，服务尚未运行", "settings.socks5.api.healthProbeUnavailable", nil)
		return
	}
	if !started {
		utils.OkDetailedI18n(c, "SOCKS5 健康探测正在执行", gin.H{"started": false, "health": socks5service.DefaultManager().HealthSnapshot()}, "settings.socks5.api.healthProbeRunning", nil)
		return
	}
	utils.OkDetailedI18n(c, "SOCKS5 健康探测已启动", gin.H{"started": true, "health": socks5service.DefaultManager().HealthSnapshot()}, "settings.socks5.api.healthProbeStarted", nil)
}
