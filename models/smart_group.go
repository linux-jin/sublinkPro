package models

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sublink/database"
	"time"
)

// SmartGroup is an independent, dynamic view of nodes. It never modifies Node.Group.
type SmartGroup struct {
	ID          int       `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:191;uniqueIndex;not null" json:"name"`
	Countries   string    `gorm:"size:191;not null" json:"countries"` // comma-separated ISO 3166-1 alpha-2 codes
	MaxDelay    int       `json:"maxDelay"`
	MinSpeed    float64   `json:"minSpeed"`
	MaxAgeHours int       `json:"maxAgeHours"` // 0 disables the freshness check
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (g *SmartGroup) Validate() error {
	g.Name = strings.TrimSpace(g.Name)
	if g.Name == "" || len(g.Name) > 191 {
		return fmt.Errorf("智能分组名称不能为空且不能超过191字符")
	}
	seen := map[string]bool{}
	countries := make([]string, 0)
	for _, part := range strings.Split(g.Countries, ",") {
		code := strings.ToUpper(strings.TrimSpace(part))
		if code == "" {
			continue
		}
		if len(code) != 2 || code[0] < 'A' || code[0] > 'Z' || code[1] < 'A' || code[1] > 'Z' {
			return fmt.Errorf("国家必须是两位 ISO 代码")
		}
		if !seen[code] {
			countries = append(countries, code)
			seen[code] = true
		}
	}
	if len(countries) == 0 || len(countries) > 32 {
		return fmt.Errorf("请选择1至32个国家")
	}
	if g.MaxDelay < 0 || g.MinSpeed < 0 || g.MaxAgeHours < 0 || g.MaxAgeHours > 8760 {
		return fmt.Errorf("检测阈值不能为负，结果有效期不能超过一年")
	}
	g.Countries = strings.Join(countries, ",")
	return nil
}

func ListSmartGroups() ([]SmartGroup, error) {
	var groups []SmartGroup
	err := database.DB.Order("id ASC").Find(&groups).Error
	return groups, err
}

func GetSmartGroup(id int) (SmartGroup, error) {
	var group SmartGroup
	err := database.DB.First(&group, id).Error
	return group, err
}

// Candidates intentionally ignores current health so failed nodes can be retested.
func (g SmartGroup) Candidates() ([]Node, error) {
	return (&Node{}).ListWithFilters(NodeFilter{Countries: strings.Split(g.Countries, ",")})
}

func (g SmartGroup) CandidateIDs() ([]int, error) {
	nodes, err := g.Candidates()
	if err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(nodes))
	for _, node := range nodes {
		ids = append(ids, node.ID)
	}
	sort.Ints(ids)
	return ids, nil
}

// Members evaluates the latest persisted test result at read time, across all source groups.
func (g SmartGroup) Members() ([]Node, error) {
	nodes, err := (&Node{}).ListWithFilters(NodeFilter{
		Countries: strings.Split(g.Countries, ","), DelayStatus: "success", SpeedStatus: "success",
		MaxDelay: g.MaxDelay,
	})
	if err != nil {
		return nil, err
	}
	cutoff := time.Now().Add(-time.Duration(g.MaxAgeHours) * time.Hour)
	members := make([]Node, 0, len(nodes))
	for _, node := range nodes {
		if node.DelayTime <= 0 || node.Speed <= 0 || node.Speed < g.MinSpeed {
			continue
		}
		if g.MaxAgeHours > 0 {
			delayAt, delayErr := time.ParseInLocation("2006-01-02 15:04:05", node.LatencyCheckAt, time.Local)
			speedAt, speedErr := time.ParseInLocation("2006-01-02 15:04:05", node.SpeedCheckAt, time.Local)
			if delayErr != nil || speedErr != nil || delayAt.Before(cutoff) || speedAt.Before(cutoff) {
				continue
			}
		}
		members = append(members, node)
	}
	sort.Slice(members, func(i, j int) bool { return members[i].ID < members[j].ID })
	return members, nil
}

// ParseSmartGroupIDs parses the subscription's compact, ordered ID list.
func ParseSmartGroupIDs(value string) []int {
	seen := map[int]bool{}
	ids := make([]int, 0)
	for _, part := range strings.Split(value, ",") {
		id, err := strconv.Atoi(strings.TrimSpace(part))
		if err == nil && id > 0 && !seen[id] {
			ids = append(ids, id)
			seen[id] = true
		}
	}
	return ids
}

func NormalizeSmartGroupIDs(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", nil
	}
	for _, part := range strings.Split(value, ",") {
		id, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || id <= 0 {
			return "", fmt.Errorf("智能分组ID无效")
		}
	}
	ids := ParseSmartGroupIDs(value)
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		if _, err := GetSmartGroup(id); err != nil {
			return "", fmt.Errorf("智能分组 %d 不存在", id)
		}
		parts = append(parts, strconv.Itoa(id))
	}
	return strings.Join(parts, ","), nil
}
