package top.hcode.hoj.manager.admin.problem;

import cn.hutool.json.JSONArray;
import cn.hutool.json.JSONObject;
import cn.hutool.json.JSONUtil;
import org.springframework.stereotype.Component;
import top.hcode.hoj.pojo.dto.ProblemAiValidateDTO;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.pojo.entity.problem.ProblemAiConfig;
import top.hcode.hoj.pojo.entity.problem.ProblemCase;

import javax.annotation.Resource;
import java.util.*;
import java.util.concurrent.*;

@Component
public class ProblemAiValidationRunner {
    private static final int POINTS_PER_BATCH = 12;
    private static final ExecutorService AI_BATCH_POOL = Executors.newFixedThreadPool(3, runnable -> {
        Thread thread = new Thread(runnable, "problem-ai-test-point");
        thread.setDaemon(true);
        return thread;
    });
    @Resource private ProblemAiGateway gateway;
    @Resource private ProblemAiPromptFactory promptFactory;
    @Resource private ProblemAiValidationContext contextService;
    @Resource private ProblemAiTestlibRunner testlibRunner;
    @Resource private ProblemAiMultiLanguageRunner multiLanguageRunner;

    public String run(ProblemAiConfig config, Problem problem, List<ProblemCase> cases,
                      ProblemAiValidateDTO input) throws Exception {
        ProblemAiValidationContext.Snapshot snapshot = contextService.load(input);
        testlibRunner.run(problem, input.getValidatorLanguage(), input.getValidatorCode(), snapshot.getPoints());
        CompletionService<Object> completions = new ExecutorCompletionService<>(AI_BATCH_POOL);
        Future<Object> overview = completions.submit((Callable<Object>) () -> complete(config,
                promptFactory.validationPrompt(config, problem, cases, input, snapshot)));
        Map<Integer, JSONObject> aiResults = new HashMap<>();
        List<ProblemAiTestPoint> points = snapshot.getPoints();
        List<Future<Object>> batches = new ArrayList<>();
        for (int start = 0; start < points.size(); start += POINTS_PER_BATCH) {
            List<ProblemAiTestPoint> batch = new ArrayList<>(
                    points.subList(start, Math.min(points.size(), start + POINTS_PER_BATCH)));
            batches.add(completions.submit((Callable<Object>) () -> parseObject(complete(config,
                    promptFactory.testPointPrompt(config, problem, input, batch, points.size())))));
        }
        try {
            String overviewContent = null;
            int remaining = batches.size() + 1;
            while (remaining-- > 0) {
                Future<Object> completed = completions.take();
                if (completed == overview) {
                    overviewContent = (String) completed.get();
                    continue;
                }
                JSONObject batchResult = (JSONObject) completed.get();
                JSONArray items = batchResult.getJSONArray("testPointResults");
                if (items == null) continue;
                for (Object value : items) {
                    JSONObject item = JSONUtil.parseObj(value);
                    Integer index = item.getInt("index");
                    if (index != null) aiResults.put(index, item);
                }
            }
            JSONObject result = normalized(overviewContent);

            JSONArray merged = new JSONArray();
            for (ProblemAiTestPoint point : points) {
                JSONObject item = aiResults.getOrDefault(point.getIndex(), new JSONObject());
                String aiStatus = item.getStr("status", "WARN");
                String aiDetail = item.getStr("detail", "AI 未返回该点的完整分析");
                item.set("index", point.getIndex())
                        .set("judgeStatus", point.getJudgeStatusText())
                        .set("timeMs", point.getTime())
                        .set("memoryKb", point.getMemory())
                        .set("executed", point.getExecuted())
                        .set("inputAvailable", point.getInputAvailable())
                        .set("expectedOutputAvailable", point.getExpectedOutputAvailable())
                        .set("testlibValidator", new JSONObject()
                                .set("status", point.getValidatorStatus())
                                .set("statusText", point.getValidatorStatusText())
                                .set("timeMs", point.getValidatorTime())
                                .set("memoryKb", point.getValidatorMemory())
                                .set("stderr", emptyAsMarker(point.getValidatorStderr())))
                        .set("stderr", emptyAsMarker(point.getStderr()));
                boolean concrete = Boolean.TRUE.equals(point.getConcrete());
                boolean formallyAccepted = isFormallyAccepted(point);
                item.set("aiStatus", aiStatus)
                        .set("formalValidation", new JSONObject()
                                .set("mode", "testlib_registerValidation_strict")
                                .set("status", formallyAccepted ? "PASS" : "FAIL")
                                .set("reason", formallyAccepted
                                        ? "正式判题 Accepted 且未发生 Runtime Error（RE）"
                                        : "正式判题未 Accepted、未执行或发生 Runtime Error（RE）"));
                if (!concrete && Boolean.TRUE.equals(point.getExecuted()) && Objects.equals(point.getJudgeStatus(), 0)) {
                    item.set("status", "PASS")
                            .set("detail", "该点是动态或样例占位测试点，没有独立输入文件，已跳过 testlib 校验；正式判题状态为 Accepted。原始分析：" + aiDetail);
                } else if (formallyAccepted) {
                    // A missing/truncated testcase payload is not a defect when
                    // the concrete testcase has already passed the formal judge.
                    // The formal result also takes precedence over an AI FAIL
                    // caused by the same unavailable evidence.
                    item.set("status", "PASS")
                            .set("detail", "正式判题 Accepted 且无 RE；按 testlib/registerValidation 严格规则该测试点通过。"
                                    + " AI 未展开文件内容不构成 WARN。原始提示：" + aiDetail);
                } else if (!formallyAccepted) {
                    // Never hide a concrete formal failure behind an AI WARN.
                    item.set("status", "FAIL")
                            .set("detail", "正式判题状态为 " + point.getJudgeStatusText()
                                    + "，testlib 校验器状态为 " + point.getValidatorStatusText()
                                    + "（或该点未执行），按严格测试点规则判定失败。原始分析：" + aiDetail);
                } else {
                    item.set("status", aiStatus).set("detail", aiDetail);
                }
                merged.add(item);
            }
            result.set("testPointResults", merged);
            result.set("execution", new JSONObject().set("submitId", snapshot.getSubmitId())
                    .set("status", snapshot.getStatusText()).set("testPointCount", points.size())
                    .set("judgeError", snapshot.getJudgeError()));
            result.set("standardProgram", new JSONObject().set("language", input.getLanguage())
                    .set("code", input.getStandardProgram()).set("aiGenerated", input.getAiGenerated())
                    .set("algorithm", input.getAlgorithmSummary())
                    .set("validatorLanguage", input.getValidatorLanguage())
                    .set("validatorCode", input.getValidatorCode())
                    .set("multiLanguagePrograms", input.getMultiLanguagePrograms()));
            if (input.getMultiLanguagePrograms() != null && !input.getMultiLanguagePrograms().isEmpty()) {
                JSONArray multiLanguageResults = multiLanguageRunner.run(problem,
                        input.getMultiLanguagePrograms(), points);
                result.set("multiLanguageResults", multiLanguageResults);
                enforceMultiLanguageFailure(result, multiLanguageResults);
            }
            result.set("testPointValidation", new JSONObject()
                    .set("mode", "testlib_registerValidation_strict")
                    .set("rule", "Accepted 且无 RE 通过；RE、非 Accepted 或未执行不通过；不因拿不到文件内容而 WARN"));
            mergeTestPointConclusion(result, points, merged);
            enforceRuntimeFailure(result, snapshot);
            return result.toString();
        } catch (Exception e) {
            cancel(overview);
            for (Future<Object> future : batches) cancel(future);
            throw unwrap(e);
        }
    }

