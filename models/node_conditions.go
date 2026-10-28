package models

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sublink/utils"
)

// NodeConditions 条件表达式结构
type NodeConditions struct {
	Logic      string          `json:"logic"`      // "and" 或 "or"
	Conditions []NodeCondition `json:"conditions"` // 条件列表
}

// NodeCondition 单个条件
type NodeCondition struct {
	Field    string `json:"field"`    // 字段名
	Operator string `json:"operator"` // 操作符
	Value    any    `json:"value"`    // 比较值
}

// ParseConditions preserves the tag-rule parser entry point and permissive legacy semantics.
func ParseConditions(raw string) (*TagConditions, error) { return ParseNodeConditions(raw) }

// ========== 条件评估 ==========

// ParseNodeConditions parses the existing flat node-condition format.
func ParseNodeConditions(conditionsJSON string) (*NodeConditions, error) {
	if conditionsJSON == "" {
		return nil, fmt.Errorf("empty conditions")
	}
	var conditions NodeConditions
	if err := json.Unmarshal([]byte(conditionsJSON), &conditions); err != nil {
		return nil, err
	}
	return &conditions, nil
}

// EvaluateNode 对节点评估条件
func (tc *NodeConditions) EvaluateNode(node Node) bool {
	if len(tc.Conditions) == 0 {
		return false
	}

	logic := strings.ToLower(strings.TrimSpace(tc.Logic))
	if logic == "or" {
		for i := 0; i < len(tc.Conditions); {
			matched, next := evaluateConditionUnit(node, tc.Conditions, i)
			if matched {
				return true
			}
			i = next
		}
		return false
	}

	for i := 0; i < len(tc.Conditions); {
		matched, next := evaluateConditionUnit(node, tc.Conditions, i)
		if !matched {
			return false
		}
		i = next
	}
	return true
}

func evaluateConditionUnit(node Node, conditions []TagCondition, index int) (bool, int) {
	if shouldStartUnlockConditionGroup(conditions, index) {
		group, next := collectUnlockConditionGroup(conditions, index)
		return evaluateUnlockConditionGroup(node, group), next
	}
	return evaluateCondition(node, conditions[index]), index + 1
}

func shouldStartUnlockConditionGroup(conditions []TagCondition, index int) bool {
	if index < 0 || index >= len(conditions) {
		return false
	}
	switch conditions[index].Field {
	case "unlock_provider", "unlock_status":
		return true
	case "unlock_keyword":
		group, _ := collectUnlockConditionGroup(conditions, index)
		for _, cond := range group {
			if cond.Field == "unlock_provider" || cond.Field == "unlock_status" {
				return true
			}
		}
	}
	return false
}

func collectUnlockConditionGroup(conditions []TagCondition, start int) ([]TagCondition, int) {
	group := make([]TagCondition, 0, 3)
	seen := make(map[string]bool, 3)
	for i := start; i < len(conditions); i++ {
		field := conditions[i].Field
		if !isGroupedUnlockConditionField(field) || seen[field] {
			break
		}
		group = append(group, conditions[i])
		seen[field] = true
	}
	return group, start + len(group)
}

func isGroupedUnlockConditionField(field string) bool {
	return field == "unlock_provider" || field == "unlock_status" || field == "unlock_keyword"
}

func evaluateUnlockConditionGroup(node Node, group []TagCondition) bool {
	if len(group) == 0 {
		return false
	}
	if len(group) == 1 && group[0].Field == "unlock_keyword" {
		return evaluateCondition(node, group[0])
	}

	summary := ParseUnlockSummary(node.UnlockSummary)
	for _, result := range summary.Providers {
		matched := true
		for _, cond := range group {
			if !evaluateUnlockProviderResultCondition(result, cond) {
				matched = false
				break
			}
		}
		if matched {
			return true
		}
	}
	return false
}

func evaluateUnlockProviderResultCondition(result UnlockProviderResult, cond TagCondition) bool {
	compareValue := fmt.Sprintf("%v", cond.Value)

	switch cond.Field {
	case "unlock_provider":
		return matchUnlockProviderCondition(result.Provider, cond.Operator, compareValue)
	case "unlock_status":
		return matchUnlockStatusCondition(result.Status, cond.Operator, compareValue)
	case "unlock_keyword":
		return matchTextCondition(buildUnlockResultSearchText(result), cond.Operator, compareValue)
	default:
		return false
	}
}

