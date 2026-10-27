package top.hcode.hoj.manager.admin.problem;

import cn.hutool.json.JSONArray;
import cn.hutool.json.JSONObject;
import cn.hutool.json.JSONUtil;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Test;
import org.springframework.test.util.ReflectionTestUtils;
import top.hcode.hoj.pojo.dto.ProblemAiValidateDTO;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.pojo.entity.problem.ProblemAiConfig;
import top.hcode.hoj.pojo.vo.ProblemAiLanguageProgramVO;

import java.util.*;
import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.atomic.AtomicReference;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

class ProblemAiValidationRunnerTest {
    private final ProblemAiGateway gateway = mock(ProblemAiGateway.class);
    private final ProblemAiValidationContext context = mock(ProblemAiValidationContext.class);
    private final ProblemAiMultiLanguageRunner languages = mock(ProblemAiMultiLanguageRunner.class);
    private final ProblemAiValidationRunner runner = new ProblemAiValidationRunner();
    private final ProblemAiConfig config = new ProblemAiConfig();
    private final Problem problem = new Problem().setId(5L).setJudgeMode("default");
    private final ProblemAiValidateDTO input = new ProblemAiValidateDTO();
    private String overview;
    private final ExecutorService caller = Executors.newSingleThreadExecutor();

    @BeforeEach
    void setup() {
        ReflectionTestUtils.setField(runner, "gateway", gateway);
        ReflectionTestUtils.setField(runner, "contextService", context);
        ReflectionTestUtils.setField(runner, "promptFactory", new ProblemAiPromptFactory());
        ReflectionTestUtils.setField(runner, "multiLanguageRunner", languages);
        input.setPid(5L);
        input.setLanguage("C++ 17");
        input.setStandardProgram("strict std source");
        overview = "{\"overall\":\"PASS\",\"issues\":[],\"summary\":\"题面检查通过\"}";
        when(gateway.complete(eq(config), anyList())).thenAnswer(call -> {
            java.util.List<java.util.Map<String, String>> messages = call.getArgument(1);
            return messages.get(1).get("content").contains("逐测试点阶段")
                    ? "{\"testPointResults\":[{\"index\":1,\"status\":\"WARN\",\"detail\":\"文件未展开\"}]}"
                    : overview;
        });
    }

    @AfterEach
    void shutdown() {
        caller.shutdownNow();
        runner.shutdown();
    }

    @Test
    void specialJudgeReviewIsExecutedAndCannotBeOverriddenByAcceptedStandard() throws Exception {
        problem.setJudgeMode("spj").setSpjLanguage("C++").setSpjCode("checker source");
        ProblemAiJudgeProgramRunner checks = mock(ProblemAiJudgeProgramRunner.class);
        ReflectionTestUtils.setField(runner, "judgeProgramRunner", checks);
        snapshot(0, point(0, true));
        overview = "{\"overall\":\"PASS\",\"judgeProgramReview\":{\"status\":\"PASS\"}}";
        when(checks.run(eq(problem), any(), anyString(), anyString()))
                .thenReturn(new JSONObject().set("status", "FAIL").set("summary", "非法输出被接受"));
        JSONObject result = run();
        assertEquals("FAIL", result.getStr("overall"));
        assertEquals("FAIL", result.getJSONObject("judgeProgramReview").getStr("status"));
        verify(checks).run(eq(problem), any(), anyString(), anyString());
    }

    @Test
    void interactiveReportDoesNotClaimStaticTestlibValidation() throws Exception {
        problem.setJudgeMode("interactive");
        ProblemAiJudgeProgramRunner checks = mock(ProblemAiJudgeProgramRunner.class);
        ReflectionTestUtils.setField(runner, "judgeProgramRunner", checks);
        snapshot(0, point(0, true));
        when(checks.run(eq(problem), any(), anyString(), anyString()))
                .thenReturn(new JSONObject().set("status", "WARN").set("summary", "未完成交互器检查"));
        JSONObject result = run();
        assertEquals("WARN", result.getStr("overall"));
        assertEquals("interactive_protocol", result.getJSONObject("standardProgram").getStr("validationMode"));
        assertEquals("interactive_protocol", result.getJSONObject("testPointValidation").getStr("mode"));
    }

    @Test
    void acceptedStdIsBothValidationAndSolutionWithoutAnotherValidator() throws Exception {
        ProblemAiTestPoint point = point(0, true).setInputAvailable(false).setExpectedOutputAvailable(false);
        snapshot(0, point);
        JSONObject report = run();
        assertEquals("PASS", report.getStr("overall"));
        JSONObject test = report.getJSONArray("testPointResults").getJSONObject(0);
        assertEquals("PASS", test.getStr("status"));
        assertEquals("WARN", test.getStr("aiStatus"));
        assertEquals("integrated_standard_program", test.getJSONObject("formalValidation").getStr("mode"));
        assertEquals("PASS", test.getJSONObject("testlibValidator").getStr("status"));
        assertFalse(report.getJSONObject("standardProgram").containsKey("validatorCode"));
        verifyNoInteractions(languages);
    }

