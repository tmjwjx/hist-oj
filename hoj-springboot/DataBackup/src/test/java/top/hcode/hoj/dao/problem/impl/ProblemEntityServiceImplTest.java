package top.hcode.hoj.dao.problem.impl;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.test.util.ReflectionTestUtils;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.dao.problem.CodeTemplateEntityService;
import top.hcode.hoj.dao.problem.ProblemCaseEntityService;
import top.hcode.hoj.dao.problem.ProblemLanguageEntityService;
import top.hcode.hoj.dao.problem.ProblemTagEntityService;
import top.hcode.hoj.dao.problem.TagEntityService;
import top.hcode.hoj.mapper.ProblemMapper;
import top.hcode.hoj.pojo.dto.ProblemDTO;
import top.hcode.hoj.pojo.entity.problem.Language;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.pojo.entity.problem.ProblemCase;
import top.hcode.hoj.pojo.entity.problem.ProblemLanguage;

import java.util.Collections;
import java.util.List;
import java.util.function.Consumer;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

class ProblemEntityServiceImplTest {
    private final ProblemEntityServiceImpl service = spy(new ProblemEntityServiceImpl());
    private final ProblemMapper problems = mock(ProblemMapper.class);
    private final ProblemCaseEntityService cases = mock(ProblemCaseEntityService.class);
    private final ProblemLanguageEntityService languages = mock(ProblemLanguageEntityService.class);
    private final ProblemTagEntityService problemTags = mock(ProblemTagEntityService.class);
    private final TagEntityService tags = mock(TagEntityService.class);
    private final CodeTemplateEntityService templates = mock(CodeTemplateEntityService.class);
    private final Problem saved = new Problem().setId(12L).setProblemId("P12")
            .setIsRemote(false).setIsGroup(true).setIsUploadCase(true)
            .setType(0).setJudgeMode("default").setJudgeCaseMode("default").setCaseVersion("verified-version");

    @BeforeEach
    void setUp() {
        ReflectionTestUtils.setField(service, "problemMapper", problems);
        ReflectionTestUtils.setField(service, "problemCaseEntityService", cases);
        ReflectionTestUtils.setField(service, "problemLanguageEntityService", languages);
        ReflectionTestUtils.setField(service, "problemTagEntityService", problemTags);
        ReflectionTestUtils.setField(service, "tagEntityService", tags);
        ReflectionTestUtils.setField(service, "codeTemplateEntityService", templates);
        when(problems.selectById(12L)).thenReturn(saved);
        when(problems.selectOne(any())).thenReturn(saved);
        when(problems.updateById(any())).thenReturn(1);
        when(cases.listByMap(anyMap())).thenReturn(Collections.singletonList(
                new ProblemCase().setId(31L).setPid(12L).setInput("1.in").setOutput("1.out")));
        when(cases.saveOrUpdateBatch(anyCollection())).thenReturn(true);
        when(languages.listByMap(anyMap())).thenReturn(Collections.singletonList(
                new ProblemLanguage().setPid(12L).setLid(1L)));
        when(problemTags.listByMap(anyMap())).thenReturn(Collections.emptyList());
        when(templates.listByMap(anyMap())).thenReturn(Collections.emptyList());
        // Prevent filesystem writes while observing which data representation is selected.
        doNothing().when(service).initUploadTestCase(anyString(), anyString(), anyString(),
                anyLong(), nullable(String.class), anyList());
        doNothing().when(service).initHandTestCase(anyString(), anyString(), anyString(), anyLong(), anyList());
        doNothing().when(service).initUploadTestCase(anyString(), anyString(), anyString(),
                anyLong(), nullable(String.class), anyList(), anyBoolean());
        doNothing().when(service).initHandTestCase(anyString(), anyString(), anyString(), anyLong(), anyList(), anyBoolean());
    }

    @Test
    void repeatedMetadataSavesIgnoreStaleClientVersionAndFalseChangeFlags() throws Exception {
        ProblemDTO request = unchangedDataRequest().setChangeModeCode(true).setChangeJudgeCaseMode(true);
        request.getProblem().setTitle("新标题").setDescription("新题面").setCaseVersion("stale-client-version");

        assertTrue(service.adminUpdateProblem(request));
        assertEquals("verified-version", request.getProblem().getCaseVersion());
        assertTrue(service.adminUpdateProblem(request));
        assertEquals("verified-version", request.getProblem().getCaseVersion());

        verifyNoTestDataInitialization();
        verify(cases, never()).saveOrUpdateBatch(anyCollection());
    }

    @Test
    void omittedClientVersionKeepsExistingVersion() throws Exception {
        ProblemDTO request = unchangedDataRequest();

        assertTrue(service.adminUpdateProblem(request));

        assertEquals("verified-version", request.getProblem().getCaseVersion());
        verifyNoTestDataInitialization();
    }

