package top.hcode.hoj.manager.admin.problem;

import cn.hutool.json.JSONUtil;
import org.apache.shiro.subject.Subject;
import org.apache.shiro.util.ThreadContext;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.ArgumentCaptor;
import org.springframework.test.util.ReflectionTestUtils;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.dao.problem.ProblemAiConfigEntityService;
import top.hcode.hoj.pojo.entity.problem.ProblemAiConfig;
import top.hcode.hoj.pojo.entity.problem.ProblemAiModelConfig;
import top.hcode.hoj.pojo.vo.ProblemAiConfigVO;

import java.util.Arrays;
import java.util.Collections;
import java.util.List;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

class ProblemAiConfigManagerTest {
    private static final String PRIMARY_ID = "aaaaaaaa-1111-4111-8111-111111111111";
    private static final String SECONDARY_ID = "bbbbbbbb-2222-4222-8222-222222222222";
    private static final String LEGACY_ID = "00000000-0000-4000-8000-000000000001";
    private final ProblemAiConfigEntityService service = mock(ProblemAiConfigEntityService.class);
    private final Subject subject = mock(Subject.class);
    private final ProblemAiConfigManager manager = new ProblemAiConfigManager();

    @BeforeEach
    void setup() {
        ReflectionTestUtils.setField(manager, "configService", service);
        ReflectionTestUtils.setField(manager, "promptFactory", new ProblemAiPromptFactory());
        ThreadContext.bind(subject);
        when(subject.isAuthenticated()).thenReturn(true);
        when(subject.hasRole("admin")).thenReturn(true);
        when(service.saveOrUpdate(any(ProblemAiConfig.class))).thenReturn(true);
    }

    @AfterEach
    void cleanup() {
        ThreadContext.remove();
    }

    @Test
    void newDisabledConfigIsEditableAndDoesNotWriteOnRead() throws Exception {
        ProblemAiConfigVO config = manager.getConfig();
        assertFalse(config.getEnabled());
        assertEquals(1, config.getModels().size());
        assertEquals(LEGACY_ID, config.getModels().get(0).getId());
        assertEquals("openai_chat", config.getProtocol());
        assertEquals("openai_chat", config.getModels().get(0).getProtocol());
        assertTrue(config.getModels().get(0).getPrimary());
        assertFalse(config.getApiKeyConfigured());
        assertEquals("", config.getApiKey());
        assertEquals(Integer.valueOf(120), config.getTimeoutSeconds());
        assertThrows(StatusFailException.class, manager::requireConfig);
        verify(service, never()).saveOrUpdate(any());
    }

    @Test
    void legacyConfigRetainsStableIdentityAndNeverReturnsSecret() throws Exception {
        ProblemAiConfig stored = legacy();
        when(service.getById(1L)).thenReturn(stored);
        ProblemAiConfigVO config = manager.getConfig();
        assertEquals(LEGACY_ID, config.getModels().get(0).getId());
        assertEquals("gpt-6-astra", config.getModel());
        assertEquals("openai_chat", config.getProtocol());
        assertEquals("", config.getApiKey());
        assertEquals("", config.getModels().get(0).getApiKey());
        assertTrue(config.getApiKeyConfigured());
        assertTrue(config.getModels().get(0).getApiKeyConfigured());
        assertFalse(JSONUtil.toJsonStr(config).contains("legacy-secret"));
        assertEquals("legacy-secret", manager.requireConfig().getApiKey());
        assertEquals("legacy-secret", stored.getApiKey());
        assertNull(stored.getModels());
    }

