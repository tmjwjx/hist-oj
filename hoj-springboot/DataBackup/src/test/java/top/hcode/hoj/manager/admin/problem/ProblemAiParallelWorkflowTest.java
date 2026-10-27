package top.hcode.hoj.manager.admin.problem;

import cn.hutool.json.JSONArray;
import cn.hutool.json.JSONObject;
import cn.hutool.json.JSONUtil;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.test.util.ReflectionTestUtils;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.dao.problem.ProblemAiRecordEntityService;
import top.hcode.hoj.dao.problem.ProblemCaseEntityService;
import top.hcode.hoj.dao.problem.ProblemEntityService;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.pojo.entity.problem.ProblemAiConfig;
import top.hcode.hoj.pojo.entity.problem.ProblemAiModelConfig;
import top.hcode.hoj.pojo.entity.problem.ProblemAiRecord;
import top.hcode.hoj.pojo.vo.ProblemAiGeneratedProgramVO;

import java.util.Arrays;
import java.util.Collections;
import java.util.List;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.atomic.AtomicLong;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

class ProblemAiParallelWorkflowTest {
    private final ProblemAiManager manager = new ProblemAiManager();
    private final ProblemAiConfigManager configs = mock(ProblemAiConfigManager.class);
    private final ProblemAiGateway gateway = mock(ProblemAiGateway.class);
    private final ProblemVerificationManager verification = mock(ProblemVerificationManager.class);
    private final ProblemAiValidationRunner runner = mock(ProblemAiValidationRunner.class);
    private final ProblemAiRecordEntityService records = mock(ProblemAiRecordEntityService.class);
    private final ProblemEntityService problems = mock(ProblemEntityService.class);
    private final ProblemCaseEntityService cases = mock(ProblemCaseEntityService.class);
    private final Problem problem = new Problem().setId(7L).setCaseVersion("v1").setJudgeMode("spj");
    private final ProblemAiConfig config = new ProblemAiConfig().setModel("primary-model");
    private final ProblemAiConfig secondaryConfig = new ProblemAiConfig().setModel("secondary-model");
    private final ProblemAiConfig thirdConfig = new ProblemAiConfig().setModel("third-model");
    private final ProblemAiModelConfig primary = model("primary", true);
    private final ProblemAiModelConfig secondary = model("secondary", false);
    private final ProblemAiModelConfig third = model("third", false);
    private final ProblemAiModelConfig unselected = model("unselected", false);
    private final ProblemAiRecord record = new ProblemAiRecord().setId(50L).setStatus("running");

    @BeforeEach
    void setup() throws Exception {
        ReflectionTestUtils.setField(manager, "problemAiConfigManager", configs);
        ReflectionTestUtils.setField(manager, "gateway", gateway);
        ReflectionTestUtils.setField(manager, "verificationManager", verification);
        ReflectionTestUtils.setField(manager, "validationRunner", runner);
        ReflectionTestUtils.setField(manager, "recordService", records);
        ReflectionTestUtils.setField(manager, "problemService", problems);
        ReflectionTestUtils.setField(manager, "caseService", cases);
        ReflectionTestUtils.setField(manager, "promptFactory", new ProblemAiPromptFactory());
        config.setModels(Arrays.asList(primary, secondary, third, unselected));
        when(problems.getById(7L)).thenReturn(problem);
        when(cases.list(any())).thenReturn(Collections.emptyList());
        when(configs.forModel(config, secondary)).thenReturn(secondaryConfig);
        when(configs.forModel(config, third)).thenReturn(thirdConfig);
        when(gateway.complete(any(), anyList())).thenReturn("{\"code\":\"int main(){return 0;}\"}");
        when(verification.submitForAi(any(), eq("admin"), eq("admin"))).thenReturn(100L);
        AtomicLong comparisonId = new AtomicLong(100L);
        when(verification.submitComparisonForAi(any(), eq("admin"), eq("admin")))
                .thenAnswer(call -> comparisonId.incrementAndGet());
        when(verification.comparisonResult(eq(7L), anyLong())).thenAnswer(call -> formal(call.getArgument(1)));
        when(runner.run(eq(config), eq(problem), anyList(), any(), eq("admin"), eq("admin")))
                .thenReturn("{\"overall\":\"PASS\",\"summary\":\"主流程完成\"}");
    }