    @Test
    void runtimeFailureIncludesActualStdStderr() throws Exception {
        snapshot(3, point(3, true).setStderr("Integer n violates range [1, 100]"));
        JSONObject report = run();
        assertEquals("FAIL", report.getStr("overall"));
        JSONObject test = report.getJSONArray("testPointResults").getJSONObject(0);
        assertEquals("FAIL", test.getStr("status"));
        assertEquals("Integer n violates range [1, 100]", test.getJSONObject("formalValidation").getStr("stderr"));
    }

    @Test
    void unexecutedPointsCannotPass() throws Exception {
        snapshot(0, point(0, false));
        assertEquals("FAIL", run().getStr("overall"));
    }

    @Test
    void noTestPointsCannotPassEvenWhenAiSaysPass() throws Exception {
        when(context.load(input)).thenReturn(new ProblemAiValidationContext.Snapshot(
                90L, -2, "Compile Error", "compilation failed", Collections.emptyList()));
        JSONObject report = run();
        assertEquals("FAIL", report.getStr("overall"));
        assertEquals("compilation failed", report.getJSONObject("execution").getStr("judgeError"));
    }

    @Test
    void acceptedSamplesUseSameIntegratedValidation() throws Exception {
        snapshot(0, point(0, true).setConcrete(false));
        assertEquals("PASS", run().getStr("overall"));
    }

    @Test
    void acceptedJudgeDoesNotEraseStatementWarnings() throws Exception {
        overview = "{\"overall\":\"WARN\",\"issues\":[{\"severity\":\"WARNING\",\"detail\":\"约束不完整\"}]}";
        snapshot(0, point(0, true));
        assertEquals("WARN", run().getStr("overall"));
    }

    @Test
    void confirmedStatementErrorPreventsPass() throws Exception {
        overview = "{\"overall\":\"PASS\",\"issues\":[{\"severity\":\"ERROR\",\"detail\":\"题面矛盾\"}]}";
        snapshot(0, point(0, true));
        assertEquals("FAIL", run().getStr("overall"));
    }

    @Test
    void requiredLanguageFailuresAreRetained() throws Exception {
        snapshot(0, point(0, true));
        input.setMultiLanguagePrograms(Collections.singletonList(
                new ProblemAiLanguageProgramVO().setLanguage("Python3").setCode("print(1)")));
        when(languages.run(eq(problem), eq(input.getMultiLanguagePrograms()), anyList(), eq("uid"), eq("admin")))
                .thenReturn(new JSONArray().put(new JSONObject().set("language", "Python3")
                        .set("status", "FAIL").set("passed", 0).set("total", 1)));
        JSONObject report = run();
        assertEquals("FAIL", report.getStr("overall"));
        assertTrue(report.getJSONArray("issues").toString().contains("Python3"));
    }

    @Test
    void languagesStartAlongsideAiAndAllResultsAreJoinedBeforeReportingFailure() throws Exception {
        prepareLanguages();
        CountDownLatch aiStarted = new CountDownLatch(2);
        CountDownLatch aiFinished = new CountDownLatch(2);
        CountDownLatch languageStarted = new CountDownLatch(1);
        CountDownLatch releaseAi = new CountDownLatch(1);
        CountDownLatch releaseLanguage = new CountDownLatch(1);
        when(gateway.complete(eq(config), anyList())).thenAnswer(call -> {
            aiStarted.countDown();
            await(releaseAi);
            aiFinished.countDown();
            return response(call.getArgument(1));
        });
        when(languages.run(any(), anyList(), anyList(), anyString(), anyString())).thenAnswer(call -> {
            languageStarted.countDown();
            await(releaseLanguage);
            return new JSONArray().put(new JSONObject().set("language", "Python3")
                    .set("status", "FAIL").set("passed", 0).set("total", 1));
        });
        Future<JSONObject> result = caller.submit(this::run);
        await(aiStarted);
        await(languageStarted);
        releaseAi.countDown();
        await(aiFinished);
        assertFalse(result.isDone(), "必须等待仍在执行的多语言正式判题");
        releaseLanguage.countDown();
        JSONObject report = result.get(5, TimeUnit.SECONDS);
        assertEquals("FAIL", report.getStr("overall"));
        assertEquals(1, report.getJSONArray("testPointResults").size());
        assertTrue(report.getJSONArray("issues").toString().contains("Python3"));
    }