    @Test
    void persistsModelListAndMirrorsExactlyThePrimaryForLegacyCallers() throws Exception {
        ProblemAiConfig input = multi();
        input.getModels().get(0).setPrimary(false);
        input.getModels().get(1).setPrimary(true);
        manager.saveConfig(input);
        ProblemAiConfig saved = saved();
        assertEquals(Long.valueOf(1), saved.getId());
        assertEquals("secondary-secret", saved.getApiKey());
        assertEquals("secondary-model", saved.getModel());
        assertEquals("https://secondary.example/v1", saved.getApiUrl());
        assertEquals(2, JSONUtil.parseArray(saved.getModelsJson()).size());
        when(service.getById(1L)).thenReturn(saved);
        ProblemAiConfigVO returned = manager.getConfig();
        assertEquals(2, returned.getModels().size());
        assertFalse(JSONUtil.toJsonStr(returned).contains("-secret"));
        assertTrue(returned.getModels().stream().allMatch(model -> model.getApiKey().isEmpty()
                && Boolean.TRUE.equals(model.getApiKeyConfigured())));
        assertEquals("primary-secret", input.getModels().get(0).getApiKey());
    }

    @Test
    void blankKeysAreRetainedByModelIdEvenAfterReordering() throws Exception {
        store(multi());
        ProblemAiConfig input = multi();
        input.getModels().get(0).setApiKey("  ").setId(PRIMARY_ID.toUpperCase());
        input.getModels().get(1).setApiKey(null);
        Collections.reverse(input.getModels());
        manager.saveConfig(input);
        List<ProblemAiModelConfig> models = saved().getModels();
        assertEquals(SECONDARY_ID, models.get(0).getId());
        assertEquals("secondary-secret", models.get(0).getApiKey());
        assertEquals(PRIMARY_ID, models.get(1).getId());
        assertEquals("primary-secret", models.get(1).getApiKey());
    }

    @Test
    void persistedModelsWithoutProtocolsDefaultToChatCompletions() throws Exception {
        store(multi());
        ProblemAiConfig config = manager.requireConfig();
        assertEquals("openai_chat", config.getProtocol());
        assertTrue(config.getModels().stream().allMatch(model -> "openai_chat".equals(model.getProtocol())));
        verify(service, never()).saveOrUpdate(any());
    }

    @Test
    void supportedProtocolsPersistAndPrimaryIsMirrored() throws Exception {
        for (String protocol : Arrays.asList("openai_chat", "openai_responses", "gemini", "anthropic")) {
            clearInvocations(service);
            ProblemAiConfig input = multi();
            input.getModels().get(0).setProtocol(protocol);
            input.getModels().get(1).setProtocol("anthropic");
            manager.saveConfig(input);
            ProblemAiConfig saved = saved();
            assertEquals(protocol, saved.getProtocol());
            assertEquals(protocol, JSONUtil.parseArray(saved.getModelsJson()).getJSONObject(0).getStr("protocol"));
            when(service.getById(1L)).thenReturn(saved);
            ProblemAiConfigVO returned = manager.getConfig();
            assertEquals(protocol, returned.getProtocol());
            assertEquals(protocol, returned.getModels().get(0).getProtocol());
            assertEquals("anthropic", returned.getModels().get(1).getProtocol());
        }
    }

    @Test
    void omittedProtocolsStayWithTheirModelIdentityWhenOldClientReordersModels() throws Exception {
        ProblemAiConfig stored = multi();
        stored.getModels().get(0).setProtocol("gemini");
        stored.getModels().get(1).setProtocol("anthropic");
        store(stored);
        ProblemAiConfig input = multi();
        input.getModels().get(0).setId(PRIMARY_ID.toUpperCase());
        Collections.reverse(input.getModels());
        manager.saveConfig(input);
        ProblemAiConfig saved = saved();
        assertEquals("anthropic", saved.getModels().get(0).getProtocol());
        assertEquals("gemini", saved.getModels().get(1).getProtocol());
        assertEquals("gemini", saved.getProtocol());
        assertNull(input.getModels().get(0).getProtocol());
    }

    @Test
    void legacySingleModelClientDoesNotResetNativeProtocol() throws Exception {
        ProblemAiConfig stored = multi();
        stored.setModels(Collections.singletonList(stored.getModels().get(0).setProtocol("gemini")));
        store(stored);
        manager.saveConfig(legacy().setApiKey(""));
        assertEquals("gemini", saved().getProtocol());
    }

