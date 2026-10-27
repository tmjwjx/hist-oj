package top.hcode.hoj.manager.admin.problem;

import cn.hutool.json.JSONObject;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.test.util.ReflectionTestUtils;
import top.hcode.hoj.dao.judge.JudgeCaseEntityService;
import top.hcode.hoj.dao.judge.JudgeEntityService;
import top.hcode.hoj.dao.problem.ProblemCaseEntityService;
import top.hcode.hoj.dao.problem.ProblemEntityService;
import top.hcode.hoj.dao.problem.ProblemVerificationEntityService;
import top.hcode.hoj.judge.self.JudgeDispatcher;
import top.hcode.hoj.pojo.dto.ProblemVerificationSubmitDTO;
import top.hcode.hoj.pojo.entity.judge.Judge;
import top.hcode.hoj.pojo.entity.judge.JudgeCase;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.pojo.entity.problem.ProblemVerification;
import top.hcode.hoj.utils.ProblemVerificationConstants;

import java.util.Arrays;
import java.util.Collections;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

class ProblemAiComparisonTest {
    private final ProblemVerificationManager manager = new ProblemVerificationManager();
    private final ProblemEntityService problems = mock(ProblemEntityService.class);
    private final ProblemCaseEntityService cases = mock(ProblemCaseEntityService.class);
    private final JudgeEntityService judges = mock(JudgeEntityService.class);
    private final JudgeCaseEntityService judgeCases = mock(JudgeCaseEntityService.class);
    private final ProblemVerificationEntityService verification = mock(ProblemVerificationEntityService.class);
    private final JudgeDispatcher dispatcher = mock(JudgeDispatcher.class);
    private final Problem problem = new Problem().setId(7L).setProblemId("P7").setCaseVersion("v1")
            .setExamples("<input>1\n</input><output>1\n</output>");
    private final Judge judge = new Judge().setSubmitId(100L).setPid(7L).setIsProblemVerification(true).setStatus(0);
    private final ProblemVerification binding = new ProblemVerification().setId(1L).setPid(7L)
            .setCaseVersion("v1").setSyncStatus(1).setSubmitId(99L).setVerificationStatus(2);

    @BeforeEach
    void setup() {
        ReflectionTestUtils.setField(manager, "problemEntityService", problems);
        ReflectionTestUtils.setField(manager, "problemCaseEntityService", cases);
        ReflectionTestUtils.setField(manager, "judgeEntityService", judges);
        ReflectionTestUtils.setField(manager, "judgeCaseEntityService", judgeCases);
        ReflectionTestUtils.setField(manager, "verificationService", verification);
        ReflectionTestUtils.setField(manager, "judgeDispatcher", dispatcher);
        when(problems.getById(7L)).thenReturn(problem);
        when(cases.count(any())).thenReturn(2);
        when(judges.getById(100L)).thenReturn(judge);
        when(verification.getOne(any(), eq(false))).thenReturn(binding);
        when(judgeCases.list(any())).thenReturn(Collections.emptyList());
    }

    @Test
    void auxiliaryUsesFullJudgeButNeverChangesAuthoritativeBinding() throws Exception {
        when(judges.save(any(Judge.class))).thenAnswer(call -> {
            ((Judge) call.getArgument(0)).setSubmitId(100L);
            return true;
        });
        ProblemVerificationSubmitDTO request = new ProblemVerificationSubmitDTO();
        request.setPid(7L); request.setCode("int main(){}"); request.setLanguage("C++ 17");
        assertEquals(100L, manager.submitComparisonForAi(request, "admin", "admin"));
        manager.waitForAiJudge(7L, 100L, 1000);
        verify(judges).save(argThat((Judge row) -> Boolean.TRUE.equals(row.getIsProblemVerification())
                && ProblemVerificationConstants.AI_COMPARISON.equals(row.getSubmissionType())
                && Boolean.FALSE.equals(row.getShare())));
        verify(dispatcher).sendTask(100L, 7L, false);
        verify(verification, never()).updateById(any());
        assertEquals(99L, binding.getSubmitId());
    }

    @Test
    void compileErrorStillReportsFullDatasetAndSamplesNotRun() throws Exception {
        judge.setStatus(-2).setErrorMessage("compile error");
        JSONObject result = manager.comparisonResult(7L, 100L);
        assertEquals("FAILED", result.getStr("status"));
        assertEquals(3, result.getInt("totalCases"));
        assertEquals(0, result.getInt("passedCases"));
        assertEquals(3, result.getInt("pendingCases"));
        assertEquals("compile error", result.getStr("errorMessage"));
    }

    @Test
    void aggregateAcceptedCannotHideMissingOrFailedCases() throws Exception {
        when(judgeCases.list(any())).thenReturn(Arrays.asList(
                new JudgeCase().setSeq(1).setStatus(0), new JudgeCase().setSeq(3).setStatus(-1).setStderr("wrong")));
        JSONObject result = manager.comparisonResult(7L, 100L);
        assertEquals("FAILED", result.getStr("status"));
        assertEquals(1, result.getInt("passedCases"));
        assertEquals(1, result.getInt("failedCases"));
        assertEquals(1, result.getInt("pendingCases"));
        assertEquals("wrong", result.getJSONArray("caseResults").getJSONObject(2).getStr("stderr"));
    }

    @Test
    void allConcreteCasesAndSamplesMustPass() throws Exception {
        when(judgeCases.list(any())).thenReturn(Arrays.asList(new JudgeCase().setSeq(1).setStatus(0),
                new JudgeCase().setSeq(2).setStatus(0), new JudgeCase().setSeq(3).setStatus(0)));
        JSONObject result = manager.comparisonResult(7L, 100L);
        assertEquals("SUCCESS", result.getStr("status"));
        assertEquals(3, result.getInt("passedCases"));
        judge.setStatus(4);
        assertEquals("FAILED", manager.comparisonResult(7L, 100L).getStr("status"));
    }

    @Test
    void interactiveDisplayTranscriptIsNotCountedAsHiddenInput() throws Exception {
        problem.setJudgeMode("interactive");
        assertEquals(2, manager.comparisonResult(7L, 100L).getInt("totalCases"));
    }

    @Test
    void checkerDiagnosticsArePreservedWhenStoredAsUserOutput() throws Exception {
        when(judgeCases.list(any())).thenReturn(Collections.singletonList(
                new JudgeCase().setSeq(1).setStatus(-1).setUserOutput("checker: illegal output")));
        assertEquals("checker: illegal output", manager.comparisonResult(7L, 100L)
                .getJSONArray("caseResults").getJSONObject(0).getStr("stderr"));
    }

    @Test
    void interactiveAcceptedDoesNotClaimDisplayTranscriptWasExecuted() {
        problem.setJudgeMode("interactive");
        binding.setSubmitId(100L).setVerificationStatus(1);
        ReflectionTestUtils.invokeMethod(manager, "refreshJudgeStatus", binding);
        assertEquals(2, binding.getVerificationStatus());
        assertFalse(binding.getSampleVerified());
        assertTrue(binding.getSyncMessage().contains("未作为隐藏输入执行"));
    }

    @Test
    void wrongProblemOrOrdinarySubmissionCannotSupplyEvidence() {
        judge.setPid(8L);
        assertThrows(Exception.class, () -> manager.comparisonResult(7L, 100L));
        judge.setPid(7L).setIsProblemVerification(false);
        assertThrows(Exception.class, () -> manager.comparisonResult(7L, 100L));
    }
}
