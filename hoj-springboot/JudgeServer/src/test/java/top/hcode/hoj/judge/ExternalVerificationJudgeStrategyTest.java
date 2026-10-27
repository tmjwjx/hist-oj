package top.hcode.hoj.judge;

import cn.hutool.json.JSONArray;
import cn.hutool.json.JSONObject;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.test.util.ReflectionTestUtils;
import top.hcode.hoj.common.exception.SystemError;
import top.hcode.hoj.dao.JudgeCaseEntityService;
import top.hcode.hoj.dao.JudgeEntityService;
import top.hcode.hoj.judge.entity.LanguageConfig;
import top.hcode.hoj.pojo.entity.judge.Judge;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.util.Constants;

import java.util.Arrays;
import java.util.Map;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

class ExternalVerificationJudgeStrategyTest {
    private final JudgeStrategy strategy = new JudgeStrategy();
    private final ProblemTestCaseUtils data = mock(ProblemTestCaseUtils.class);
    private final JudgeRun run = mock(JudgeRun.class);
    private final VerificationSampleCaseService samples = mock(VerificationSampleCaseService.class);
    private final Problem problem = new Problem().setId(31L).setType(Constants.Contest.TYPE_OI.getCode())
            .setJudgeMode("default").setJudgeCaseMode("default").setCaseVersion("v1").setIoScore(100).setDifficulty(2);
    private final Judge judge = new Judge().setSubmitId(11L).setUid("creator").setLanguage("JavaScript")
            .setCode("answer").setIsProblemVerification(true).setSubmissionType("external_verification");
    private final JSONObject info = new JSONObject().set("testCases", new JSONArray());

    @BeforeEach
    void setup() throws Exception {
        LanguageConfigLoader languages = mock(LanguageConfigLoader.class);
        JudgeCaseEntityService cases = mock(JudgeCaseEntityService.class);
        ReflectionTestUtils.setField(strategy, "judgeEntityService", mock(JudgeEntityService.class));
        ReflectionTestUtils.setField(strategy, "JudgeCaseEntityService", cases);
        ReflectionTestUtils.setField(strategy, "languageConfigLoader", languages);
        ReflectionTestUtils.setField(strategy, "problemTestCaseUtils", data);
        ReflectionTestUtils.setField(strategy, "judgeRun", run);
        ReflectionTestUtils.setField(strategy, "verificationSampleCaseService", samples);
        when(languages.getLanguageConfigByName("JavaScript")).thenReturn(new LanguageConfig());
        when(data.loadTestCaseInfo(eq(31L), anyString(), eq("v1"), eq("default"), any())).thenReturn(info);
        when(cases.saveBatch(anyCollection())).thenReturn(true);
        when(run.judgeAllCase(eq(11L), same(problem), eq("JavaScript"), anyString(), same(info),
                isNull(), eq("answer"), eq(false), anyString())).thenReturn(Arrays.asList(
                result(1, 1, 0), result(2, 1, -1), result(3, 2, 0)));
    }

    @Test
    void externalSubmissionKeepsLowestAndAverageSubtaskScoringWithoutExtraSamples() throws Exception {
        for (String mode : Arrays.asList("subtask_lowest", "subtask_average")) {
            info.set("judgeCaseMode", mode);
            Map<String, Object> external = strategy.judge(problem, judge);
            assertEquals(mode.equals("subtask_lowest") ? 50 : 75, external.get("score"));
            assertEquals(Constants.Judge.STATUS_PARTIAL_ACCEPTED.getStatus(), external.get("code"));
            verify(run).judgeAllCase(eq(11L), same(problem), eq("JavaScript"), anyString(), same(info),
                    isNull(), eq("answer"), eq(false), eq(mode));
        }
        verify(samples, never()).append(any(), any(), any(), any());
    }

    @Test
    void creatorAiComparisonAndHistoricalVerificationStillAppendSamplesAndRunAllCases() throws Exception {
        info.set("judgeCaseMode", "subtask_lowest");
        for (String type : Arrays.asList(null, "creator_validation", "ai_validation", "ai_comparison")) {
            Map<String, Object> result = strategy.judge(problem, judge.setSubmissionType(type));
            assertEquals(100, result.get("score"));
        }
        verify(samples, times(4)).append(same(problem), eq(11L), anyString(), same(info));
        verify(run, times(4)).judgeAllCase(eq(11L), same(problem), eq("JavaScript"), anyString(), same(info),
                isNull(), eq("answer"), eq(false), eq("default"));
    }

    @Test
    void externalSystemFailureDoesNotIncludePrivateJudgeDiagnostics() throws Exception {
        when(data.loadTestCaseInfo(anyLong(), anyString(), anyString(), anyString(), any()))
                .thenThrow(new SystemError("private path", "private output", "private checker"));
        Map<String, Object> result = strategy.judge(problem, judge);
        assertEquals(Constants.Judge.STATUS_SYSTEM_ERROR.getStatus(), result.get("code"));
        assertFalse(result.get("errMsg").toString().contains("private"));
    }

    private JSONObject result(int sequence, int group, int status) {
        return new JSONObject().set("seq", sequence).set("groupNum", group).set("score", 50)
                .set("status", status).set("time", 1).set("memory", 1);
    }
}
