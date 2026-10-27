package top.hcode.hoj.manager.oj;

import org.apache.shiro.subject.Subject;
import org.apache.shiro.util.ThreadContext;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.InOrder;
import org.springframework.aop.framework.ProxyFactory;
import org.springframework.test.util.ReflectionTestUtils;
import org.springframework.transaction.PlatformTransactionManager;
import org.springframework.transaction.annotation.AnnotationTransactionAttributeSource;
import org.springframework.transaction.interceptor.TransactionInterceptor;
import org.springframework.transaction.support.SimpleTransactionStatus;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.service.problem.ProblemPendingSaveGuard;
import top.hcode.hoj.dao.judge.JudgeEntityService;
import top.hcode.hoj.dao.problem.ProblemEntityService;
import top.hcode.hoj.manager.classroom.HomeworkManager;
import top.hcode.hoj.pojo.entity.judge.Judge;
import top.hcode.hoj.pojo.entity.problem.Problem;

import static org.junit.jupiter.api.Assertions.assertSame;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.*;

class BeforeDispatchClassroomTest {
    private final JudgeEntityService judges = mock(JudgeEntityService.class);
    private final ProblemEntityService problems = mock(ProblemEntityService.class);
    private final HomeworkManager homework = mock(HomeworkManager.class);
    private final TrainingManager training = mock(TrainingManager.class);
    private final PlatformTransactionManager transactions = mock(PlatformTransactionManager.class);
    private final SimpleTransactionStatus transaction = new SimpleTransactionStatus();
    private BeforeDispatchInitManager manager;
    private Judge judge;

    @BeforeEach
    void setup() {
        BeforeDispatchInitManager target = new BeforeDispatchInitManager();
        ReflectionTestUtils.setField(target, "judgeEntityService", judges);
        ReflectionTestUtils.setField(target, "problemEntityService", problems);
        ReflectionTestUtils.setField(target, "pendingSaveGuard", mock(ProblemPendingSaveGuard.class));
        ReflectionTestUtils.setField(target, "homeworkManager", homework);
        ReflectionTestUtils.setField(target, "trainingManager", training);
        ThreadContext.bind(mock(Subject.class));
        when(problems.getOne(any(), eq(false))).thenReturn(new Problem().setId(23L)
                .setProblemId("P23").setAuth(1).setIsGroup(false));
        when(transactions.getTransaction(any())).thenReturn(transaction);
        judge = new Judge().setSubmitId(12L).setHomeworkId(8L).setUid("student").setCid(0L);
        when(judges.save(judge)).thenReturn(true);
        ProxyFactory factory = new ProxyFactory(target);
        factory.addAdvice(new TransactionInterceptor(transactions, new AnnotationTransactionAttributeSource()));
        manager = (BeforeDispatchInitManager) factory.getProxy();
    }

    @AfterEach
    void cleanup() {
        ThreadContext.remove();
    }

    @Test
    void commitsBothRecordsBeforeReturningForDispatch() throws Exception {
        manager.initClassroomSubmission("P23", null, judge);
        InOrder order = inOrder(transactions, judges, homework);
        order.verify(transactions).getTransaction(any());
        order.verify(homework).checkProgrammingSubmission(8L, "P23");
        order.verify(judges).save(judge);
        order.verify(homework).recordProgrammingSubmission(judge);
        order.verify(transactions).commit(transaction);
        verify(transactions, never()).rollback(any());
        verifyNoInteractions(training);
    }

    @Test
    void recordFailureRollsBackJudgeAndPropagatesToPreventDispatch() throws Exception {
        IllegalStateException failure = new IllegalStateException("record failed");
        when(homework.recordProgrammingSubmission(judge)).thenThrow(failure);
        assertSame(failure, assertThrows(IllegalStateException.class,
                () -> manager.initClassroomSubmission("P23", null, judge)));
        InOrder order = inOrder(transactions, judges, homework);
        order.verify(transactions).getTransaction(any());
        order.verify(homework).checkProgrammingSubmission(8L, "P23");
        order.verify(judges).save(judge);
        order.verify(homework).recordProgrammingSubmission(judge);
        order.verify(transactions).rollback(transaction);
        verify(transactions, never()).commit(any());
        verifyNoInteractions(training);
    }

    @Test
    void examClosedBeforeTransactionDoesNotPersistAnotherSubmission() throws Exception {
        StatusFailException failure = new StatusFailException("考试已经交卷，不能继续提交代码");
        doThrow(failure).when(homework).checkProgrammingSubmission(8L, "P23");
        assertSame(failure, assertThrows(StatusFailException.class,
                () -> manager.initClassroomSubmission("P23", null, judge)));
        InOrder order = inOrder(transactions, homework);
        order.verify(transactions).getTransaction(any());
        order.verify(homework).checkProgrammingSubmission(8L, "P23");
        order.verify(transactions).rollback(transaction);
        verify(transactions, never()).commit(any());
        verify(homework, never()).recordProgrammingSubmission(any());
        verifyNoInteractions(judges, problems, training);
    }
}