func matchUnlockProviderCondition(provider string, operator string, compareValue string) bool {
	targetProvider := NormalizeUnlockProvider(compareValue)
	if targetProvider == "" {
		return false
	}
	providerKey := NormalizeUnlockProvider(provider)

	switch operator {
	case "equals":
		return providerKey == targetProvider
	case "not_equals":
		return providerKey != targetProvider
	case "contains":
		return strings.Contains(strings.ToLower(provider), strings.ToLower(compareValue))
	case "not_contains":
		return !strings.Contains(strings.ToLower(provider), strings.ToLower(compareValue))
	default:
		return false
	}
}

func matchUnlockStatusCondition(status string, operator string, compareValue string) bool {
	targetStatus := NormalizeUnlockStatus(compareValue)
	if targetStatus == "" {
		targetStatus = strings.TrimSpace(compareValue)
	}
	statusKey := NormalizeUnlockStatus(status)
	if statusKey == "" {
		statusKey = strings.TrimSpace(status)
	}

	switch operator {
	case "equals":
		return statusKey == targetStatus
	case "not_equals":
		return statusKey != targetStatus
	case "contains":
		return strings.Contains(strings.ToLower(status), strings.ToLower(compareValue))
	case "not_contains":
		return !strings.Contains(strings.ToLower(status), strings.ToLower(compareValue))
	default:
		return false
	}
}

func buildUnlockResultSearchText(result UnlockProviderResult) string {
	providerMeta := GetUnlockProviderMeta(result.Provider)
	return strings.Join([]string{
		providerMeta.Label,
		result.Provider,
		result.Status,
		GetUnlockStatusLabel(result.Status),
		result.Region,
		result.Reason,
		result.Detail,
	}, " ")
}

func matchTextCondition(fieldValue string, operator string, compareValue string) bool {
	switch operator {
	case "equals":
		return fieldValue == compareValue
	case "not_equals":
		return fieldValue != compareValue
	case "contains":
		return strings.Contains(strings.ToLower(fieldValue), strings.ToLower(compareValue))
	case "not_contains":
		return !strings.Contains(strings.ToLower(fieldValue), strings.ToLower(compareValue))
	case "regex":
		re, err := regexp.Compile(compareValue)
		if err != nil {
			utils.Error("正则表达式编译失败: %s, error: %v", compareValue, err)
			return false
		}
		return re.MatchString(fieldValue)
	default:
		return false
	}
}

// evaluateCondition 评估单个条件
func evaluateCondition(node Node, cond TagCondition) bool {
	// 解锁字段需要特殊处理：检查所有 providers 而不仅仅是第一个
	if cond.Field == "unlock_provider" || cond.Field == "unlock_status" {
		return evaluateUnlockCondition(node, cond)
	}

	fieldValue := getNodeFieldValue(node, cond.Field)
	compareValue := cond.Value

	switch cond.Operator {
	case "equals":
		return fmt.Sprintf("%v", fieldValue) == fmt.Sprintf("%v", compareValue)
	case "not_equals":
		return fmt.Sprintf("%v", fieldValue) != fmt.Sprintf("%v", compareValue)
	case "contains":
		return strings.Contains(strings.ToLower(fmt.Sprintf("%v", fieldValue)), strings.ToLower(fmt.Sprintf("%v", compareValue)))
	case "not_contains":
		return !strings.Contains(strings.ToLower(fmt.Sprintf("%v", fieldValue)), strings.ToLower(fmt.Sprintf("%v", compareValue)))
	case "regex":
		pattern := fmt.Sprintf("%v", compareValue)
		re, err := regexp.Compile(pattern)
		if err != nil {
			utils.Error("正则表达式编译失败: %s, error: %v", pattern, err)
			return false
		}
		return re.MatchString(fmt.Sprintf("%v", fieldValue))
	case "greater_than":
		return compareNumeric(fieldValue, compareValue) > 0
	case "less_than":
		return compareNumeric(fieldValue, compareValue) < 0
	case "greater_or_equal":
		return compareNumeric(fieldValue, compareValue) >= 0
	case "less_or_equal":
		return compareNumeric(fieldValue, compareValue) <= 0
	default:
		return false
	}
}

