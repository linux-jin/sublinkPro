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
	ID           int       `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:191;uniqueIndex;not null" json:"name"`
	Countries    string    `gorm:"size:191;not null" json:"countries"` // comma-separated ISO 3166-1 alpha-2 codes
	Keyword      string    `gorm:"size:191" json:"keyword"`
	SourceGroups []string  `gorm:"serializer:json;type:text" json:"sourceGroups"`
	MaxDelay     int       `json:"maxDelay"`
	MinSpeed     float64   `json:"minSpeed"`
	MaxAgeHours  int       `json:"maxAgeHours"` // 0 disables the freshness check
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
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
	g.Keyword = strings.TrimSpace(g.Keyword)
	if len(g.Keyword) > 191 {
		return fmt.Errorf("关键词不能超过191字符")
	}
	groups := make([]string, 0, len(g.SourceGroups))
	seenGroups := make(map[string]bool)
	for _, source := range g.SourceGroups {
		source = strings.TrimSpace(source)
		if source == "" || len(source) > 191 {
			return fmt.Errorf("原分组名称不能为空且不能超过191字符")
		}
		key := strings.ToLower(source)
		if !seenGroups[key] {
			groups = append(groups, source)
			seenGroups[key] = true
		}
	}
	if len(groups) > 64 {
		return fmt.Errorf("最多选择64个原分组")
	}
	g.SourceGroups = groups
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
	nodes, err := (&Node{}).ListWithFilters(NodeFilter{Countries: strings.Split(g.Countries, ",")})
	if err != nil {
		return nil, err
	}
	groups := make(map[string]bool, len(g.SourceGroups))
	for _, source := range g.SourceGroups {
		groups[strings.ToLower(source)] = true
	}
	keyword := strings.ToLower(g.Keyword)
	candidates := make([]Node, 0, len(nodes))
	for _, node := range nodes {
		if len(groups) > 0 && !groups[strings.ToLower(node.Group)] {
			continue
		}
		if keyword != "" && !strings.Contains(strings.ToLower(node.EffectiveName()), keyword) &&
			!strings.Contains(strings.ToLower(node.Name), keyword) && !strings.Contains(strings.ToLower(node.LinkName), keyword) {
			continue
		}
		candidates = append(candidates, node)
	}
	return candidates, nil
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

// Members evaluates the latest persisted test result at read time.
// A speed threshold of zero requires only a successful latency test; TCP-only profiles can then populate the view.
func (g SmartGroup) Members() ([]Node, error) {
	nodes, err := g.Candidates()
	if err != nil {
		return nil, err
	}
	cutoff := time.Now().Add(-time.Duration(g.MaxAgeHours) * time.Hour)
	members := make([]Node, 0, len(nodes))
	for _, node := range nodes {
		if node.DelayStatus != "success" || node.DelayTime <= 0 || (g.MaxDelay > 0 && node.DelayTime > g.MaxDelay) {
			continue
		}
		if g.MinSpeed > 0 && (node.SpeedStatus != "success" || node.Speed < g.MinSpeed) {
			continue
		}
		if g.MaxAgeHours > 0 {
			delayAt, delayErr := time.ParseInLocation("2006-01-02 15:04:05", node.LatencyCheckAt, time.Local)
			if delayErr != nil || delayAt.Before(cutoff) {
				continue
			}
			if g.MinSpeed > 0 {
				speedAt, speedErr := time.ParseInLocation("2006-01-02 15:04:05", node.SpeedCheckAt, time.Local)
				if speedErr != nil || speedAt.Before(cutoff) {
					continue
				}
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
