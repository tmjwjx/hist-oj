package top.hcode.hoj.manager.admin.problem;

import cn.hutool.json.JSONArray;
import cn.hutool.json.JSONObject;
import cn.hutool.json.JSONUtil;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;
import top.hcode.hoj.dao.problem.ProblemEntityService;
import top.hcode.hoj.pojo.dto.ProblemVerificationSubmitDTO;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.service.problem.ProblemVerificationChangeDetector;

import javax.annotation.Resource;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.*;

/** Checks the checker itself using bounded, isolated full-judge submissions. */
@Component
public class ProblemAiJudgeProgramRunner {
    @Resource private ProblemVerificationManager verificationManager;
    @Resource private ProblemEntityService problemService;

    public JSONObject run(Problem problem, JSONObject proposed, String uid, String username) throws Exception {
        boolean interactive = "interactive".equals(problem.getJudgeMode());
        JSONObject result = new JSONObject().set("mode", problem.getJudgeMode()).set("language", problem.getSpjLanguage())
                .set("sourceSha256", digest(problem.getSpjCode())).set("status", "PASS")
                .set("summary", "源码专项审查与有限正反例验证；不能替代判题程序正确性的完整证明")
                .set("checks", new JSONArray()).set("probes", new JSONArray());
        if (!StringUtils.hasText(problem.getSpjCode()) || !StringUtils.hasText(problem.getSpjLanguage())) {
            return result.set("status", "FAIL").set("summary", "缺少判题程序源码或语言，无法完成验证");
        }
        if (proposed == null) return result.set("status", "WARN").set("summary", "AI 未返回判题程序专项审查，不能认定判题程序验证通过");
        String declared = proposed.getStr("status", "WARN");
        if (!Arrays.asList("PASS", "WARN", "FAIL").contains(declared)) declared = "WARN";
        result.set("status", declared).set("summary", limited(proposed.getStr("summary", result.getStr("summary")), 8000));
        Set<String> required = new LinkedHashSet<>(interactive
                ? Arrays.asList("protocol", "limits", "termination", "verdict")
                : Arrays.asList("acceptance", "rejection", "numeric", "termination"));
        JSONArray checks = array(proposed, "checks");
        JSONArray verifiedChecks = result.getJSONArray("checks");
        if (checks != null && checks.size() <= 20) for (Object value : checks) {
            JSONObject check = object(value);
            String id = check.getStr("id", "");
            String status = check.getStr("status", "WARN");
            String detail = limited(check.getStr("detail", ""), 8000);
            if (!required.remove(id) || !Arrays.asList("PASS", "WARN", "FAIL").contains(status) || detail.trim().isEmpty()) {
                worsen(result, "WARN");
                continue;
            }
            verifiedChecks.add(new JSONObject().set("id", id).set("status", status).set("detail", detail));
            worsen(result, status);
        }
        for (String id : required) {
            verifiedChecks.add(new JSONObject().set("id", id).set("status", "WARN").set("detail", "未返回该项必要检查"));
            worsen(result, "WARN");
        }
        Map<String, String> expected = new LinkedHashMap<>();
        if (interactive) {
            expected.put("protocol_violation", "REJECT"); expected.put("premature_exit", "REJECT");
        } else {
            expected.put("alternate_valid", "ACCEPT"); expected.put("invalid_output", "REJECT");
        }
        JSONArray probes = array(proposed, "probes");
        JSONArray outcomes = result.getJSONArray("probes");
        if (probes != null && probes.size() <= 2) for (Object value : probes) {
            JSONObject probe = object(value);
            String category = probe.getStr("category", "");
            String expectation = expected.get(category);
            String code = probe.getStr("code", "");
            String reason = limited(probe.getStr("reason", ""), 8000);
            if (expectation == null || !expectation.equals(probe.getStr("expectation"))
                    || code.trim().isEmpty() || code.length() > 65535 || reason.trim().isEmpty()) {
                worsen(result, "WARN");
                continue;
            }
            expected.remove(category);
            JSONObject outcome = new JSONObject().set("id", category).set("category", category)
                    .set("expectation", expectation).set("code", code).set("reason", reason);
            outcomes.add(outcome);
            try {
                verifyUnchanged(problem);
                ProblemVerificationSubmitDTO dto = new ProblemVerificationSubmitDTO();
                dto.setPid(problem.getId()); dto.setLanguage("C++ 17"); dto.setCode(code);
                Long submitId = verificationManager.submitComparisonForAi(dto, uid, username);
                outcome.set("submitId", submitId);
                verificationManager.waitForAiJudge(problem.getId(), submitId, 5 * 60 * 1000L);
                verifyUnchanged(problem);
                JSONObject formal = verificationManager.comparisonResult(problem.getId(), submitId);
                outcome.putAll(formal);
                String verdict = evaluate(expectation, formal);
                outcome.set("status", verdict);
                if ("FAIL".equals(verdict)) outcome.set("errorMessage", "ACCEPT".equals(expectation)
                        ? "预期合法的候选程序未通过；请复核候选程序和判题程序"
                        : "预期非法的候选程序获得通过或部分分，判题程序存在漏判风险，请复核反例依据");
                else if ("INCONCLUSIVE".equals(verdict)) outcome.set("errorMessage", "未取得明确的判题程序接受/拒绝证据；编译错误、超时或崩溃不视为正确拒绝");
                worsen(result, "INCONCLUSIVE".equals(verdict) ? "WARN" : verdict);
            } catch (Exception e) {
                if (e instanceof InterruptedException) Thread.currentThread().interrupt();
                if (Thread.currentThread().isInterrupted()) throw e;
                outcome.set("status", "INCONCLUSIVE").set("errorMessage", "判题程序反例验证未完成，请检查判题服务或题目版本后重试");
                worsen(result, "WARN");
            }
        }
        for (Map.Entry<String, String> missing : expected.entrySet()) {
            outcomes.add(new JSONObject().set("id", missing.getKey()).set("category", missing.getKey())
                    .set("expectation", missing.getValue()).set("status", "INCONCLUSIVE")
                    .set("errorMessage", "未提供可执行且有题意依据的验证程序"));
            worsen(result, "WARN");
        }
        if (problem.getSpjCode().length() + (problem.getJudgeExtraFile() == null ? 0 : problem.getJudgeExtraFile().length()) > 60000) {
            worsen(result, "WARN");
            result.set("summary", result.getStr("summary", "") + "；判题源码及附件较大，不能仅依据本轮 AI 回复认定已完整审查");
        }
        return result;
    }

