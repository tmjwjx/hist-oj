package top.hcode.hoj.manager.admin.problem;

import cn.hutool.json.JSONArray;
import cn.hutool.json.JSONObject;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.ArgumentCaptor;
import org.springframework.test.util.ReflectionTestUtils;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.pojo.dto.ProblemVerificationSubmitDTO;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.pojo.vo.ProblemAiLanguageProgramVO;

import java.util.Collections;
import java.util.Arrays;
import java.util.List;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

class ProblemAiMultiLanguageRunnerTest {
    private final ProblemVerificationManager verification = mock(ProblemVerificationManager.class);
    private final ProblemAiMultiLanguageRunner runner = new ProblemAiMultiLanguageRunner();
    private final Problem problem = new Problem().setId(12L).setJudgeMode("spj");
    private final List<ProblemAiTestPoint> points = Collections.singletonList(new ProblemAiTestPoint()
            .setIndex(1).setCaseId(71L).setConcrete(true).setInputAvailable(false));
    private final List<ProblemAiLanguageProgramVO> programs = Collections.singletonList(
            new ProblemAiLanguageProgramVO().setLanguage("Python3").setCode("print(7)"));

    @BeforeEach
    void setup() throws Exception {
        ReflectionTestUtils.setField(runner, "verificationManager", verification);
        when(verification.submitComparisonForAi(any(), eq("uid"), eq("admin"))).thenReturn(91L);
    }

    @Test
    void usesFullJudgeEvenWhenInputFilesAreNotLocallyReadable() throws Exception {
        when(verification.comparisonResult(12L, 91L)).thenReturn(result("SUCCESS", 0, 0, 1));
        JSONObject language = run().getJSONObject(0);
        assertEquals("PASS", language.getStr("status"));
        assertEquals(1, language.getInt("passed"));
        assertEquals(1, language.getInt("total"));
        ArgumentCaptor<ProblemVerificationSubmitDTO> submission = ArgumentCaptor.forClass(ProblemVerificationSubmitDTO.class);
        verify(verification).submitComparisonForAi(submission.capture(), eq("uid"), eq("admin"));
        assertEquals(12L, submission.getValue().getPid());
        assertEquals("Python3", submission.getValue().getLanguage());
        assertEquals("print(7)", submission.getValue().getCode());
        verify(verification).waitForAiJudge(12L, 91L, 300000L);
        verify(verification, never()).submitForAi(any(), anyString(), anyString());
    }

    @Test
    void reportsNonAcceptedAndMissingCasesAsFailures() throws Exception {
        JSONObject result = result("FAILED", -1, -1, 0);
        result.getJSONArray("caseResults").add(new JSONObject().set("index", 2).set("status", "NOT_RUN"));
        result.set("totalCases", 2);
        when(verification.comparisonResult(12L, 91L)).thenReturn(result);
        JSONObject language = run().getJSONObject(0);
        assertEquals("FAIL", language.getStr("status"));
        assertEquals(2, language.getInt("total"));
        assertEquals("FAIL", language.getJSONArray("testPoints").getJSONObject(0).getStr("status"));
        assertEquals("FAIL", language.getJSONArray("testPoints").getJSONObject(1).getStr("status"));
        assertEquals("checker rejected answer", language.getJSONArray("testPoints").getJSONObject(0).getStr("stderr"));
    }

    @Test
    void timeoutKeepsSubmissionIdAndCannotBeReportedAsPass() throws Exception {
        doThrow(new StatusFailException("判题等待超时")).when(verification).waitForAiJudge(12L, 91L, 300000L);
        JSONObject language = run().getJSONObject(0);
        assertEquals("FAIL", language.getStr("status"));
        assertEquals(91L, language.getLong("submitId"));
        assertEquals("判题等待超时", language.getStr("errorMessage"));
        assertEquals("NOT_RUN", language.getJSONArray("testPoints").getJSONObject(0).getStr("judgeStatus"));
    }

    @Test
    void compilationFailureUsesRealTotalAndDiagnostics() throws Exception {
        when(verification.comparisonResult(12L, 91L)).thenReturn(result("FAILED", -2, null, 0)
                .set("errorMessage", "SyntaxError: invalid syntax"));
        JSONObject language = run().getJSONObject(0);
        assertEquals("FAIL", language.getStr("status"));
        assertEquals(1, language.getInt("total"));
        assertEquals("SyntaxError: invalid syntax", language.getStr("errorMessage"));
    }

    @Test
    void interruptedJudgeWaitDoesNotContinueWithAnotherLanguage() throws Exception {
        doAnswer(call -> { throw new InterruptedException("cancelled"); })
                .when(verification).waitForAiJudge(12L, 91L, 300000L);
        try {
            assertThrows(InterruptedException.class, () -> runner.run(problem,
                    Arrays.asList(programs.get(0), programs.get(0)), points, "uid", "admin"));
            assertTrue(Thread.currentThread().isInterrupted());
            verify(verification, times(1)).submitComparisonForAi(any(), eq("uid"), eq("admin"));
            verify(verification, never()).comparisonResult(anyLong(), anyLong());
        } finally {
            Thread.interrupted();
        }
    }

    @Test
    void alreadyInterruptedRunDoesNotSubmitAnyLanguage() {
        try {
            Thread.currentThread().interrupt();
            assertThrows(InterruptedException.class, this::run);
            verifyNoInteractions(verification);
        } finally {
            Thread.interrupted();
        }
    }

    private JSONArray run() throws Exception {
        return runner.run(problem, programs, points, "uid", "admin");
    }

    private JSONObject result(String state, Integer judgeStatus, Integer caseStatus, int passed) {
        JSONArray cases = new JSONArray();
        cases.add(new JSONObject().set("index", 1).set("caseId", 71L)
                .set("status", caseStatus == null ? "NOT_RUN" : caseStatus == 0 ? "Accepted" : "Wrong Answer")
                .set("judgeStatus", caseStatus).set("stderr", "checker rejected answer").set("timeMs", 12));
        return new JSONObject().set("status", state).set("judgeStatus", judgeStatus)
                .set("passedCases", passed).set("totalCases", 1).set("caseResults", cases);
    }
}