    private void cancel(Future<?> future) {
        if (future != null && !future.isDone()) future.cancel(true);
    }

    private String complete(ProblemAiConfig config, String prompt) {
        List<Map<String, String>> messages = new ArrayList<>();
        messages.add(message("system", promptFactory.systemPrompt(config.getSystemPrompt())));
        messages.add(message("user", prompt));
        return gateway.complete(config, messages);
    }

    private void enforceRuntimeFailure(JSONObject result, ProblemAiValidationContext.Snapshot snapshot) {
        boolean failed = snapshot.getPoints().stream().filter(point -> Boolean.TRUE.equals(point.getConcrete())).anyMatch(point -> Boolean.FALSE.equals(point.getExecuted())
                || (point.getJudgeStatus() != null && point.getJudgeStatus() != 0)
                || !"PASS".equals(point.getValidatorStatus()));
        if (!failed) return;
        result.set("overall", "FAIL");
        JSONArray issues = result.getJSONArray("issues");
        if (issues == null) issues = new JSONArray();
        issues.add(new JSONObject().set("severity", "ERROR").set("location", "正式全测试点判题")
                .set("detail", "至少一个测试点未通过正式判题或 testlib 输入校验，不能发布题目")
                .set("suggestion", "结合正式判题、testlib 校验器编译/运行状态和 stderr 修复题目后重新验题"));
        result.set("issues", issues);
    }

    private void enforceMultiLanguageFailure(JSONObject result, JSONArray results) {
        if (results == null) return;
        for (Object value : results) {
            JSONObject item = JSONUtil.parseObj(value);
            if (!"PASS".equals(item.getStr("status"))) {
                result.set("overall", "FAIL");
                JSONArray issues = result.getJSONArray("issues");
                if (issues == null) issues = new JSONArray();
                issues.add(new JSONObject().set("severity", "ERROR")
                        .set("location", "多语言全测试点判题")
                        .set("detail", item.getStr("language", "未知语言") + " 未通过全部具体测试点（"
                                + item.getInt("passed", 0) + "/" + item.getInt("total", 0) + "）")
                        .set("suggestion", "检查该语言的源码、输入输出格式及运行时兼容性后重新验题"));
            }
        }
    }