    @Test
    void completedLanguageJudgeDoesNotSkipPendingAiAnalysis() throws Exception {
        prepareLanguages();
        CountDownLatch aiStarted = new CountDownLatch(2);
        CountDownLatch languageFinished = new CountDownLatch(1);
        CountDownLatch releaseAi = new CountDownLatch(1);
        when(gateway.complete(eq(config), anyList())).thenAnswer(call -> {
            aiStarted.countDown();
            await(releaseAi);
            return response(call.getArgument(1));
        });
        when(languages.run(any(), anyList(), anyList(), anyString(), anyString())).thenAnswer(call -> {
            languageFinished.countDown();
            return new JSONArray().put(new JSONObject().set("language", "Python3").set("status", "PASS"));
        });
        Future<JSONObject> result = caller.submit(this::run);
        await(aiStarted);
        await(languageFinished);
        assertFalse(result.isDone(), "必须等待全部 AI 检查返回");
        releaseAi.countDown();
        assertEquals("PASS", result.get(5, TimeUnit.SECONDS).getStr("overall"));
    }

    @Test
    void aiFailureCancelsRunningLanguageJudgeAndOtherAiBatches() throws Exception {
        prepareLanguages();
        CountDownLatch aiStarted = new CountDownLatch(2);
        CountDownLatch languageStarted = new CountDownLatch(1);
        CountDownLatch cancelled = new CountDownLatch(2);
        CountDownLatch blocked = new CountDownLatch(1);
        when(gateway.complete(eq(config), anyList())).thenAnswer(call -> {
            aiStarted.countDown();
            await(aiStarted);
            await(languageStarted);
            if (!isBatch(call.getArgument(1))) throw new IllegalStateException("AI request failed");
            try { blocked.await(); }
            catch (InterruptedException error) { cancelled.countDown(); throw error; }
            return response(call.getArgument(1));
        });
        when(languages.run(any(), anyList(), anyList(), anyString(), anyString())).thenAnswer(call -> {
            languageStarted.countDown();
            try { blocked.await(); }
            catch (InterruptedException error) { cancelled.countDown(); throw error; }
            return new JSONArray();
        });
        Future<JSONObject> result = caller.submit(this::run);
        ExecutionException error = assertThrows(ExecutionException.class, () -> result.get(5, TimeUnit.SECONDS));
        assertEquals("AI request failed", error.getCause().getMessage());
        await(cancelled);
    }

    @Test
    void languageFailureIsObservedWhileAiIsPendingAndCancelsBothAiCalls() throws Exception {
        prepareLanguages();
        CountDownLatch aiStarted = new CountDownLatch(2);
        CountDownLatch cancelled = new CountDownLatch(2);
        CountDownLatch blocked = new CountDownLatch(1);
        when(gateway.complete(eq(config), anyList())).thenAnswer(call -> {
            aiStarted.countDown();
            try { blocked.await(); }
            catch (InterruptedException error) { cancelled.countDown(); throw error; }
            return response(call.getArgument(1));
        });
        when(languages.run(any(), anyList(), anyList(), anyString(), anyString())).thenAnswer(call -> {
            await(aiStarted);
            throw new IllegalStateException("judge failed");
        });
        Future<JSONObject> result = caller.submit(this::run);
        ExecutionException error = assertThrows(ExecutionException.class, () -> result.get(5, TimeUnit.SECONDS));
        assertEquals("judge failed", error.getCause().getMessage());
        await(cancelled);
    }

    @Test
    void callerInterruptionPreservesFlagAndCancelsAllStartedTasks() throws Exception {
        prepareLanguages();
        CountDownLatch started = new CountDownLatch(3);
        CountDownLatch cancelled = new CountDownLatch(3);
        CountDownLatch blocked = new CountDownLatch(1);
        AtomicReference<Thread> callerThread = new AtomicReference<>();
        AtomicBoolean interrupted = new AtomicBoolean();
        when(gateway.complete(eq(config), anyList())).thenAnswer(call -> {
            started.countDown();
            try { blocked.await(); }
            catch (InterruptedException error) { cancelled.countDown(); throw error; }
            return response(call.getArgument(1));
        });
        when(languages.run(any(), anyList(), anyList(), anyString(), anyString())).thenAnswer(call -> {
            started.countDown();
            try { blocked.await(); }
            catch (InterruptedException error) { cancelled.countDown(); throw error; }
            return new JSONArray();
        });
        Future<JSONObject> result = caller.submit(() -> {
            callerThread.set(Thread.currentThread());
            try { return run(); }
            finally { interrupted.set(Thread.currentThread().isInterrupted()); }
        });
        await(started);
        callerThread.get().interrupt();
        ExecutionException error = assertThrows(ExecutionException.class, () -> result.get(5, TimeUnit.SECONDS));
        assertTrue(error.getCause() instanceof InterruptedException);
        assertTrue(interrupted.get());
        await(cancelled);
    }

