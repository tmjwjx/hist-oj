package top.hcode.hoj.manager.classroom;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.test.util.ReflectionTestUtils;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.dao.judge.JudgeEntityService;
import top.hcode.hoj.dao.problem.ProblemEntityService;
import top.hcode.hoj.mapper.classroom.ClassroomHomeworkMapper;
import top.hcode.hoj.mapper.classroom.HomeworkQuestionMapper;
import top.hcode.hoj.mapper.classroom.HomeworkSubmitMapper;
import top.hcode.hoj.pojo.entity.classroom.Classroom;
import top.hcode.hoj.pojo.entity.classroom.ClassroomHomework;
import top.hcode.hoj.pojo.entity.classroom.HomeworkQuestion;
import top.hcode.hoj.pojo.entity.classroom.HomeworkSubmit;
import top.hcode.hoj.pojo.entity.judge.Judge;
import top.hcode.hoj.pojo.entity.problem.Problem;

import java.util.Arrays;
import java.util.Collections;
import java.util.Date;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

class HomeworkProgrammingTest {
    private HomeworkManager manager;
    private ClassroomHomeworkMapper homeworks;
    private HomeworkQuestionMapper questions;
    private HomeworkSubmitMapper submissions;
    private JudgeEntityService judges;
    private ProblemEntityService problems;
    private ClassroomAccessManager access;
    private ClassroomHomework homework;
    private HomeworkQuestion question;

    @BeforeEach
    void setUp() throws Exception {
        manager = new HomeworkManager();
        homeworks = mock(ClassroomHomeworkMapper.class);
        questions = mock(HomeworkQuestionMapper.class);
        submissions = mock(HomeworkSubmitMapper.class);
        judges = mock(JudgeEntityService.class);
        problems = mock(ProblemEntityService.class);
        access = mock(ClassroomAccessManager.class);
        ReflectionTestUtils.setField(manager, "homeworkMapper", homeworks);
        ReflectionTestUtils.setField(manager, "questionMapper", questions);
        ReflectionTestUtils.setField(manager, "submitMapper", submissions);
        ReflectionTestUtils.setField(manager, "judgeService", judges);
        ReflectionTestUtils.setField(manager, "problemService", problems);
        ReflectionTestUtils.setField(manager, "access", access);
        homework = new ClassroomHomework().setId(1L).setClassroomId(2L).setStatus(2).setIsExamMode(0)
                .setStartTime(new Date(System.currentTimeMillis() - 60_000))
                .setEndTime(new Date(System.currentTimeMillis() + 60_000));
        question = new HomeworkQuestion().setId(3L).setHomeworkId(1L).setProblemId("P1000").setScore(20);
        when(homeworks.selectById(1L)).thenReturn(homework);
        when(access.uid()).thenReturn("student");
        when(access.requireClassroom(2L)).thenReturn(new Classroom().setId(2L));
        when(access.isStudent(2L, "student")).thenReturn(true);
        when(questions.selectCount(any())).thenReturn(1);
        when(questions.selectById(3L)).thenReturn(question);
        when(questions.selectList(any())).thenReturn(Collections.singletonList(question));
    }

    private Judge judge(long id, int status) {
        return new Judge().setSubmitId(id).setHomeworkId(1L).setUid("student").setPid(4L)
                .setDisplayPid("P1000").setStatus(status).setCode("int main() {}").setLanguage("C++ 17")
                .setSubmitTime(new Date());
    }

    private HomeworkSubmit submission(long id) {
        return new HomeworkSubmit().setId(5L).setHomeworkId(1L).setUid("student").setProblemId("P1000")
                .setSubmitId(id).setScore(0d).setIsScored(0).setIsOfficiallySubmitted(1);
    }

    @Test
    void partialAcceptanceFinishesAndUsesActualProblemTotal() {
        HomeworkSubmit row = submission(10);
        when(judges.listByIds(anyCollection())).thenReturn(Collections.singletonList(judge(10, 8).setScore(60)));
        when(problems.getById(4L)).thenReturn(new Problem().setIoScore(200));
        manager.syncProgrammingResults(1L, Collections.singletonList(row));
        assertEquals("PAC", row.getJudgeResult());
        assertEquals(1, row.getIsScored());
        assertEquals(6d, row.getScore());
        verify(submissions).update(isNull(), any());
    }

