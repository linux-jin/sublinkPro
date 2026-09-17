package api

import (
	"net/http"
	"strconv"
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
		Enabled                      bool      `json:"enabled"`
		ListenAddress                string    `json:"listenAddress"`
		Port                         int       `json:"port"`
		Username                     string    `json:"username"`
		Password                     string    `json:"password"`
		ClearPassword                bool      `json:"clearPassword"`
		NodeID                       int       `json:"nodeId"`
		Selection                    string    `json:"selection"`
		RequireAuth                  *bool     `json:"requireAuth"`
		MaxAttempts                  *int      `json:"maxAttempts"`
		DialTimeoutSeconds           *int      `json:"dialTimeoutSeconds"`
		FailureCooldownSeconds       *int      `json:"failureCooldownSeconds"`
		SpecificFallback             *bool     `json:"specificFallback"`
		MaxConnections               *int      `json:"maxConnections"`
		MaxConnectionsPerClient      *int      `json:"maxConnectionsPerClient"`
		IdleTimeoutSeconds           *int      `json:"idleTimeoutSeconds"`
		MaxConnectionDurationSeconds *int      `json:"maxConnectionDurationSeconds"`
		HealthCheckEnabled           *bool     `json:"healthCheckEnabled"`
		HealthCheckIntervalSeconds   *int      `json:"healthCheckIntervalSeconds"`
		HealthCheckTimeoutSeconds    *int      `json:"healthCheckTimeoutSeconds"`
		CandidateGroups              *[]string `json:"candidateGroups"`
		CandidateSources             *[]string `json:"candidateSources"`
		CandidateProtocols           *[]string `json:"candidateProtocols"`
		CandidateCountries           *[]string `json:"candidateCountries"`
		StickySessionEnabled         *bool     `json:"stickySessionEnabled"`
		StickySessionMode            *string   `json:"stickySessionMode"`
		StickySessionTTLSeconds      *int      `json:"stickySessionTtlSeconds"`
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
	candidateGroups := current.CandidateGroups
	if req.CandidateGroups != nil {
		candidateGroups = *req.CandidateGroups
	}
	candidateSources := current.CandidateSources
	if req.CandidateSources != nil {
		candidateSources = *req.CandidateSources
	}
	candidateProtocols := current.CandidateProtocols
	if req.CandidateProtocols != nil {
		candidateProtocols = *req.CandidateProtocols
	}
	candidateCountries := current.CandidateCountries
	if req.CandidateCountries != nil {
		candidateCountries = *req.CandidateCountries
	}
	stickySessionEnabled := current.StickySessionEnabled
	if req.StickySessionEnabled != nil {
		stickySessionEnabled = *req.StickySessionEnabled
	}
	stickySessionMode := current.StickySessionMode
	if req.StickySessionMode != nil {
		stickySessionMode = *req.StickySessionMode
	}
	stickySessionTTLSeconds := current.StickySessionTTLSeconds
	if req.StickySessionTTLSeconds != nil {
		stickySessionTTLSeconds = *req.StickySessionTTLSeconds
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
		CandidateGroups:              candidateGroups,
		CandidateSources:             candidateSources,
		CandidateProtocols:           candidateProtocols,
		CandidateCountries:           candidateCountries,
		StickySessionEnabled:         stickySessionEnabled,
		StickySessionMode:            stickySessionMode,
		StickySessionTTLSeconds:      stickySessionTTLSeconds,
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

type socks5RoutingProfileRequest struct {
	ID                      string   `json:"id"`
	Name                    string   `json:"name"`
	Enabled                 *bool    `json:"enabled"`
	Selection               string   `json:"selection"`
	NodeID                  int      `json:"nodeId"`
	MaxAttempts             int      `json:"maxAttempts"`
	DialTimeoutSeconds      int      `json:"dialTimeoutSeconds"`
	FailureCooldownSeconds  int      `json:"failureCooldownSeconds"`
	SpecificFallback        bool     `json:"specificFallback"`
	CandidateGroups         []string `json:"candidateGroups"`
	CandidateSources        []string `json:"candidateSources"`
	CandidateProtocols      []string `json:"candidateProtocols"`
	CandidateCountries      []string `json:"candidateCountries"`
	StickySessionEnabled    bool     `json:"stickySessionEnabled"`
	StickySessionTTLSeconds int      `json:"stickySessionTtlSeconds"`
}

func (req socks5RoutingProfileRequest) profile(id string, defaultEnabled bool) socks5service.RoutingProfile {
	enabled := defaultEnabled
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	if id == "" {
		id = req.ID
	}
	return socks5service.RoutingProfile{
		ID: id, Name: req.Name, Enabled: enabled, Selection: req.Selection, NodeID: req.NodeID,
		MaxAttempts: req.MaxAttempts, DialTimeoutSeconds: req.DialTimeoutSeconds,
		FailureCooldownSeconds: req.FailureCooldownSeconds, SpecificFallback: req.SpecificFallback,
		CandidateGroups: req.CandidateGroups, CandidateSources: req.CandidateSources,
		CandidateProtocols: req.CandidateProtocols, CandidateCountries: req.CandidateCountries,
		StickySessionEnabled: req.StickySessionEnabled, StickySessionTTLSeconds: req.StickySessionTTLSeconds,
	}
}

func ListSocks5RoutingProfiles(c *gin.Context) {
	if !requireSocks5Admin(c) {
		return
	}
	cfg, err := socks5service.LoadConfig()
	if err != nil {
		utils.FailWithI18n(c, "读取 SOCKS5 设置失败: "+err.Error(), "settings.socks5.api.loadFailed", map[string]any{"message": err.Error()})
		return
	}
	profiles, err := socks5service.ListRoutingProfiles(cfg)
	if err != nil {
		utils.FailWithI18n(c, "读取 SOCKS5 路由配置失败: "+err.Error(), "settings.socks5.api.profileLoadFailed", map[string]any{"message": err.Error()})
		return
	}
	utils.OkDetailedI18n(c, "SOCKS5 路由配置已加载", profiles, "settings.socks5.api.profilesLoaded", nil)
}

func CreateSocks5RoutingProfile(c *gin.Context) {
	if !requireSocks5Admin(c) {
		return
	}
	var req socks5RoutingProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithI18n(c, "参数错误: "+err.Error(), "settings.socks5.api.invalidRequest", nil)
		return
	}
	cfg, err := socks5service.LoadConfig()
	if err != nil {
		utils.FailWithI18n(c, "读取 SOCKS5 设置失败: "+err.Error(), "settings.socks5.api.loadFailed", map[string]any{"message": err.Error()})
		return
	}
	profile, err := socks5service.CreateRoutingProfile(cfg, req.profile("", true))
	if err != nil {
		utils.FailWithI18n(c, "创建 SOCKS5 路由配置失败: "+err.Error(), "settings.socks5.api.profileSaveFailed", map[string]any{"message": err.Error()})
		return
	}
	if err := socks5service.DefaultManager().ReloadRoutingProfiles(); err != nil {
		utils.FailWithI18n(c, "应用 SOCKS5 路由配置失败: "+err.Error(), "settings.socks5.api.profileApplyFailed", map[string]any{"message": err.Error()})
		return
	}
	utils.OkDetailedI18n(c, "SOCKS5 路由配置已创建", profile, "settings.socks5.api.profileCreated", nil)
}

func UpdateSocks5RoutingProfile(c *gin.Context) {
	if !requireSocks5Admin(c) {
		return
	}
	id := strings.TrimSpace(c.Param("id"))
	if id == "" || len(id) > 32 {
		utils.FailWithI18n(c, "SOCKS5 路由配置 ID 无效", "settings.socks5.api.invalidProfile", nil)
		return
	}
	var req socks5RoutingProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithI18n(c, "参数错误: "+err.Error(), "settings.socks5.api.invalidRequest", nil)
		return
	}
	cfg, err := socks5service.LoadConfig()
	if err != nil {
		utils.FailWithI18n(c, "读取 SOCKS5 设置失败: "+err.Error(), "settings.socks5.api.loadFailed", map[string]any{"message": err.Error()})
		return
	}
	profile, err := socks5service.UpdateRoutingProfile(cfg, id, req.profile(id, false))
	if err != nil {
		utils.FailWithI18n(c, "更新 SOCKS5 路由配置失败: "+err.Error(), "settings.socks5.api.profileSaveFailed", map[string]any{"message": err.Error()})
		return
	}
	if err := socks5service.DefaultManager().ReloadRoutingProfiles(); err != nil {
		utils.FailWithI18n(c, "应用 SOCKS5 路由配置失败: "+err.Error(), "settings.socks5.api.profileApplyFailed", map[string]any{"message": err.Error()})
		return
	}
	utils.OkDetailedI18n(c, "SOCKS5 路由配置已更新", profile, "settings.socks5.api.profileUpdated", nil)
}

