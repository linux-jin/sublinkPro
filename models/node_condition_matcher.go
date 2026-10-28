package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
)

const (
	NodeConditionSmartCandidate = "smart-group-candidate"
	NodeConditionSmartMember    = "smart-group-member"
)

// Candidate fields must not depend on a previous check passing. Check-dependent
// predicates belong to membership so failed and untested candidates can recover.
func isSmartCandidateField(field string) bool {
	switch field {
	case "name", "effective_name", "link_name", "link_country", "protocol", "group", "source", "link_address", "link_host", "link_port", "dialer_proxy_name", "link":
		return true
	default:
		return false
	}
}

// CompileNodeConditionMatcher is strict for newly configured smart-group rules.
// Legacy tag rules keep ParseConditions/EvaluateNode without a silent migration.
// Empty raw text is the only disabled form; malformed saved rules never match all.
func CompileNodeConditionMatcher(raw, surface string) (func(Node) bool, error) {
	if strings.TrimSpace(raw) == "" {
		return func(Node) bool { return true }, nil
	}
	if len(raw) > 16384 {
		return nil, fmt.Errorf("附加条件不能超过16KB")
	}
	decoder := json.NewDecoder(bytes.NewBufferString(raw))
	decoder.DisallowUnknownFields()
	var conditions NodeConditions
	if err := decoder.Decode(&conditions); err != nil {
		return nil, fmt.Errorf("附加条件格式无效: %w", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return nil, fmt.Errorf("附加条件必须是单个JSON对象")
	}
	conditions.Logic = strings.ToLower(strings.TrimSpace(conditions.Logic))
	if conditions.Logic != "and" && conditions.Logic != "or" {
		return nil, fmt.Errorf("附加条件的组合方式必须为and或or")
	}
	if len(conditions.Conditions) == 0 || len(conditions.Conditions) > 32 {
		return nil, fmt.Errorf("启用附加条件时需配置1至32个条件")
	}
	metadata := make(map[string]NodeConditionFieldMeta)
	for _, field := range GetNodeConditionFields() {
		metadata[field.Value] = field
	}
	for _, condition := range conditions.Conditions {
		field, ok := metadata[condition.Field]
		if !ok {
			return nil, fmt.Errorf("不支持的条件字段: %s", condition.Field)
		}
		if surface == NodeConditionSmartCandidate && !isSmartCandidateField(condition.Field) {
			return nil, fmt.Errorf("候选条件不能使用检测结果字段: %s，请放在入组条件中", condition.Field)
		}
		allowed := false
		for _, operator := range field.Operators {
			if condition.Operator == operator {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, fmt.Errorf("字段%s不支持操作符%s", condition.Field, condition.Operator)
		}
		if condition.Value == nil {
			return nil, fmt.Errorf("条件值不能为空")
		}
		value := fmt.Sprint(condition.Value)
		if len(value) > 2048 {
			return nil, fmt.Errorf("条件值不能超过2048字符")
		}
		switch condition.Value.(type) {
		case string, float64:
		default:
			return nil, fmt.Errorf("条件值必须为文本或数字")
		}
		if field.DataType == "number" {
			number, err := strconv.ParseFloat(value, 64)
			if err != nil || math.IsNaN(number) || math.IsInf(number, 0) {
				return nil, fmt.Errorf("字段%s的条件值必须为有限数字", condition.Field)
			}
		}
		if field.OptionSource == "unlockStatuses" && NormalizeUnlockStatus(value) == "" {
			return nil, fmt.Errorf("字段%s的条件值无效", condition.Field)
		}
		if field.OptionSource == "unlockProviders" {
			found := false
			for _, provider := range GetUnlockProviderValues() {
				if NormalizeUnlockProvider(provider) == NormalizeUnlockProvider(value) {
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("字段%s的条件值无效", condition.Field)
			}
		}
		if field.DataType == "enum" {
			if strings.TrimSpace(value) == "" {
				return nil, fmt.Errorf("字段%s请选择条件值", condition.Field)
			}
			if len(field.Options) > 0 {
				found := false
				for _, option := range field.Options {
					if option.Value == value {
						found = true
						break
					}
				}
				if !found {
					return nil, fmt.Errorf("字段%s的条件值无效", condition.Field)
				}
			}
		}
		if condition.Operator == "regex" || condition.Operator == "contains" || condition.Operator == "not_contains" {
			if strings.TrimSpace(value) == "" {
				return nil, fmt.Errorf("匹配内容不能为空")
			}
		}
		if condition.Operator == "regex" {
			if _, err := regexp.Compile(value); err != nil {
				return nil, fmt.Errorf("正则表达式无效: %w", err)
			}
		}
	}
	// Build independent units in their saved order. Consecutive unlock conditions
	// retain legacy same-provider AND grouping, even under an outer OR expression.
	units := make([]func(Node) bool, 0, len(conditions.Conditions))
	for index := 0; index < len(conditions.Conditions); {
		if shouldStartUnlockConditionGroup(conditions.Conditions, index) {
			group, next := collectUnlockConditionGroup(conditions.Conditions, index)
			units = append(units, func(node Node) bool { return evaluateUnlockConditionGroup(node, group) })
			index = next
			continue
		}
		condition := conditions.Conditions[index]
		if condition.Operator == "regex" {
			expression, err := regexp.Compile(fmt.Sprint(condition.Value))
			if err != nil {
				return nil, err
			}
			units = append(units, func(node Node) bool {
				return expression.MatchString(fmt.Sprint(getNodeFieldValue(node, condition.Field)))
			})
		} else {
			units = append(units, func(node Node) bool { return evaluateCondition(node, condition) })
		}
		index++
	}
	return func(node Node) bool {
		for _, unit := range units {
			matches := unit(node)
			if conditions.Logic == "or" && matches {
				return true
			}
			if conditions.Logic == "and" && !matches {
				return false
			}
		}
		return conditions.Logic == "and"
	}, nil
}
