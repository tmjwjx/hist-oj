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

type compositeSubQuestion struct {
	ID      string   `json:"id"`
	Content string   `json:"content"`
	Options []string `json:"options"`
	Score   int      `json:"score"`
}

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
	case "composite":
		if rawOptions == nil || strings.TrimSpace(*rawOptions) == "" {
			return "", nil, fmt.Errorf("组合题必须提供子题配置")
		}

		subQuestions, normalizedOptions, err := normalizeCompositeSubQuestions(*rawOptions, true)
		if err != nil {
			return "", nil, err
		}

		normalizedAnswer, err := normalizeCompositeCorrectAnswers(rawAnswer, subQuestions)
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
	case "fill_blank":
		normalizedAnswers, err := normalizeFillBlankAnswerList(rawAnswer, false)
		if err != nil {
			return "", nil, err
		}
		answerBytes, err := json.Marshal(normalizedAnswers)
		if err != nil {
			return "", nil, fmt.Errorf("序列化填空题答案失败")
		}
		return string(answerBytes), nil, nil
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
	case "fill_blank":
		return normalizeFillBlankStorageValue(rawAnswer), nil
	case "composite":
		subQuestions, _, err := parseCompositeSubQuestionsFromStored(question.Options)
		if err != nil {
			if strings.TrimSpace(rawAnswer) == "" {
				return "{}", nil
			}
			return "", fmt.Errorf("组合题配置错误")
		}
		normalizedAnswer, err := normalizeCompositeStudentAnswers(rawAnswer, subQuestions, true)
		if err != nil {
			return "", err
		}
		return normalizedAnswer, nil
	default:
		return rawAnswer, nil
	}
}

func normalizeCompositeSubQuestions(raw string, strict bool) ([]compositeSubQuestion, string, error) {
	type rawSubQuestion struct {
		ID            string          `json:"id"`
		Content       string          `json:"content"`
		Options       json.RawMessage `json:"options"`
		ChoiceOptions json.RawMessage `json:"choiceOptions"`
		Score         *int            `json:"score"`
		SubScore      *int            `json:"subScore"`
	}

	var rawItems []rawSubQuestion
	if err := json.Unmarshal([]byte(raw), &rawItems); err != nil {
		return nil, "", fmt.Errorf("组合题子题必须是JSON数组")
	}
	if len(rawItems) == 0 {
		return nil, "", fmt.Errorf("组合题至少需要一个子题")
	}

	subQuestions := make([]compositeSubQuestion, 0, len(rawItems))
	seenID := make(map[string]struct{}, len(rawItems))

	for idx, item := range rawItems {
		subID := strings.TrimSpace(item.ID)
		if subID == "" {
			subID = fmt.Sprintf("sq_%d", idx+1)
		}
		if _, exists := seenID[subID]; exists {
			return nil, "", fmt.Errorf("组合题子题ID重复: %s", subID)
		}
		seenID[subID] = struct{}{}

		content := strings.TrimSpace(item.Content)
		if content == "" {
			return nil, "", fmt.Errorf("第%d个子题题干不能为空", idx+1)
		}

		score := 0
		if item.SubScore != nil {
			score = *item.SubScore
		} else if item.Score != nil {
			score = *item.Score
		}
		if score <= 0 {
			return nil, "", fmt.Errorf("第%d个子题分值必须大于0", idx+1)
		}

		optionsRaw := item.Options
		if len(optionsRaw) == 0 {
			optionsRaw = item.ChoiceOptions
		}
		if len(optionsRaw) == 0 {
			return nil, "", fmt.Errorf("第%d个子题必须提供4个选项", idx+1)
		}

		var options []string
		if err := json.Unmarshal(optionsRaw, &options); err != nil {
			return nil, "", fmt.Errorf("第%d个子题选项格式错误", idx+1)
		}
		if len(options) != 4 {
			return nil, "", fmt.Errorf("第%d个子题必须且仅能有4个选项", idx+1)
		}

		normalizedOptions := make([]string, 0, 4)
		for optionIndex, option := range options {
			cleanContent := normalizeChoiceOptionContent(option)
			if cleanContent == "" {
				return nil, "", fmt.Errorf("第%d个子题第%d个选项不能为空", idx+1, optionIndex+1)
			}
			normalizedOptions = append(normalizedOptions,
				fmt.Sprintf("%s. %s", string(rune('A'+optionIndex)), cleanContent))
		}

		subQuestions = append(subQuestions, compositeSubQuestion{
			ID:      subID,
			Content: content,
			Options: normalizedOptions,
			Score:   score,
		})
	}

	if strict {
		totalScore := 0
		for _, subQuestion := range subQuestions {
			totalScore += subQuestion.Score
		}
		if totalScore <= 0 {
			return nil, "", fmt.Errorf("组合题总分必须大于0")
		}
	}

	normalizedBytes, err := json.Marshal(subQuestions)
	if err != nil {
		return nil, "", fmt.Errorf("序列化组合题子题失败")
	}
	return subQuestions, string(normalizedBytes), nil
}

