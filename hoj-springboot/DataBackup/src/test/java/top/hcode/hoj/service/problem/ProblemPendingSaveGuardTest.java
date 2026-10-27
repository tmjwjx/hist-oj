package top.hcode.hoj.service.problem;

import com.baomidou.mybatisplus.core.conditions.Wrapper;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.test.util.ReflectionTestUtils;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.dao.problem.ProblemCaseEntityService;
import top.hcode.hoj.pojo.dto.ProblemDTO;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.pojo.entity.problem.ProblemCase;

import java.util.Collections;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

class ProblemPendingSaveGuardTest {
    private final ProblemPendingSaveGuard guard = new ProblemPendingSaveGuard();
    private final JdbcTemplate jdbc = mock(JdbcTemplate.class);
    private final ProblemCaseEntityService cases = mock(ProblemCaseEntityService.class);

    @BeforeEach
    void setup() {
        ReflectionTestUtils.setField(guard, "jdbcTemplate", jdbc);
        ReflectionTestUtils.setField(guard, "cases", cases);
        when(cases.list(any(Wrapper.class))).thenReturn(Collections.singletonList(testCase()));
    }

    @Test
    void newLocalProblemCannotBeFinalSavedWithoutVerification() throws Exception {
        assertTrue(guard.requiresVerification(dto(), null));
        verifyNoInteractions(cases);
    }

    @Test
    void acceptedMetadataAndExplanationEditCanSaveDespiteStaleClientFlags() throws Exception {
        ProblemDTO dto = dto();
        dto.getProblem().setTitle("new title").setDescription("new statement").setCaseVersion("forged")
                .setExamples("<input>1 2\n</input><output>3\n</output><explanation>new explanation</explanation>");
        dto.setChangeModeCode(true).setChangeJudgeCaseMode(true);
        assertFalse(guard.requiresVerification(dto, problem()));
    }

    @Test
    void constraintsCasesAndPackagesMustBeVerifiedBeforeSave() throws Exception {
        ProblemDTO dto = dto(); dto.getProblem().setTimeLimit(2000);
        assertTrue(guard.requiresVerification(dto, problem()));
        dto = dto(); dto.getSamples().get(0).setOutput("4\n");
        assertTrue(guard.requiresVerification(dto, problem()));
        dto = dto().setUploadTestcaseDir("/hoj/file/zip/new");
        assertTrue(guard.requiresVerification(dto, problem()));
        dto = dto(); dto.getSamples().get(0).setId(null);
        assertTrue(guard.requiresVerification(dto, problem()));
    }

    @Test
    void existingFormalMetadataCanSaveWithoutHistoricalVerification() throws Exception {
        ProblemDTO dto = dto();
        dto.getProblem().setDescription("edited description only");
        assertFalse(guard.requiresVerification(dto, problem()));
        // An unchanged resave must not create another candidate either.
        assertFalse(guard.requiresVerification(dto(), problem()));
        // The same existing problem still needs a candidate for changed judging inputs.
        dto.getProblem().setMemoryLimit(512);
        assertTrue(guard.requiresVerification(dto, problem()));
    }

    @Test
    void remoteFlagCannotBypassLocalVerification() {
        ProblemDTO dto = dto(); dto.getProblem().setIsRemote(true);
        assertThrows(StatusFailException.class, () -> guard.requiresVerification(dto, problem()));
    }

    @Test
    void candidateCannotUseFormalSaveAndIncompleteDataCannotEraseExistingCases() {
        when(jdbc.queryForObject(anyString(), eq(Integer.class), eq(12L))).thenReturn(1);
        assertThrows(StatusFailException.class, () -> guard.requiresVerification(dto(), problem()));
        when(jdbc.queryForObject(anyString(), eq(Integer.class), eq(12L))).thenReturn(0);
        ProblemDTO dto = dto().setSamples(null);
        assertThrows(StatusFailException.class, () -> guard.requiresVerification(dto, problem()));
    }

    private ProblemDTO dto() {
        return new ProblemDTO().setProblem(problem()).setIsUploadTestCase(false)
                .setSamples(Collections.singletonList(testCase()));
    }

    private Problem problem() {
        return new Problem().setId(12L).setIsRemote(false).setIsGroup(false).setIsUploadCase(false)
                .setType(0).setCaseVersion("v1").setJudgeMode("default").setJudgeCaseMode("default")
                .setTimeLimit(1000).setMemoryLimit(256).setStackLimit(128)
                .setExamples("<input>1 2\n</input><output>3\n</output>")
                .setIsFileIO(false).setIsRemoveEndBlank(false);
    }

    private ProblemCase testCase() {
        return new ProblemCase().setId(11L).setPid(12L).setInput("217 309\n").setOutput("526\n");
    }
}
