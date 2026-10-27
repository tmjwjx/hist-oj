package top.hcode.hoj.service.problem;

import cn.hutool.json.JSONUtil;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.utils.Constants;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Objects;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/** 判断题目改动是否会影响标准程序验题结果。 */
public final class ProblemVerificationChangeDetector {
    private static final Pattern SAMPLE_PATTERN = Pattern.compile(
            "<input>([\\s\\S]*?)</input>\\s*<output>([\\s\\S]*?)</output>"
                    + "(?:\\s*<explanation>[\\s\\S]*?</explanation>)?");

    private ProblemVerificationChangeDetector() {
    }

    public static boolean changed(Problem before, Problem after) {
        if (before == null || after == null) return false;
        return !Objects.equals(before.getCaseVersion(), after.getCaseVersion())
                || !Objects.equals(sampleData(before.getExamples()), sampleData(after.getExamples()))
                || judgeConfigurationChanged(before, after)
                || !Objects.equals(before.getTimeLimit(), after.getTimeLimit())
                || !Objects.equals(before.getMemoryLimit(), after.getMemoryLimit())
                || !Objects.equals(before.getStackLimit(), after.getStackLimit())
                || !Objects.equals(before.getIsRemoveEndBlank(), after.getIsRemoveEndBlank())
                || !Objects.equals(before.getIsFileIO(), after.getIsFileIO())
                || (Boolean.TRUE.equals(after.getIsFileIO())
                    && (!Objects.equals(before.getIoReadFileName(), after.getIoReadFileName())
                        || !Objects.equals(before.getIoWriteFileName(), after.getIoWriteFileName())))
                || !Objects.equals(extraFiles(before.getUserExtraFile()), extraFiles(after.getUserExtraFile()));
    }

    /** 这些配置需要重建测试数据 info；客户端的 changeModeCode 标记不能代替实际比较。 */
    public static boolean judgeConfigurationChanged(Problem before, Problem after) {
        return !Objects.equals(before.getType(), after.getType())
                || !Objects.equals(before.getJudgeMode(), after.getJudgeMode())
                || !Objects.equals(before.getJudgeCaseMode(), after.getJudgeCaseMode())
                || (!Constants.JudgeMode.DEFAULT.getMode().equals(after.getJudgeMode())
                    && (!Objects.equals(before.getSpjCode(), after.getSpjCode())
                        || !Objects.equals(before.getSpjLanguage(), after.getSpjLanguage())))
                || !Objects.equals(extraFiles(before.getJudgeExtraFile()), extraFiles(after.getJudgeExtraFile()));
    }

    private static Object sampleData(String examples) {
        if (examples == null || examples.trim().isEmpty()) return Collections.emptyList();
        List<String> data = new ArrayList<>();
        Matcher matcher = SAMPLE_PATTERN.matcher(examples);
        int cursor = 0;
        while (matcher.find()) {
            // 无法完整解析时按原文比较，不能把损坏的样例当作已通过验证。
            if (!examples.substring(cursor, matcher.start()).trim().isEmpty()) return examples;
            data.add(matcher.group(1).replace("\r\n", "\n").replace('\r', '\n'));
            data.add(matcher.group(2).replace("\r\n", "\n").replace('\r', '\n'));
            cursor = matcher.end();
        }
        return cursor == 0 || !examples.substring(cursor).trim().isEmpty() ? examples : data;
    }

    private static Object extraFiles(String value) {
        if (value == null || value.trim().isEmpty()) return Collections.emptyMap();
        try {
            return JSONUtil.parseObj(value);
        } catch (RuntimeException invalidJson) {
            return value;
        }
    }
}
