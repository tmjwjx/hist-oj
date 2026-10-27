package top.hcode.hoj.manager.admin.problem;

import org.junit.jupiter.api.Test;
import top.hcode.hoj.pojo.dto.ProblemAiValidateDTO;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.pojo.entity.problem.ProblemAiConfig;

import java.util.Collections;

import static org.junit.jupiter.api.Assertions.*;

class ProblemAiPromptFactoryTest {
    private final ProblemAiPromptFactory prompts = new ProblemAiPromptFactory();

    @Test
    void legacyDefaultIsNotAppendedAsConflictingCustomInstruction() {
        String oldDefault = prompts.defaultSystemPrompt()
                .replace("并检查标准程序是否按题面输入格式执行 testlib 的", "并按题面输入格式执行等价于 testlib 的")
                .replace("一体化标准程序校验失败", "验证器失败");
        assertFalse(prompts.programSystemPrompt(oldDefault).contains("管理员补充规则"));
        assertFalse(prompts.systemPrompt(oldDefault).contains("验证器失败"));
    }

    @Test
    void customInstructionsRemainButCurrentWorkflowRequirementsComeLast() {
        String prompt = prompts.programSystemPrompt("管理员要求检查整数溢出，同时生成独立 validator");
        assertTrue(prompt.contains("管理员要求检查整数溢出"));
        assertTrue(prompt.indexOf("只生成一份完整标准程序") > prompt.indexOf("同时生成独立 validator"));
        assertTrue(prompt.endsWith("禁止生成分离校验器或 validator 字段。"));
    }

    @Test
    void primaryAndAuxiliaryGenerationReceiveDistinctRequirements() {
        ProblemAiConfig config = new ProblemAiConfig();
        Problem problem = new Problem().setJudgeMode("default").setTitle("独立推导测试");
        String primary = prompts.standardProgramPrompt(config, problem, "C++ 17", 0);
        String auxiliary = prompts.auxiliaryProgramPrompt(config, problem);
        assertTrue(primary.contains("registerValidation(argc, argv)"));
        assertTrue(primary.contains("inf.readEof()"));
        assertTrue(auxiliary.contains("不使用其他模型的代码、结论或难度估计"));
        assertTrue(auxiliary.contains("不需要 testlib 或 validator"));
        assertFalse(auxiliary.contains("参考标准程序"));
    }

    @Test
    void interactiveGenerationUsesProtocolClientRatherThanStaticInputValidator() {
        Problem problem = new Problem().setJudgeMode("interactive");
        ProblemAiConfig config = new ProblemAiConfig();
        String primary = prompts.standardProgramPrompt(config, problem, "C++ 17", 0);
        assertTrue(primary.contains("禁止 registerValidation、inf.readEof"));
        assertFalse(primary.contains("必须 #include \"testlib.h\""));
        assertFalse(primary.contains("最开始调用 registerValidation"));
        assertTrue(primary.contains("不能读取测试文件、隐藏答案"));
        assertTrue(primary.contains("每次查询和最终答案后及时 flush"));
        assertTrue(primary.contains("显式控制查询上限"));
        assertTrue(prompts.programSystemPrompt(null, problem).contains("当前为交互题，主模型 std 生成普通交互协议客户端"));
        assertTrue(prompts.systemPrompt(null, problem).contains("当前为交互题，std 是协议客户端"));
        assertTrue(prompts.auxiliaryProgramPrompt(config, problem).contains("每次输出立即 flush"));
        assertTrue(prompts.multiLanguageProgramPrompt(config, problem, "Java", "client")
                .contains("不能读取隐藏测试输入或答案"));
    }

