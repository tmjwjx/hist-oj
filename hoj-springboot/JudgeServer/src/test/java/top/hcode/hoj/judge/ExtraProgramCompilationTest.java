package top.hcode.hoj.judge;

import cn.hutool.extra.spring.SpringUtil;
import cn.hutool.json.JSONArray;
import cn.hutool.json.JSONObject;
import cn.hutool.json.JSONUtil;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;
import org.springframework.context.support.StaticApplicationContext;
import org.springframework.http.MediaType;
import org.springframework.http.client.ClientHttpRequestFactory;
import org.springframework.mock.http.client.MockClientHttpRequest;
import org.springframework.test.util.ReflectionTestUtils;
import org.springframework.test.web.client.MockRestServiceServer;
import top.hcode.hoj.common.exception.SystemError;
import top.hcode.hoj.judge.entity.LanguageConfig;
import top.hcode.hoj.pojo.entity.judge.Judge;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.util.Constants;

import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Method;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Collections;
import java.util.Map;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;
import java.util.concurrent.TimeUnit;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;
import static org.springframework.test.web.client.match.MockRestRequestMatchers.requestTo;
import static org.springframework.test.web.client.response.MockRestResponseCreators.withSuccess;

class ExtraProgramCompilationTest {
    @TempDir Path directory;
    private final JudgeStrategy strategy = new JudgeStrategy();
    private final LanguageConfigLoader languages = mock(LanguageConfigLoader.class);
    private MockRestServiceServer sandbox;
    private ClientHttpRequestFactory originalFactory;
    private Object originalBeanFactory;
    private String originalSpjDir;
    private String originalInteractiveDir;

    @BeforeEach
    void setup() {
        originalSpjDir = Constants.JudgeDir.SPJ_WORKPLACE_DIR.getContent();
        originalInteractiveDir = Constants.JudgeDir.INTERACTIVE_WORKPLACE_DIR.getContent();
        ReflectionTestUtils.setField(Constants.JudgeDir.SPJ_WORKPLACE_DIR, "content", directory.resolve("spj").toString());
        ReflectionTestUtils.setField(Constants.JudgeDir.INTERACTIVE_WORKPLACE_DIR, "content", directory.resolve("interactive").toString());
        originalBeanFactory = ReflectionTestUtils.getField(SpringUtil.class, "beanFactory");
        StaticApplicationContext context = new StaticApplicationContext();
        context.getBeanFactory().registerSingleton("languages", languages);
        new SpringUtil().postProcessBeanFactory(context.getBeanFactory());
        ReflectionTestUtils.setField(strategy, "languageConfigLoader", languages);
        when(languages.getLanguageConfigByName("SPJ-C++")).thenReturn(language("spj"));
        when(languages.getLanguageConfigByName("INTERACTIVE-C++")).thenReturn(language("interactor"));
        originalFactory = SandboxRun.getRestTemplate().getRequestFactory();
        sandbox = MockRestServiceServer.bindTo(SandboxRun.getRestTemplate()).build();
    }

    @AfterEach
    void cleanup() {
        SandboxRun.getRestTemplate().setRequestFactory(originalFactory);
        ReflectionTestUtils.setField(SpringUtil.class, "beanFactory", originalBeanFactory);
        ReflectionTestUtils.setField(Constants.JudgeDir.SPJ_WORKPLACE_DIR, "content", originalSpjDir);
        ReflectionTestUtils.setField(Constants.JudgeDir.INTERACTIVE_WORKPLACE_DIR, "content", originalInteractiveDir);
    }

    @Test
    void unchangedSpjIsCompiledOnceAndReused() throws Exception {
        Problem problem = problem("spj");
        expectCompile("spj", "spj", true);
        assertTrue(check(problem));
        assertTrue(check(problem));
        assertEquals(64, read(version("spj")).length());
        sandbox.verify();
    }

    @Test
    void legacyInteractiveVersionUsesInteractorCompilerAndDirectory() throws Exception {
        write(directory.resolve("interactive/23/interactor"), "old binary");
        write(version("interactive"), "old-version");
        expectCompile("interactive", "interactor", true);
        assertTrue(check(problem("interactive")));
        assertFalse(Files.exists(directory.resolve("spj/23/spj")));
        sandbox.verify();
    }

    @Test
    void sourceExtraFilesLanguageAndDataChangesEachInvalidateCache() throws Exception {
        Problem problem = problem("spj");
        compileAndReset(problem, "spj");
        String firstVersion = read(version("spj"));
        problem.setSpjCode("int main() { return 1; }");
        compileAndReset(problem, "spj");
        assertNotEquals(firstVersion, read(version("spj")));
        problem.setJudgeExtraFile("{\"helper.h\":\"#define X 1\"}");
        compileAndReset(problem, "spj");
        when(languages.getLanguageConfigByName("SPJ-C++17")).thenReturn(language("spj"));
        problem.setSpjLanguage("C++17");
        compileAndReset(problem, "spj");
        problem.setCaseVersion("v2");
        compileAndReset(problem, "spj");
    }

    @Test
    void failedCompilationCannotLeaveOldCacheMarkedValid() throws Exception {
        Problem problem = problem("spj");
        compileAndReset(problem, "spj");
        problem.setSpjCode("broken");
        expectCompile("spj", "spj", false);
        SystemError error = assertThrows(SystemError.class, () -> check(problem));
        assertEquals("checker.cpp: error", error.getStderr());
        assertFalse(Files.exists(version("spj")));
        sandbox.verify();
        sandbox.reset();
        problem.setSpjCode("int main() {}");
        compileAndReset(problem, "spj");
    }

