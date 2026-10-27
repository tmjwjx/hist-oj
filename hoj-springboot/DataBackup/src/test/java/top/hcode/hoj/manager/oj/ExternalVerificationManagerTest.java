package top.hcode.hoj.manager.oj;

import cn.hutool.crypto.digest.DigestUtil;
import com.baomidou.mybatisplus.core.conditions.Wrapper;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.apache.shiro.subject.Subject;
import org.apache.shiro.util.ThreadContext;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.ArgumentCaptor;
import org.springframework.test.util.ReflectionTestUtils;
import org.springframework.transaction.PlatformTransactionManager;
import org.springframework.transaction.support.TransactionSynchronization;
import org.springframework.transaction.support.TransactionSynchronizationManager;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.service.problem.ProblemPendingSaveGuard;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.dao.judge.JudgeCaseEntityService;
import top.hcode.hoj.dao.judge.JudgeEntityService;
import top.hcode.hoj.dao.problem.*;
import top.hcode.hoj.judge.self.JudgeDispatcher;
import top.hcode.hoj.mapper.ExternalVerificationSubmissionMapper;
import top.hcode.hoj.mapper.ProblemExternalVerificationMapper;
import top.hcode.hoj.pojo.dto.ExternalVerificationSubmitDTO;
import top.hcode.hoj.pojo.dto.ExternalVerificationSettingDTO;
import top.hcode.hoj.pojo.entity.judge.ExternalVerificationSubmission;
import top.hcode.hoj.pojo.entity.judge.Judge;
import top.hcode.hoj.pojo.entity.judge.JudgeCase;
import top.hcode.hoj.pojo.entity.problem.*;
import top.hcode.hoj.pojo.vo.ExternalVerificationSubmissionVO;
import top.hcode.hoj.service.problem.JudgeProgressService;
import top.hcode.hoj.shiro.AccountProfile;
import top.hcode.hoj.utils.ProblemVerificationConstants;

import java.util.Collections;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

class ExternalVerificationManagerTest {
    private static final String TOKEN = String.join("", Collections.nCopies(64, "a"));
    private static final String VISITOR = String.join("", Collections.nCopies(64, "b"));
    private final ExternalVerificationManager manager = new ExternalVerificationManager();
    private final ProblemExternalVerificationMapper links = mock(ProblemExternalVerificationMapper.class);
    private final ExternalVerificationSubmissionMapper submissions = mock(ExternalVerificationSubmissionMapper.class);
    private final ProblemEntityService problems = mock(ProblemEntityService.class);
    private final ProblemVerificationEntityService verifications = mock(ProblemVerificationEntityService.class);
    private final ProblemCaseEntityService cases = mock(ProblemCaseEntityService.class);
    private final ProblemLanguageEntityService problemLanguages = mock(ProblemLanguageEntityService.class);
    private final LanguageEntityService languages = mock(LanguageEntityService.class);
    private final JudgeEntityService judges = mock(JudgeEntityService.class);
    private final JudgeCaseEntityService judgeCases = mock(JudgeCaseEntityService.class);
    private final JudgeDispatcher dispatcher = mock(JudgeDispatcher.class);
    private final JudgeProgressService progress = mock(JudgeProgressService.class);
    private final ExternalVerificationRateLimiter limiter = mock(ExternalVerificationRateLimiter.class);
    private final Subject subject = mock(Subject.class);
    private final Problem problem = new Problem().setId(10L).setProblemId("P10").setTitle("Example")
            .setAuthor("author").setIsRemote(false).setCaseVersion("v1").setSpjCode("secret standard");
    private final ProblemExternalVerification link = new ProblemExternalVerification().setPid(10L)
            .setToken(TOKEN).setEnabled(true).setOwnerUid("owner").setOwnerUsername("author");