func normalizeCompositeCorrectAnswers(rawAnswer string, subQuestions []compositeSubQuestion) (string, error) {
	trimmed := strings.TrimSpace(rawAnswer)
	if trimmed == "" {
		return "", fmt.Errorf("组合题答案不能为空")
	}

	var rawAnswerMap map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed), &rawAnswerMap); err != nil {
		return "", fmt.Errorf("组合题答案必须是JSON对象")
	}

	subQuestionMap := make(map[string]compositeSubQuestion, len(subQuestions))
	for _, subQuestion := range subQuestions {
		subQuestionMap[subQuestion.ID] = subQuestion
	}

	normalized := make(map[string]string, len(subQuestions))
	for _, subQuestion := range subQuestions {
		rawValue, exists := rawAnswerMap[subQuestion.ID]
		if !exists {
			return "", fmt.Errorf("子题 %s 缺少正确答案", subQuestion.ID)
		}
		optionID := extractOptionID(fmt.Sprintf("%v", rawValue), -1)
		if optionID == "" || optionID < "A" || optionID > "D" {
			return "", fmt.Errorf("子题 %s 的正确答案必须是A-D", subQuestion.ID)
		}
		normalized[subQuestion.ID] = optionID
	}

	for key := range rawAnswerMap {
		if _, exists := subQuestionMap[key]; !exists {
			return "", fmt.Errorf("组合题答案包含未知子题ID: %s", key)
		}
	}

	answerBytes, err := json.Marshal(normalized)
	if err != nil {
		return "", fmt.Errorf("序列化组合题答案失败")
	}
	return string(answerBytes), nil
}

func normalizeCompositeStudentAnswers(rawAnswer string, subQuestions []compositeSubQuestion, allowEmpty bool) (string, error) {
	trimmed := strings.TrimSpace(rawAnswer)
	if trimmed == "" {
		if allowEmpty {
			return "{}", nil
		}
		return "", fmt.Errorf("组合题答案不能为空")
	}

	var rawAnswerMap map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed), &rawAnswerMap); err != nil {
		return "", fmt.Errorf("组合题答案格式错误")
	}

	optionSetMap := make(map[string]map[string]struct{}, len(subQuestions))
	subQuestionMap := make(map[string]struct{}, len(subQuestions))
	for _, subQuestion := range subQuestions {
		subQuestionMap[subQuestion.ID] = struct{}{}
		optionSet := make(map[string]struct{}, 4)
		for optionIndex, option := range subQuestion.Options {
			optionID := extractOptionID(option, optionIndex)
			if optionID != "" {
				optionSet[optionID] = struct{}{}
			}
		}
		optionSetMap[subQuestion.ID] = optionSet
	}

	normalized := make(map[string]string, len(rawAnswerMap))
	for subID, rawValue := range rawAnswerMap {
		if _, exists := subQuestionMap[subID]; !exists {
			return "", fmt.Errorf("组合题答案包含未知子题ID: %s", subID)
		}
		optionID := extractOptionID(fmt.Sprintf("%v", rawValue), -1)
		if optionID == "" {
			continue
		}
		if optionSet, exists := optionSetMap[subID]; exists && len(optionSet) > 0 {
			if _, ok := optionSet[optionID]; !ok {
				return "", fmt.Errorf("子题 %s 的答案不在选项范围内", subID)
			}
		}
		normalized[subID] = optionID
	}

	answerBytes, err := json.Marshal(normalized)
	if err != nil {
		return "", fmt.Errorf("序列化组合题答案失败")
	}
	return string(answerBytes), nil
}

func parseCompositeSubQuestionsFromStored(options *string) ([]compositeSubQuestion, map[string]compositeSubQuestion, error) {
	if options == nil || strings.TrimSpace(*options) == "" {
		return nil, nil, fmt.Errorf("组合题配置为空")
	}

	subQuestions, _, err := normalizeCompositeSubQuestions(*options, false)
	if err != nil {
		return nil, nil, err
	}

	subQuestionMap := make(map[string]compositeSubQuestion, len(subQuestions))
	for _, subQuestion := range subQuestions {
		subQuestionMap[subQuestion.ID] = subQuestion
	}
	return subQuestions, subQuestionMap, nil
}

func calculateCompositeQuestionScore(question *model.QuestionBank, studentAnswer string) (float64, error) {
	subQuestions, subQuestionMap, err := parseCompositeSubQuestionsFromStored(question.Options)
	if err != nil {
		return 0, err
	}

	var correctAnswerMap map[string]string
	if err := json.Unmarshal([]byte(strings.TrimSpace(question.Answer)), &correctAnswerMap); err != nil {
		return 0, fmt.Errorf("组合题标准答案格式错误")
	}

	normalizedStudentAnswer, err := normalizeCompositeStudentAnswers(studentAnswer, subQuestions, true)
	if err != nil {
		return 0, err
	}

	studentAnswerMap := make(map[string]string)
	if strings.TrimSpace(normalizedStudentAnswer) != "" {
		if err := json.Unmarshal([]byte(normalizedStudentAnswer), &studentAnswerMap); err != nil {
			return 0, fmt.Errorf("组合题学生答案格式错误")
		}
	}

	total := 0.0
	for subID, correctOption := range correctAnswerMap {
		subQuestion, exists := subQuestionMap[subID]
		if !exists {
			return 0, fmt.Errorf("组合题标准答案包含未知子题ID: %s", subID)
		}
		if studentAnswerMap[subID] == correctOption {
			total += float64(subQuestion.Score)
		}
	}

	return total, nil
}

