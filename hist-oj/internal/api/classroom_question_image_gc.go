package api

import (
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

const (
	questionImageUploadPrefix = "/uploads/classroom/questions/"
	questionImageUploadDir    = "./uploads/classroom/questions"
)

var (
	questionImageMarkdownRegexp = regexp.MustCompile(`!\[[^\]]*\]\(([^)]+)\)`)
	questionImageHTMLRegexp     = regexp.MustCompile(`(?i)<img[^>]*src\s*=\s*["']([^"']+)["'][^>]*>`)
	questionImageURLWithTitle   = regexp.MustCompile(`^(\S+)\s+["'][^"']*["']$`)
)

func normalizeQuestionImageURL(raw string) (string, bool) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", false
	}

	if strings.HasPrefix(value, "<") && strings.HasSuffix(value, ">") {
		value = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(value, "<"), ">"))
	}
	if value == "" {
		return "", false
	}

	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		parsed, err := url.Parse(value)
		if err != nil {
			return "", false
		}
		value = parsed.Path
	} else if strings.HasPrefix(value, "//") {
		parsed, err := url.Parse("https:" + value)
		if err != nil {
			return "", false
		}
		value = parsed.Path
	}

	value = strings.Split(value, "?")[0]
	value = strings.Split(value, "#")[0]
	if !strings.HasPrefix(value, questionImageUploadPrefix) {
		return "", false
	}

	filename := strings.TrimSpace(strings.TrimPrefix(value, questionImageUploadPrefix))
	if filename == "" {
		return "", false
	}
	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") || strings.Contains(filename, "..") {
		return "", false
	}
	if filepath.Base(filename) != filename {
		return "", false
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
		return "", false
	}

	return questionImageUploadPrefix + filename, true
}

func extractQuestionImageURLsFromText(text string) []string {
	if strings.TrimSpace(text) == "" {
		return nil
	}

	urls := make([]string, 0)

	markdownMatches := questionImageMarkdownRegexp.FindAllStringSubmatch(text, -1)
	for _, match := range markdownMatches {
		if len(match) < 2 {
			continue
		}
		candidate := strings.TrimSpace(match[1])
		if candidate == "" {
			continue
		}
		if mdURLWithTitle := questionImageURLWithTitle.FindStringSubmatch(candidate); len(mdURLWithTitle) > 1 {
			candidate = mdURLWithTitle[1]
		}
		if normalized, ok := normalizeQuestionImageURL(candidate); ok {
			urls = append(urls, normalized)
		}
	}

	htmlMatches := questionImageHTMLRegexp.FindAllStringSubmatch(text, -1)
	for _, match := range htmlMatches {
		if len(match) < 2 {
			continue
		}
		if normalized, ok := normalizeQuestionImageURL(match[1]); ok {
			urls = append(urls, normalized)
		}
	}

	return uniqueQuestionImageURLs(urls)
}

func uniqueQuestionImageURLs(urls []string) []string {
	if len(urls) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(urls))
	result := make([]string, 0, len(urls))
	for _, raw := range urls {
		normalized, ok := normalizeQuestionImageURL(raw)
		if !ok {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func extractQuestionImageURLsFromFields(title, content, analysis string, options *string) []string {
	all := make([]string, 0)
	all = append(all, extractQuestionImageURLsFromText(title)...)
	all = append(all, extractQuestionImageURLsFromText(content)...)
	all = append(all, extractQuestionImageURLsFromText(analysis)...)
	if options != nil {
		all = append(all, extractQuestionImageURLsFromText(*options)...)
	}
	return uniqueQuestionImageURLs(all)
}

func diffRemovedQuestionImageURLs(oldURLs, newURLs []string) []string {
	oldList := uniqueQuestionImageURLs(oldURLs)
	newList := uniqueQuestionImageURLs(newURLs)
	if len(oldList) == 0 {
		return nil
	}

	newSet := make(map[string]struct{}, len(newList))
	for _, item := range newList {
		newSet[item] = struct{}{}
	}

	removed := make([]string, 0)
	for _, item := range oldList {
		if _, exists := newSet[item]; !exists {
			removed = append(removed, item)
		}
	}
	return removed
}

func deleteQuestionImagesByURLs(logger *zap.Logger, urls []string) (deleted int, skipped int) {
	targetURLs := uniqueQuestionImageURLs(urls)
	if len(targetURLs) == 0 {
		return 0, 0
	}

	for _, imageURL := range targetURLs {
		filename := strings.TrimPrefix(imageURL, questionImageUploadPrefix)
		fullPath := filepath.Join(questionImageUploadDir, filename)
		err := os.Remove(fullPath)
		if err == nil {
			deleted++
			continue
		}
		if os.IsNotExist(err) {
			skipped++
			continue
		}

		logger.Warn("删除题目图片失败", zap.String("path", fullPath), zap.Error(err))
		skipped++
	}

	return deleted, skipped
}