    @Test
    void manualCompilationInvalidatesStoredFingerprint() throws Exception {
        Problem problem = problem("spj");
        compileAndReset(problem, "spj");
        expectCompile("spj", "spj", true);
        assertTrue(Compiler.compileSpj("different code", 23L, "C++", null));
        assertFalse(Files.exists(version("spj")));
        sandbox.verify();
        sandbox.reset();
        compileAndReset(problem, "spj");
    }

    @Test
    void concurrentJudgesShareOneCompilation() throws Exception {
        expectCompile("spj", "spj", true);
        ExecutorService executor = Executors.newFixedThreadPool(2);
        CountDownLatch ready = new CountDownLatch(2);
        CountDownLatch start = new CountDownLatch(1);
        try {
            Future<Boolean> first = executor.submit(() -> { ready.countDown(); start.await(); return check(problem("spj")); });
            Future<Boolean> second = executor.submit(() -> { ready.countDown(); start.await(); return check(problem("spj")); });
            assertTrue(ready.await(5, TimeUnit.SECONDS));
            start.countDown();
            assertTrue(first.get(10, TimeUnit.SECONDS));
            assertTrue(second.get(10, TimeUnit.SECONDS));
            sandbox.verify();
        } finally {
            start.countDown();
            executor.shutdownNow();
        }
    }

    @Test
    void verificationReceivesCompilerDiagnosticsButOrdinarySubmissionDoesNot() throws Exception {
        ProblemTestCaseUtils cases = mock(ProblemTestCaseUtils.class);
        when(cases.loadTestCaseInfo(anyLong(), anyString(), anyString(), anyString(), any()))
                .thenReturn(new JSONObject());
        ReflectionTestUtils.setField(strategy, "problemTestCaseUtils", cases);
        ReflectionTestUtils.setField(strategy, "verificationSampleCaseService", mock(VerificationSampleCaseService.class));
        when(languages.getLanguageConfigByName("JavaScript")).thenReturn(new LanguageConfig());
        expectCompile("spj", "spj", false);
        Judge judge = new Judge().setSubmitId(41L).setLanguage("JavaScript").setCode("").setIsProblemVerification(true);
        Map<String, Object> result = strategy.judge(problem("spj"), judge);
        assertEquals(Constants.Judge.STATUS_SYSTEM_ERROR.getStatus(), result.get("code"));
        assertTrue(result.get("errMsg").toString().contains("checker.cpp: error"));
        assertTrue(result.get("errMsg").toString().contains("compiler output"));
        sandbox.verify();
        sandbox.reset();
        expectCompile("spj", "spj", false);
        result = strategy.judge(problem("spj"), judge.setIsProblemVerification(false));
        assertFalse(result.get("errMsg").toString().contains("checker.cpp"));
        sandbox.verify();
    }

    @Test
    void missingSourceIsRejectedAndDefaultNeedsNoChecker() throws Exception {
        assertFalse(check(problem("spj").setSpjCode(" \n")));
        assertFalse(check(problem("interactive").setSpjCode(null)));
        assertTrue(check(problem("default").setSpjCode(null)));
        assertThrows(SystemError.class, () -> check(problem("spj").setSpjLanguage("unknown")));
    }

    private void compileAndReset(Problem problem, String executable) throws Exception {
        expectCompile("spj", executable, true);
        assertTrue(check(problem));
        sandbox.verify();
        sandbox.reset();
    }

    private void expectCompile(String mode, String executable, boolean accepted) {
        sandbox.expect(requestTo("http://localhost:5050/run")).andRespond(request -> {
            JSONObject cmd = JSONUtil.parseObj(((MockClientHttpRequest) request).getBodyAsString())
                    .getJSONArray("cmd").getJSONObject(0);
            assertEquals(directory.resolve(mode + "/23").toString(), cmd.getStr("copyOutDir"));
            assertTrue(cmd.getJSONObject("copyIn").containsKey(executable + ".cpp"));
            if (accepted) write(directory.resolve(mode + "/23/" + executable), "compiled");
            String body = new JSONArray().put(new JSONObject().set("status", accepted ? "Accepted" : "Nonzero Exit Status")
                    .set("files", new JSONObject().set("stdout", "compiler output").set("stderr", "checker.cpp: error"))).toString();
            return withSuccess(body, MediaType.APPLICATION_JSON).createResponse(request);
        });
    }

    private LanguageConfig language(String executable) {
        LanguageConfig result = new LanguageConfig();
        result.setSrcName(executable + ".cpp");
        result.setExeName(executable);
        result.setCompileCommand("g++ " + executable + ".cpp -o " + executable);
        result.setCompileEnvs(Collections.emptyList());
        result.setMaxCpuTime(1000L);
        result.setMaxRealTime(2000L);
        result.setMaxMemory(1024L * 1024);
        return result;
    }

    private Problem problem(String mode) {
        return new Problem().setId(23L).setJudgeMode(mode).setSpjCode("int main() {}")
                .setSpjLanguage("C++").setCaseVersion("v1");
    }

    private boolean check(Problem problem) throws Exception {
        Method method = JudgeStrategy.class.getDeclaredMethod("checkOrCompileExtraProgram", Problem.class);
        method.setAccessible(true);
        try {
            return (Boolean) method.invoke(strategy, problem);
        } catch (InvocationTargetException error) {
            if (error.getCause() instanceof Exception) throw (Exception) error.getCause();
            throw error;
        }
    }

    private Path version(String mode) { return directory.resolve(mode + "/23/version"); }
    private String read(Path path) throws Exception { return new String(Files.readAllBytes(path), StandardCharsets.UTF_8); }
    private void write(Path path, String value) throws java.io.IOException {
        Files.createDirectories(path.getParent());
        Files.write(path, value.getBytes(StandardCharsets.UTF_8));
    }
}