    @Test
    void generationsOverlapPrimaryJudgeAndReportWaitsForEverySelectedComparison() throws Exception {
        CountDownLatch primaryWaiting = new CountDownLatch(1);
        CountDownLatch generationsStarted = new CountDownLatch(4);
        CountDownLatch releasePrimary = new CountDownLatch(1);
        CountDownLatch releaseGenerations = new CountDownLatch(1);
        CountDownLatch auxiliaryJudgesWaiting = new CountDownLatch(2);
        CountDownLatch releaseAuxiliaryJudges = new CountDownLatch(1);
        CountDownLatch analysisStarted = new CountDownLatch(1);
        CountDownLatch releaseAnalysis = new CountDownLatch(1);
        CountDownLatch mainReportPersisted = new CountDownLatch(1);
        AtomicBoolean primaryFinished = new AtomicBoolean();
        AtomicLong comparisonId = new AtomicLong(100L);
        when(verification.comparisonResult(7L, 100L)).thenAnswer(call -> {
            primaryFinished.set(true);
            return formal(100L);
        });
        when(verification.submitComparisonForAi(any(), eq("admin"), eq("admin"))).thenAnswer(call -> {
            assertTrue(primaryFinished.get(), "辅助模型必须等主模型完成首次正式判题后才提交");
            return comparisonId.incrementAndGet();
        });
        when(gateway.complete(any(), anyList())).thenAnswer(call -> {
            generationsStarted.countDown();
            await(releaseGenerations);
            return "{\"code\":\"int main(){return 0;}\"}";
        });
        doAnswer(call -> {
            primaryWaiting.countDown();
            await(releasePrimary);
            return null;
        }).when(verification).waitForAiJudge(eq(7L), eq(100L), anyLong());
        doAnswer(call -> {
            auxiliaryJudgesWaiting.countDown();
            await(releaseAuxiliaryJudges);
            return null;
        }).when(verification).waitForAiJudge(eq(7L), longThat(id -> id > 100L), anyLong());
        when(runner.run(eq(config), eq(problem), anyList(), any(), eq("admin"), eq("admin")))
                .thenAnswer(call -> {
                    analysisStarted.countDown();
                    await(releaseAnalysis);
                    return "{\"overall\":\"PASS\",\"summary\":\"主流程完成\"}";
                });
        when(records.updateById(any())).thenAnswer(call -> {
            if (record.getResponse().contains("主流程完成") && "running".equals(record.getStatus())) {
                mainReportPersisted.countDown();
            }
            return true;
        });
        ExecutorService orchestration = Executors.newSingleThreadExecutor();
        Future<?> task = orchestration.submit(() -> run(Arrays.asList(primary, secondary, third)));
        try {
            await(primaryWaiting);
            await(generationsStarted);
            verify(verification, never()).submitComparisonForAi(any(), anyString(), anyString());
            releaseGenerations.countDown();
            releasePrimary.countDown();
            await(auxiliaryJudgesWaiting);
            await(analysisStarted);
            assertFalse(task.isDone());
            releaseAnalysis.countDown();
            await(mainReportPersisted);
            assertFalse(task.isDone(), "报告必须等待所有已选择模型完成正式判题");
            releaseAuxiliaryJudges.countDown();
            task.get(10, TimeUnit.SECONDS);
            JSONObject report = JSONUtil.parseObj(record.getResponse());
            assertEquals("success", record.getStatus());
            assertEquals("PASS", report.getStr("overall"));
            JSONArray models = report.getJSONArray("modelResults");
            assertEquals(3, models.size());
            for (Object value : models) {
                JSONObject result = JSONUtil.parseObj(value);
                assertEquals("SUCCESS", result.getStr("status"));
                assertEquals(3, result.getInt("passedCases"));
                assertEquals(3, result.getJSONArray("caseResults").size());
            }
            verify(configs, never()).forModel(config, unselected);
            verify(gateway, times(2)).complete(eq(config), anyList());
            verify(gateway, times(1)).complete(eq(secondaryConfig), anyList());
            verify(gateway, times(1)).complete(eq(thirdConfig), anyList());
            verify(runner, times(1)).run(eq(config), eq(problem), anyList(), argThat(input ->
                    input.getMultiLanguagePrograms().size() == 2
                            && "Java".equals(input.getMultiLanguagePrograms().get(0).getLanguage())
                            && "PyPy3".equals(input.getMultiLanguagePrograms().get(1).getLanguage())),
                    eq("admin"), eq("admin"));
        } finally {
            releasePrimary.countDown();
            releaseGenerations.countDown();
            releaseAuxiliaryJudges.countDown();
            releaseAnalysis.countDown();
            task.cancel(true);
            orchestration.shutdownNow();
        }
    }