    @Test
    void actualJudgeModeChangeRefreshesMetadataEvenWhenClientClaimsUnchanged() throws Exception {
        ProblemDTO request = unchangedDataRequest().setChangeModeCode(false);
        request.getProblem().setJudgeMode("spj").setSpjLanguage("C++").setSpjCode("checker");

        assertTrue(service.adminUpdateProblem(request));

        assertNotEquals("verified-version", request.getProblem().getCaseVersion());
        assertEquals("spj", request.getJudgeMode());
        verify(service).initUploadTestCase(eq("spj"), eq("default"), anyString(),
                eq(12L), isNull(), eq(request.getSamples()), eq(false));
        verify(service, never()).initUploadTestCase(anyString(), anyString(), anyString(),
                anyLong(), nullable(String.class), anyList());
    }

    @Test
    void actualCaseModeChangeRefreshesMetadataEvenWhenClientClaimsUnchanged() throws Exception {
        ProblemDTO request = unchangedDataRequest().setChangeJudgeCaseMode(false);
        request.getProblem().setJudgeCaseMode("ergodic_without_error");

        assertTrue(service.adminUpdateProblem(request));

        assertNotEquals("verified-version", request.getProblem().getCaseVersion());
        verify(service).initUploadTestCase(eq("default"), eq("ergodic_without_error"), anyString(),
                eq(12L), isNull(), eq(request.getSamples()), eq(false));
    }

    @Test
    void newUploadedPackageWithSameFilenamesStillInvalidatesVerification() throws Exception {
        ProblemDTO request = unchangedDataRequest().setUploadTestcaseDir("/temporary/new-package");

        assertTrue(service.adminUpdateProblem(request));

        assertNotEquals("verified-version", request.getProblem().getCaseVersion());
        verify(service).initUploadTestCase(eq("default"), eq("default"), anyString(),
                eq(12L), eq("/temporary/new-package"), eq(request.getSamples()));
    }

    @Test
    void manualDataContentChangeUpdatesVersion() throws Exception {
        saved.setIsUploadCase(false);
        when(cases.listByMap(anyMap())).thenReturn(Collections.singletonList(
                new ProblemCase().setId(31L).setPid(12L).setInput("1 2\n").setOutput("3\n")));
        ProblemDTO request = request().setIsUploadTestCase(false);
        request.getProblem().setIsUploadCase(false);
        request.getSamples().get(0).setInput("2 3\n").setOutput("5\n");

        assertTrue(service.adminUpdateProblem(request));

        assertNotEquals("verified-version", request.getProblem().getCaseVersion());
        verify(service).initHandTestCase(eq("default"), eq("default"), anyString(),
                eq(12L), eq(request.getSamples()));
    }

    @Test
    void unchangedManualDataDoesNotResetVersion() throws Exception {
        saved.setIsUploadCase(false);
        ProblemDTO request = unchangedDataRequest().setIsUploadTestCase(false).setChangeModeCode(true);
        request.getProblem().setIsUploadCase(false);

        assertTrue(service.adminUpdateProblem(request));

        assertEquals("verified-version", request.getProblem().getCaseVersion());
        verifyNoTestDataInitialization();
    }

    @Test
    void missingFlagsKeepStoredGroupAndUploadedDataInsteadOfWritingFilenamesAsInput() throws Exception {
        ProblemDTO request = request();
        request.setIsUploadTestCase(null);
        request.getProblem().setIsUploadCase(null);

        assertTrue(service.adminUpdateProblem(request));

        assertEquals(Boolean.FALSE, request.getProblem().getIsRemote());
        assertEquals(Boolean.TRUE, request.getProblem().getIsGroup());
        assertEquals(Boolean.TRUE, request.getProblem().getIsUploadCase());
        assertEquals(Boolean.TRUE, request.getIsUploadTestCase());
        verify(service).initUploadTestCase(eq("default"), eq("default"), anyString(),
                eq(12L), isNull(), eq(request.getSamples()));
        verify(service, never()).initHandTestCase(anyString(), anyString(), anyString(), anyLong(), anyList());
        verify(problems).updateById(request.getProblem());
    }

    @Test
    void missingRemoteFlagPreservesRemoteProblemAndDoesNotGenerateLocalTestData() throws Exception {
        saved.setIsRemote(true).setProblemId("HDU-1000");
        ProblemDTO request = request();
        request.getProblem().setProblemId("HDU-1000");
        request.setSamples(null).setIsUploadTestCase(null);

        assertTrue(service.adminUpdateProblem(request));

        assertEquals(Boolean.TRUE, request.getProblem().getIsRemote());
        assertEquals(Boolean.TRUE, request.getProblem().getIsGroup());
        verify(cases, never()).saveOrUpdateBatch(anyCollection());
        verify(cases, never()).removeByIds(anyCollection());
        verify(service, never()).initUploadTestCase(anyString(), anyString(), anyString(),
                anyLong(), nullable(String.class), anyList());
        verify(service, never()).initHandTestCase(anyString(), anyString(), anyString(), anyLong(), anyList());
    }