    @Test
    void spjRetainsIntegratedValidationAndReceivesRequiredReviewChecksAndProbes() {
        Problem problem = new Problem().setJudgeMode("spj").setSpjLanguage("C++ 17")
                .setSpjCode("registerTestlibCmd(argc, argv);");
        String primary = prompts.standardProgramPrompt(new ProblemAiConfig(), problem, "C++ 17", 0);
        assertTrue(primary.contains("最开始调用 registerValidation(argc, argv)"));
        assertTrue(primary.contains("末尾必须 inf.readEof()"));
        String prompt = validationPrompt(problem);
        assertTrue(prompt.contains("\"judgeProgramReview\""));
        assertTrue(prompt.contains("acceptance、rejection、numeric、termination"));
        assertTrue(prompt.contains("alternate_valid（expectation=ACCEPT"));
        assertTrue(prompt.contains("invalid_output（expectation=REJECT"));
        assertTrue(prompt.contains("NaN/Inf"));
        assertTrue(prompt.contains("registerTestlibCmd 与 argv/inf/ouf/ans"));
        assertTrue(prompt.contains("不能把编译错误、运行异常、超时或系统错误当作拒错成功"));
        assertTrue(prompt.contains("不得以 std Accepted 替代"));
        assertTrue(prompt.contains("status 至少 WARN"));
    }

    @Test
    void interactiveReviewChecksDeadlockPrematureExitAndVerdict() {
        Problem problem = new Problem().setJudgeMode("interactive");
        String prompt = validationPrompt(problem);
        assertTrue(prompt.contains("protocol、limits、termination、verdict"));
        assertTrue(prompt.contains("protocol_violation（expectation=REJECT"));
        assertTrue(prompt.contains("premature_exit（expectation=REJECT"));
        assertTrue(prompt.contains("registerInteraction 与 argv/inf/ouf/ans"));
        assertTrue(prompt.contains("隐藏输入/答案泄露"));
        assertTrue(prompt.contains("不能把选手异常终止或交互器异常当作正确答案"));
        assertTrue(prompt.contains("禁止读取隐藏输入或使用 registerValidation/inf.readEof"));
        String points = prompts.testPointPrompt(new ProblemAiConfig(), problem, new ProblemAiValidateDTO(),
                Collections.emptyList(), 0);
        assertTrue(points.contains("不能要求 std 使用 registerValidation、inf.readEof 或读取隐藏输入"));
        assertFalse(points.contains("输入格式检查必须采用 testlib/registerValidation 的严格规则"));
    }

    @Test
    void ordinaryProblemsDoNotAcquireJudgeProgramReviewRequirements() {
        String prompt = validationPrompt(new Problem().setJudgeMode("default"));
        assertFalse(prompt.contains("\"judgeProgramReview\""));
        assertFalse(prompt.contains("\"probes\""));
        assertFalse(prompt.contains("判题程序完整源码"));
    }

    @Test
    void completeJudgeSourceAndExtraFilesAreSentEvenBeyondOldTruncationLimit() {
        String source = String.join("", Collections.nCopies(14000, "x")) + "SOURCE_END";
        String extra = "{\"helper.h\":\"" + String.join("", Collections.nCopies(14000, "y")) + "EXTRA_END\"}";
        Problem problem = new Problem().setJudgeMode("spj").setSpjCode(source).setJudgeExtraFile(extra);
        String prompt = validationPrompt(problem);
        assertTrue(prompt.contains(source));
        assertTrue(prompt.contains(extra));
        assertFalse(prompt.contains("[内容已截断]"));
        String generation = prompts.standardProgramPrompt(new ProblemAiConfig(), problem, "C++ 17", 0);
        assertTrue(generation.contains(source));
        assertTrue(generation.contains(extra));
        assertTrue(prompt.contains("内容超过模型可审查范围或无法确认关键路径时标记 WARN"));
    }

    private String validationPrompt(Problem problem) {
        return prompts.validationPrompt(new ProblemAiConfig(), problem, Collections.emptyList(),
                new ProblemAiValidateDTO(),
                new ProblemAiValidationContext.Snapshot(1L, 0, "Accepted", "", Collections.emptyList()));
    }
}