    @Test
    void unchangedVerdictPreservesTeacherScoreButLaterRejudgeRefreshesIt() {
        HomeworkSubmit row = submission(10).setScore(17d).setIsScored(1).setJudgeResult("PAC")
                .setUpdateTime(new Date(20_000));
        Judge verdict = judge(10, 8).setScore(50).setGmtModified(new Date(10_000));
        when(judges.listByIds(anyCollection())).thenReturn(Collections.singletonList(verdict));
        when(problems.getById(4L)).thenReturn(new Problem().setIoScore(100));
        manager.syncProgrammingResults(1L, Collections.singletonList(row));
        assertEquals(17d, row.getScore());
        verify(submissions, never()).update(any(), any());
        verdict.setGmtModified(new Date(30_000));
        manager.syncProgrammingResults(1L, Collections.singletonList(row));
        assertEquals(10d, row.getScore());
        verify(submissions).update(isNull(), any());
    }

    @Test
    void staleStoredVerdictIsCorrectedEvenWhenAlreadyScored() {
        HomeworkSubmit row = submission(10).setScore(0d).setIsScored(1).setJudgeResult("WA");
        when(judges.listByIds(anyCollection())).thenReturn(Collections.singletonList(judge(10, 1)));
        manager.syncProgrammingResults(1L, Collections.singletonList(row));
        assertEquals("TLE", row.getJudgeResult());
        verify(submissions).update(isNull(), any());
    }

    @Test
    void ownResultSynchronizesEvenAfterBrowserStoppedPolling() throws Exception {
        HomeworkSubmit row = submission(10);
        when(submissions.selectList(any())).thenReturn(Collections.singletonList(row));
        when(judges.listByIds(anyCollection())).thenReturn(Collections.singletonList(judge(10, 0)));
        Map<String, Object> result = manager.myDetail(1L);
        assertEquals(20d, result.get("score"));
        assertEquals(false, result.get("hasUngraded"));
        assertEquals("AC", row.getJudgeResult());
    }

    @Test
    void pendingRejudgeClearsPreviousGradeAndFinalErrorsFinish() {
        for (int status : new int[]{5, 6, 7, 9, 15}) {
            HomeworkSubmit row = submission(10).setScore(20d).setIsScored(1).setJudgeResult("AC");
            when(judges.listByIds(anyCollection())).thenReturn(Collections.singletonList(judge(10, status)));
            manager.syncProgrammingResults(1L, Collections.singletonList(row));
            assertEquals("PENDING", row.getJudgeResult());
            assertEquals(0, row.getIsScored());
            assertEquals(0d, row.getScore());
        }
        for (int status : new int[]{-5, -4, -3, -2, -1, 1, 2, 3, 4, 10}) {
            HomeworkSubmit row = submission(10);
            when(judges.listByIds(anyCollection())).thenReturn(Collections.singletonList(judge(10, status)));
            manager.syncProgrammingResults(1L, Collections.singletonList(row));
            assertNotEquals("PENDING", row.getJudgeResult());
            assertEquals(1, row.getIsScored());
            assertEquals(0d, row.getScore());
        }
    }

    @Test
    void unrelatedJudgeNeverSuppliesCodeOrGrade() throws Exception {
        HomeworkSubmit row = submission(10);
        Judge other = judge(10, 0).setUid("other-student");
        when(submissions.selectList(any())).thenReturn(Collections.singletonList(row));
        when(judges.listByIds(anyCollection())).thenReturn(Collections.singletonList(other));
        when(judges.getById(10L)).thenReturn(other);
        Map<String, String> params = new HashMap<>();
        params.put("homeworkId", "1");
        params.put("homeworkQuestionId", "3");
        assertTrue(manager.programmingSubmissions(params).isEmpty());
        assertEquals(0, row.getIsScored());
        verify(submissions, never()).update(any(), any());
    }

    @Test
    void historyRetainsEarlierJudgesAfterLatestAnswerChanges() throws Exception {
        HomeworkSubmit row = submission(11);
        Judge latest = judge(11, 0);
        Judge previous = judge(10, -1);
        when(submissions.selectList(any())).thenReturn(Collections.singletonList(row));
        when(judges.listByIds(anyCollection())).thenReturn(Collections.singletonList(latest));
        when(judges.list(any())).thenReturn(Arrays.asList(latest, previous));
        Map<String, String> params = new HashMap<>();
        params.put("homeworkId", "1");
        params.put("homeworkQuestionId", "3");
        List<Map<String, Object>> result = manager.programmingSubmissions(params);
        assertEquals(2, result.size());
        assertEquals(11L, result.get(0).get("submitId"));
        assertEquals("AC", result.get(0).get("result"));
        assertEquals("WA", result.get(1).get("result"));
    }