    @Test
    void newModelDefaultsToChatAndExplicitChangeOverridesStoredProtocol() throws Exception {
        ProblemAiConfig stored = multi();
        stored.getModels().get(0).setProtocol("gemini");
        stored.getModels().get(1).setProtocol("anthropic");
        store(stored);
        ProblemAiConfig input = multi();
        input.getModels().get(0).setProtocol("openai_responses");
        input.getModels().get(1).setId("cccccccc-3333-4333-8333-333333333333");
        manager.saveConfig(input);
        assertEquals("openai_responses", saved().getProtocol());
        assertEquals("openai_chat", saved().getModels().get(1).getProtocol());
    }

    @Test
    void rejectsUnknownAndExplicitlyEmptyProtocolsWithoutPersisting() {
        for (String protocol : Arrays.asList("unknown", "", " ", "OpenAI", "gemini\nprivate-value")) {
            ProblemAiConfig input = multi();
            input.getModels().get(1).setProtocol(protocol);
            StatusFailException error = assertThrows(StatusFailException.class, () -> manager.saveConfig(input));
            assertFalse(error.getMessage().contains("private-value"));
        }
        verify(service, never()).saveOrUpdate(any());
    }

    @Test
    void rejectsUnknownStoredProtocolBeforeCallingProvider() {
        ProblemAiConfig stored = multi();
        stored.getModels().get(0).setProtocol("unsupported");
        store(stored);
        assertThrows(StatusFailException.class, manager::requireConfig);
    }

    @Test
    void newModelCannotInheritAnotherModelsKey() {
        store(multi());
        ProblemAiConfig input = multi();
        input.getModels().get(1).setId("cccccccc-3333-4333-8333-333333333333").setApiKey("");
        assertThrows(StatusFailException.class, () -> manager.saveConfig(input));
        verify(service, never()).saveOrUpdate(any());
    }

    @Test
    void legacyClientPreservesMaskedKeyButCannotDeleteMultipleModels() throws Exception {
        when(service.getById(1L)).thenReturn(legacy());
        manager.saveConfig(legacy().setApiKey("***已配置***"));
        assertEquals("legacy-secret", saved().getApiKey());
        store(multi());
        clearInvocations(service);
        assertThrows(StatusFailException.class, () -> manager.saveConfig(legacy()));
        verify(service, never()).saveOrUpdate(any());
    }

    @Test
    void primarySelectionIsMandatoryAndNullMeansOnlyPrimary() throws Exception {
        ProblemAiConfig config = multi();
        List<ProblemAiModelConfig> selected = manager.selectModels(config, null);
        assertEquals(1, selected.size());
        assertEquals(PRIMARY_ID, selected.get(0).getId());
        selected.get(0).setApiKey("changed");
        assertEquals("primary-secret", config.getModels().get(0).getApiKey());
        assertEquals(2, manager.selectModels(config, Arrays.asList(SECONDARY_ID, PRIMARY_ID)).size());
        assertThrows(StatusFailException.class, () -> manager.selectModels(config, Collections.emptyList()));
        assertThrows(StatusFailException.class, () -> manager.selectModels(config, Collections.singletonList(SECONDARY_ID)));
        assertThrows(StatusFailException.class, () -> manager.selectModels(config, Arrays.asList(PRIMARY_ID, "missing")));
        assertThrows(StatusFailException.class, () -> manager.selectModels(config, Arrays.asList(PRIMARY_ID, PRIMARY_ID)));
        assertThrows(StatusFailException.class, () -> manager.selectModels(config, Arrays.asList(PRIMARY_ID, null)));
    }