    private void mergeTestPointConclusion(JSONObject result, List<ProblemAiTestPoint> points, JSONArray merged) {
        int accepted = 0, formalAccepted = 0, validatorPassed = 0, stderrCount = 0, warn = 0, fail = 0;
        int concreteCount = 0;
        for (int i = 0; i < points.size(); i++) {
            ProblemAiTestPoint point = points.get(i);
            if (Boolean.TRUE.equals(point.getConcrete())) concreteCount++;
            if (Boolean.TRUE.equals(point.getExecuted()) && Objects.equals(point.getJudgeStatus(), 0)) formalAccepted++;
            if (Boolean.TRUE.equals(point.getConcrete()) && "PASS".equals(point.getValidatorStatus())) validatorPassed++;
            if (isFormallyAccepted(point)) accepted++;
            if (point.getStderr() != null && !point.getStderr().trim().isEmpty()) stderrCount++;
            String aiStatus = merged.getJSONObject(i).getStr("status", "WARN");
            if ("FAIL".equals(aiStatus)) fail++;
            else if ("WARN".equals(aiStatus)) warn++;
        }
        if (fail > 0) result.set("overall", "FAIL");
        else if (warn > 0 && "PASS".equals(result.getStr("overall"))) result.set("overall", "WARN");
        else if (warn == 0 && concreteCount > 0 && accepted == concreteCount
                && !hasErrorIssue(result) && !"FAIL".equals(result.getStr("overall"))) {
            result.set("overall", "PASS");
        }
        String evidence = "正式判题已核验 " + formalAccepted + "/" + points.size()
                + " 个测试点 Accepted；独立正式测试点的 testlib 输入校验器通过 " + validatorPassed + "/" + concreteCount
                + "；逐点 stderr 已全部采集，其中 " + stderrCount + " 个非空。";
        result.set("summary", evidence + " " + result.getStr("summary", ""));
        JSONArray steps = result.getJSONArray("steps");
        if (steps == null) steps = new JSONArray();
        steps.add(new JSONObject().set("name", "全部测试点结果与 stderr")
                .set("status", fail > 0 ? "FAIL" : warn > 0 ? "WARN" : "PASS")
                .set("detail", evidence + " AI 逐点结论：" + fail + " 个 FAIL，" + warn + " 个 WARN。"));
        result.set("steps", steps);
    }

    private boolean isFormallyAccepted(ProblemAiTestPoint point) {
        if (!Boolean.TRUE.equals(point.getConcrete())) {
            return Boolean.TRUE.equals(point.getExecuted()) && Objects.equals(point.getJudgeStatus(), 0);
        }
        return Boolean.TRUE.equals(point.getExecuted()) && Objects.equals(point.getJudgeStatus(), 0)
                && "PASS".equals(point.getValidatorStatus());
    }

    private boolean hasErrorIssue(JSONObject result) {
        JSONArray issues = result.getJSONArray("issues");
        if (issues == null) return false;
        for (Object value : issues) {
            JSONObject issue = JSONUtil.parseObj(value);
            if ("ERROR".equalsIgnoreCase(issue.getStr("severity"))) return true;
        }
        return false;
    }

    private JSONObject normalized(String content) {
        JSONObject result = parseObject(content);
        if (!result.containsKey("overall")) result.set("overall", "WARN");
        if (!result.containsKey("summary")) result.set("summary", "AI 已完成检查，但未提供摘要");
        if (!result.containsKey("steps")) result.set("steps", new JSONArray());
        if (!result.containsKey("issues")) result.set("issues", new JSONArray());
        if (!result.containsKey("sampleResults")) result.set("sampleResults", new JSONArray());
        return result;
    }

    private JSONObject parseObject(String content) {
        String value = content == null ? "" : content.trim();
        if (value.startsWith("```")) value = value.replaceFirst("^```(?:json)?\\s*", "").replaceFirst("\\s*```$", "");
        int start = value.indexOf('{'), end = value.lastIndexOf('}');
        if (start >= 0 && end > start) value = value.substring(start, end + 1);
        try { return JSONUtil.parseObj(value); }
        catch (Exception ignored) { return new JSONObject().set("overall", "WARN").set("summary", value); }
    }

    private Map<String, String> message(String role, String content) {
        Map<String, String> value = new HashMap<>();
        value.put("role", role); value.put("content", content); return value;
    }

    private String emptyAsMarker(String value) {
        return value == null || value.trim().isEmpty() ? "[stderr 为空]" : value;
    }

    private Exception unwrap(Exception error) {
        Throwable cause = error;
        while ((cause instanceof ExecutionException || cause instanceof CompletionException)
                && cause.getCause() != null) {
            cause = cause.getCause();
        }
        return cause instanceof Exception ? (Exception) cause : error;
    }
}
