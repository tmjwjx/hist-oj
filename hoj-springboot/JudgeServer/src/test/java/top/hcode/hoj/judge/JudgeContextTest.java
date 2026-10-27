package top.hcode.hoj.judge;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.test.util.ReflectionTestUtils;
import top.hcode.hoj.dao.ContestRecordEntityService;
import top.hcode.hoj.dao.JudgeEntityService;
import top.hcode.hoj.dao.UserAcproblemEntityService;
import top.hcode.hoj.pojo.entity.judge.Judge;
import top.hcode.hoj.pojo.entity.user.UserAcproblem;

import static org.mockito.ArgumentMatchers.argThat;
import static org.mockito.Mockito.*;

class JudgeContextTest {
    private final JudgeContext context = new JudgeContext();
    private final JudgeEntityService judges = mock(JudgeEntityService.class);
    private final UserAcproblemEntityService accepted = mock(UserAcproblemEntityService.class);
    private final ContestRecordEntityService contests = mock(ContestRecordEntityService.class);

    @BeforeEach
    void setup() {
        ReflectionTestUtils.setField(context, "judgeEntityService", judges);
        ReflectionTestUtils.setField(context, "userAcproblemEntityService", accepted);
        ReflectionTestUtils.setField(context, "contestRecordEntityService", contests);
    }

    @Test
    void classroomAcceptanceDoesNotPublishSolvedProblem() {
        when(judges.getById(12L)).thenReturn(new Judge().setSubmitId(12L).setHomeworkId(8L));
        update(0, 0L, null);
        verifyZeroInteractions(accepted, contests);
    }

    @Test
    void practiceAcceptanceStillPublishesSolvedProblem() {
        when(judges.getById(12L)).thenReturn(new Judge().setSubmitId(12L));
        update(0, 0L, null);
        verify(accepted).saveOrUpdate(argThat((UserAcproblem row) ->
                row.getSubmitId().equals(12L) && row.getPid().equals(23L) && row.getUid().equals("student")));
        verifyZeroInteractions(contests);
    }

    @Test
    void deletedOrVerificationSubmissionDoesNotPublishSolvedProblem() {
        update(0, 0L, null);
        when(judges.getById(12L)).thenReturn(new Judge().setSubmitId(12L).setIsProblemVerification(true));
        update(0, 0L, null);
        verifyZeroInteractions(accepted, contests);
    }

    @Test
    void contestScoringStillUpdatesWithoutExtraLookup() {
        update(0, 6L, null);
        verify(contests).updateContestRecord(100, 0, 12L, 15);
        verifyZeroInteractions(judges, accepted);
    }

    @Test
    void unsuccessfulAndGroupSubmissionsNeedNoPublicStatisticsLookup() {
        update(-1, 0L, null);
        update(0, 0L, 7L);
        verifyZeroInteractions(judges, accepted, contests);
    }

    private void update(int status, Long cid, Long gid) {
        context.updateOtherTable(12L, status, cid, "student", 23L, gid, 100, 15);
    }
}