// evaluateUnlockCondition 评估解锁相关条件
// 需要遍历所有 providers 进行匹配，而不仅仅是第一个
func evaluateUnlockCondition(node Node, cond TagCondition) bool {
	summary := ParseUnlockSummary(node.UnlockSummary)
	compareValue := fmt.Sprintf("%v", cond.Value)

	if cond.Field == "unlock_provider" {
		// 检查是否有任何 provider 匹配
		targetProvider := NormalizeUnlockProvider(compareValue)
		if targetProvider == "" {
			return false
		}

		for _, result := range summary.Providers {
			providerKey := NormalizeUnlockProvider(result.Provider)
			match := false

			switch cond.Operator {
			case "equals":
				match = providerKey == targetProvider
			case "not_equals":
				match = providerKey != targetProvider
			case "contains":
				match = strings.Contains(strings.ToLower(result.Provider), strings.ToLower(compareValue))
			case "not_contains":
				match = !strings.Contains(strings.ToLower(result.Provider), strings.ToLower(compareValue))
			}

			if match {
				return true
			}
		}
		return false
	}

	if cond.Field == "unlock_status" {
		// 检查是否有任何 provider 的状态匹配
		targetStatus := NormalizeUnlockStatus(compareValue)
		if targetStatus == "" {
			// 如果规范化后为空，尝试直接匹配原始值
			targetStatus = strings.TrimSpace(compareValue)
		}

		for _, result := range summary.Providers {
			statusKey := NormalizeUnlockStatus(result.Status)
			if statusKey == "" {
				statusKey = strings.TrimSpace(result.Status)
			}

			match := false
			switch cond.Operator {
			case "equals":
				match = statusKey == targetStatus
			case "not_equals":
				match = statusKey != targetStatus
			case "contains":
				match = strings.Contains(strings.ToLower(result.Status), strings.ToLower(compareValue))
			case "not_contains":
				match = !strings.Contains(strings.ToLower(result.Status), strings.ToLower(compareValue))
			}

			if match {
				return true
			}
		}
		return false
	}

	return false
}

// getNodeFieldValue 获取节点字段值
func getNodeFieldValue(node Node, field string) any {
	switch field {
	case "name":
		return node.EffectiveName()
	case "effective_name":
		return node.EffectiveName()
	case "link_name":
		return node.LinkName
	case "name_mode":
		return NormalizeNodeNameMode(node.NameMode)
	case "link_address":
		return node.LinkAddress
	case "link_host":
		return node.LinkHost
	case "link_port":
		return node.LinkPort
	case "link_country":
		return node.LinkCountry
	case "protocol":
		return node.Protocol
	case "source":
		return node.Source
	case "group":
		return node.Group
	case "speed":
		return node.Speed
	case "speed_status":
		return node.SpeedStatus
	case "delay_time":
		return node.DelayTime
	case "delay_status":
		return node.DelayStatus
	case "dialer_proxy_name":
		return node.DialerProxyName
	case "fraud_score":
		return node.FraudScore
	case "quality_status":
		return getNodeQualityStatusValue(node)
	case "unlock_provider":
		summary := ParseUnlockSummary(node.UnlockSummary)
		if result, ok := GetUnlockResult(summary, ""); ok {
			return NormalizeUnlockProvider(result.Provider)
		}
		return ""
	case "unlock_status":
		return BuildPrimaryUnlockStatus(node.UnlockSummary)
	case "unlock_result":
		return BuildUnlockRenameValue(node.UnlockSummary, "")
	case "unlock_keyword":
		return BuildUnlockSearchText(node.UnlockSummary)
	case "ip_type":
		return getNodeIPTypeValue(node)
	case "residential_type":
		return getNodeResidentialTypeValue(node)
	case "link":
		return node.Link
	case "tags":
		return node.Tags
	default:
		return ""
	}
}

// compareNumeric 数值比较，返回 -1, 0, 1
func compareNumeric(a, b any) int {
	aFloat := toFloat64(a)
	bFloat := toFloat64(b)
	if aFloat > bFloat {
		return 1
	} else if aFloat < bFloat {
		return -1
	}
	return 0
}

// toFloat64 转换为float64
func toFloat64(v any) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	default:
		return 0
	}
}