    @Test
    void primaryJudgeFailureReleasesAuxiliaryGateAndRetainsItsResults() throws Exception {
        doThrow(new StatusFailException("主模型判题失败"))
                .when(verification).waitForAiJudge(eq(7L), eq(100L), anyLong());
        ExecutorService orchestration = Executors.newSingleThreadExecutor();
        Future<?> task = orchestration.submit(() -> run(Arrays.asList(primary, secondary)));
        try {
            task.get(10, TimeUnit.SECONDS);
            JSONObject report = JSONUtil.parseObj(record.getResponse());
            assertEquals("FAIL", report.getStr("overall"));
            assertEquals("ERROR", report.getJSONArray("modelResults").getJSONObject(0).getStr("status"));
            assertEquals("SUCCESS", report.getJSONArray("modelResults").getJSONObject(1).getStr("status"));
            assertEquals(3, report.getJSONArray("modelResults").getJSONObject(1).getInt("passedCases"));
            verifyNoInteractions(runner);
        } finally {
            task.cancel(true);
            orchestration.shutdownNow();
        }
    }

    @Test
    void statementChangeWithoutCaseVersionChangeStopsBeforeLaunchingWorkers() {
        when(problems.getById(7L)).thenReturn(new Problem().setId(7L).setCaseVersion("v1")
                .setJudgeMode("spj").setInput("new input constraints"));
        run(Arrays.asList(primary, secondary));
        JSONObject report = JSONUtil.parseObj(record.getResponse());
        assertEquals("FAIL", report.getStr("overall"));
        assertTrue(report.getStr("summary").contains("题目或测试数据已变化"));
        verifyNoInteractions(gateway, verification, runner);
    }

    @Test
    void translationsUseOnlyPrimarySourceAndAuxiliaryPromptRemainsIndependent() {
        run(Arrays.asList(primary, secondary));
        verify(gateway, times(2)).complete(eq(config), argThat(messages ->
                messages.stream().anyMatch(message -> message.get("content").contains("canonical std source"))));
        verify(gateway).complete(eq(secondaryConfig), argThat(messages ->
                messages.stream().noneMatch(message -> message.get("content").contains("canonical std source"))));
        verify(configs, never()).forModel(config, third);
        verify(configs, never()).forModel(config, unselected);
    }

    private void run(List<ProblemAiModelConfig> models) {
        ProblemAiGeneratedProgramVO standard = new ProblemAiGeneratedProgramVO()
                .setLanguage("C++ 17").setCode("canonical std source");
        ReflectionTestUtils.invokeMethod(manager, "runFullValidation", config, models, problem,
                standard, record, "admin", "admin");
    }

    private static void await(CountDownLatch latch) throws InterruptedException {
        assertTrue(latch.await(10, TimeUnit.SECONDS), "并行阶段未在等待期限内到达");
    }

    private static ProblemAiModelConfig model(String id, boolean primary) {
        return new ProblemAiModelConfig().setId(id).setName(id).setModel(id + "-model").setPrimary(primary);
    }

    private static JSONObject formal(Long submitId) {
        JSONArray cases = new JSONArray();
        for (int i = 1; i <= 3; i++) cases.add(new JSONObject().set("index", i).set("judgeStatus", 0));
        return new JSONObject().set("submitId", submitId).set("status", "SUCCESS").set("judgeStatus", 0)
                .set("totalCases", 3).set("passedCases", 3).set("failedCases", 0)
                .set("pendingCases", 0).set("caseResults", cases);
    }
}