    @Test
    void explicitManualModeStillOverridesStoredUploadMode() throws Exception {
        ProblemDTO request = request();
        request.setIsUploadTestCase(false);
        request.getProblem().setIsUploadCase(false);
        request.getSamples().get(0).setInput("1 2\n").setOutput("3\n");

        assertTrue(service.adminUpdateProblem(request));

        verify(service).initHandTestCase(eq("default"), eq("default"), anyString(),
                eq(12L), eq(request.getSamples()));
        verify(service, never()).initUploadTestCase(anyString(), anyString(), anyString(),
                anyLong(), nullable(String.class), anyList());
    }

    @Test
    void conflictingUploadModesRejectBeforeDeletingOrUpdatingData() {
        ProblemDTO request = request();
        request.getProblem().setIsUploadCase(false);

        StatusFailException error = assertThrows(StatusFailException.class,
                () -> service.adminUpdateProblem(request));

        assertTrue(error.getMessage().contains("上传方式不一致"));
        assertNoWrites();
    }

    @Test
    void missingUploadModeWithNoStoredValueRejectsInsteadOfGuessingManualMode() {
        saved.setIsUploadCase(null);
        ProblemDTO request = request();
        request.setIsUploadTestCase(null);
        request.getProblem().setIsUploadCase(null);

        StatusFailException error = assertThrows(StatusFailException.class,
                () -> service.adminUpdateProblem(request));

        assertTrue(error.getMessage().contains("上传方式缺失"));
        assertNoWrites();
    }

    @Test
    void missingAssociationListsOrTestDataRejectBeforeExistingDataCanBeCleared() {
        List<Consumer<ProblemDTO>> incompleteRequests = java.util.Arrays.asList(
                dto -> dto.setTags(null), dto -> dto.setCodeTemplates(null),
                dto -> dto.setLanguages(null), dto -> dto.setLanguages(Collections.emptyList()),
                dto -> dto.setSamples(null), dto -> dto.setSamples(Collections.emptyList()));
        for (Consumer<ProblemDTO> makeIncomplete : incompleteRequests) {
            ProblemDTO request = request();
            makeIncomplete.accept(request);
            assertThrows(StatusFailException.class, () -> service.adminUpdateProblem(request));
        }
        assertNoWrites();
    }

    @Test
    void missingProblemAndUnknownRemoteModeReturnExplicitFailuresBeforeAnyWrites() {
        when(problems.selectById(12L)).thenReturn(null);
        assertThrows(StatusFailException.class, () -> service.adminUpdateProblem(request()));
        when(problems.selectById(12L)).thenReturn(saved.setIsRemote(null));
        assertThrows(StatusFailException.class, () -> service.adminUpdateProblem(request()));
        assertNoWrites();
    }

    private void assertNoWrites() {
        verify(problems, never()).updateById(any());
        verifyNoInteractions(cases, languages, problemTags, tags, templates);
    }

    private void verifyNoTestDataInitialization() {
        verify(service, never()).initUploadTestCase(anyString(), anyString(), anyString(),
                anyLong(), nullable(String.class), anyList());
        verify(service, never()).initUploadTestCase(anyString(), anyString(), anyString(),
                anyLong(), nullable(String.class), anyList(), anyBoolean());
        verify(service, never()).initHandTestCase(anyString(), anyString(), anyString(), anyLong(), anyList());
        verify(service, never()).initHandTestCase(anyString(), anyString(), anyString(), anyLong(), anyList(), anyBoolean());
    }

    private ProblemDTO unchangedDataRequest() {
        ProblemDTO request = request();
        request.getSamples().get(0).setOutput("1.out");
        return request;
    }

    private ProblemDTO request() {
        return new ProblemDTO()
                .setProblem(new Problem().setId(12L).setProblemId("P12").setType(0)
                        .setJudgeMode("default").setJudgeCaseMode("default").setIsUploadCase(true))
                .setJudgeMode("default").setIsUploadTestCase(true)
                .setTags(Collections.emptyList()).setCodeTemplates(Collections.emptyList())
                .setLanguages(Collections.singletonList(new Language().setId(1L)))
                .setSamples(Collections.singletonList(new ProblemCase().setId(31L).setPid(12L)
                        .setInput("1.in").setOutput("updated.out")));
    }
}
