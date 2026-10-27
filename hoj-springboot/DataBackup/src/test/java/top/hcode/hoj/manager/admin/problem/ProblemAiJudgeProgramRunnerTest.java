package top.hcode.hoj.manager.admin.problem;

import cn.hutool.json.JSONArray;
import cn.hutool.json.JSONObject;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.test.util.ReflectionTestUtils;
import top.hcode.hoj.dao.problem.ProblemEntityService;
import top.hcode.hoj.pojo.entity.problem.Problem;

import java.util.Arrays;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

class ProblemAiJudgeProgramRunnerTest {
    private final ProblemAiJudgeProgramRunner runner = new ProblemAiJudgeProgramRunner();
    private final ProblemVerificationManager verification = mock(ProblemVerificationManager.class);
    private final ProblemEntityService problems = mock(ProblemEntityService.class);
    private final Problem problem = new Problem().setId(7L).setJudgeMode("spj").setSpjLanguage("C++")
            .setSpjCode("int main(){return 0;}").setCaseVersion("v1");

    @BeforeEach
    void setup() throws Exception {
        ReflectionTestUtils.setField(runner, "verificationManager", verification);
        ReflectionTestUtils.setField(runner, "problemService", problems);
        when(problems.getById(7L)).thenReturn(problem);
        when(verification.submitComparisonForAi(any(), eq("admin"), eq("admin"))).thenReturn(100L, 101L);
        when(verification.comparisonResult(7L, 100L)).thenReturn(formal(0, 0, 0));
        when(verification.comparisonResult(7L, 101L)).thenReturn(formal(-1, -1, -3));
    }

    @Test
    void spjRequiresCodeReviewAndPositiveNegativeFullJudgeEvidence() throws Exception {
        JSONObject result = run(proposal(false));
        assertEquals("PASS", result.getStr("status"));
        assertEquals(64, result.getStr("sourceSha256").length());
        assertEquals(4, result.getJSONArray("checks").size());
        assertEquals(2, result.getJSONArray("probes").size());
        assertEquals("PASS", result.getJSONArray("probes").getJSONObject(1).getStr("status"));
        verify(verification, times(2)).submitComparisonForAi(any(), anyString(), anyString());
    }

    @Test
    void permissiveCheckerAcceptingInvalidOutputFailsEvenIfAiSaysPass() throws Exception {
        when(verification.comparisonResult(7L, 101L)).thenReturn(formal(0, -1, 0));
        assertEquals("FAIL", run(proposal(false)).getStr("status"));
        when(verification.comparisonResult(7L, 101L)).thenReturn(formal(8, -1, 8));
        assertEquals("FAIL", run(proposal(false)).getStr("status"));
    }

    @Test
    void compileCrashTimeoutSystemFailureOrMissingCasesNeverProveRejection() throws Exception {
        for (Integer failure : Arrays.asList(-2, 1, 2, 3, 4, 5, 10, 15)) {
            when(verification.comparisonResult(eq(7L), anyLong())).thenReturn(formal(failure, failure, failure));
            JSONObject result = run(proposal(false));
            assertEquals("WARN", result.getStr("status"), "failure=" + failure);
            assertEquals("INCONCLUSIVE", result.getJSONArray("probes").getJSONObject(1).getStr("status"));
        }
        when(verification.comparisonResult(eq(7L), anyLong())).thenReturn(formal(5, 5, 5).set("pendingCases", 1));
        assertEquals("WARN", run(proposal(false)).getStr("status"));
    }

    @Test
    void absentOrIncompleteReviewCannotPassAndCannotRunUnboundedProbes() throws Exception {
        assertEquals("WARN", run(null).getStr("status"));
        JSONObject review = proposal(false);
        review.getJSONArray("checks").remove(0);
        assertEquals("WARN", run(review).getStr("status"));
        clearInvocations(verification);
        review = proposal(false);
        review.getJSONArray("probes").add(new JSONObject().set("code", "extra"));
        assertEquals("WARN", run(review).getStr("status"));
        verifyNoInteractions(verification);
    }

