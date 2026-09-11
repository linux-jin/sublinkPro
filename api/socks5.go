package api

import (
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
		Enabled       bool   `json:"enabled"`
		ListenAddress string `json:"listenAddress"`
		Port          int    `json:"port"`
		Username      string `json:"username"`
		Password      string `json:"password"`
		ClearPassword bool   `json:"clearPassword"`
		NodeID        int    `json:"nodeId"`
		Selection     string `json:"selection"`
		RequireAuth   *bool  `json:"requireAuth"`
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
	cfg, err := socks5service.SaveConfig(socks5service.Config{
		Enabled:       req.Enabled,
		ListenAddress: req.ListenAddress,
		Port:          req.Port,
		Username:      req.Username,
		Password:      password,
		NodeID:        req.NodeID,
		Selection:     req.Selection,
		RequireAuth:   requireAuth,
		ClearPassword: req.ClearPassword,
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
