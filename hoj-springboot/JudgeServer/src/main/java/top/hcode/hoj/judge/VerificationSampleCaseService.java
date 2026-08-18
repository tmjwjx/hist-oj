package top.hcode.hoj.judge;

import cn.hutool.core.io.FileUtil;
import cn.hutool.core.io.file.FileWriter;
import cn.hutool.core.util.CharsetUtil;
import cn.hutool.json.JSONArray;
import cn.hutool.json.JSONObject;
import org.springframework.stereotype.Component;
import org.springframework.util.DigestUtils;
import org.springframework.util.StringUtils;
import top.hcode.hoj.common.exception.SystemError;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.util.Constants;

import java.io.File;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.List;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/** 将题面样例临时加入标准程序的正式评测，不影响普通用户提交。 */
@Component
public class VerificationSampleCaseService {
    private static final Pattern SAMPLE_PATTERN = Pattern.compile(
            "<input>([\\s\\S]*?)</input>\\s*<output>([\\s\\S]*?)</output>"
                    + "(?:\\s*<explanation>[\\s\\S]*?</explanation>)?");

    public List<String> append(Problem problem, Long submitId, String testCasesDir,
                               JSONObject testCasesInfo) throws SystemError {
        List<Sample> samples = parse(problem.getExamples());
        if (samples.isEmpty()) {
            throw error("题面没有可验证的样例，请先添加样例输入输出");
        }
        JSONArray cases = testCasesInfo.getJSONArray("testCases");
        if (cases == null) {
            throw error("测试数据 info 文件缺少 testCases");
        }

        List<String> files = new ArrayList<>();
        try {
            for (int i = 0; i < samples.size(); i++) {
                appendOne(problem, submitId, testCasesDir, cases, files, samples.get(i), i);
            }
            testCasesInfo.set("testCasesSize", cases.size());
            return files;
        } catch (Exception e) {
            cleanup(testCasesDir, files);
            throw error("创建样例正式评测数据失败：" + e.getMessage());
        }
    }

    public void cleanup(String testCasesDir, List<String> files) {
        if (files == null) return;
        for (String name : files) {
            FileUtil.del(testCasesDir + File.separator + name);
        }
    }

    private void appendOne(Problem problem, Long submitId, String dir, JSONArray cases,
                           List<String> files, Sample sample, int index) {
        String prefix = "__verification_sample_" + submitId + "_" + (index + 1);
        String inputName = prefix + ".in";
        String outputName = prefix + ".out";
        write(dir, inputName, sample.input); files.add(inputName);
        write(dir, outputName, sample.output); files.add(outputName);

        JSONObject item = new JSONObject();
        item.set("inputName", inputName).set("outputName", outputName)
                .set("score", 0).set("groupNum", -100000 - index)
                .set("outputSize", sample.output.getBytes(StandardCharsets.UTF_8).length);
        if (Constants.JudgeMode.DEFAULT.getMode().equals(problem.getJudgeMode())) {
            item.set("outputMd5", md5(sample.output))
                    .set("allStrippedOutputMd5", md5(sample.output.replaceAll("\\s+", "")))
                    .set("EOFStrippedOutputMd5", md5(ProblemTestCaseUtils.rtrim(sample.output)));
        }
        cases.add(item);
    }

    private void write(String dir, String name, String content) {
        new FileWriter(dir + File.separator + name, CharsetUtil.UTF_8).write(content);
    }

    private List<Sample> parse(String examples) throws SystemError {
        List<Sample> result = new ArrayList<>();
        if (StringUtils.isEmpty(examples)) return result;
        Matcher matcher = SAMPLE_PATTERN.matcher(examples);
        int cursor = 0;
        while (matcher.find()) {
            if (!examples.substring(cursor, matcher.start()).trim().isEmpty()) {
                throw error("题面样例格式不完整，请检查每组 input/output");
            }
            result.add(new Sample(normalize(matcher.group(1)), normalize(matcher.group(2))));
            cursor = matcher.end();
        }
        if (result.isEmpty() || !examples.substring(cursor).trim().isEmpty()) {
            throw error("题面样例格式不完整，请检查每组 input/output");
        }
        return result;
    }

    private String normalize(String value) {
        return value.replace("\r\n", "\n").replace('\r', '\n');
    }

    private String md5(String value) {
        return DigestUtils.md5DigestAsHex(value.getBytes(StandardCharsets.UTF_8));
    }

    private SystemError error(String message) {
        return new SystemError(message, null, message);
    }

    private static class Sample {
        private final String input;
        private final String output;
        private Sample(String input, String output) { this.input = input; this.output = output; }
    }
}