    @BeforeEach
    void setup() {
        ReflectionTestUtils.setField(manager, "linkMapper", links);
        ReflectionTestUtils.setField(manager, "submissionMapper", submissions);
        ReflectionTestUtils.setField(manager, "problemService", problems);
        ReflectionTestUtils.setField(manager, "pendingSaveGuard", mock(ProblemPendingSaveGuard.class));
        ReflectionTestUtils.setField(manager, "verificationService", verifications);
        ReflectionTestUtils.setField(manager, "problemCaseService", cases);
        ReflectionTestUtils.setField(manager, "problemLanguageService", problemLanguages);
        ReflectionTestUtils.setField(manager, "languageService", languages);
        ReflectionTestUtils.setField(manager, "judgeService", judges);
        ReflectionTestUtils.setField(manager, "judgeCaseService", judgeCases);
        ReflectionTestUtils.setField(manager, "judgeDispatcher", dispatcher);
        ReflectionTestUtils.setField(manager, "progressService", progress);
        ReflectionTestUtils.setField(manager, "rateLimiter", limiter);
        ReflectionTestUtils.setField(manager, "transactionManager", mock(PlatformTransactionManager.class));
        ThreadContext.bind(subject);
        when(problems.getById(10L)).thenReturn(problem);
        when(links.selectOne(any())).thenReturn(link);
        when(links.lockByToken(TOKEN)).thenReturn(link);
        when(progress.resolve(anyLong(), any(), isNull())).thenReturn(new JudgeProgressService.Progress("Accepted", null, "secret diagnostic"));
        when(problemLanguages.list(any(Wrapper.class))).thenReturn(Collections.singletonList(new ProblemLanguage().setLid(1L)));
        when(languages.list(any(Wrapper.class))).thenReturn(Collections.singletonList(new Language().setName("GNU G++17").setContentType("text/x-c++src")));
        when(verifications.getOne(any(Wrapper.class), eq(false))).thenReturn(new ProblemVerification()
                .setCaseVersion("v1").setSyncStatus(ProblemVerificationConstants.SYNC_SUCCESS));
        when(cases.count(any(Wrapper.class))).thenReturn(1);
    }

    @AfterEach
    void cleanup() {
        ThreadContext.unbindSubject();
        if (TransactionSynchronizationManager.isSynchronizationActive()) TransactionSynchronizationManager.clearSynchronization();
    }

    @Test
    void everyAnonymousEndpointRejectsDisabledLinkBeforeReadingOrSavingAnything() {
        link.setEnabled(false);
        assertThrows(StatusForbiddenException.class, () -> manager.problem(TOKEN));
        assertThrows(StatusForbiddenException.class, () -> manager.submissions(TOKEN, VISITOR, 1, 20));
        assertThrows(StatusForbiddenException.class, () -> manager.submission(TOKEN, 77L, VISITOR));
        assertThrows(StatusForbiddenException.class, () -> manager.submit(TOKEN, dto(), "203.0.113.1"));
        verifyNoInteractions(submissions, judges, dispatcher, limiter);
    }

    @Test
    void malformedLinkCannotReachDatabase() {
        assertThrows(StatusForbiddenException.class, () -> manager.problem("invalid"));
        verifyNoInteractions(links);
    }

    @Test
    void ordinaryAdministratorCannotReadOrChangeAnotherAuthorsExternalVerification() throws Exception {
        AccountProfile profile = new AccountProfile();
        profile.setUsername("someone-else");
        when(subject.getPrincipal()).thenReturn(profile);
        when(subject.hasRole("admin")).thenReturn(true);
        ExternalVerificationSettingDTO setting = new ExternalVerificationSettingDTO();
        setting.setPid(10L); setting.setEnabled(true);
        assertThrows(StatusForbiddenException.class, () -> manager.setting(10L));
        assertThrows(StatusForbiddenException.class, () -> manager.saveSetting(setting));
        setting.setEnabled(false);
        assertThrows(StatusForbiddenException.class, () -> manager.saveSetting(setting));
        assertThrows(StatusForbiddenException.class, () -> manager.adminSubmissions(10L, 1, 20));
        assertThrows(StatusForbiddenException.class, () -> manager.adminSubmission(10L, 77L));
        verifyNoInteractions(links, submissions, judges, judgeCases);
        profile.setUsername("author");
        assertFalse(manager.setting(10L).getEnabled());
        assertNull(manager.setting(10L).getToken());
        verify(links, never()).saveSetting(any());
        when(links.selectById(10L)).thenReturn(link);
        assertEquals(TOKEN, manager.saveSetting(setting).getToken());
        verify(links).saveSetting(any());
        when(submissions.submissions(any(), eq(10L), isNull())).thenReturn(new Page<>());
        assertNotNull(manager.adminSubmissions(10L, 1, 20));
        profile.setUsername("someone-else");
        when(subject.hasRole("problem_admin")).thenReturn(true);
        assertNotNull(manager.setting(10L));
        assertNotNull(manager.adminSubmissions(10L, 1, 20));
        when(subject.hasRole("problem_admin")).thenReturn(false);
        when(subject.hasRole("root")).thenReturn(true);
        assertNotNull(manager.setting(10L));
        assertNotNull(manager.adminSubmissions(10L, 1, 20));
    }

