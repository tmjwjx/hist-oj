package top.hcode.hoj.manager.admin.problem;

import cn.hutool.json.JSONArray;
import cn.hutool.json.JSONObject;
import cn.hutool.json.JSONUtil;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.test.util.ReflectionTestUtils;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.dao.problem.*;
import top.hcode.hoj.pojo.dto.ProblemAiValidateDTO;
import top.hcode.hoj.pojo.entity.problem.*;
import top.hcode.hoj.pojo.vo.ProblemAiGeneratedProgramVO;

import java.util.*;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

class ProblemAiMultiModelTest {
    private final ProblemAiManager manager = new ProblemAiManager();
    private final ProblemAiConfigManager configs = mock(ProblemAiConfigManager.class);
    private final ProblemAiGateway gateway = mock(ProblemAiGateway.class);
    private final ProblemVerificationManager verification = mock(ProblemVerificationManager.class);
    private final ProblemAiValidationRunner runner = mock(ProblemAiValidationRunner.class);
    private final ProblemAiRecordEntityService records = mock(ProblemAiRecordEntityService.class);
    private final ProblemEntityService problems = mock(ProblemEntityService.class);
    private final ProblemCaseEntityService cases = mock(ProblemCaseEntityService.class);
    private final Problem problem = new Problem().setId(7L).setCaseVersion("v1").setJudgeMode("default");
    private final ProblemAiConfig config = new ProblemAiConfig().setModel("primary-model");
    private final ProblemAiConfig secondaryConfig = new ProblemAiConfig().setModel("secondary-model");
    private final ProblemAiModelConfig primary = new ProblemAiModelConfig().setId("primary").setName("主模型")
            .setModel("primary-model").setPrimary(true);
    private final ProblemAiModelConfig secondary = new ProblemAiModelConfig().setId("secondary").setName("辅助模型")
            .setModel("secondary-model").setPrimary(false);
    private final ProblemAiRecord record = new ProblemAiRecord().setId(50L).setStatus("running");
    private static final String STRICT = "#include \"testlib.h\"\nint main(int argc,char** argv){registerValidation(argc,argv);"
            + "int n=inf.readInt(1,10,\"n\");inf.readEoln();inf.readEof();printf(\"%d\\n\",n);}";

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
        when(problems.getById(7L)).thenReturn(problem);
        when(cases.list(any())).thenReturn(Collections.emptyList());
        when(configs.forModel(config, secondary)).thenReturn(secondaryConfig);
        when(gateway.complete(eq(config), anyList())).thenReturn("{\"code\":\"translated standard\"}");
        when(gateway.complete(eq(secondaryConfig), anyList())).thenReturn(new JSONObject().set("code", "int main(){return 0;}")
                .set("difficulty", difficulty()).toString());
        when(verification.submitForAi(any(), eq("admin"), eq("admin"))).thenReturn(100L);
        when(verification.submitComparisonForAi(any(), eq("admin"), eq("admin"))).thenReturn(101L);
        when(verification.comparisonResult(7L, 100L)).thenReturn(formal("SUCCESS", 3));
        when(verification.comparisonResult(7L, 101L)).thenReturn(formal("FAILED", 1));
        when(runner.run(eq(config), eq(problem), anyList(), any(), eq("admin"), eq("admin")))
                .thenReturn("{\"overall\":\"PASS\",\"summary\":\"主流程完成\"}");
    }

    @Test
    void selectedAuxiliaryGetsOneIndependentGenerationAndNeverRunsFullAnalysis() throws Exception {
        run(Arrays.asList(primary, secondary));
        JSONObject report = JSONUtil.parseObj(record.getResponse());
        assertEquals("PASS", report.getStr("overall"));
        assertEquals("success", record.getStatus());
        JSONArray models = report.getJSONArray("modelResults");
        assertEquals(2, models.size());
        assertEquals(3, models.getJSONObject(0).getInt("passedCases"));
        assertEquals(1, models.getJSONObject(1).getInt("passedCases"));
        assertEquals("FAILED", models.getJSONObject(1).getStr("status"));
        assertEquals(1400, models.getJSONObject(1).getJSONObject("difficulty").getInt("rating"));
        verify(runner, times(1)).run(eq(config), eq(problem), anyList(), any(), eq("admin"), eq("admin"));
        verify(gateway, times(1)).complete(eq(secondaryConfig), argThat(messages ->
                messages.stream().noneMatch(message -> message.get("content").contains(STRICT))));
        verify(verification, times(1)).submitForAi(any(), anyString(), anyString());
        verify(verification, times(1)).submitComparisonForAi(any(), anyString(), anyString());
    }

    @Test
    void onlyPrimarySelectionDoesNotCallOtherModels() throws Exception {
        run(Collections.singletonList(primary));
        verify(gateway, never()).complete(eq(secondaryConfig), anyList());
        verify(verification, never()).submitComparisonForAi(any(), anyString(), anyString());
        assertEquals(1, JSONUtil.parseObj(record.getResponse()).getJSONArray("modelResults").size());
    }

    @Test
    void auxiliaryProviderFailureKeepsPrimaryReport() throws Exception {
        when(gateway.complete(eq(secondaryConfig), anyList())).thenThrow(new IllegalStateException("provider unavailable"));
        run(Arrays.asList(primary, secondary));
        JSONObject report = JSONUtil.parseObj(record.getResponse());
        assertEquals("PASS", report.getStr("overall"));
        assertEquals("ERROR", report.getJSONArray("modelResults").getJSONObject(1).getStr("status"));
        verify(verification, never()).submitComparisonForAi(any(), anyString(), anyString());
    }

    @Test
    void primaryAnalysisFailureDoesNotDiscardFormalResultsOrSkipAuxiliary() throws Exception {
        when(runner.run(any(), any(), anyList(), any(), anyString(), anyString()))
                .thenThrow(new IllegalStateException("analysis unavailable"));
        run(Arrays.asList(primary, secondary));
        JSONObject report = JSONUtil.parseObj(record.getResponse());
        assertEquals("FAIL", report.getStr("overall"));
        assertEquals(3, report.getJSONArray("modelResults").getJSONObject(0).getInt("passedCases"));
        verify(gateway).complete(eq(secondaryConfig), anyList());
    }

    @Test
    void strictStandardRequiredOnlyForPrimaryAndDifficultyNeverInvented() {
        String plain = new JSONObject().set("code", "int main(){return 0;}").toString();
        assertThrows(Exception.class, () -> ReflectionTestUtils.invokeMethod(manager, "parseProgram", plain, true));
        ProblemAiGeneratedProgramVO auxiliary = ReflectionTestUtils.invokeMethod(manager, "parseProgram", plain, false);
        assertNull(auxiliary.getDifficulty());
        ProblemAiGeneratedProgramVO strict = ReflectionTestUtils.invokeMethod(manager, "parseProgram",
                new JSONObject().set("code", STRICT).set("difficulty", difficulty()).toString(), true);
        assertEquals(STRICT, strict.getCode());
        assertNull(strict.getValidatorCode());
        assertEquals(1400, strict.getDifficulty().getInt("rating"));
    }

    @Test
    void changedTestDataCannotProducePassingReport() throws Exception {
        when(problems.getById(7L)).thenReturn(new Problem().setId(7L).setCaseVersion("v2"));
        run(Arrays.asList(primary, secondary));
        assertEquals("FAIL", JSONUtil.parseObj(record.getResponse()).getStr("overall"));
        verifyZeroInteractions(verification, runner, gateway);
    }

    private void run(List<ProblemAiModelConfig> models) {
        ProblemAiGeneratedProgramVO standard = new ProblemAiGeneratedProgramVO().setRecordId(49L)
                .setLanguage("C++ 17").setCode(STRICT).setDifficulty(difficulty());
        ReflectionTestUtils.invokeMethod(manager, "runFullValidation", config, models, problem,
                standard, record, "admin", "admin");
    }

    private JSONObject formal(String status, int passed) {
        return new JSONObject().set("status", status).set("judgeStatus", "SUCCESS".equals(status) ? 0 : -1)
                .set("totalCases", 3).set("passedCases", passed).set("failedCases", 3 - passed)
                .set("pendingCases", 0).set("caseResults", new JSONArray());
    }

    private JSONObject difficulty() {
        return new JSONObject().set("rating", 1400).set("rangeLow", 1300).set("rangeHigh", 1500)
                .set("confidence", "medium").set("reason", "基础建模与边界处理");
    }
}
