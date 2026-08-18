package top.hcode.hoj.manager.classroom;

import top.hcode.hoj.common.exception.StatusFailException;

import java.util.Arrays;
import java.util.stream.Collectors;

final class QuestionAnswerNormalizer {
    private static final String[] TYPES = {"single_choice", "multiple_choice", "judge", "fill_blank", "subjective", "programming", "composite"};

    private QuestionAnswerNormalizer() { }

    static String[] normalize(String type, String answer, String options) throws StatusFailException {
        if (!Arrays.asList(TYPES).contains(type)) throw new StatusFailException("题型不合法");
        if (("single_choice".equals(type) || "multiple_choice".equals(type) || "composite".equals(type))
                && blank(options)) throw new StatusFailException("该题型必须提供选项配置");
        if ("judge".equals(type)) {
            String value = answer == null ? "" : answer.trim().toLowerCase();
            if ("是".equals(value) || "正确".equals(value) || "1".equals(value)) value = "true";
            if ("否".equals(value) || "错误".equals(value) || "0".equals(value)) value = "false";
            if (!"true".equals(value) && !"false".equals(value)) throw new StatusFailException("判断题答案必须是 true 或 false");
            answer = value;
        } else if ("multiple_choice".equals(type)) {
            answer = answer == null ? "" : Arrays.stream(answer.toUpperCase().replace("，", ",").split("[,、;；\\s]+"))
                    .filter(v -> !v.isEmpty()).distinct().sorted().collect(Collectors.joining(","));
        } else if (answer != null) {
            answer = answer.trim();
        }
        if ("fill_blank".equals(type) && answer != null && !answer.trim().startsWith("[")) {
            answer = "[\"" + answer.replace("\"", "\\\"").replace("|", "\",\"") + "\"]";
        }
        return new String[]{answer == null ? "" : answer, options == null || options.trim().isEmpty() ? null : options};
    }

    private static boolean blank(String value) { return value == null || value.trim().isEmpty(); }
}
