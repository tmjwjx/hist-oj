package top.hcode.hoj.service.problem;

import org.junit.jupiter.api.Test;
import top.hcode.hoj.pojo.entity.problem.Problem;

import java.util.Arrays;
import java.util.List;
import java.util.function.Consumer;

import static org.junit.jupiter.api.Assertions.*;

class ProblemVerificationChangeDetectorTest {
    @Test
    void ordinaryMetadataAndSampleExplanationsKeepVerification() {
        Problem before = problem();
        Problem after = problem().setProblemId("NEW-ID").setTitle("新标题").setAuthor("new-author")
                .setDescription("新题面").setInput("输入说明").setOutput("输出说明")
                .setHint("新提示").setSource("新来源").setAuth(2).setDifficulty(3)
                .setCodeShare(true).setOpenCaseResult(true).setModifiedUser("editor")
                .setExamples("<input>1 2\n</input><output>3\n</output><explanation>新的样例说明</explanation>");

        assertFalse(ProblemVerificationChangeDetector.changed(before, after));
        assertFalse(ProblemVerificationChangeDetector.changed(after, after));
    }

    @Test
    void eachJudgingInputInvalidatesVerification() {
        List<Consumer<Problem>> changes = Arrays.asList(
                value -> value.setCaseVersion("new-package"),
                value -> value.setExamples("<input>2 3\n</input><output>3\n</output>"),
                value -> value.setExamples("<input>1 2\n</input><output>5\n</output>"),
                value -> value.setExamples(""),
                value -> value.setTimeLimit(2000),
                value -> value.setMemoryLimit(512),
                value -> value.setStackLimit(256),
                value -> value.setIsRemoveEndBlank(true),
                value -> value.setIsFileIO(true),
                value -> value.setType(1),
                value -> value.setJudgeMode("spj"),
                value -> value.setJudgeCaseMode("ergodic_without_error"),
                value -> value.setUserExtraFile("{\"header.h\":\"new contents\"}"),
                value -> value.setJudgeExtraFile("{\"checker.h\":\"new contents\"}"));

        for (int i = 0; i < changes.size(); i++) {
            Problem after = problem();
            changes.get(i).accept(after);
            assertTrue(ProblemVerificationChangeDetector.changed(problem(), after), "critical change " + i);
        }
    }

    @Test
    void fileNamesInvalidateOnlyWhenFileIoIsEnabled() {
        Problem before = problem();
        Problem after = problem().setIoReadFileName("new.in").setIoWriteFileName("new.out");
        assertFalse(ProblemVerificationChangeDetector.changed(before, after));
        before.setIsFileIO(true);
        after.setIsFileIO(true);
        assertTrue(ProblemVerificationChangeDetector.changed(before, after));
    }

    @Test
    void checkerCodeAndLanguageInvalidateSpecialJudging() {
        Problem before = problem().setJudgeMode("spj").setSpjCode("old checker").setSpjLanguage("C++");
        Problem after = problem().setJudgeMode("spj").setSpjCode("new checker").setSpjLanguage("C++");
        assertTrue(ProblemVerificationChangeDetector.changed(before, after));
        after.setSpjCode("old checker").setSpjLanguage("C++17");
        assertTrue(ProblemVerificationChangeDetector.changed(before, after));
    }

    @Test
    void obsoleteCheckerFieldsDoNotInvalidateDefaultJudging() {
        assertFalse(ProblemVerificationChangeDetector.changed(
                problem().setSpjCode("").setSpjLanguage("C++"), problem()));
    }

    @Test
    void equivalentExtraFilesDoNotInvalidateFromJsonFormattingOrOrdering() {
        Problem before = problem().setUserExtraFile("{\"a.h\":\"a\",\"b.h\":\"b\"}").setJudgeExtraFile(null);
        Problem after = problem().setUserExtraFile("{ \"b.h\" : \"b\", \"a.h\" : \"a\" }").setJudgeExtraFile("{}");
        assertFalse(ProblemVerificationChangeDetector.changed(before, after));
        after.setUserExtraFile("{\"a.h\":\"changed\",\"b.h\":\"b\"}");
        assertTrue(ProblemVerificationChangeDetector.changed(before, after));
    }

    @Test
    void sampleMarkupSpacingAndLineEndingsDoNotInvalidateButDataWhitespaceDoes() {
        Problem after = problem().setExamples("\n<input>1 2\r\n</input>\n<output>3\r\n</output>\n");
        assertFalse(ProblemVerificationChangeDetector.changed(problem(), after));
        after.setExamples("<input>1 2 \n</input><output>3\n</output>");
        assertTrue(ProblemVerificationChangeDetector.changed(problem(), after));
    }

    @Test
    void malformedSamplesCannotReuseVerification() {
        Problem after = problem().setExamples("<input>1 2\n</input><output>3\n</output>trailing garbage");
        assertTrue(ProblemVerificationChangeDetector.changed(problem(), after));
        assertTrue(ProblemVerificationChangeDetector.changed(
                problem().setExamples("invalid-before"), problem().setExamples("invalid-after")));
    }

    private Problem problem() {
        return new Problem().setId(12L).setProblemId("P12").setCaseVersion("verified-version")
                .setType(0).setJudgeMode("default").setJudgeCaseMode("default")
                .setTimeLimit(1000).setMemoryLimit(256).setStackLimit(128)
                .setIsRemoveEndBlank(false).setIsFileIO(false)
                .setExamples("<input>1 2\n</input><output>3\n</output>");
    }
}
