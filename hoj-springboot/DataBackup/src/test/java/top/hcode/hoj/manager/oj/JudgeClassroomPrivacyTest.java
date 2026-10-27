package top.hcode.hoj.manager.oj;

import com.baomidou.mybatisplus.core.MybatisConfiguration;
import com.baomidou.mybatisplus.core.metadata.TableInfoHelper;
import org.apache.ibatis.builder.MapperBuilderAssistant;
import org.apache.shiro.subject.Subject;
import org.apache.shiro.util.ThreadContext;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.InOrder;
import org.springframework.mock.web.MockHttpServletRequest;
import org.springframework.test.util.ReflectionTestUtils;
import org.springframework.web.context.request.RequestContextHolder;
import org.springframework.web.context.request.ServletRequestAttributes;
import top.hcode.hoj.common.exception.StatusAccessDeniedException;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.dao.judge.JudgeCaseEntityService;
import top.hcode.hoj.dao.judge.JudgeEntityService;
import top.hcode.hoj.dao.problem.ProblemEntityService;
import top.hcode.hoj.config.NacosSwitchConfig;
import top.hcode.hoj.config.SwitchConfig;
import top.hcode.hoj.judge.self.JudgeDispatcher;
import top.hcode.hoj.manager.classroom.ClassroomAccessManager;
import top.hcode.hoj.manager.classroom.HomeworkManager;
import top.hcode.hoj.mapper.classroom.ClassroomHomeworkMapper;
import top.hcode.hoj.pojo.dto.SubmitIdListDTO;
import top.hcode.hoj.pojo.dto.SubmitJudgeDTO;
import top.hcode.hoj.pojo.entity.classroom.Classroom;
import top.hcode.hoj.pojo.entity.classroom.ClassroomHomework;
import top.hcode.hoj.pojo.entity.judge.Judge;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.pojo.vo.SubmissionInfoVO;
import top.hcode.hoj.shiro.AccountProfile;
import top.hcode.hoj.validator.AccessValidator;
import top.hcode.hoj.validator.JudgeValidator;

import java.util.Arrays;
import java.util.Collections;
import java.util.Map;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

class JudgeClassroomPrivacyTest {
    private final JudgeManager manager = new JudgeManager();
    private final JudgeEntityService judges = mock(JudgeEntityService.class);
    private final JudgeCaseEntityService cases = mock(JudgeCaseEntityService.class);
    private final ProblemEntityService problems = mock(ProblemEntityService.class);
    private final ClassroomHomeworkMapper homework = mock(ClassroomHomeworkMapper.class);
    private final ClassroomAccessManager access = mock(ClassroomAccessManager.class);
    private final HomeworkManager homeworkManager = mock(HomeworkManager.class);
    private final Subject subject = mock(Subject.class);
    private Judge classroomJudge;

    @BeforeEach
    void setUp() {
        TableInfoHelper.initTableInfo(new MapperBuilderAssistant(new MybatisConfiguration(), ""), Judge.class);
        ReflectionTestUtils.setField(manager, "judgeEntityService", judges);
        ReflectionTestUtils.setField(manager, "judgeCaseEntityService", cases);
        ReflectionTestUtils.setField(manager, "problemEntityService", problems);
        ReflectionTestUtils.setField(manager, "classroomHomeworkMapper", homework);
        ReflectionTestUtils.setField(manager, "classroomAccessManager", access);
        ReflectionTestUtils.setField(manager, "homeworkManager", homeworkManager);
        ReflectionTestUtils.setField(manager, "accessValidator", mock(AccessValidator.class));
        ThreadContext.bind(subject);
        classroomJudge = new Judge().setSubmitId(11L).setHomeworkId(21L).setPid(31L)
                .setCid(0L).setUid("author").setShare(true).setCode("private answer").setStatus(0);
        when(judges.getById(11L)).thenReturn(classroomJudge);
        when(judges.getOne(any())).thenReturn(classroomJudge);
        when(problems.getById(31L)).thenReturn(new Problem().setCodeShare(true)
                .setOpenCaseResult(true).setIsRemote(false));
        when(cases.list(any())).thenReturn(Collections.emptyList());
    }

