package top.hcode.hoj.manager.admin.problem;

import cn.hutool.json.JSONArray;
import cn.hutool.json.JSONObject;
import cn.hutool.json.JSONUtil;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;
import top.hcode.hoj.pojo.dto.ProblemVerificationSubmitDTO;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.pojo.vo.ProblemAiLanguageProgramVO;

import javax.annotation.Resource;
import java.util.List;
import java.util.Objects;

/** Full judge submissions preserve testcase bytes, SPJ and interactive judging. */
@Component
public class ProblemAiMultiLanguageRunner {
    private static final long WAIT_TIMEOUT_MS = 5 * 60 * 1000L;

    @Resource private ProblemVerificationManager verificationManager;

    public JSONArray run(Problem problem, List<ProblemAiLanguageProgramVO> programs,
                         List<ProblemAiTestPoint> points, String uid, String username) throws Exception {
        JSONArray languages = new JSONArray();
        if (programs == null) return languages;
        for (ProblemAiLanguageProgramVO program : programs) {
            if (Thread.currentThread().isInterrupted()) throw new InterruptedException("AI 多语言判题已中断");
            JSONObject languageResult = new JSONObject()
                    .set("language", program == null ? "未指定" : program.getLanguage());
            try {
                if (program == null || !StringUtils.hasText(program.getLanguage())
                        || !StringUtils.hasText(program.getCode())) {
                    throw new IllegalArgumentException("未生成有效语言和源码");
                }
                ProblemVerificationSubmitDTO submission = new ProblemVerificationSubmitDTO();
                submission.setPid(problem.getId());
                submission.setLanguage(program.getLanguage());
                submission.setCode(program.getCode());
                Long submitId = verificationManager.submitComparisonForAi(submission, uid, username);
                languageResult.set("submitId", submitId);
                verificationManager.waitForAiJudge(problem.getId(), submitId, WAIT_TIMEOUT_MS);
                JSONObject formal = verificationManager.comparisonResult(problem.getId(), submitId);
                int passed = formal.getInt("passedCases", 0);
                int total = formal.getInt("totalCases", 0);
                boolean allPassed = "SUCCESS".equals(formal.getStr("status"))
                        && total > 0 && passed == total;
                JSONArray tests = new JSONArray();
                JSONArray caseResults = formal.getJSONArray("caseResults");
                if (caseResults != null) for (Object value : caseResults) {
                    JSONObject point = JSONUtil.parseObj(value);
                    tests.add(new JSONObject().set("index", point.getInt("index"))
                            .set("caseId", point.getLong("caseId"))
                            .set("status", Objects.equals(point.getInt("judgeStatus"), 0) ? "PASS" : "FAIL")
                            .set("judgeStatus", point.getStr("status", "NOT_RUN"))
                            .set("timeMs", point.get("timeMs"))
                            .set("memoryKb", point.get("memoryKb"))
                            .set("stderr", point.getStr("stderr", "")));
                }
                languageResult.set("status", allPassed ? "PASS" : "FAIL")
                        .set("judgeStatus", formal.getInt("judgeStatus"))
                        .set("passed", passed).set("total", total)
                        .set("message", allPassed ? "全部正式测试点通过" : "存在未通过或未执行测试点")
                        .set("errorMessage", formal.getStr("errorMessage", ""))
                        .set("testPoints", tests);
            } catch (Exception error) {
                if (error instanceof InterruptedException || Thread.currentThread().isInterrupted()) {
                    Thread.currentThread().interrupt();
                    throw error;
                }
                JSONArray tests = new JSONArray();
                for (ProblemAiTestPoint point : points) {
                    tests.add(new JSONObject().set("index", point.getIndex()).set("caseId", point.getCaseId())
                            .set("status", "FAIL").set("judgeStatus", "NOT_RUN")
                            .set("stderr", error.getMessage()));
                }
                languageResult.set("status", "FAIL").set("passed", 0).set("total", points.size())
                        .set("message", "正式判题未完成").set("errorMessage", error.getMessage())
                        .set("testPoints", tests);
            }
            languages.add(languageResult);
        }
        return languages;
    }
}