func DeleteSocks5RoutingProfile(c *gin.Context) {
	if !requireSocks5Admin(c) {
		return
	}
	id := strings.TrimSpace(c.Param("id"))
	if id == "" || len(id) > 32 {
		utils.FailWithI18n(c, "SOCKS5 路由配置 ID 无效", "settings.socks5.api.invalidProfile", nil)
		return
	}
	cfg, err := socks5service.LoadConfig()
	if err != nil {
		utils.FailWithI18n(c, "读取 SOCKS5 设置失败: "+err.Error(), "settings.socks5.api.loadFailed", map[string]any{"message": err.Error()})
		return
	}
	if err := socks5service.DeleteRoutingProfile(cfg, id); err != nil {
		utils.FailWithI18n(c, "删除 SOCKS5 路由配置失败: "+err.Error(), "settings.socks5.api.profileDeleteFailed", map[string]any{"message": err.Error()})
		return
	}
	if err := socks5service.DefaultManager().ReloadRoutingProfiles(); err != nil {
		utils.FailWithI18n(c, "应用 SOCKS5 路由配置失败: "+err.Error(), "settings.socks5.api.profileApplyFailed", map[string]any{"message": err.Error()})
		return
	}
	utils.OkDetailedI18n(c, "SOCKS5 路由配置已删除", nil, "settings.socks5.api.profileDeleted", nil)
}

func GetSocks5RoutingSnapshot(c *gin.Context) {
	if !requireSocks5Admin(c) {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "25"))
	snapshot, err := socks5service.DefaultManager().RoutingSnapshot(socks5service.RoutingSnapshotQuery{
		ProfileID: c.Query("profileId"), Keyword: c.Query("keyword"), Status: c.Query("status"), SortBy: c.Query("sortBy"),
		SortOrder: c.Query("sortOrder"), Page: page, PageSize: pageSize,
	})
	if err != nil {
		utils.FailWithI18n(c, "读取 SOCKS5 节点路由统计失败: "+err.Error(), "settings.socks5.api.routingLoadFailed", map[string]any{"message": err.Error()})
		return
	}
	utils.OkDetailedI18n(c, "SOCKS5 节点路由统计已加载", snapshot, "settings.socks5.api.routingLoaded", nil)
}

func ResetSocks5RuntimeStats(c *gin.Context) {
	if !requireSocks5Admin(c) {
		return
	}
	socks5service.DefaultManager().ResetNodeRuntimeStats()
	utils.OkDetailedI18n(c, "SOCKS5 节点运行统计已重置", nil, "settings.socks5.api.runtimeReset", nil)
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