    private String evaluate(String expectation, JSONObject result) {
        JSONArray cases = array(result, "caseResults");
        int total = result.getInt("totalCases", 0);
        Integer overall = result.getInt("judgeStatus");
        if (cases == null || total <= 0) return "INCONCLUSIVE";
        boolean accepted = Objects.equals(overall, 0), rejected = Objects.equals(overall, -1) || Objects.equals(overall, -3);
        for (Object value : cases) {
            Integer status = object(value).getInt("judgeStatus");
            // A concrete counterexample is still meaningful when another testcase did not run.
            if ("REJECT".equals(expectation) && (Objects.equals(status, 0) || Objects.equals(status, 8))) return "FAIL";
            if ("ACCEPT".equals(expectation) && (Objects.equals(status, -1) || Objects.equals(status, -3) || Objects.equals(status, 8))) return "FAIL";
            accepted &= Objects.equals(status, 0);
            rejected &= Objects.equals(status, -1) || Objects.equals(status, -3);
        }
        if (cases.size() != total || result.getInt("pendingCases", 0) > 0) return "INCONCLUSIVE";
        if ("ACCEPT".equals(expectation)) return accepted ? "PASS" : "INCONCLUSIVE";
        return rejected ? "PASS" : "INCONCLUSIVE";
    }

    private void verifyUnchanged(Problem snapshot) {
        if (Thread.currentThread().isInterrupted()) throw new IllegalStateException("验题中断");
        Problem current = problemService.getById(snapshot.getId());
        if (current == null || ProblemVerificationChangeDetector.changed(snapshot, current)
                || !Objects.equals(snapshot.getDescription(), current.getDescription())
                || !Objects.equals(snapshot.getInput(), current.getInput()) || !Objects.equals(snapshot.getOutput(), current.getOutput())) {
            throw new IllegalStateException("题目或判题程序已变化");
        }
    }

    private void worsen(JSONObject result, String status) {
        if ("FAIL".equals(status) || ("WARN".equals(status) && !"FAIL".equals(result.getStr("status")))) result.set("status", status);
    }

    private JSONArray array(JSONObject object, String key) {
        try { return object.getJSONArray(key); } catch (RuntimeException e) { return null; }
    }

    private JSONObject object(Object value) {
        try { return value instanceof Map ? JSONUtil.parseObj(value) : new JSONObject(); }
        catch (RuntimeException e) { return new JSONObject(); }
    }

    private String digest(String source) throws Exception {
        byte[] bytes = MessageDigest.getInstance("SHA-256").digest((source == null ? "" : source).getBytes(StandardCharsets.UTF_8));
        StringBuilder hex = new StringBuilder();
        for (byte value : bytes) hex.append(String.format("%02x", value));
        return hex.toString();
    }

    private String limited(String text, int limit) { return text.length() > limit ? text.substring(0, limit) : text; }
}