    @AfterEach
    void tearDown() {
        ThreadContext.remove();
        RequestContextHolder.resetRequestAttributes();
    }

    private void login(String uid) {
        AccountProfile profile = new AccountProfile();
        profile.setUid(uid);
        when(subject.getPrincipal()).thenReturn(profile);
    }

    @Test
    void anonymousCannotReadClassroomDetailsOrTestOutputEvenWhenShared() {
        assertThrows(StatusAccessDeniedException.class, () -> manager.getSubmission(11L));
        assertThrows(StatusForbiddenException.class, () -> manager.getALLCaseResult(11L));
        verifyNoInteractions(problems, cases);
    }

    @Test
    void ordinarySharedSubmissionRemainsPublic() throws Exception {
        classroomJudge.setHomeworkId(null);
        SubmissionInfoVO detail = manager.getSubmission(11L);
        assertEquals("private answer", detail.getSubmission().getCode());
        assertTrue(detail.getCodeShare());
        assertNotNull(manager.getALLCaseResult(11L));
        verifyNoInteractions(homework, access);
    }

    @Test
    void otherStudentAndProblemAdministratorCannotReadClassroomSubmission() throws Exception {
        login("classmate");
        when(subject.hasRole("problem_admin")).thenReturn(true);
        Classroom classroom = new Classroom().setId(41L);
        when(homework.selectById(21L)).thenReturn(new ClassroomHomework().setClassroomId(41L));
        when(access.requireClassroom(41L)).thenReturn(classroom);
        assertThrows(StatusAccessDeniedException.class, () -> manager.getSubmission(11L));
        assertThrows(StatusForbiddenException.class, () -> manager.getALLCaseResult(11L));
        verify(access, times(2)).canManage(classroom, "classmate");
    }

    @Test
    void authorCanReadOwnCodeAndResultsButCannotEnableSharing() throws Exception {
        login("author");
        SubmissionInfoVO detail = manager.getSubmission(11L);
        assertEquals("private answer", detail.getSubmission().getCode());
        assertFalse(detail.getSubmission().getShare());
        assertFalse(detail.getCodeShare());
        assertNotNull(manager.getALLCaseResult(11L));
        assertThrows(StatusForbiddenException.class,
                () -> manager.updateSubmission(new Judge().setSubmitId(11L).setShare(true)));
        verify(judges, never()).update(any());
    }

    @Test
    void assignedTeacherCanReadStudentsCode() throws Exception {
        login("teacher");
        Classroom classroom = new Classroom().setId(41L);
        when(homework.selectById(21L)).thenReturn(new ClassroomHomework().setClassroomId(41L));
        when(access.requireClassroom(41L)).thenReturn(classroom);
        when(access.canManage(classroom, "teacher")).thenReturn(true);
        assertEquals("private answer", manager.getSubmission(11L).getSubmission().getCode());
        assertNotNull(manager.getALLCaseResult(11L));
    }

    @Test
    void administratorCanReadOrphanedClassroomSubmission() throws Exception {
        login("admin");
        when(access.isAdmin()).thenReturn(true);
        assertEquals("private answer", manager.getSubmission(11L).getSubmission().getCode());
        verifyNoInteractions(homework);
    }

    @Test
    void deletedClassroomDoesNotMakeSubmissionsPublic() throws Exception {
        login("outsider");
        when(homework.selectById(21L)).thenReturn(new ClassroomHomework().setClassroomId(41L));
        when(access.requireClassroom(41L)).thenThrow(new StatusNotFoundException("班级不存在"));
        assertThrows(StatusAccessDeniedException.class, () -> manager.getSubmission(11L));
    }