    @Test
    void missingCheckerOrChangedSourceCannotBeValidated() throws Exception {
        problem.setSpjCode("");
        assertEquals("FAIL", run(proposal(false)).getStr("status"));
        problem.setSpjCode("original");
        when(problems.getById(7L)).thenReturn(new Problem().setId(7L).setCaseVersion("v1").setSpjCode("changed"));
        assertEquals("WARN", run(proposal(false)).getStr("status"));
        verifyNoInteractions(verification);
    }

    @Test
    void interactiveChecksProtocolViolationsAndEarlyExitWithRealInteractor() throws Exception {
        problem.setJudgeMode("interactive");
        when(verification.comparisonResult(eq(7L), anyLong())).thenReturn(formal(-1, -1, -3));
        JSONObject result = run(proposal(true));
        assertEquals("PASS", result.getStr("status"));
        assertEquals("interactive", result.getStr("mode"));
        for (Object value : result.getJSONArray("probes")) {
            JSONObject probe = (JSONObject) value;
            assertEquals("REJECT", probe.getStr("expectation"));
            assertEquals(2, probe.getJSONArray("caseResults").size());
        }
    }

    @Test
    void mixedAcceptRejectOrUnfinishedCasesCannotHideConcreteCounterexample() throws Exception {
        when(verification.comparisonResult(7L, 100L)).thenReturn(formal(-1, 0, -1));
        assertEquals("FAIL", run(proposal(false)).getStr("status"));
        when(verification.comparisonResult(eq(7L), anyLong())).thenReturn(formal(0, 0, 5).set("pendingCases", 1));
        assertEquals("FAIL", run(proposal(false)).getStr("status"));
    }

    @Test
    void largeJudgeProgramCannotClaimCompleteReviewFromOneModelReply() throws Exception {
        problem.setSpjCode(String.join("", java.util.Collections.nCopies(60001, "a")));
        assertEquals("WARN", run(proposal(false)).getStr("status"));
    }

    @Test
    void sourceReviewFailureSurvivesPassingRuntimeProbes() throws Exception {
        JSONObject review = proposal(false);
        review.getJSONArray("checks").getJSONObject(0).set("status", "FAIL");
        assertEquals("FAIL", run(review).getStr("status"));
    }

    private JSONObject run(JSONObject proposed) throws Exception { return runner.run(problem, proposed, "admin", "admin"); }

    private JSONObject proposal(boolean interactive) {
        JSONArray checks = new JSONArray();
        for (String id : interactive ? new String[]{"protocol", "limits", "termination", "verdict"}
                : new String[]{"acceptance", "rejection", "numeric", "termination"}) {
            checks.add(new JSONObject().set("id", id).set("status", "PASS").set("detail", "基于源码的检查依据"));
        }
        JSONArray probes = new JSONArray();
        String[] categories = interactive ? new String[]{"protocol_violation", "premature_exit"}
                : new String[]{"alternate_valid", "invalid_output"};
        for (int i = 0; i < 2; i++) probes.add(new JSONObject().set("id", categories[i]).set("category", categories[i])
                .set("expectation", !interactive && i == 0 ? "ACCEPT" : "REJECT")
                .set("code", "int main(){return 0;}").set("reason", "依据题意构造可验证行为"));
        return new JSONObject().set("status", "PASS").set("summary", "检查通过").set("checks", checks).set("probes", probes);
    }

    private JSONObject formal(int overall, int first, int second) {
        return new JSONObject().set("judgeStatus", overall).set("totalCases", 2).set("pendingCases", 0)
                .set("caseResults", new JSONArray().put(new JSONObject().set("index", 1).set("judgeStatus", first))
                        .put(new JSONObject().set("index", 2).set("judgeStatus", second)));
    }
}
