package top.hcode.hoj.manager.classroom;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import top.hcode.hoj.pojo.entity.classroom.HomeworkQuestion;
import top.hcode.hoj.pojo.entity.classroom.HomeworkSubmit;
import top.hcode.hoj.pojo.entity.classroom.QuestionBank;

import java.util.*;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

final class HomeworkAnalysisSupport {
    private static final ObjectMapper JSON = new ObjectMapper();
    private static final Pattern OPTION_PREFIX = Pattern.compile("^\\s*([A-Za-z0-9]+)[.、:：]\\s*(.*)$");

    private HomeworkAnalysisSupport() { }

    static List<Map<String, Object>> distributions(QuestionBank question,
                                                    HomeworkQuestion relation,
                                                    List<HomeworkSubmit> submissions,
                                                    Map<String, Map<String, Object>> students) {
        String type = question == null ? "programming" : text(question.getType());
        if ("fill_blank".equals(type)) return fillBlank(submissions, students);
        if (isChoice(type)) return choices(question, submissions, students);
        return scores(relation.getScore(), submissions, students);
    }

    static List<String> answerList(String raw, String type) {
        if (blank(raw)) return Collections.emptyList();
        try {
            JsonNode value = JSON.readTree(raw);
            if (value != null && value.isArray()) {
                List<String> result = new ArrayList<>();
                for (JsonNode item : value) addAnswer(result, nodeText(item));
                return result;
            }
            if (value != null && value.isValueNode()) {
                String parsed = nodeText(value);
                if (!blank(parsed)) return Collections.singletonList(parsed.trim());
            }
        } catch (Exception ignored) { }
        String value = raw.trim();
        if ("multiple_choice".equals(type)) return split(value, "[,，、;；\\s]+");
        if ("fill_blank".equals(type) && value.contains("|")) return split(value, "\\|");
        return Collections.singletonList(value);
    }

    static String displayAnswer(String raw, String type) {
        List<String> values = answerList(raw, type);
        if ("multiple_choice".equals(type)) {
            Set<String> normalized = new TreeSet<>();
            for (String value : values) {
                String item = normalizeChoice(value);
                if (!blank(item)) normalized.add(item);
            }
            return String.join(",", normalized);
        }
        if ("single_choice".equals(type) && !values.isEmpty()) return normalizeChoice(values.get(0));
        if ("fill_blank".equals(type)) return String.join(" / ", values);
        return values.isEmpty() ? "" : values.get(0);
    }

    private static List<Map<String, Object>> fillBlank(List<HomeworkSubmit> submissions,
                                                        Map<String, Map<String, Object>> students) {
        Map<String, List<HomeworkSubmit>> groups = new HashMap<>();
        for (HomeworkSubmit row : submissions) {
            List<String> answers = answerList(row.getAnswer(), "fill_blank");
            String label = answers.isEmpty() ? "未作答" : text(answers.get(0)).trim();
            if (blank(label)) label = "未作答";
            groups.computeIfAbsent(label, key -> new ArrayList<>()).add(row);
        }
        List<String> labels = new ArrayList<>(groups.keySet());
        labels.sort((left, right) -> {
            int count = Integer.compare(groups.get(right).size(), groups.get(left).size());
            return count == 0 ? left.compareTo(right) : count;
        });
        List<Map<String, Object>> result = new ArrayList<>();
        for (String label : labels) {
            List<HomeworkSubmit> rows = groups.get(label);
            result.add(option(label, label, rows, submissions.size(), students));
        }
        return result;
    }

    private static List<Map<String, Object>> choices(QuestionBank question,
                                                      List<HomeworkSubmit> submissions,
                                                      Map<String, Map<String, Object>> students) {
        List<Choice> options = parseOptions(question);
        List<Map<String, Object>> result = new ArrayList<>();
        for (Choice option : options) {
            List<HomeworkSubmit> selected = new ArrayList<>();
            for (HomeworkSubmit row : submissions) {
                if (selected(question.getType(), option.label, answerList(row.getAnswer(), question.getType()))) {
                    selected.add(row);
                }
            }
            result.add(option(option.label, option.content, selected, submissions.size(), students));
        }
        return result;
    }

    private static List<Map<String, Object>> scores(Integer fullScoreValue,
                                                     List<HomeworkSubmit> submissions,
                                                     Map<String, Map<String, Object>> students) {
        double fullScore = fullScoreValue == null ? 0d : Math.max(0d, fullScoreValue.doubleValue());
        List<ScoreBucket> buckets = Arrays.asList(
                new ScoreBucket("满分", fullScore, Double.POSITIVE_INFINITY),
                new ScoreBucket("80%以上", fullScore * 0.8d, fullScore),
                new ScoreBucket("60%-80%", fullScore * 0.6d, fullScore * 0.8d),
                new ScoreBucket("60%以下", Double.NEGATIVE_INFINITY, fullScore * 0.6d));
        for (HomeworkSubmit row : submissions) {
            double score = row.getScore() == null ? 0d : row.getScore();
            for (ScoreBucket bucket : buckets) {
                if (bucket.accept(score, fullScore)) {
                    bucket.rows.add(row);
                    break;
                }
            }
        }
        List<Map<String, Object>> result = new ArrayList<>();
        for (ScoreBucket bucket : buckets) {
            result.add(option(bucket.label, bucket.label, bucket.rows, submissions.size(), students));
        }
        return result;
    }

