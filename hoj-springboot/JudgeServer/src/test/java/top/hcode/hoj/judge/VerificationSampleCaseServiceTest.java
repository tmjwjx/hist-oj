package top.hcode.hoj.judge;

import cn.hutool.json.JSONArray;
import cn.hutool.json.JSONObject;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;
import org.springframework.util.DigestUtils;
import top.hcode.hoj.judge.entity.JudgeDTO;
import top.hcode.hoj.judge.entity.JudgeGlobalDTO;
import top.hcode.hoj.judge.entity.SandBoxRes;
import top.hcode.hoj.judge.task.DefaultJudge;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.util.Constants;

import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.List;

import static org.junit.jupiter.api.Assertions.*;

class VerificationSampleCaseServiceTest {
    @TempDir Path directory;
    private final VerificationSampleCaseService service = new VerificationSampleCaseService();

    @Test
    void addsMissingFinalLineEndingsAndKeepsWhitespaceAndFormalDataUntouched() throws Exception {
        Path original = directory.resolve("1.in");
        Files.write(original, "9  \n\n".getBytes(StandardCharsets.UTF_8));
        Files.write(directory.resolve("1.out"), "3".getBytes(StandardCharsets.UTF_8));
        JSONObject info = info();
        String originalMetadata = info.getJSONArray("testCases").getJSONObject(0).toString();
        List<String> created = service.append(problem("1  2 ", "3 "), 71L, directory.toString(), info);
        assertEquals("1  2 \n", read(created.get(0)));
        assertEquals("3 \n", read(created.get(1)));
        assertEquals("9  \n\n", read("1.in"));
        assertEquals("3", read("1.out"));
        assertEquals(originalMetadata, info.getJSONArray("testCases").getJSONObject(0).toString());
        assertEquals(2, info.getInt("testCasesSize"));
    }

    @Test
    void existingFinalLineEndingIsNotDuplicated() throws Exception {
        List<String> created = service.append(problem("1\n", "1\n"), 72L, directory.toString(), info());
        assertEquals("1\n", read(created.get(0)));
        assertEquals("1\n", read(created.get(1)));
    }

    @Test
    void existingBlankLinesArePreserved() throws Exception {
        List<String> created = service.append(problem(" 1 \n\n", " 1 \n\n"), 73L, directory.toString(), info());
        assertEquals(" 1 \n\n", read(created.get(0)));
        assertEquals(" 1 \n\n", read(created.get(1)));
    }

    @Test
    void carriageReturnLineEndingsNormalizeWithoutDroppingBlankLines() throws Exception {
        List<String> created = service.append(problem("1\r\n2\r\n\r\n", "1\r\n"), 74L, directory.toString(), info());
        assertEquals("1\n2\n\n", read(created.get(0)));
        assertEquals("1\n", read(created.get(1)));
    }

    @Test
    void emptyOutputRemainsEmptyAndAcceptsNoOutput() throws Exception {
        JSONObject info = info();
        List<String> created = service.append(problem("1", ""), 77L, directory.toString(), info);
        assertEquals("", read(created.get(1)));
        assertEquals(0, info.getJSONArray("testCases").getJSONObject(1).getInt("outputSize"));
        assertEquals(Constants.Judge.STATUS_ACCEPTED.getStatus(), judgeStatus(info, 2, ""));
    }

    @Test
    void sampleAcceptsFinalLineEndingButRejectsWrongValuesAndInternalWhitespace() throws Exception {
        JSONObject info = info();
        service.append(problem("1", "3 4"), 78L, directory.toString(), info);
        assertEquals(Constants.Judge.STATUS_ACCEPTED.getStatus(), judgeStatus(info, 2, "3 4\n"));
        assertEquals(Constants.Judge.STATUS_WRONG_ANSWER.getStatus(), judgeStatus(info, 2, "3 5\n"));
        assertEquals(Constants.Judge.STATUS_PRESENTATION_ERROR.getStatus(), judgeStatus(info, 2, "3  4\n"));
    }

    @Test
    void formalOutputStillRequiresExactFinalLineEnding() throws Exception {
        JSONObject info = info();
        service.append(problem("1", "3"), 79L, directory.toString(), info);
        assertEquals(Constants.Judge.STATUS_ACCEPTED.getStatus(), judgeStatus(info, 1, "3"));
        assertEquals(Constants.Judge.STATUS_PRESENTATION_ERROR.getStatus(), judgeStatus(info, 1, "3\n"));
        assertEquals(Constants.Judge.STATUS_ACCEPTED.getStatus(), judgeStatus(info, 2, "3\n"));
    }

    @Test
    void interactiveConversationDoesNotBecomeHiddenInput() throws Exception {
        JSONObject info = info().set("testCasesSize", 1);
        Problem problem = problem("judge: 3\n", "contestant: ? 1 2\n").setJudgeMode("interactive");
        assertTrue(service.append(problem, 75L, directory.toString(), info).isEmpty());
        assertEquals(1, info.getJSONArray("testCases").size());
        assertEquals(1, info.getInt("testCasesSize"));
        assertFalse(Files.exists(directory.resolve("__verification_sample_75_1.in")));
        problem.setExamples("对话示例无需转换为静态 input/output");
        assertTrue(service.append(problem, 76L, directory.toString(), info).isEmpty());
    }

    private Problem problem(String input, String output) {
        return new Problem().setJudgeMode("default")
                .setExamples("<input>" + input + "</input><output>" + output + "</output>");
    }

    private JSONObject info() {
        String outputMd5 = DigestUtils.md5DigestAsHex("3".getBytes(StandardCharsets.UTF_8));
        return new JSONObject().set("testCases", new JSONArray().put(new JSONObject()
                .set("inputName", "1.in").set("outputName", "1.out")
                .set("outputMd5", outputMd5).set("allStrippedOutputMd5", outputMd5)
                .set("EOFStrippedOutputMd5", outputMd5)));
    }

    private Integer judgeStatus(JSONObject info, int caseNumber, String output) {
        SandBoxRes result = SandBoxRes.builder().status(Constants.Judge.STATUS_ACCEPTED.getStatus())
                .time(1L).memory(1L).stdout(output).build();
        JudgeDTO judge = JudgeDTO.builder().testCaseNum(caseNumber).build();
        JudgeGlobalDTO global = JudgeGlobalDTO.builder().testCaseInfo(info)
                .maxTime(1000L).maxMemory(128L).removeEOLBlank(false).needUserOutputFile(false).build();
        return new DefaultJudge().checkResult(result, judge, global).getInt("status");
    }

    private String read(String name) throws Exception {
        return new String(Files.readAllBytes(directory.resolve(name)), StandardCharsets.UTF_8);
    }
}
