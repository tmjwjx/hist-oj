package top.hcode.hoj.manager.admin.problem;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.test.util.ReflectionTestUtils;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.dao.judge.JudgeCaseEntityService;
import top.hcode.hoj.dao.judge.JudgeEntityService;
import top.hcode.hoj.dao.problem.ProblemCaseEntityService;
import top.hcode.hoj.dao.problem.ProblemEntityService;
import top.hcode.hoj.dao.problem.ProblemVerificationEntityService;
import top.hcode.hoj.pojo.dto.ProblemAiValidateDTO;
import top.hcode.hoj.pojo.entity.judge.Judge;
import top.hcode.hoj.pojo.entity.judge.JudgeCase;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.pojo.entity.problem.ProblemCase;

import java.util.Arrays;
import java.util.Collections;
import java.util.List;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

class ProblemAiValidationContextTest {
    private final ProblemVerificationEntityService verification = mock(ProblemVerificationEntityService.class);
    private final ProblemCaseEntityService cases = mock(ProblemCaseEntityService.class);
    private final ProblemEntityService problems = mock(ProblemEntityService.class);
    private final JudgeEntityService judges = mock(JudgeEntityService.class);
    private final JudgeCaseEntityService results = mock(JudgeCaseEntityService.class);
    private final ProblemAiValidationContext context = new ProblemAiValidationContext();
    private final ProblemAiValidateDTO input = new ProblemAiValidateDTO();
    private Judge judge;

    @BeforeEach
    void setup() {
        ReflectionTestUtils.setField(context, "verificationService", verification);
        ReflectionTestUtils.setField(context, "problemCaseService", cases);
        ReflectionTestUtils.setField(context, "problemService", problems);
        ReflectionTestUtils.setField(context, "judgeService", judges);
        ReflectionTestUtils.setField(context, "judgeCaseService", results);
        input.setPid(5L); input.setSubmitId(91L); input.setLanguage("C++ 17"); input.setStandardProgram("std");
        judge = new Judge().setSubmitId(91L).setPid(5L).setStatus(-2).setCode("std")
                .setLanguage("C++ 17").setIsProblemVerification(true).setErrorMessage("compile error");
        when(judges.getById(91L)).thenReturn(judge);
        when(problems.getById(5L)).thenReturn(new Problem().setId(5L)
                .setExamples("<input>2\r\n</input><output>4\r\n</output>"));
        when(cases.list(any())).thenReturn(Collections.singletonList(
                new ProblemCase().setId(51L).setInput("1\n").setOutput("2\n")));
        when(results.list(any())).thenReturn(Collections.emptyList());
    }

    @Test
    void compileErrorIncludesAllFormalAndSamplePointsAsNotExecuted() throws Exception {
        ProblemAiValidationContext.Snapshot snapshot = context.load(input);
        assertEquals(2, snapshot.getPoints().size());
        assertEquals("compile error", snapshot.getJudgeError());
        for (ProblemAiTestPoint point : snapshot.getPoints()) {
            assertEquals(false, point.getExecuted());
            assertEquals("NOT_RUN", point.getJudgeStatusText());
        }
        assertEquals(51L, snapshot.getPoints().get(0).getCaseId());
        assertEquals("2\n", snapshot.getPoints().get(1).getInput());
        assertEquals("4\n", snapshot.getPoints().get(1).getExpectedOutput());
        assertEquals(true, snapshot.getPoints().get(1).getInputAvailable());
    }

    @Test
    void sparseJudgeResultsRetainHighestSequenceAndMarkHoles() throws Exception {
        judge.setStatus(-1);
        when(results.list(any())).thenReturn(Arrays.asList(
                new JudgeCase().setSeq(1).setStatus(0),
                new JudgeCase().setSeq(4).setStatus(-1).setStderr("bad answer")));
        List<ProblemAiTestPoint> points = context.load(input).getPoints();
        assertEquals(4, points.size());
        assertEquals(true, points.get(0).getExecuted());
        assertEquals(false, points.get(2).getExecuted());
        assertEquals(-1, points.get(3).getJudgeStatus());
        assertEquals("bad answer", points.get(3).getStderr());
    }

    @Test
    void placeholderStatusesAreNotClaimedAsExecuted() throws Exception {
        when(results.list(any())).thenReturn(Collections.singletonList(new JudgeCase().setSeq(1).setStatus(15)));
        assertEquals(false, context.load(input).getPoints().get(0).getExecuted());
    }

    @Test
    void explicitSubmissionStillMustBelongToProblemAndMatchProgram() {
        judge.setPid(6L);
        assertThrows(StatusFailException.class, () -> context.load(input));
        judge.setPid(5L).setCode("other");
        assertThrows(StatusFailException.class, () -> context.load(input));
    }

    @Test
    void externalOrComparisonSubmissionCannotBecomeOfficialStandardProgram() {
        for (String type : Arrays.asList("external_verification", "ai_comparison", "user_submission")) {
            judge.setSubmissionType(type);
            assertThrows(StatusFailException.class, () -> context.load(input));
        }
        verifyNoInteractions(cases, results, problems);
    }

    @Test
    void sampleEvidenceMatchesJudgeNormalizationWithoutTrimmingWhitespace() throws Exception {
        String[] raw = {"1 ", "1\n", " 1 \n\n", "1\r\n2\r\n\r\n"};
        String[] normalized = {"1 \n", "1\n", " 1 \n\n", "1\n2\n\n"};
        for (int index = 0; index < raw.length; index++) {
            when(problems.getById(5L)).thenReturn(new Problem().setId(5L)
                    .setExamples("<input>" + raw[index] + "</input><output>2 </output>"));
            List<ProblemAiTestPoint> points = context.load(input).getPoints();
            assertEquals(normalized[index], points.get(1).getInput());
            assertEquals("2 ", points.get(1).getExpectedOutput());
            assertEquals("1\n", points.get(0).getInput());
        }
    }
}