    private static Map<String, Object> option(String label, String content,
                                               List<HomeworkSubmit> rows, int total,
                                               Map<String, Map<String, Object>> students) {
        Map<String, Object> value = new LinkedHashMap<>();
        value.put("label", label);
        value.put("content", content);
        value.put("selectedCount", rows.size());
        value.put("percentage", total == 0 ? 0d : rows.size() * 100d / total);
        List<Map<String, Object>> selectedBy = new ArrayList<>();
        for (HomeworkSubmit row : rows) {
            Map<String, Object> student = students.get(row.getUid());
            if (student == null) continue;
            Map<String, Object> item = new LinkedHashMap<>(student);
            item.put("score", score(row));
            selectedBy.add(item);
        }
        value.put("selectedBy", selectedBy);
        return value;
    }

    private static List<Choice> parseOptions(QuestionBank question) {
        if (question != null && "judge".equals(question.getType())) {
            return Arrays.asList(new Choice("对", "正确"), new Choice("错", "错误"));
        }
        List<Choice> result = new ArrayList<>();
        String raw = question == null ? null : question.getOptions();
        if (!blank(raw)) {
            try {
                JsonNode options = JSON.readTree(raw);
                if (options != null && options.isArray()) {
                    int index = 0;
                    for (JsonNode item : options) {
                        result.add(parseOption(item, index++));
                    }
                }
            } catch (Exception ignored) { }
        }
        if (result.isEmpty()) {
            for (int i = 0; i < 4; i++) {
                String label = generatedLabel(i);
                result.add(new Choice(label, "选项" + label));
            }
        }
        return result;
    }

    private static Choice parseOption(JsonNode item, int index) {
        String fallback = generatedLabel(index);
        if (item != null && item.isObject()) {
            String label = first(item, "label", "letter", "value");
            String content = first(item, "content", "text", "title");
            if (blank(label)) label = fallback;
            if (blank(content)) content = label;
            return new Choice(label.trim(), content.trim());
        }
        String raw = nodeText(item).trim();
        Matcher matcher = OPTION_PREFIX.matcher(raw);
        if (matcher.matches()) return new Choice(matcher.group(1).toUpperCase(), matcher.group(2).trim());
        return new Choice(fallback, raw);
    }

    private static boolean selected(String type, String label, List<String> answers) {
        for (String answer : answers) {
            if ("judge".equals(type)) {
                Boolean expected = judge(label), actual = judge(answer);
                if (expected != null && expected.equals(actual)) return true;
            } else if (normalizeChoice(label).equals(normalizeChoice(answer))) {
                return true;
            }
        }
        return false;
    }

    private static String normalizeChoice(String value) {
        String normalized = text(value).trim().toUpperCase()
                .replaceAll("^[\\[\\(\\{\\\"']+|[\\]\\)\\}\\\"']+$", "").trim();
        Matcher matcher = OPTION_PREFIX.matcher(normalized);
        return matcher.matches() ? matcher.group(1) : normalized;
    }

    private static Boolean judge(String value) {
        String normalized = text(value).trim().toLowerCase();
        if (Arrays.asList("对", "正确", "true", "1", "是").contains(normalized)) return Boolean.TRUE;
        if (Arrays.asList("错", "错误", "false", "0", "否").contains(normalized)) return Boolean.FALSE;
        return null;
    }

    private static boolean isChoice(String type) {
        return "single_choice".equals(type) || "multiple_choice".equals(type) || "judge".equals(type);
    }

    private static List<String> split(String value, String expression) {
        List<String> result = new ArrayList<>();
        for (String part : value.split(expression)) addAnswer(result, part);
        return result;
    }

    private static void addAnswer(List<String> target, String value) {
        if (!blank(value)) target.add(value.trim());
    }

    private static String first(JsonNode value, String... keys) {
        for (String key : keys) {
            JsonNode child = value.get(key);
            if (child != null && !child.isNull() && !blank(nodeText(child))) return nodeText(child);
        }
        return "";
    }

    private static String nodeText(JsonNode value) {
        if (value == null || value.isNull()) return "";
        return value.isValueNode() ? value.asText() : value.toString();
    }

    private static String generatedLabel(int index) {
        return index < 26 ? String.valueOf((char) ('A' + index)) : String.valueOf(index + 1);
    }

    private static double score(HomeworkSubmit row) {
        return row.getScore() == null ? 0d : row.getScore();
    }

    private static boolean blank(String value) {
        return value == null || value.trim().isEmpty();
    }

    private static String text(Object value) {
        return value == null ? "" : String.valueOf(value);
    }

    private static final class Choice {
        private final String label;
        private final String content;

        private Choice(String label, String content) {
            this.label = label;
            this.content = content;
        }
    }

    private static final class ScoreBucket {
        private final String label;
        private final double min;
        private final double max;
        private final List<HomeworkSubmit> rows = new ArrayList<>();

        private ScoreBucket(String label, double min, double max) {
            this.label = label;
            this.min = min;
            this.max = max;
        }

        private boolean accept(double score, double fullScore) {
            if ("满分".equals(label)) return fullScore > 0d && score >= fullScore;
            return score >= min && score < max;
        }
    }
}
