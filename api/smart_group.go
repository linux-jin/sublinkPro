package api

import (
	"net/http"
	"strconv"
	"sublink/database"
	"sublink/models"
	"sublink/services/scheduler"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func smartGroupID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的智能分组ID"})
		return 0, false
	}
	return id, true
}

func smartGroupByID(c *gin.Context) (models.SmartGroup, bool) {
	id, ok := smartGroupID(c)
	if !ok {
		return models.SmartGroup{}, false
	}
	group, err := models.GetSmartGroup(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "智能分组不存在"})
		return models.SmartGroup{}, false
	}
	return group, true
}

func ListSmartGroups(c *gin.Context) {
	groups, err := models.ListSmartGroups()
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "获取智能分组失败"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": groups})
}

func CreateSmartGroup(c *gin.Context) {
	var group models.SmartGroup
	if err := c.ShouldBindBodyWith(&group, binding.JSON); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "无效的分组配置"})
		return
	}
	group.ID = 0
	group.MaxAgeHours = maxAgeHoursFromRequest(c, group.MaxAgeHours)
	if err := group.Validate(); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	if err := database.DB.Create(&group).Error; err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "名称已存在或保存失败"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": group})
}

// New groups default to 72 hours; explicitly passing zero disables expiration.
func maxAgeHoursFromRequest(c *gin.Context, current int) int {
	var fields map[string]any
	if err := c.ShouldBindBodyWith(&fields, binding.JSON); err == nil {
		if _, exists := fields["maxAgeHours"]; !exists {
			return 72
		}
	}
	return current
}

func UpdateSmartGroup(c *gin.Context) {
	existing, ok := smartGroupByID(c)
	if !ok {
		return
	}
	var group models.SmartGroup
	if err := c.ShouldBindBodyWith(&group, binding.JSON); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "无效的分组配置"})
		return
	}
	group.ID = existing.ID
	if err := group.Validate(); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	if err := database.DB.Model(&existing).Select("Name", "Countries", "Keyword", "SourceGroups", "MaxDelay", "MinSpeed", "MaxAgeHours").Updates(&group).Error; err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "名称已存在或保存失败"})
		return
	}
	updated, err := models.GetSmartGroup(existing.ID)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "读取更新结果失败"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": updated})
}

func DeleteSmartGroup(c *gin.Context) {
	group, ok := smartGroupByID(c)
	if !ok {
		return
	}
	// Do not silently change the output of subscriptions referencing this view.
	var subscriptions []models.Subcription
	if err := database.DB.Select("id", "smart_group_ids").Where("smart_group_ids LIKE ?", "%"+strconv.Itoa(group.ID)+"%").Find(&subscriptions).Error; err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "检查订阅引用失败"})
		return
	}
	for _, sub := range subscriptions {
		for _, id := range models.ParseSmartGroupIDs(sub.SmartGroupIDs) {
			if id == group.ID {
				c.JSON(http.StatusConflict, gin.H{"code": 409, "msg": "该智能分组已被订阅引用，请先移除订阅中的选择"})
				return
			}
		}
	}
	if err := database.DB.Delete(&group).Error; err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "删除失败"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "msg": "删除成功"})
}

func SmartGroupMembers(c *gin.Context) {
	group, ok := smartGroupByID(c)
	if !ok {
		return
	}
	candidates, err := group.Candidates()
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "查询候选节点失败"})
		return
	}
	members, err := group.Members()
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "查询节点失败"})
		return
	}
	ids := make([]int, 0, len(members))
	details := make([]gin.H, 0, min(len(members), 100))
	for _, node := range members {
		ids = append(ids, node.ID)
		if len(details) < 100 {
			details = append(details, gin.H{"id": node.ID, "name": node.EffectiveName(), "group": node.Group, "country": node.LinkCountry, "delay": node.DelayTime, "speed": node.Speed})
		}
	}
	c.JSON(200, gin.H{"code": 200, "data": gin.H{"ids": ids, "count": len(ids), "candidateCount": len(candidates), "nodes": details}})
}

func CheckSmartGroupCandidates(c *gin.Context) {
	group, ok := smartGroupByID(c)
	if !ok {
		return
	}
	var req struct {
		ProfileID int `json:"profileId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ProfileID <= 0 {
		c.JSON(400, gin.H{"code": 400, "msg": "请选择节点检测策略"})
		return
	}
	if _, err := models.GetNodeCheckProfileByID(req.ProfileID); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "检测策略不存在"})
		return
	}
	ids, err := group.CandidateIDs()
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "获取候选节点失败"})
		return
	}
	if len(ids) == 0 {
		c.JSON(400, gin.H{"code": 400, "msg": "当前国家、关键词和原分组条件下没有候选节点"})
		return
	}
	go scheduler.ExecuteNodeCheckWithProfile(req.ProfileID, ids, models.TaskTriggerManual)
	c.JSON(200, gin.H{"code": 200, "data": gin.H{"count": len(ids)}})
}