    @Test
    void legacySaveRejectsOtherUsersProblemsAndOrdinaryPractice() throws Exception {
        Map<String, Object> request = new HashMap<>();
        request.put("homeworkId", 1L);
        request.put("problemId", "P1000");
        request.put("submitId", 10L);
        for (Judge invalid : Arrays.asList(judge(10, 0).setUid("other-student"),
                judge(10, 0).setHomeworkId(null), judge(10, 0).setHomeworkId(9L),
                judge(10, 0).setDisplayPid("P1001"))) {
            when(judges.getById(10L)).thenReturn(invalid);
            assertThrows(StatusForbiddenException.class, () -> manager.saveProgramming(request));
        }
        verify(submissions, never()).insert(any());
        verify(submissions, never()).update(any(), any());
    }

    @Test
    void recordedCodeComesFromJudgeAndRepeatedRequestsKeepGrade() {
        HomeworkSubmit row = manager.recordProgrammingSubmission(judge(10, 5));
        assertEquals(10L, row.getSubmitId());
        assertTrue(row.getAnswer().contains("int main() {}"));
        verify(submissions).insert(row);
        row.setScore(20d).setIsScored(1).setJudgeResult("AC");
        when(submissions.selectOne(any())).thenReturn(row);
        assertSame(row, manager.recordProgrammingSubmission(judge(10, 5)));
        assertSame(row, manager.recordProgrammingSubmission(judge(9, 5)));
        assertEquals(20d, row.getScore());
        verify(submissions, never()).update(any(), any());
    }

    @Test
    void examRejectsBeforeStartAfterDeadlineAndAfterHandIn() throws Exception {
        homework.setIsExamMode(1).setExamDuration(1);
        assertThrows(StatusFailException.class, () -> manager.checkProgrammingSubmission(1L, "P1000"));
        HomeworkSubmit marker = new HomeworkSubmit().setExamStartTime(new Date(System.currentTimeMillis() - 30_000))
                .setExamEndTime(new Date(System.currentTimeMillis() + 30_000)).setIsOfficiallySubmitted(0);
        when(submissions.selectList(any())).thenReturn(Collections.singletonList(marker));
        assertDoesNotThrow(() -> manager.checkProgrammingSubmission(1L, "P1000"));
        marker.setExamStartTime(new Date(System.currentTimeMillis() - 120_000)).setExamEndTime(null);
        assertThrows(StatusFailException.class, () -> manager.checkProgrammingSubmission(1L, "P1000"));
        marker.setExamStartTime(new Date()).setIsOfficiallySubmitted(1);
        assertThrows(StatusFailException.class, () -> manager.checkProgrammingSubmission(1L, "P1000"));
    }

    @Test
    void submissionChecksStudentMembershipQuestionAndHomeworkWindow() {
        when(access.isStudent(2L, "student")).thenReturn(false);
        assertThrows(StatusForbiddenException.class, () -> manager.checkProgrammingSubmission(1L, "P1000"));
        when(access.isStudent(2L, "student")).thenReturn(true);
        when(questions.selectCount(any())).thenReturn(0);
        assertThrows(StatusNotFoundException.class, () -> manager.checkProgrammingSubmission(1L, "P1000"));
        when(questions.selectCount(any())).thenReturn(1);
        homework.setEndTime(new Date(System.currentTimeMillis() - 1));
        assertThrows(StatusFailException.class, () -> manager.checkProgrammingSubmission(1L, "P1000"));
    }

    @Test
    void teacherCannotGradeSubmissionFromAnotherHomework() {
        Map<String, Object> request = new HashMap<>();
        request.put("homeworkId", 1L);
        request.put("submissionId", 5L);
        request.put("score", 20d);
        when(submissions.selectById(5L)).thenReturn(submission(10).setHomeworkId(9L));
        assertThrows(StatusForbiddenException.class, () -> manager.grade(request));
        verify(submissions, never()).update(any(), any());
    }

    @Test
    void handInCannotReplaceProgrammingAnswerOrResetItsScore() throws Exception {
        Map<String, Object> request = new HashMap<>();
        request.put("homeworkId", 1L);
        request.put("answers", Collections.singletonMap("P1000", "forged answer"));
        assertEquals(true, manager.submit(request).get("submitted"));
        verify(submissions, never()).updateById(any());
        verify(submissions, never()).insert(any());
        verify(submissions).update(isNull(), argThat(wrapper -> !wrapper.getSqlSet().contains("score")));
    }
}
