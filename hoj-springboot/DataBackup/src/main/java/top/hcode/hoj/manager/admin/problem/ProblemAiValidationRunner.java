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
import javax.annotation.PreDestroy;
import java.util.*;
import java.util.concurrent.*;

@Component
public class ProblemAiValidationRunner {
    private static final int POINTS_PER_BATCH = 12;
    private final ExecutorService aiBatchPool = Executors.newFixedThreadPool(3, runnable -> {
        Thread thread = new Thread(runnable, "problem-ai-test-point");
        thread.setDaemon(true);
        return thread;
    });
    private final ExecutorService multiLanguagePool = Executors.newFixedThreadPool(2, runnable -> {
        Thread thread = new Thread(runnable, "problem-ai-language-judge");
        thread.setDaemon(true);
        return thread;
    });
    @Resource private ProblemAiGateway gateway;
    @Resource private ProblemAiPromptFactory promptFactory;
    @Resource private ProblemAiValidationContext contextService;
    @Resource private ProblemAiMultiLanguageRunner multiLanguageRunner;
    @Resource private ProblemAiJudgeProgramRunner judgeProgramRunner;

    public String run(ProblemAiConfig config, Problem problem, List<ProblemCase> cases,
                      ProblemAiValidateDTO input, String uid, String username) throws Exception {
        ProblemAiValidationContext.Snapshot snapshot = contextService.load(input);
        boolean interactive = "interactive".equals(problem.getJudgeMode());
        boolean special = interactive || "spj".equals(problem.getJudgeMode());
        String validationMode = interactive ? "interactive_protocol" : "integrated_standard_program";
        BlockingQueue<Future<Object>> completedTasks = new LinkedBlockingQueue<>();
        CompletionService<Object> completions = new ExecutorCompletionService<>(aiBatchPool, completedTasks);
        CompletionService<Object> languageCompletions = new ExecutorCompletionService<>(multiLanguagePool, completedTasks);
        Map<Integer, JSONObject> aiResults = new HashMap<>();
        List<ProblemAiTestPoint> points = snapshot.getPoints();
        List<Future<Object>> tasks = new ArrayList<>();
        try {
            Future<Object> overview = completions.submit(() -> complete(config, problem,
                    promptFactory.validationPrompt(config, problem, cases, input, snapshot)));
            tasks.add(overview);
            Future<Object> languages = null;
            if (input.getMultiLanguagePrograms() != null && !input.getMultiLanguagePrograms().isEmpty()) {
                languages = languageCompletions.submit(() -> multiLanguageRunner.run(problem,
                        input.getMultiLanguagePrograms(), points, uid, username));
                tasks.add(languages);
            }
            for (int start = 0; start < points.size(); start += POINTS_PER_BATCH) {
                List<ProblemAiTestPoint> batch = new ArrayList<>(
                        points.subList(start, Math.min(points.size(), start + POINTS_PER_BATCH)));
                tasks.add(completions.submit(() -> checkedBatch(complete(config, problem,
                        promptFactory.testPointPrompt(config, problem, input, batch, points.size())), batch)));
            }
            String overviewContent = null;
            JSONArray multiLanguageResults = null;
            Future<Object> judgeProgram = null;
            JSONObject judgeProgramReview = null;
            int remaining = tasks.size();
            while (remaining-- > 0) {
                Future<Object> completed = completedTasks.take();
                if (completed == overview) {
                    overviewContent = (String) completed.get();
                    if (special) {
                        JSONObject overviewReport = normalized(overviewContent);
                        JSONObject proposed;
                        try { proposed = overviewReport.getJSONObject("judgeProgramReview"); }
                        catch (RuntimeException e) { proposed = null; }
                        final JSONObject candidateReview = proposed;
                        judgeProgram = languageCompletions.submit(() -> judgeProgramRunner.run(problem, candidateReview, uid, username));
                        tasks.add(judgeProgram);
                        remaining++;
                    }
                    continue;
                }
                if (completed == judgeProgram) {
                    judgeProgramReview = (JSONObject) completed.get();
                    continue;
                }
                if (completed == languages) {
                    multiLanguageResults = (JSONArray) completed.get();
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
                boolean formallyAccepted = isFormallyAccepted(point);
                JSONObject formal = new JSONObject().set("mode", validationMode)
                        .set("status", formallyAccepted ? "PASS" : "FAIL")
                        .set("statusText", point.getJudgeStatusText())
                        .set("timeMs", point.getTime()).set("memoryKb", point.getMemory())
                        .set("stderr", emptyAsMarker(point.getStderr()))
                        .set("reason", formallyAccepted
                                ? (interactive ? "交互客户端与交互器完成正式判题；协议与隐藏数据校验另见判题程序审查"
                                    : "标准程序正式判题 Accepted；输入校验覆盖是否完整另见程序分析")
                                : "标准程序未 Accepted 或未执行，请检查输入校验与求解结果");
                item.set("index", point.getIndex()).set("caseId", point.getCaseId())
                        .set("judgeStatus", point.getJudgeStatusText())
                        .set("timeMs", point.getTime()).set("memoryKb", point.getMemory())
                        .set("executed", point.getExecuted())
                        .set("inputAvailable", point.getInputAvailable())
                        .set("expectedOutputAvailable", point.getExpectedOutputAvailable())
                        .set("formalValidation", formal)
                        // Legacy report consumers: this is the same std execution, not another validator.
                        .set("testlibValidator", formal)
                        .set("stderr", emptyAsMarker(point.getStderr()))
                        .set("aiStatus", aiStatus)
                        .set("status", formallyAccepted ? "PASS" : "FAIL")
                        .set("detail", (formallyAccepted
                                ? (interactive ? "交互客户端通过正式判题；协议及交互器正确性另见专项审查。"
                                    : "一体化标准程序通过正式判题；格式与范围校验的完整性另见程序分析。")
                                : "一体化标准程序正式判题状态为 " + point.getJudgeStatusText()
                                        + "（或未执行），该测试点未通过。")
                                + " AI 分析：" + aiDetail);
                merged.add(item);
            }
            result.set("testPointResults", merged);
            result.set("execution", new JSONObject().set("submitId", snapshot.getSubmitId())
                    .set("status", snapshot.getStatusText()).set("testPointCount", points.size())
                    .set("judgeError", snapshot.getJudgeError()));
            result.set("standardProgram", new JSONObject().set("language", input.getLanguage())
                    .set("code", input.getStandardProgram()).set("aiGenerated", input.getAiGenerated())
                    .set("algorithm", input.getAlgorithmSummary())
                    .set("validationMode", validationMode)
                    .set("multiLanguagePrograms", input.getMultiLanguagePrograms()));
            if (languages != null) {
                result.set("multiLanguageResults", multiLanguageResults);
                enforceMultiLanguageFailure(result, multiLanguageResults);
            }
            result.set("testPointValidation", new JSONObject()
                    .set("mode", validationMode)
                    .set("rule", interactive ? "交互客户端逐轮校验回复并遵循交互协议；隐藏数据由交互器读取；展示样例仅作协议推演，不作为隐藏输入"
                            : "同一份 std 执行严格输入格式与范围校验并输出答案；正式判题 Accepted 通过，非 Accepted 或未执行不通过"));
            mergeTestPointConclusion(result, points, merged);
            enforceRuntimeFailure(result, snapshot);
            if (special) {
                if (judgeProgramReview == null) judgeProgramReview = new JSONObject().set("status", "WARN")
                        .set("summary", "判题程序专项检查未完成");
                result.set("judgeProgramReview", judgeProgramReview);
                String status = judgeProgramReview.getStr("status", "WARN");
                if ("FAIL".equals(status)) result.set("overall", "FAIL");
                else if (!"PASS".equals(status) && !"FAIL".equals(result.getStr("overall"))) result.set("overall", "WARN");
                result.getJSONArray("steps").add(new JSONObject().set("name", "判题程序专项审查与正反例验证")
                        .set("status", status).set("detail", judgeProgramReview.getStr("summary", "")));
                if (!"PASS".equals(status)) result.getJSONArray("issues").add(new JSONObject()
                        .set("severity", "FAIL".equals(status) ? "ERROR" : "WARNING")
                        .set("location", "判题程序").set("detail", judgeProgramReview.getStr("summary", ""))
                        .set("suggestion", "检查源码审查与正反例结果，不可仅凭标准解 Accepted 认定判题程序正确"));
            }
            if (hasErrorIssue(result)) result.set("overall", "FAIL");
            return result.toString();
        } catch (Exception e) {
            if (e instanceof InterruptedException) Thread.currentThread().interrupt();
            throw unwrap(e);
        } finally {
            for (Future<Object> future : tasks) cancel(future);
        }
    }

    @PreDestroy
    void shutdown() {
        aiBatchPool.shutdownNow();
        multiLanguagePool.shutdownNow();
    }

    private JSONObject checkedBatch(String content, List<ProblemAiTestPoint> batch) {
        JSONObject result = parseObject(content);
        JSONArray items = result.getJSONArray("testPointResults");
        Set<Integer> missing = new HashSet<>();
        for (ProblemAiTestPoint point : batch) missing.add(point.getIndex());
        if (items != null) for (Object value : items) {
            Integer index = JSONUtil.parseObj(value).getInt("index");
            if (index == null || !missing.remove(index)) {
                throw new IllegalStateException("AI 逐点分析返回重复或不属于本批次的测试点编号：" + index);
            }
        }
        if (items == null || !missing.isEmpty()) {
            throw new IllegalStateException("AI 逐点分析未覆盖全部测试点，缺失编号：" + new TreeSet<>(missing));
        }
        return result;
    }

    private void cancel(Future<?> future) {
        if (future != null && !future.isDone()) future.cancel(true);
    }

    private String complete(ProblemAiConfig config, Problem problem, String prompt) {
        List<Map<String, String>> messages = new ArrayList<>();
        messages.add(message("system", promptFactory.systemPrompt(config.getSystemPrompt(), problem)));
        messages.add(message("user", prompt));
        return gateway.complete(config, messages);
    }

    private void enforceRuntimeFailure(JSONObject result, ProblemAiValidationContext.Snapshot snapshot) {
        boolean failed = !Objects.equals(snapshot.getStatus(), 0) || snapshot.getPoints().isEmpty()
                || snapshot.getPoints().stream().anyMatch(point -> !isFormallyAccepted(point));
        if (!failed) return;
        result.set("overall", "FAIL");
        JSONArray issues = result.getJSONArray("issues");
        if (issues == null) issues = new JSONArray();
        issues.add(new JSONObject().set("severity", "ERROR").set("location", "正式全测试点判题")
                .set("detail", "一体化标准程序未通过全部正式测试点，不能发布题目")
                .set("suggestion", "结合标准程序的编译/运行状态、stderr 和分析问题修复题目后重新验题"));
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
                        .set("detail", item.getStr("language", "未知语言") + " 未通过全部正式测试点（"
                                + item.getInt("passed", 0) + "/" + item.getInt("total", 0) + "）")
                        .set("suggestion", "检查该语言的源码、输入输出格式及运行时兼容性后重新验题"));
                result.set("issues", issues);
            }
        }
    }

    private void mergeTestPointConclusion(JSONObject result, List<ProblemAiTestPoint> points, JSONArray merged) {
        int accepted = 0, stderrCount = 0, fail = 0;
        for (int i = 0; i < points.size(); i++) {
            ProblemAiTestPoint point = points.get(i);
            if (isFormallyAccepted(point)) accepted++;
            if (point.getStderr() != null && !point.getStderr().trim().isEmpty()) stderrCount++;
            String status = merged.getJSONObject(i).getStr("status", "WARN");
            if ("FAIL".equals(status)) fail++;
        }
        if (fail > 0) result.set("overall", "FAIL");
        String evidence = "一体化标准程序正式判题通过 " + accepted + "/" + points.size()
                + " 个测试点，输入校验与求解使用同一份 std，校验覆盖完整性以程序分析为准；逐点 stderr 已采集，其中 "
                + stderrCount + " 个非空。";
        result.set("summary", evidence + " " + result.getStr("summary", ""));
        JSONArray steps = result.getJSONArray("steps");
        if (steps == null) steps = new JSONArray();
        steps.add(new JSONObject().set("name", "全部测试点结果与 stderr")
                .set("status", fail > 0 || points.isEmpty() ? "FAIL" : "PASS")
                .set("detail", evidence + " 正式逐点结论：" + accepted + " 个 PASS，" + fail + " 个 FAIL。"));
        result.set("steps", steps);
    }

    private boolean isFormallyAccepted(ProblemAiTestPoint point) {
        return Boolean.TRUE.equals(point.getExecuted()) && Objects.equals(point.getJudgeStatus(), 0);
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