    @Test
    void authorizedProblemCannotBeUsedToReadAnotherProblemsCode() {
        AccountProfile profile = new AccountProfile();
        profile.setUsername("author");
        when(subject.getPrincipal()).thenReturn(profile);
        when(subject.hasRole("admin")).thenReturn(true);
        when(submissions.selectById(77L)).thenReturn(new ExternalVerificationSubmission().setPid(11L));
        assertThrows(StatusForbiddenException.class, () -> manager.adminSubmission(10L, 77L));
        verifyNoInteractions(judges, judgeCases);
    }

    @Test
    void visitorCannotReadAnotherVisitorOrProblemSubmission() {
        when(submissions.selectById(77L)).thenReturn(new ExternalVerificationSubmission().setPid(10L)
                .setVisitorTokenHash(DigestUtil.sha256Hex("different visitor")));
        assertThrows(StatusForbiddenException.class, () -> manager.submission(TOKEN, 77L, VISITOR));
        when(submissions.selectById(77L)).thenReturn(new ExternalVerificationSubmission().setPid(11L)
                .setVisitorTokenHash(DigestUtil.sha256Hex(VISITOR)));
        assertThrows(StatusForbiddenException.class, () -> manager.submission(TOKEN, 77L, VISITOR));
        verifyNoInteractions(judges, judgeCases);
    }

    @Test
    void visitorListUsesHashedIdentityAndBoundedPagination() throws Exception {
        when(submissions.submissions(any(), eq(10L), eq(DigestUtil.sha256Hex(VISITOR)))).thenReturn(new Page<>());
        assertNotNull(manager.submissions(TOKEN, VISITOR, 1, 20));
        verify(submissions).submissions(any(), eq(10L), eq(DigestUtil.sha256Hex(VISITOR)));
        assertThrows(StatusFailException.class, () -> manager.submissions(TOKEN, VISITOR, 0, 20));
        assertThrows(StatusFailException.class, () -> manager.submissions(TOKEN, VISITOR, 1, 101));
        assertThrows(StatusFailException.class, () -> manager.submissions(TOKEN, "bad", 1, 20));
    }

    @Test
    void statementAndSubmissionProjectionExcludeSecretsButRetainCompilationErrors() throws Exception {
        problem.setIsFileIO(true).setIoReadFileName("sum.in").setIoWriteFileName("sum.out");
        String statement = new ObjectMapper().writeValueAsString(manager.problem(TOKEN));
        com.fasterxml.jackson.databind.JsonNode publicProblem = new ObjectMapper().readTree(statement).get("problem");
        assertTrue(publicProblem.get("isFileIO").asBoolean());
        assertEquals("sum.in", publicProblem.get("ioReadFileName").asText());
        assertEquals("sum.out", publicProblem.get("ioWriteFileName").asText());
        assertFalse(statement.contains("secret standard"));
        assertFalse(statement.contains("spjCode"));
        assertFalse(statement.contains("author"));
        when(submissions.selectById(77L)).thenReturn(new ExternalVerificationSubmission().setPid(10L)
                .setNickname("Visitor").setVisitorTokenHash(DigestUtil.sha256Hex(VISITOR)));
        Judge judge = new Judge().setSubmitId(77L).setPid(10L).setUid("secret owner").setUsername("secret author")
                .setCode("submitted code").setStatus(10).setErrorMessage("secret server path")
                .setIsProblemVerification(true).setSubmissionType(ProblemVerificationConstants.EXTERNAL_VERIFICATION);
        when(judges.getById(77L)).thenReturn(judge);
        when(judgeCases.list(any(Wrapper.class))).thenReturn(Collections.singletonList(new JudgeCase()
                .setSeq(1).setStatus(10).setInputData("secret input").setOutputData("secret answer")
                .setUserOutput("secret user output").setStderr("secret stderr")));
        ExternalVerificationSubmissionVO result = manager.submission(TOKEN, 77L, VISITOR);
        assertEquals("submitted code", result.getCode());
        String serialized = new ObjectMapper().writeValueAsString(result);
        assertFalse(serialized.contains("secret"));
        assertFalse(serialized.contains("inputData"));
        assertFalse(serialized.contains("stderr"));
        judge.setStatus(-2).setErrorMessage("compiler error");
        assertEquals("compiler error", manager.submission(TOKEN, 77L, VISITOR).getErrorMessage());
        when(subject.hasRole("root")).thenReturn(true);
        when(subject.getPrincipal()).thenReturn(new AccountProfile());
        assertEquals("secret stderr", manager.adminSubmission(10L, 77L).getCaseResults().get(0).getStderr());
    }