    @Test
    void modelRequestConfigDoesNotMutateSharedConfiguration() {
        ProblemAiConfig config = multi().setApiKey("primary-secret").setModel("primary-model")
                .setApiUrl("https://primary.example/v1").setTimeoutSeconds(300).setProtocol("gemini");
        config.getModels().get(0).setProtocol("gemini");
        config.getModels().get(1).setProtocol("anthropic");
        ProblemAiConfig request = manager.forModel(config, config.getModels().get(1));
        assertEquals("secondary-secret", request.getApiKey());
        assertEquals("secondary-model", request.getModel());
        assertEquals(Integer.valueOf(300), request.getTimeoutSeconds());
        assertEquals("anthropic", request.getProtocol());
        request.getModels().get(0).setName("changed");
        assertEquals("primary-secret", config.getApiKey());
        assertEquals("primary-model", config.getModel());
        assertEquals("gemini", config.getProtocol());
        assertEquals("主模型", config.getModels().get(0).getName());
    }

    @Test
    void selectedModelsKeepTheirProtocolsAndCopiesAreIsolated() throws Exception {
        ProblemAiConfig config = multi();
        config.getModels().get(0).setProtocol("gemini");
        config.getModels().get(1).setProtocol("anthropic");
        List<ProblemAiModelConfig> selected = manager.selectModels(config, Arrays.asList(SECONDARY_ID, PRIMARY_ID));
        assertEquals("gemini", selected.get(0).getProtocol());
        assertEquals("anthropic", selected.get(1).getProtocol());
        selected.get(0).setProtocol("openai_responses");
        assertEquals("gemini", config.getModels().get(0).getProtocol());
    }

    @Test
    void requiresExactlyOnePrimaryEvenWhenDisabled() {
        ProblemAiConfig input = multi().setEnabled(false);
        input.getModels().forEach(model -> model.setPrimary(false));
        assertThrows(StatusFailException.class, () -> manager.saveConfig(input));
        input.getModels().forEach(model -> model.setPrimary(true));
        assertThrows(StatusFailException.class, () -> manager.saveConfig(input));
        verify(service, never()).saveOrUpdate(any());
    }

    @Test
    void rejectsEmptyMalformedOrDuplicateModelDefinitions() {
        assertThrows(StatusFailException.class, () -> manager.saveConfig(multi().setModels(Collections.emptyList())));
        assertThrows(StatusFailException.class, () -> manager.saveConfig(multi().setModels(Collections.singletonList(null))));
        ProblemAiConfig malformed = multi();
        malformed.getModels().get(0).setId("1-1-1-1-1");
        assertThrows(StatusFailException.class, () -> manager.saveConfig(malformed));
        ProblemAiConfig duplicate = multi();
        duplicate.getModels().get(1).setId(PRIMARY_ID.toUpperCase());
        assertThrows(StatusFailException.class, () -> manager.saveConfig(duplicate));
        ProblemAiConfig unnamed = multi();
        unnamed.getModels().get(0).setName("  ");
        assertThrows(StatusFailException.class, () -> manager.saveConfig(unnamed));
        ProblemAiConfig noModel = multi();
        noModel.getModels().get(0).setModel("  ");
        assertThrows(StatusFailException.class, () -> manager.saveConfig(noModel));
        verify(service, never()).saveOrUpdate(any());
    }

    @Test
    void rejectsInvalidEndpointsWithoutEchoingTheirCredentials() {
        for (String url : Arrays.asList("file:///tmp/key", "https://user:secret@example.com/v1",
                "https://example.com/#secret", "https://example.com:65536/v1", "https:///v1", "bad url")) {
            ProblemAiConfig input = multi();
            input.getModels().get(0).setApiUrl(url);
            StatusFailException failure = assertThrows(StatusFailException.class, () -> manager.saveConfig(input));
            assertFalse(failure.getMessage().contains(url));
            assertNull(failure.getCause());
        }
        verify(service, never()).saveOrUpdate(any());
    }

