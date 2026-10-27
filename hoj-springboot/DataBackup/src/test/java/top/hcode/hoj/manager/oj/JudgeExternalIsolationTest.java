package top.hcode.hoj.manager.oj;

import com.baomidou.mybatisplus.core.MybatisConfiguration;
import com.baomidou.mybatisplus.core.metadata.TableInfoHelper;
import org.apache.ibatis.builder.MapperBuilderAssistant;
import org.apache.shiro.subject.Subject;
import org.apache.shiro.util.ThreadContext;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.test.util.ReflectionTestUtils;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.controller.admin.AdminJudgeController;
import top.hcode.hoj.dao.contest.ContestEntityService;
import top.hcode.hoj.dao.judge.JudgeCaseEntityService;
import top.hcode.hoj.dao.judge.JudgeEntityService;
import top.hcode.hoj.manager.admin.rejudge.RejudgeManager;
import top.hcode.hoj.pojo.dto.SubmitIdListDTO;
import top.hcode.hoj.pojo.entity.contest.Contest;
import top.hcode.hoj.pojo.entity.judge.Judge;
import top.hcode.hoj.service.admin.rejudge.RejudgeService;
import top.hcode.hoj.shiro.AccountProfile;
import top.hcode.hoj.utils.ProblemVerificationConstants;
import top.hcode.hoj.validator.ContestValidator;

import java.util.Arrays;
import java.util.Collections;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

class JudgeExternalIsolationTest {
    private final JudgeManager manager = new JudgeManager();
    private final JudgeEntityService judges = mock(JudgeEntityService.class);
    private final JudgeCaseEntityService cases = mock(JudgeCaseEntityService.class);
    private final Subject subject = mock(Subject.class);
    private final Judge external = new Judge().setSubmitId(11L).setUid("creator").setPid(31L)
            .setCid(0L).setShare(true).setStatus(0).setCode("external answer")
            .setIsProblemVerification(true).setSubmissionType(ProblemVerificationConstants.EXTERNAL_VERIFICATION);

    @BeforeEach
    void setUp() {
        TableInfoHelper.initTableInfo(new MapperBuilderAssistant(new MybatisConfiguration(), ""), Judge.class);
        ReflectionTestUtils.setField(manager, "judgeEntityService", judges);
        ReflectionTestUtils.setField(manager, "judgeCaseEntityService", cases);
        when(judges.getById(11L)).thenReturn(external);
        when(judges.getOne(any())).thenReturn(external);
        ThreadContext.bind(subject);
    }

    @AfterEach
    void tearDown() {
        ThreadContext.remove();
    }

    @Test
    void ordinaryEndpointsRejectExternalSubmissionForAnonymousCreatorAndRoot() {
        for (String uid : Arrays.asList(null, "creator", "root")) {
            AccountProfile profile = new AccountProfile();
            profile.setUid(uid);
            when(subject.getPrincipal()).thenReturn(uid == null ? null : profile);
            when(subject.hasRole("root")).thenReturn("root".equals(uid));
            assertThrows(StatusNotFoundException.class, () -> manager.getSubmission(11L));
            assertThrows(StatusNotFoundException.class, () -> manager.getALLCaseResult(11L));
            assertThrows(StatusNotFoundException.class, () -> manager.resubmit(11L));
            assertThrows(StatusForbiddenException.class,
                    () -> manager.updateSubmission(new Judge().setSubmitId(11L).setShare(true)));
        }
        verify(judges, never()).update(any());
        verify(judges, never()).updateById(any());
        verifyNoInteractions(cases);
    }

    @Test
    void ordinaryPollingCannotRevealExternalStatusOrOwner() {
        Judge ordinary = new Judge().setSubmitId(12L).setUid("ordinary").setStatus(0);
        when(judges.list(any())).thenReturn(Arrays.asList(external, ordinary));
        SubmitIdListDTO request = new SubmitIdListDTO();
        request.setSubmitIds(Arrays.asList(11L, 12L));
        assertEquals(Collections.singleton(12L), manager.checkCommonJudgeResult(request).keySet());
        assertEquals("external answer", external.getCode());
    }

    @Test
    void contestPollingCannotBeUsedToReadExternalStatus() throws Exception {
        AccountProfile profile = new AccountProfile();
        profile.setUid("creator");
        when(subject.getPrincipal()).thenReturn(profile);
        ContestEntityService contests = mock(ContestEntityService.class);
        ReflectionTestUtils.setField(manager, "contestEntityService", contests);
        ReflectionTestUtils.setField(manager, "contestValidator", mock(ContestValidator.class));
        when(contests.getById(21L)).thenReturn(new Contest().setId(21L).setUid("creator").setIsGroup(false));
        when(judges.list(any())).thenReturn(Collections.singletonList(external));
        SubmitIdListDTO request = new SubmitIdListDTO();
        request.setCid(21L);
        request.setSubmitIds(Collections.singletonList(11L));
        assertTrue(manager.checkContestJudgeResult(request).isEmpty());
    }

    @Test
    void administrativeRejudgeManualResultAndCancellationCannotChangeExternalRecord() {
        AdminJudgeController controller = new AdminJudgeController();
        RejudgeService rejudge = mock(RejudgeService.class);
        ReflectionTestUtils.setField(controller, "judgeEntityService", judges);
        ReflectionTestUtils.setField(controller, "rejudgeService", rejudge);
        assertEquals(Integer.valueOf(400), controller.rejudge(11L).getStatus());
        verifyNoInteractions(rejudge);

        RejudgeManager admin = new RejudgeManager();
        ReflectionTestUtils.setField(admin, "judgeEntityService", judges);
        assertThrows(StatusFailException.class, () -> admin.manualJudge(11L, 0, 100));
        assertThrows(StatusFailException.class, () -> admin.cancelJudge(11L));
        verify(judges, never()).update(any());
    }
}
