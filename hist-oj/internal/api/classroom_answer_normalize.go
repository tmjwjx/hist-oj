package api

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/hoj/hist-oj/internal/model"
)

var choiceOptionPrefixPattern = regexp.MustCompile(`^\s*([A-Za-z])[\.\)、:：]`)

// normalizeQuestionAnswerForStorage 规范化题库标准答案与选项存储格式
func normalizeQuestionAnswerForStorage(questionType string, rawAnswer string, rawOptions *string) (string, *string, error) {
	switch questionType {
	case "single_choice", "multiple_choice":
		if rawOptions == nil || strings.TrimSpace(*rawOptions) == "" {
			return "", nil, fmt.Errorf("选择题必须提供选项")
		}
		normalizedOptions, optionSet, err := normalizeChoiceOptions(*rawOptions)
		if err != nil {
			return "", nil, err
		}

		var normalizedAnswer string
		if questionType == "single_choice" {
			normalizedAnswer, err = normalizeSingleChoiceAnswer(rawAnswer, optionSet, false)
		} else {
			normalizedAnswer, err = normalizeMultipleChoiceAnswer(rawAnswer, optionSet, false)
		}
		if err != nil {
			return "", nil, err
		}

		return normalizedAnswer, &normalizedOptions, nil
	case "judge":
		normalized := normalizeJudgeAnswer(rawAnswer)
		if normalized != "true" && normalized != "false" {
			return "", nil, fmt.Errorf("判断题答案必须是 true 或 false")
		}
		return normalized, nil, nil
	case "subjective":
		return strings.TrimSpace(rawAnswer), nil, nil
	case "programming":
		return "", nil, nil
	default:
		return "", nil, fmt.Errorf("不支持的题型: %s", questionType)
	}
}

// normalizeStudentAnswerForStorage 规范化学生提交答案格式（用于草稿和正式提交）
func normalizeStudentAnswerForStorage(question *model.QuestionBank, rawAnswer string) (string, error) {
	switch question.Type {
	case "single_choice":
		optionSet := parseChoiceOptionSetLenient(question.Options)
		return normalizeSingleChoiceAnswer(rawAnswer, optionSet, true)
	case "multiple_choice":
		optionSet := parseChoiceOptionSetLenient(question.Options)
		return normalizeMultipleChoiceAnswer(rawAnswer, optionSet, true)
	case "judge":
		if strings.TrimSpace(rawAnswer) == "" {
			return "", nil
		}
		normalized := normalizeJudgeAnswer(rawAnswer)
		if normalized != "true" && normalized != "false" {
			return "", fmt.Errorf("判断题答案格式错误")
		}
		return normalized, nil
	case "subjective":
		return strings.TrimSpace(rawAnswer), nil
	default:
		return rawAnswer, nil
	}
}

func normalizeChoiceOptions(raw string) (string, map[string]struct{}, error) {
	var options []string
	if err := json.Unmarshal([]byte(raw), &options); err != nil {
		return "", nil, fmt.Errorf("选项必须是JSON数组")
	}
	if len(options) == 0 {
		return "", nil, fmt.Errorf("选项不能为空")
	}

	normalizedOptions := make([]string, 0, len(options))
	optionSet := make(map[string]struct{}, len(options))
	for idx, opt := range options {
		trimmed := strings.TrimSpace(opt)
		if trimmed == "" {
			return "", nil, fmt.Errorf("第%d个选项为空", idx+1)
		}
		normalizedOptions = append(normalizedOptions, trimmed)

		optionID := extractOptionID(trimmed, idx)
		if optionID != "" {
			optionSet[optionID] = struct{}{}
		}
	}

	normalizedBytes, err := json.Marshal(normalizedOptions)
	if err != nil {
		return "", nil, fmt.Errorf("序列化选项失败")
	}
	return string(normalizedBytes), optionSet, nil
}

func normalizeSingleChoiceAnswer(raw string, optionSet map[string]struct{}, allowEmpty bool) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		if allowEmpty {
			return "", nil
		}
		return "", fmt.Errorf("单选题答案不能为空")
	}

	optionID := extractOptionID(trimmed, -1)
	if optionID == "" {
		return "", fmt.Errorf("单选题答案格式错误")
	}
	if len(optionSet) > 0 {
		if _, ok := optionSet[optionID]; !ok {
			return "", fmt.Errorf("单选题答案不在选项范围内")
		}
	}

	return optionID, nil
}

func normalizeMultipleChoiceAnswer(raw string, optionSet map[string]struct{}, allowEmpty bool) (string, error) {
	optionIDs, err := parseMultipleChoiceAnswerIDs(raw)
	if err != nil {
		return "", fmt.Errorf("多选题答案格式错误")
	}
	if len(optionIDs) == 0 {
		if allowEmpty {
			return "[]", nil
		}
		return "", fmt.Errorf("多选题答案不能为空")
	}

	if len(optionSet) > 0 {
		for _, optionID := range optionIDs {
			if _, ok := optionSet[optionID]; !ok {
				return "", fmt.Errorf("多选题答案包含无效选项")
			}
		}
	}

	sort.Strings(optionIDs)
	answerBytes, err := json.Marshal(optionIDs)
	if err != nil {
		return "", fmt.Errorf("序列化多选题答案失败")
	}
	return string(answerBytes), nil
}

func parseMultipleChoiceAnswerIDs(raw string) ([]string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return []string{}, nil
	}

	var rawItems []string
	if strings.HasPrefix(trimmed, "[") {
		var arr []interface{}
		if err := json.Unmarshal([]byte(trimmed), &arr); err != nil {
			return nil, err
		}
		rawItems = make([]string, 0, len(arr))
		for _, item := range arr {
			rawItems = append(rawItems, fmt.Sprintf("%v", item))
		}
	} else {
		rawItems = strings.Split(trimmed, ",")
	}

	unique := make(map[string]struct{}, len(rawItems))
	for _, item := range rawItems {
		optionID := extractOptionID(item, -1)
		if optionID == "" {
			continue
		}
		unique[optionID] = struct{}{}
	}

	result := make([]string, 0, len(unique))
	for optionID := range unique {
		result = append(result, optionID)
	}
	return result, nil
}

func parseChoiceOptionSetLenient(options *string) map[string]struct{} {
	if options == nil || strings.TrimSpace(*options) == "" {
		return map[string]struct{}{}
	}
	_, optionSet, err := normalizeChoiceOptions(*options)
	if err != nil {
		// 老数据异常时不阻塞提交，仅跳过选项范围校验
		return map[string]struct{}{}
	}
	return optionSet
}

func extractOptionID(raw string, fallbackIndex int) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	if matches := choiceOptionPrefixPattern.FindStringSubmatch(trimmed); len(matches) == 2 {
		return strings.ToUpper(matches[1])
	}

	for _, r := range trimmed {
		if unicode.IsLetter(r) {
			upper := strings.ToUpper(string(r))
			if len(upper) == 1 && upper[0] >= 'A' && upper[0] <= 'Z' {
				return upper
			}
		}
	}

	if fallbackIndex >= 0 && fallbackIndex < 26 {
		return string(rune('A' + fallbackIndex))
	}

	return ""
}