func calculateAutoObjectiveScore(question *model.QuestionBank, studentAnswer string, questionFullScore float64) (float64, bool, error) {
	switch question.Type {
	case "single_choice":
		if strings.TrimSpace(studentAnswer) == strings.TrimSpace(question.Answer) {
			return questionFullScore, true, nil
		}
		return 0, true, nil
	case "judge":
		if compareJudgeAnswers(studentAnswer, question.Answer) {
			return questionFullScore, true, nil
		}
		return 0, true, nil
	case "multiple_choice":
		studentAnswers := parseChoiceAnswersLenient(studentAnswer)
		correctAnswers := parseChoiceAnswersLenient(question.Answer)
		if compareArrays(studentAnswers, correctAnswers) {
			return questionFullScore, true, nil
		}
		return 0, true, nil
	case "composite":
		score, err := calculateCompositeQuestionScore(question, studentAnswer)
		if err != nil {
			return 0, true, err
		}
		return score, true, nil
	case "fill_blank":
		if isFillBlankAnswerCorrect(studentAnswer, question.Answer) {
			return questionFullScore, true, nil
		}
		return 0, true, nil
	default:
		return 0, false, nil
	}
}

func normalizeFillBlankStorageValue(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	return strings.Join(strings.Fields(trimmed), " ")
}

func normalizeFillBlankAnswerList(raw string, allowEmpty bool) ([]string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		if allowEmpty {
			return []string{}, nil
		}
		return nil, fmt.Errorf("填空题至少需要一个有效答案")
	}

	values := make([]string, 0)
	if strings.HasPrefix(trimmed, "[") {
		var arr []interface{}
		if err := json.Unmarshal([]byte(trimmed), &arr); err != nil {
			return nil, fmt.Errorf("填空题答案格式错误")
		}
		for _, item := range arr {
			normalized := normalizeFillBlankStorageValue(fmt.Sprintf("%v", item))
			if normalized != "" {
				values = append(values, normalized)
			}
		}
	} else {
		normalized := normalizeFillBlankStorageValue(trimmed)
		if normalized != "" {
			values = append(values, normalized)
		}
	}

	unique := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		unique = append(unique, value)
	}

	if len(unique) == 0 {
		if allowEmpty {
			return []string{}, nil
		}
		return nil, fmt.Errorf("填空题至少需要一个有效答案")
	}

	return unique, nil
}

func isFillBlankAnswerCorrect(studentAnswer string, correctAnswerRaw string) bool {
	studentAnswers, err := normalizeFillBlankAnswerList(studentAnswer, true)
	if err != nil || len(studentAnswers) == 0 {
		return false
	}

	correctAnswers, err := normalizeFillBlankAnswerList(correctAnswerRaw, false)
	if err != nil || len(correctAnswers) == 0 {
		return false
	}

	correctExact := make(map[string]struct{}, len(correctAnswers))
	correctFold := make(map[string]struct{}, len(correctAnswers))
	for _, answer := range correctAnswers {
		correctExact[answer] = struct{}{}
		correctFold[strings.ToLower(answer)] = struct{}{}
	}

	for _, answer := range studentAnswers {
		if _, ok := correctExact[answer]; ok {
			return true
		}
		if _, ok := correctFold[strings.ToLower(answer)]; ok {
			return true
		}
	}

	return false
}

func isAutoScoredObjectiveQuestionType(questionType string) bool {
	switch questionType {
	case "single_choice", "multiple_choice", "judge", "fill_blank", "composite":
		return true
	default:
		return false
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

func normalizeChoiceOptionContent(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	if matches := choiceOptionPrefixPattern.FindStringSubmatch(trimmed); len(matches) == 2 {
		prefix := matches[0]
		trimmed = strings.TrimSpace(trimmed[len(prefix):])
	}

	return trimmed
}

func parseChoiceAnswersLenient(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return []string{}
	}

	var values []string
	if strings.HasPrefix(trimmed, "[") {
		var arr []interface{}
		if err := json.Unmarshal([]byte(trimmed), &arr); err == nil {
			for _, item := range arr {
				optionID := extractOptionID(fmt.Sprintf("%v", item), -1)
				if optionID != "" {
					values = append(values, optionID)
				}
			}
			if len(values) > 0 {
				return values
			}
		}
	}

	for _, part := range strings.Split(trimmed, ",") {
		optionID := extractOptionID(part, -1)
		if optionID != "" {
			values = append(values, optionID)
		}
	}

	return values
}
