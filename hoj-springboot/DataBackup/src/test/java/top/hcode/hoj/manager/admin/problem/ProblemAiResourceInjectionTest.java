package top.hcode.hoj.manager.admin.problem;

import org.junit.jupiter.api.Test;
import org.springframework.context.annotation.AnnotationConfigApplicationContext;
import top.hcode.hoj.dao.problem.ProblemAiRecordEntityService;
import top.hcode.hoj.dao.problem.ProblemCaseEntityService;
import top.hcode.hoj.dao.problem.ProblemEntityService;
import top.hcode.hoj.manager.admin.system.ConfigManager;
import top.hcode.hoj.pojo.vo.ProblemAiConfigVO;

import java.util.Collections;

import static org.junit.jupiter.api.Assertions.assertSame;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

class ProblemAiResourceInjectionTest {
    @Test
    void resourceInjectionDoesNotConfuseSystemAndProblemAiConfiguration() throws Exception {
        ConfigManager systemConfig = mock(ConfigManager.class);
        ProblemAiConfigManager aiConfig = mock(ProblemAiConfigManager.class);
        ProblemAiRecordEntityService records = mock(ProblemAiRecordEntityService.class);
        ProblemAiConfigVO expected = new ProblemAiConfigVO().setEnabled(false);
        when(aiConfig.getConfig()).thenReturn(expected);
        when(records.list(any())).thenReturn(Collections.emptyList());

        try (AnnotationConfigApplicationContext context = new AnnotationConfigApplicationContext()) {
            context.getBeanFactory().registerSingleton("configManager", systemConfig);
            context.getBeanFactory().registerSingleton("problemAiConfigManager", aiConfig);
            context.getBeanFactory().registerSingleton("recordService", records);
            context.getBeanFactory().registerSingleton("problemService", mock(ProblemEntityService.class));
            context.getBeanFactory().registerSingleton("caseService", mock(ProblemCaseEntityService.class));
            context.getBeanFactory().registerSingleton("promptFactory", mock(ProblemAiPromptFactory.class));
            context.getBeanFactory().registerSingleton("gateway", mock(ProblemAiGateway.class));
            context.getBeanFactory().registerSingleton("validationRunner", mock(ProblemAiValidationRunner.class));
            context.getBeanFactory().registerSingleton("validationContext", mock(ProblemAiValidationContext.class));
            context.getBeanFactory().registerSingleton("verificationManager", mock(ProblemVerificationManager.class));
            context.refresh();
            assertTrue(context.isActive());

            // Apply the real @Resource and @PostConstruct processors without registering
            // destruction for this instance: its static executor is shared by other tests.
            ProblemAiManager manager = new ProblemAiManager();
            context.getAutowireCapableBeanFactory().autowireBean(manager);
            Object initialized = context.getAutowireCapableBeanFactory().initializeBean(manager, "problemAiManager");
            assertSame(manager, initialized);
            assertSame(expected, manager.getConfig());
            verify(aiConfig).getConfig();
            verify(records).list(any());
            verifyNoMoreInteractions(records);
            verifyNoInteractions(systemConfig);
        }
    }
}