    @Test
    void rejectsInvalidCredentialsAndTimeoutWithoutPersisting() {
        for (String key : Arrays.asList("", "token\r\ninjected", "token\u0000suffix", "token\tsuffix",
                String.join("", Collections.nCopies(501, "x")))) {
            ProblemAiConfig input = multi();
            input.getModels().get(0).setApiKey(key);
            StatusFailException failure = assertThrows(StatusFailException.class, () -> manager.saveConfig(input));
            assertFalse(failure.getMessage().contains("token"));
        }
        assertThrows(StatusFailException.class, () -> manager.saveConfig(multi().setTimeoutSeconds(0)));
        assertThrows(StatusFailException.class, () -> manager.saveConfig(multi().setTimeoutSeconds(3601)));
        verify(service, never()).saveOrUpdate(any());
    }

    @Test
    void malformedStoredJsonDoesNotExposeSecretOrParserCause() {
        when(service.getById(1L)).thenReturn(legacy().setModelsJson("{\"apiKey\":\"private-secret\"}"));
        StatusFailException failure = assertThrows(StatusFailException.class, manager::getConfig);
        assertFalse(failure.getMessage().contains("private-secret"));
        assertNull(failure.getCause());
        assertFalse(legacy().toString().contains("legacy-secret"));
        assertFalse(multi().getModels().get(0).toString().contains("primary-secret"));
    }

    @Test
    void requiresAuthenticatedProblemAdministratorBeforeReadingOrWriting() {
        when(subject.isAuthenticated()).thenReturn(false);
        assertThrows(StatusForbiddenException.class, manager::getConfig);
        assertThrows(StatusForbiddenException.class, () -> manager.saveConfig(multi()));
        when(subject.isAuthenticated()).thenReturn(true);
        when(subject.hasRole("admin")).thenReturn(false);
        assertThrows(StatusForbiddenException.class, manager::getConfig);
        assertThrows(StatusForbiddenException.class, () -> manager.saveConfig(multi()));
        verifyNoInteractions(service);
    }

    @Test
    void reportsFailedPersistenceInsteadOfClaimingSuccess() {
        when(service.saveOrUpdate(any(ProblemAiConfig.class))).thenReturn(false);
        assertThrows(StatusFailException.class, () -> manager.saveConfig(multi()));
    }

    @Test
    void databaseFailuresCannotExposeCredentialsInControllerErrorMessages() {
        when(service.saveOrUpdate(any(ProblemAiConfig.class)))
                .thenThrow(new IllegalStateException("bound parameter: primary-secret"));
        StatusFailException writeFailure = assertThrows(StatusFailException.class, () -> manager.saveConfig(multi()));
        assertFalse(writeFailure.getMessage().contains("primary-secret"));
        assertNull(writeFailure.getCause());
        when(service.getById(1L)).thenThrow(new IllegalStateException("corrupt value: primary-secret"));
        StatusFailException readFailure = assertThrows(StatusFailException.class, manager::getConfig);
        assertFalse(readFailure.getMessage().contains("primary-secret"));
        assertNull(readFailure.getCause());
    }

    private void store(ProblemAiConfig config) {
        config.setModelsJson(JSONUtil.toJsonStr(config.getModels()));
        when(service.getById(1L)).thenReturn(config);
    }

    private ProblemAiConfig saved() {
        ArgumentCaptor<ProblemAiConfig> captor = ArgumentCaptor.forClass(ProblemAiConfig.class);
        verify(service).saveOrUpdate(captor.capture());
        return captor.getValue();
    }

    private ProblemAiConfig legacy() {
        return new ProblemAiConfig().setId(1L).setEnabled(true).setApiUrl("https://legacy.example/v1")
                .setApiKey("legacy-secret").setModel("gpt-6-astra").setTimeoutSeconds(300);
    }

    private ProblemAiConfig multi() {
        return new ProblemAiConfig().setEnabled(true).setModels(Arrays.asList(
                new ProblemAiModelConfig().setId(PRIMARY_ID).setName("主模型").setPrimary(true)
                        .setApiUrl("https://primary.example/v1").setApiKey("primary-secret").setModel("primary-model"),
                new ProblemAiModelConfig().setId(SECONDARY_ID).setName("辅助模型").setPrimary(false)
                        .setApiUrl("https://secondary.example/v1").setApiKey("secondary-secret").setModel("secondary-model")));
    }
}
