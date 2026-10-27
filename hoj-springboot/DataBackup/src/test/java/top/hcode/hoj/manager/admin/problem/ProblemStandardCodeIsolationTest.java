package top.hcode.hoj.manager.admin.problem;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import org.apache.shiro.subject.Subject;
import org.apache.shiro.util.ThreadContext;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.test.util.ReflectionTestUtils;
import top.hcode.hoj.dao.judge.JudgeEntityService;
import top.hcode.hoj.dao.problem.ProblemEntityService;
import top.hcode.hoj.dao.problem.ProblemVerificationEntityService;
import top.hcode.hoj.pojo.entity.judge.Judge;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.pojo.entity.problem.ProblemVerification;
import top.hcode.hoj.utils.ProblemVerificationConstants;

import java.util.Arrays;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

class ProblemStandardCodeIsolationTest {
    private final ProblemVerificationManager manager = new ProblemVerificationManager();
    private final JudgeEntityService judges = mock(JudgeEntityService.class);
    private final ProblemEntityService problems = mock(ProblemEntityService.class);
    private final ProblemVerificationEntityService verification = mock(ProblemVerificationEntityService.class);
    private final Judge bound = new Judge().setSubmitId(11L).setPid(31L).setStatus(0)
            .setIsProblemVerification(true).setCode("official answer");

    @BeforeEach
    void setup() {
        ReflectionTestUtils.setField(manager, "judgeEntityService", judges);
        ReflectionTestUtils.setField(manager, "problemEntityService", problems);
        ReflectionTestUtils.setField(manager, "verificationService", verification);
        when(problems.getById(31L)).thenReturn(new Problem().setId(31L).setCaseVersion("v1"));
        when(verification.getOne(any(), eq(false))).thenReturn(new ProblemVerification()
                .setSubmitId(11L).setPid(31L).setCaseVersion("v1"));
        when(judges.getById(11L)).thenReturn(bound);
        Subject subject = mock(Subject.class);
        when(subject.hasRole("root")).thenReturn(true);
        ThreadContext.bind(subject);
    }

    @AfterEach
    void tearDown() {
        ThreadContext.remove();
    }

    @Test
    void officialCreatorAiAndHistoricalBindingsRemainAvailable() throws Exception {
        for (String type : Arrays.asList(null, ProblemVerificationConstants.CREATOR_VALIDATION,
                ProblemVerificationConstants.AI_VALIDATION)) {
            bound.setSubmissionType(type);
            assertEquals("official answer", manager.getLastPassedCode(31L).getCode());
        }
        verify(judges, never()).getOne(any(), anyBoolean());
    }

    @Test
    void externalAndAuxiliaryBindingsFallBackOnlyToOfficialSubmissionTypes() throws Exception {
        when(judges.getOne(any(), eq(false))).thenAnswer(call -> {
            QueryWrapper<Judge> query = call.getArgument(0);
            assertTrue(query.getSqlSegment().contains("submission_type IS NULL OR submission_type IN"));
            assertTrue(query.getParamNameValuePairs().containsValue(ProblemVerificationConstants.CREATOR_VALIDATION));
            assertTrue(query.getParamNameValuePairs().containsValue(ProblemVerificationConstants.AI_VALIDATION));
            assertFalse(query.getParamNameValuePairs().containsValue(ProblemVerificationConstants.EXTERNAL_VERIFICATION));
            return null;
        });
        for (String type : Arrays.asList(ProblemVerificationConstants.EXTERNAL_VERIFICATION,
                ProblemVerificationConstants.AI_COMPARISON, ProblemVerificationConstants.USER_SUBMISSION)) {
            bound.setSubmissionType(type).setCode("untrusted answer");
            assertEquals("", manager.getLastPassedCode(31L).getCode());
        }
        verify(judges, times(3)).getOne(any(), eq(false));
    }
}