    @Test
    void rejectsRemoteUnsupportedLanguageAndUnsynchronizedDataWithoutEnqueueing() {
        problem.setIsRemote(true);
        assertThrows(StatusFailException.class, () -> manager.submit(TOKEN, dto(), "203.0.113.1"));
        problem.setIsRemote(false);
        ExternalVerificationSubmitDTO unsupported = dto(); unsupported.setLanguage("unsupported");
        assertThrows(StatusFailException.class, () -> manager.submit(TOKEN, unsupported, "203.0.113.1"));
        problem.setCaseVersion("v2");
        assertThrows(StatusFailException.class, () -> manager.submit(TOKEN, dto(), "203.0.113.1"));
        verifyNoInteractions(judges, dispatcher, limiter);
    }

    @Test
    void rejectedRateLimitCannotSaveOrDispatch() throws Exception {
        doThrow(new StatusFailException("limited")).when(limiter).checkSubmission(anyString(), anyLong());
        assertThrows(StatusFailException.class, () -> manager.submit(TOKEN, dto(), "203.0.113.1"));
        verifyNoInteractions(judges, submissions, dispatcher);
    }

    @Test
    void submissionBindsVisitorBeforeDispatchAndNeverChangesAuthoritativeVerification() throws Exception {
        when(judges.save(any(Judge.class))).thenAnswer(call -> { call.<Judge>getArgument(0).setSubmitId(77L); return true; });
        when(submissions.insert(any())).thenReturn(1);
        TransactionSynchronizationManager.initSynchronization();
        ExternalVerificationSubmissionVO result = manager.submit(TOKEN, dto(), "203.0.113.1");
        assertEquals(77L, result.getSubmitId());
        ArgumentCaptor<Judge> saved = ArgumentCaptor.forClass(Judge.class);
        verify(judges).save(saved.capture());
        assertTrue(saved.getValue().getIsProblemVerification());
        assertEquals(ProblemVerificationConstants.EXTERNAL_VERIFICATION, saved.getValue().getSubmissionType());
        assertEquals("owner", saved.getValue().getUid());
        assertFalse(saved.getValue().getShare());
        ArgumentCaptor<ExternalVerificationSubmission> visitor = ArgumentCaptor.forClass(ExternalVerificationSubmission.class);
        verify(submissions).insert(visitor.capture());
        assertEquals(DigestUtil.sha256Hex(VISITOR), visitor.getValue().getVisitorTokenHash());
        verifyNoInteractions(dispatcher);
        for (TransactionSynchronization synchronization : TransactionSynchronizationManager.getSynchronizations()) synchronization.afterCommit();
        verify(dispatcher).sendTask(77L, 10L, false);
        verify(verifications, never()).updateById(any());
    }

    @Test
    void failedOwnershipInsertDoesNotDispatch() {
        when(judges.save(any(Judge.class))).thenAnswer(call -> { call.<Judge>getArgument(0).setSubmitId(77L); return true; });
        when(submissions.insert(any())).thenReturn(0);
        assertThrows(StatusFailException.class, () -> manager.submit(TOKEN, dto(), "203.0.113.1"));
        verifyNoInteractions(dispatcher);
    }

    private ExternalVerificationSubmitDTO dto() {
        ExternalVerificationSubmitDTO dto = new ExternalVerificationSubmitDTO();
        dto.setVisitorToken(VISITOR); dto.setLanguage("GNU G++17"); dto.setCode("int main() {}");
        return dto;
    }
}
