package top.hcode.hoj.manager.oj;

import org.apache.shiro.subject.Subject;
import org.apache.shiro.util.ThreadContext;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.test.util.ReflectionTestUtils;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.dao.contest.ContestProblemEntityService;
import top.hcode.hoj.dao.judge.JudgeEntityService;
import top.hcode.hoj.dao.problem.ProblemEntityService;
import top.hcode.hoj.dao.training.TrainingProblemEntityService;
import top.hcode.hoj.manager.admin.contest.AdminContestProblemManager;
import top.hcode.hoj.manager.admin.training.AdminTrainingProblemManager;
import top.hcode.hoj.manager.file.ProblemFileManager;
import top.hcode.hoj.mapper.ProblemExternalVerificationMapper;
import top.hcode.hoj.pojo.dto.ExternalVerificationSettingDTO;
import top.hcode.hoj.pojo.entity.contest.ContestProblem;
import top.hcode.hoj.pojo.entity.judge.Judge;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.pojo.entity.problem.ProblemExternalVerification;
import top.hcode.hoj.pojo.entity.training.TrainingProblem;
import top.hcode.hoj.service.problem.ProblemPendingSaveGuard;

import javax.servlet.http.HttpServletResponse;
import java.util.Collections;

import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.*;

class CandidateProblemIsolationTest {
    private final ProblemPendingSaveGuard guard = mock(ProblemPendingSaveGuard.class);
    private final ProblemEntityService problems = mock(ProblemEntityService.class);
    // Even an accidentally exposed auth flag must not publish a pending candidate.
    private final Problem candidate = new Problem().setId(31L).setProblemId("P31").setAuth(1).setIsGroup(false);

    @BeforeEach
    void setup() throws Exception {
        ThreadContext.bind(mock(Subject.class));
        when(guard.isCandidate(31L)).thenReturn(true);
        doThrow(new StatusFailException("尚未完成验题保存")).when(guard).requireFormal(31L);
        when(problems.getOne(any(), eq(false))).thenReturn(candidate);
    }

    @AfterEach
    void cleanup() {
        ThreadContext.remove();
    }

    @Test
    void ordinaryProblemDetailsHideCandidate() {
        ProblemManager manager = new ProblemManager();
        ReflectionTestUtils.setField(manager, "problemEntityService", problems);
        ReflectionTestUtils.setField(manager, "pendingSaveGuard", guard);
        assertThrows(StatusNotFoundException.class, () -> manager.getProblemInfo("P31", null));
    }

    @Test
    void ordinarySubmissionCannotPersistCandidateJudge() {
        BeforeDispatchInitManager manager = new BeforeDispatchInitManager();
        JudgeEntityService judges = mock(JudgeEntityService.class);
        ReflectionTestUtils.setField(manager, "problemEntityService", problems);
        ReflectionTestUtils.setField(manager, "pendingSaveGuard", guard);
        ReflectionTestUtils.setField(manager, "judgeEntityService", judges);
        assertThrows(StatusForbiddenException.class, () -> manager.initCommonSubmission("P31", null, new Judge()));
        verifyNoInteractions(judges);
    }

    @Test
    void candidateCannotBeAttachedToContestOrTraining() {
        AdminContestProblemManager contests = new AdminContestProblemManager();
        ContestProblemEntityService contestLinks = mock(ContestProblemEntityService.class);
        ReflectionTestUtils.setField(contests, "pendingSaveGuard", guard);
        ReflectionTestUtils.setField(contests, "contestProblemEntityService", contestLinks);
        assertThrows(StatusFailException.class, () -> contests.setContestProblem(new ContestProblem().setPid(31L)));
        AdminTrainingProblemManager trainings = new AdminTrainingProblemManager();
        TrainingProblemEntityService trainingLinks = mock(TrainingProblemEntityService.class);
        ReflectionTestUtils.setField(trainings, "pendingSaveGuard", guard);
        ReflectionTestUtils.setField(trainings, "trainingProblemEntityService", trainingLinks);
        assertThrows(StatusFailException.class, () -> trainings.updateProblem(new TrainingProblem().setPid(31L)));
        verifyNoInteractions(contestLinks, trainingLinks);
    }

    @Test
    void candidateCannotEnableExternalLinkOrUseAnExistingLink() {
        ExternalVerificationManager manager = new ExternalVerificationManager();
        ProblemExternalVerificationMapper links = mock(ProblemExternalVerificationMapper.class);
        ReflectionTestUtils.setField(manager, "pendingSaveGuard", guard);
        ReflectionTestUtils.setField(manager, "linkMapper", links);
        ExternalVerificationSettingDTO setting = new ExternalVerificationSettingDTO();
        setting.setPid(31L);
        setting.setEnabled(true);
        assertThrows(StatusFailException.class, () -> manager.saveSetting(setting));
        verifyNoInteractions(links);
        String token = String.join("", Collections.nCopies(64, "a"));
        when(links.selectOne(any())).thenReturn(new ProblemExternalVerification().setPid(31L).setEnabled(true));
        assertThrows(StatusForbiddenException.class, () -> manager.problem(token));
        assertThrows(StatusForbiddenException.class, () -> manager.submissions(token, token, 1, 10));
    }

    @Test
    void candidateCannotBeExportedAndImportedAsFormalProblem() {
        ProblemFileManager manager = new ProblemFileManager();
        ReflectionTestUtils.setField(manager, "pendingSaveGuard", guard);
        HttpServletResponse response = mock(HttpServletResponse.class);
        manager.exportProblem(Collections.singletonList(31L), response);
        verify(response).setStatus(HttpServletResponse.SC_FORBIDDEN);
    }
}
