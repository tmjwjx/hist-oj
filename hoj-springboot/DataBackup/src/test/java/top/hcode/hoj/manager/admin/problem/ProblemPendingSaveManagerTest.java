package top.hcode.hoj.manager.admin.problem;

import com.fasterxml.jackson.databind.ObjectMapper;
import org.apache.shiro.subject.Subject;
import org.apache.shiro.util.ThreadContext;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.test.util.ReflectionTestUtils;
import org.springframework.transaction.support.TransactionSynchronizationManager;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.dao.judge.JudgeCaseEntityService;
import top.hcode.hoj.dao.judge.JudgeEntityService;
import top.hcode.hoj.dao.problem.*;
import top.hcode.hoj.mapper.ProblemPendingSaveMapper;
import top.hcode.hoj.pojo.dto.ProblemDTO;
import top.hcode.hoj.pojo.dto.ProblemPendingSaveDTO;
import top.hcode.hoj.pojo.entity.problem.*;
import top.hcode.hoj.pojo.vo.ProblemVerificationVO;
import top.hcode.hoj.service.problem.ProblemPendingSaveGuard;
import top.hcode.hoj.shiro.AccountProfile;
import top.hcode.hoj.validator.ProblemValidator;

import java.util.Collections;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

class ProblemPendingSaveManagerTest {
    private final ProblemPendingSaveManager manager = new ProblemPendingSaveManager();
    private final ProblemPendingSaveMapper pending = mock(ProblemPendingSaveMapper.class);
    private final ProblemEntityService problems = mock(ProblemEntityService.class);
    private final ProblemCaseEntityService cases = mock(ProblemCaseEntityService.class);
    private final CodeTemplateEntityService templates = mock(CodeTemplateEntityService.class);
    private final ProblemVerificationManager verification = mock(ProblemVerificationManager.class);
    private final ProblemVerificationEntityService records = mock(ProblemVerificationEntityService.class);
    private final JudgeEntityService judges = mock(JudgeEntityService.class);
    private final JudgeCaseEntityService judgeCases = mock(JudgeCaseEntityService.class);
    private final ProblemPendingSaveGuard guard = mock(ProblemPendingSaveGuard.class);
    private final ObjectMapper json = new ObjectMapper();
    private final Problem candidate = problem(999999991L, "PENDING-TEST");
    private final Problem original = problem(999999992L, "P1");

    @BeforeEach
    void setup() {
        AccountProfile profile = new AccountProfile();
        profile.setUid("owner"); profile.setUsername("author");
        Subject subject = mock(Subject.class);
        when(subject.getPrincipal()).thenReturn(profile);
        ThreadContext.bind(subject);
        ReflectionTestUtils.setField(manager, "pendingMapper", pending);
        ReflectionTestUtils.setField(manager, "problems", problems);
        ReflectionTestUtils.setField(manager, "cases", cases);
        ReflectionTestUtils.setField(manager, "templates", templates);
        ReflectionTestUtils.setField(manager, "languages", mock(ProblemLanguageEntityService.class));
        ReflectionTestUtils.setField(manager, "tags", mock(ProblemTagEntityService.class));
        ReflectionTestUtils.setField(manager, "verificationManager", verification);
        ReflectionTestUtils.setField(manager, "verifications", records);
        ReflectionTestUtils.setField(manager, "judges", judges);
        ReflectionTestUtils.setField(manager, "judgeCases", judgeCases);
        ReflectionTestUtils.setField(manager, "guard", guard);
        ReflectionTestUtils.setField(manager, "validator", mock(ProblemValidator.class));
        ReflectionTestUtils.setField(manager, "objectMapper", json);
        when(problems.getById(candidate.getId())).thenReturn(candidate);
        when(problems.getById(original.getId())).thenReturn(original);
        when(pending.lockProblem(candidate.getId())).thenReturn(candidate);
        when(pending.lockProblem(original.getId())).thenReturn(original);
    }

    @AfterEach
    void cleanup() {
        ThreadContext.remove();
        if (TransactionSynchronizationManager.isSynchronizationActive()) TransactionSynchronizationManager.clearSynchronization();
    }

    @Test
    void cannotCommitBeforeAcceptedVerificationAndNeverWritesFormalProblem() throws Exception {
        ProblemPendingSave row = row(null);
        row.setCandidateFingerprint(fingerprint(candidate.getId()));
        when(pending.lock(1L)).thenReturn(row);
        when(verification.getStatus(candidate.getId())).thenReturn(new ProblemVerificationVO().setVerified(false));

        StatusFailException error = assertThrows(StatusFailException.class, () -> manager.commit(1L));

        assertTrue(error.getMessage().contains("尚未通过"));
        verify(problems, never()).updateById(any());
        verify(problems, never()).adminUpdateProblem(any());
        verifyNoInteractions(judges, judgeCases);
        verify(pending, never()).updateById(any());
    }