    @Test
    void completedOrExpiredExamCannotBeRejudgedByAuthor() throws Exception {
        login("author");
        classroomJudge.setDisplayPid("P1000");
        doThrow(new StatusFailException("考试已经交卷，不能继续提交代码"))
                .when(homeworkManager).checkProgrammingSubmission(21L, "P1000");
        assertThrows(StatusForbiddenException.class, () -> manager.resubmit(11L));
        verify(judges, never()).updateById(any());
        verifyNoInteractions(problems, cases);
        assertEquals(Integer.valueOf(0), classroomJudge.getStatus());
    }

    @Test
    void classroomSourceAndAnswerAreSavedBeforeDispatch() throws Exception {
        login("author");
        BeforeDispatchInitManager init = mock(BeforeDispatchInitManager.class);
        JudgeDispatcher dispatcher = mock(JudgeDispatcher.class);
        NacosSwitchConfig config = mock(NacosSwitchConfig.class);
        SwitchConfig switches = mock(SwitchConfig.class);
        when(config.getSwitchConfig()).thenReturn(switches);
        when(switches.getDefaultSubmitInterval()).thenReturn(0);
        ReflectionTestUtils.setField(manager, "nacosSwitchConfig", config);
        ReflectionTestUtils.setField(manager, "judgeValidator", mock(JudgeValidator.class));
        ReflectionTestUtils.setField(manager, "beforeDispatchInitManager", init);
        ReflectionTestUtils.setField(manager, "judgeDispatcher", dispatcher);
        RequestContextHolder.setRequestAttributes(new ServletRequestAttributes(new MockHttpServletRequest()));
        doAnswer(call -> {
            Judge judge = call.getArgument(2);
            assertEquals(Long.valueOf(21), judge.getHomeworkId());
            judge.setSubmitId(11L).setPid(31L);
            return null;
        }).when(init).initClassroomSubmission(eq("P1000"), isNull(), any());
        manager.submitProblemJudge(new SubmitJudgeDTO().setPid("P1000").setCid(0L)
                .setHomeworkId(21L).setCode("answer").setLanguage("C++ 17").setIsRemote(false));
        InOrder sequence = inOrder(homeworkManager, init, dispatcher);
        sequence.verify(homeworkManager).checkProgrammingSubmission(21L, "P1000");
        sequence.verify(init).initClassroomSubmission(eq("P1000"), isNull(), any());
        sequence.verify(dispatcher).sendTask(11L, 31L, false);
        verify(init, never()).initCommonSubmission(any(), any(), any());
        doThrow(new StatusFailException("考试已经交卷，不能继续提交代码"))
                .when(init).initClassroomSubmission(any(), any(), any());
        assertThrows(StatusFailException.class, () -> manager.submitProblemJudge(new SubmitJudgeDTO()
                .setPid("P1000").setCid(0L).setHomeworkId(21L).setCode("late answer")
                .setLanguage("C++ 17").setIsRemote(false)));
        verify(dispatcher, times(1)).sendTask(any(), any(), any());
    }

    @Test
    void statusPollingOmitsUnauthorizedClassroomIdsAndKeepsPublicResults() {
        Judge publicJudge = new Judge().setSubmitId(12L).setUid("someone").setStatus(0);
        when(judges.list(any())).thenReturn(Arrays.asList(classroomJudge, publicJudge));
        SubmitIdListDTO request = new SubmitIdListDTO();
        request.setSubmitIds(Arrays.asList(11L, 12L));
        Map<Long, Object> anonymousResults = manager.checkCommonJudgeResult(request);
        assertEquals(Collections.singleton(12L), anonymousResults.keySet());
        login("author");
        Map<Long, Object> ownResults = manager.checkCommonJudgeResult(request);
        assertEquals(2, ownResults.size());
        assertNull(((Judge) ownResults.get(11L)).getCode());
    }
}
