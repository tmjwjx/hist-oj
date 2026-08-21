package top.hcode.hoj.manager.admin.problem;

import cn.hutool.core.util.IdUtil;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;
import top.hcode.hoj.judge.self.JudgeDispatcher;
import top.hcode.hoj.pojo.dto.TestJudgeReq;
import top.hcode.hoj.pojo.dto.TestJudgeRes;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.utils.Constants;
import top.hcode.hoj.utils.RedisUtils;

import javax.annotation.Resource;
import java.util.HashMap;
import java.util.List;
import java.util.concurrent.TimeUnit;

/**
 * Compiles and runs the AI-generated testlib validator on every concrete
 * testcase.  The online test-judge path already has testlib.h installed and
 * gives us compile errors, RE, stderr and resource usage without changing the
 * problem's official submission or judge-case records.
 */
@Component
public class ProblemAiTestlibRunner {
    private static final long WAIT_TIMEOUT_MS = 5 * 60 * 1000L;
    private static final long POLL_MS = 300L;

    @Resource private JudgeDispatcher judgeDispatcher;
    @Resource private RedisUtils redisUtils;

    public void run(Problem problem, String validatorLanguage, String validatorCode,
                    List<ProblemAiTestPoint> points) throws Exception {
        if (StringUtils.isEmpty(validatorCode)) {
            for (ProblemAiTestPoint point : points) mark(point, "SKIPPED", "未提供 testlib 校验器源码", null);
            return;
        }
        String language = StringUtils.isEmpty(validatorLanguage) ? "C++" : validatorLanguage;
        for (ProblemAiTestPoint point : points) {
            if (!Boolean.TRUE.equals(point.getConcrete())) {
                mark(point, "SKIPPED", "动态或样例占位测试点没有独立输入文件，跳过 testlib 校验", null);
                continue;
            }
            if (!Boolean.TRUE.equals(point.getInputAvailable()) || StringUtils.isEmpty(point.getInput())) {
                mark(point, "SKIPPED", "测试点输入文件内容不可读取，未启动 testlib 校验器", null);
                continue;
            }
            TestJudgeRes result = runOne(problem, language, validatorCode, point.getInput());
            int status = result == null || result.getStatus() == null
                    ? Constants.Judge.STATUS_SYSTEM_ERROR.getStatus() : result.getStatus();
            String text = statusName(status);
            String stderr = result == null ? "" : result.getStderr();
            mark(point, status == Constants.Judge.STATUS_ACCEPTED.getStatus() ? "PASS" : "FAIL", text, stderr);
            if (result != null) {
                point.setValidatorTime(result.getTime()).setValidatorMemory(result.getMemory());
            }
        }
    }

    private TestJudgeRes runOne(Problem problem, String language, String code, String input) throws Exception {
        String key = "AI_TESTLIB_" + IdUtil.simpleUUID();
        TestJudgeReq request = new TestJudgeReq()
                .setUniqueKey(key)
                .setCode(code)
                .setLanguage(language)
                // JudgeRun.testJudgeCase appends one line ending before execution.
                // File-backed testcase contents normally already end with one, so
                // remove only trailing CR/LF here to avoid an artificial blank line
                // being rejected by validator inf.readEof().
                .setTestCaseInput(stripTrailingLineEndings(input))
                .setExpectedOutput(null)
                .setTimeLimit(Math.max(1000, problem.getTimeLimit()))
                .setMemoryLimit(Math.max(128, problem.getMemoryLimit()))
                .setStackLimit(Math.max(128, problem.getStackLimit()))
                .setIsRemoveEndBlank(false)
                .setProblemJudgeMode("default")
                .setIsFileIO(false)
                .setExtraFile(new HashMap<>());
        redisUtils.set(key, TestJudgeRes.builder().status(Constants.Judge.STATUS_PENDING.getStatus()).build(), 10 * 60);
        try {
            judgeDispatcher.sendTestJudgeTask(request);
            long deadline = System.currentTimeMillis() + WAIT_TIMEOUT_MS;
            while (System.currentTimeMillis() < deadline) {
                Object value = redisUtils.get(key);
                if (value instanceof TestJudgeRes) {
                    TestJudgeRes result = (TestJudgeRes) value;
                    if (!Constants.Judge.STATUS_PENDING.getStatus().equals(result.getStatus())) return result;
                }
                TimeUnit.MILLISECONDS.sleep(POLL_MS);
            }
            return TestJudgeRes.builder().status(Constants.Judge.STATUS_SYSTEM_ERROR.getStatus())
                    .stderr("testlib 校验器评测超时").build();
        } finally {
            redisUtils.del(key);
        }
    }

    private void mark(ProblemAiTestPoint point, String status, String statusText, String stderr) {
        point.setValidatorStatus(status).setValidatorStatusText(statusText)
                .setValidatorStderr(stderr == null ? "" : stderr);
    }

    private String statusName(Integer status) {
        for (Constants.Judge item : Constants.Judge.values()) {
            if (item.getStatus().equals(status)) return item.getName();
        }
        return "UNKNOWN(" + status + ")";
    }

    private String stripTrailingLineEndings(String input) {
        if (input == null) return "";
        int end = input.length();
        while (end > 0) {
            char ch = input.charAt(end - 1);
            if (ch != '\n' && ch != '\r') break;
            end--;
        }
        return input.substring(0, end);
    }
}