    @Test
    void anotherUserCannotCommitEvenAnAlreadyPublishedRecord() throws Exception {
        ProblemPendingSave row = row(null).setUid("someone-else").setPublishedPid(99L);
        when(pending.lock(1L)).thenReturn(row);
        assertThrows(StatusForbiddenException.class, () -> manager.commit(1L));
        verifyNoInteractions(verification, judges, judgeCases);
        verify(problems, never()).updateById(any());
    }

    @Test
    void staleOriginalRevisionRejectsCommitBeforeWritingFilesOrProblems() throws Exception {
        ProblemPendingSave row = row(original.getId()).setBaseFingerprint("old-version");
        when(pending.lock(1L)).thenReturn(row);
        StatusFailException error = assertThrows(StatusFailException.class, () -> manager.commit(1L));
        assertTrue(error.getMessage().contains("原题目已被修改"));
        verifyNoInteractions(verification, judges, judgeCases);
        verify(problems, never()).adminUpdateProblem(any());
    }

    @Test
    void repeatedCommitReturnsSamePublishedProblemWithoutRepeatingWrites() throws Exception {
        when(pending.lock(1L)).thenReturn(row(null).setPublishedPid(42L));
        assertEquals(42L, manager.commit(1L).getPid());
        assertEquals(42L, manager.commit(1L).getPid());
        verifyNoInteractions(verification, judges, judgeCases);
        verify(problems, never()).updateById(any());
        verify(pending, never()).updateById(any());
    }

    @Test
    void prepareCopiesChildrenWithoutMovingOrMutatingOriginalRows() throws Exception {
        ProblemCase originalCase = new ProblemCase().setId(20L).setPid(original.getId()).setInput("1").setOutput("1");
        CodeTemplate originalTemplate = new CodeTemplate().setId(30).setPid(original.getId()).setLid(1L).setCode("code");
        when(cases.list(any())).thenReturn(Collections.singletonList(originalCase));
        when(templates.list(any())).thenReturn(Collections.singletonList(originalTemplate));
        ProblemDTO dto = new ProblemDTO().setProblem(original).setIsUploadTestCase(false)
                .setSamples(Collections.singletonList(originalCase)).setCodeTemplates(Collections.singletonList(originalTemplate))
                .setTags(Collections.emptyList()).setLanguages(Collections.singletonList(new Language().setId(1L)));
        ProblemPendingSaveDTO request = new ProblemPendingSaveDTO(); request.setProblemDTO(dto);
        when(guard.requiresVerification(any(), any())).thenReturn(true);
        when(problems.adminAddProblem(any())).thenAnswer(invocation -> {
            ProblemDTO copy = invocation.getArgument(0);
            assertNull(copy.getProblem().getId());
            assertNull(copy.getSamples().get(0).getId());
            assertNull(copy.getSamples().get(0).getPid());
            assertNull(copy.getCodeTemplates().get(0).getId());
            assertNull(copy.getCodeTemplates().get(0).getPid());
            assertEquals(2, copy.getProblem().getAuth());
            copy.getProblem().setId(candidate.getId());
            return true;
        });
        when(pending.insert(any())).thenAnswer(invocation -> {
            ProblemPendingSave row = invocation.getArgument(0); row.setId(1L); return 1;
        });
        TransactionSynchronizationManager.initSynchronization();

        assertTrue(manager.prepare(request).getRequiresVerification());

        assertEquals(20L, originalCase.getId());
        assertEquals(original.getId(), originalCase.getPid());
        assertEquals(30, originalTemplate.getId());
        assertEquals(original.getId(), originalTemplate.getPid());
        verify(problems, never()).adminUpdateProblem(any());
        verify(problems, never()).updateById(any());
    }

    private ProblemPendingSave row(Long targetPid) throws Exception {
        ProblemPendingSaveDTO request = new ProblemPendingSaveDTO();
        request.setProblemDTO(new ProblemDTO().setProblem(problem(targetPid, "NEW")));
        return new ProblemPendingSave().setId(1L).setUid("owner").setTargetPid(targetPid)
                .setCandidatePid(candidate.getId()).setSuperseded(false).setPayload(json.writeValueAsString(request));
    }

    private String fingerprint(Long pid) { return ReflectionTestUtils.invokeMethod(manager, "fingerprint", pid); }

    private static Problem problem(Long pid, String displayId) {
        return new Problem().setId(pid).setProblemId(displayId).setAuthor("author")
                .setIsRemote(false).setIsGroup(false).setAuth(1).setIsUploadCase(false)
                .setType(0).setTimeLimit(1000).setMemoryLimit(256).setStackLimit(128)
                .setJudgeMode("default").setJudgeCaseMode("default").setCaseVersion("version");
    }
}
