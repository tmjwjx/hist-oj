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

    public String run(ProblemAiConfig config, Problem problem, List<ProblemCase> cases,
                      ProblemAiValidateDTO input) throws Exception {
        ProblemAiValidationContext.Snapshot snapshot = contextService.load(input);
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
                item.set("index", point.getIndex())
                        .set("judgeStatus", point.getJudgeStatusText())
                        .set("timeMs", point.getTime())
                        .set("memoryKb", point.getMemory())
                        .set("executed", point.getExecuted())
                        .set("stderr", emptyAsMarker(point.getStderr()));
                if (!item.containsKey("status")) item.set("status", "WARN");
                if (!item.containsKey("detail")) item.set("detail", "AI 未返回该点的完整分析，已保留正式判题数据供人工复核");
                merged.add(item);
            }
            result.set("testPointResults", merged);
            result.set("execution", new JSONObject().set("submitId", snapshot.getSubmitId())
                    .set("status", snapshot.getStatusText()).set("testPointCount", points.size())
                    .set("judgeError", snapshot.getJudgeError()));
            result.set("standardProgram", new JSONObject().set("language", input.getLanguage())
                    .set("code", input.getStandardProgram()).set("aiGenerated", input.getAiGenerated())
                    .set("algorithm", input.getAlgorithmSummary()));
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
        boolean failed = snapshot.getPoints().stream().anyMatch(point -> Boolean.FALSE.equals(point.getExecuted())
                || (point.getJudgeStatus() != null && point.getJudgeStatus() != 0));
        if (!failed) return;
        result.set("overall", "FAIL");
        JSONArray issues = result.getJSONArray("issues");
        if (issues == null) issues = new JSONArray();
        issues.add(new JSONObject().set("severity", "ERROR").set("location", "正式全测试点判题")
                .set("detail", "至少一个测试点未执行或未通过，不能发布题目")
                .set("suggestion", "结合逐测试点状态和 stderr 修复标准程序、数据或判题程序后重新验题"));
        result.set("issues", issues);
    }

    private void mergeTestPointConclusion(JSONObject result, List<ProblemAiTestPoint> points, JSONArray merged) {
        int accepted = 0, stderrCount = 0, warn = 0, fail = 0;
        for (int i = 0; i < points.size(); i++) {
            ProblemAiTestPoint point = points.get(i);
            if (Boolean.TRUE.equals(point.getExecuted()) && Objects.equals(point.getJudgeStatus(), 0)) accepted++;
            if (point.getStderr() != null && !point.getStderr().trim().isEmpty()) stderrCount++;
            String aiStatus = merged.getJSONObject(i).getStr("status", "WARN");
            if ("FAIL".equals(aiStatus)) fail++;
            else if ("WARN".equals(aiStatus)) warn++;
        }
        if (fail > 0) result.set("overall", "FAIL");
        else if (warn > 0 && "PASS".equals(result.getStr("overall"))) result.set("overall", "WARN");
        String evidence = "正式判题已核验 " + accepted + "/" + points.size()
                + " 个测试点 Accepted；逐点 stderr 已全部采集，其中 " + stderrCount + " 个非空。";
        result.set("summary", evidence + " " + result.getStr("summary", ""));
        JSONArray steps = result.getJSONArray("steps");
        if (steps == null) steps = new JSONArray();
        steps.add(new JSONObject().set("name", "全部测试点结果与 stderr")
                .set("status", fail > 0 ? "FAIL" : warn > 0 ? "WARN" : "PASS")
                .set("detail", evidence + " AI 逐点结论：" + fail + " 个 FAIL，" + warn + " 个 WARN。"));
        result.set("steps", steps);
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