    @Test
    void incompleteDuplicateAndForeignBatchIndexesCannotPassEvenWhenJudgeAccepts() throws Exception {
        snapshot(0, point(0, true));
        for (String invalid : Arrays.asList("{}", "{\"testPointResults\":[]}",
                "{\"testPointResults\":[{\"index\":1},{\"index\":1}]}",
                "{\"testPointResults\":[{\"index\":2}]}",
                "{\"testPointResults\":[{}]}")) {
            when(gateway.complete(eq(config), anyList())).thenAnswer(call ->
                    isBatch(call.getArgument(1)) ? invalid : overview);
            IllegalStateException error = assertThrows(IllegalStateException.class, this::run);
            assertTrue(error.getMessage().startsWith("AI 逐点分析"));
        }
    }

    @Test
    void allPointsAcrossBatchesAreAnalyzedWithAtMostThreeConcurrentAiCalls() throws Exception {
        List<ProblemAiTestPoint> points = new ArrayList<>();
        for (int index = 1; index <= 25; index++) points.add(point(0, true).setIndex(index));
        when(context.load(input)).thenReturn(new ProblemAiValidationContext.Snapshot(
                90L, 0, "Accepted", "", points));
        ProblemAiPromptFactory prompts = mock(ProblemAiPromptFactory.class);
        ReflectionTestUtils.setField(runner, "promptFactory", prompts);
        when(prompts.validationPrompt(any(), any(), anyList(), any(), any())).thenReturn("overview");
        List<Integer> batchSizes = new CopyOnWriteArrayList<>();
        when(prompts.testPointPrompt(any(), any(), any(), anyList(), eq(25))).thenAnswer(call -> {
            List<ProblemAiTestPoint> batch = call.getArgument(3);
            batchSizes.add(batch.size());
            JSONArray results = new JSONArray();
            for (ProblemAiTestPoint point : batch) results.add(new JSONObject()
                    .set("index", point.getIndex()).set("status", "PASS"));
            return new JSONObject().set("testPointResults", results).toString();
        });
        AtomicInteger active = new AtomicInteger();
        AtomicInteger maximum = new AtomicInteger();
        CountDownLatch firstWaveStarted = new CountDownLatch(3);
        CountDownLatch release = new CountDownLatch(1);
        when(gateway.complete(eq(config), anyList())).thenAnswer(call -> {
            maximum.accumulateAndGet(active.incrementAndGet(), Math::max);
            firstWaveStarted.countDown();
            try {
                await(release);
                List<Map<String, String>> messages = call.getArgument(1);
                String prompt = messages.get(1).get("content");
                return "overview".equals(prompt) ? overview : prompt;
            } finally {
                active.decrementAndGet();
            }
        });
        Future<JSONObject> result = caller.submit(this::run);
        await(firstWaveStarted);
        assertEquals(3, active.get());
        release.countDown();
        JSONObject report = result.get(5, TimeUnit.SECONDS);
        assertEquals("PASS", report.getStr("overall"));
        assertEquals(25, report.getJSONArray("testPointResults").size());
        Collections.sort(batchSizes);
        assertEquals(Arrays.asList(1, 12, 12), batchSizes);
        assertEquals(3, maximum.get());
        verify(gateway, times(4)).complete(eq(config), anyList());
    }

    private void prepareLanguages() throws Exception {
        snapshot(0, point(0, true));
        input.setMultiLanguagePrograms(Collections.singletonList(
                new ProblemAiLanguageProgramVO().setLanguage("Python3").setCode("print(1)")));
    }

    private boolean isBatch(List<Map<String, String>> messages) {
        return messages.get(1).get("content").contains("逐测试点阶段");
    }

    private String response(List<Map<String, String>> messages) {
        return isBatch(messages) ? "{\"testPointResults\":[{\"index\":1,\"status\":\"PASS\"}]}" : overview;
    }

    private void await(CountDownLatch latch) throws InterruptedException {
        assertTrue(latch.await(5, TimeUnit.SECONDS), "并发测试同步点未到达");
    }

    private ProblemAiTestPoint point(int status, boolean executed) {
        return new ProblemAiTestPoint().setIndex(1).setCaseId(55L).setConcrete(true)
                .setExecuted(executed).setJudgeStatus(status).setJudgeStatusText(status == 0 ? "Accepted" : "Runtime Error")
                .setTime(5).setMemory(1024).setStderr("");
    }

    private void snapshot(int status, ProblemAiTestPoint point) throws Exception {
        when(context.load(input)).thenReturn(new ProblemAiValidationContext.Snapshot(
                90L, status, status == 0 ? "Accepted" : "Runtime Error", "", Collections.singletonList(point)));
    }

    private JSONObject run() throws Exception {
        return JSONUtil.parseObj(runner.run(config, problem, Collections.emptyList(), input, "uid", "admin"));
    }
}
