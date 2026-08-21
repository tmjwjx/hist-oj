package top.hcode.hoj.manager.admin.problem;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import lombok.AllArgsConstructor;
import lombok.Getter;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.dao.judge.JudgeCaseEntityService;
import top.hcode.hoj.dao.judge.JudgeEntityService;
import top.hcode.hoj.dao.problem.ProblemCaseEntityService;
import top.hcode.hoj.dao.problem.ProblemVerificationEntityService;
import top.hcode.hoj.pojo.dto.ProblemAiValidateDTO;
import top.hcode.hoj.pojo.entity.judge.Judge;
import top.hcode.hoj.pojo.entity.judge.JudgeCase;
import top.hcode.hoj.pojo.entity.problem.ProblemCase;
import top.hcode.hoj.pojo.entity.problem.ProblemVerification;
import top.hcode.hoj.utils.Constants;
import top.hcode.hoj.utils.ProblemVerificationConstants;

import javax.annotation.Resource;
import java.io.File;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.security.MessageDigest;
import java.util.*;

@Component
public class ProblemAiValidationContext {
    @Resource private ProblemVerificationEntityService verificationService;
    @Resource private ProblemCaseEntityService problemCaseService;
    @Resource private JudgeEntityService judgeService;
    @Resource private JudgeCaseEntityService judgeCaseService;

    public Snapshot load(ProblemAiValidateDTO input) throws StatusFailException {
        ProblemVerification verification = verificationService.getOne(
                new QueryWrapper<ProblemVerification>().eq("pid", input.getPid()), false);
        if (verification == null || verification.getSubmitId() == null) {
            throw new StatusFailException("AI 验题前必须先完成标准程序的正式全测试点判题");
        }
        Judge judge = judgeService.getById(verification.getSubmitId());
        if (judge == null || ProblemVerificationConstants.isJudgeRunning(judge.getStatus())) {
            throw new StatusFailException("标准程序仍在判题，请等待全部测试点完成");
        }
        if (!Objects.equals(judge.getLanguage(), input.getLanguage())
                || !Objects.equals(judge.getCode(), input.getStandardProgram())) {
            throw new StatusFailException("AI 验题代码与最近正式判题代码不一致，请重新执行一键 AI 验题");
        }

        List<ProblemCase> cases = problemCaseService.list(new QueryWrapper<ProblemCase>()
                .eq("pid", input.getPid()).eq("status", 0).orderByAsc("id"));
        List<JudgeCase> results = judgeCaseService.list(new QueryWrapper<JudgeCase>()
                .eq("submit_id", judge.getSubmitId()).orderByAsc("seq"));
        Map<Integer, JudgeCase> resultBySeq = new HashMap<>();
        for (JudgeCase item : results) resultBySeq.put(item.getSeq(), item);

        List<ProblemAiTestPoint> points = new ArrayList<>();
        int total = Math.max(cases.size(), results.size());
        for (int i = 0; i < total; i++) {
            ProblemCase problemCase = i < cases.size() ? cases.get(i) : null;
            JudgeCase result = resultBySeq.get(i + 1);
            String stderr = result == null ? "" : firstNonBlank(result.getStderr(), result.getUserOutput());
            String inputValue = problemCase == null ? "验证样例或动态测试点，输入内容见题面样例" : problemCase.getInput();
            String expectedOutputValue = problemCase == null ? "验证样例或动态测试点，输出内容见题面样例" : problemCase.getOutput();
            String resolvedInput = resolveTestcaseFile(inputValue, input.getPid());
            String resolvedOutput = resolveTestcaseFile(expectedOutputValue, input.getPid());
            boolean concrete = problemCase != null;
            points.add(new ProblemAiTestPoint().setIndex(i + 1)
                    .setCaseId(problemCase == null ? result == null ? null : result.getCaseId() : problemCase.getId())
                    .setInput(resolvedInput)
                    .setExpectedOutput(resolvedOutput)
                    .setInputAvailable(concrete && isContentAvailable(inputValue, resolvedInput))
                    .setExpectedOutputAvailable(concrete && isContentAvailable(expectedOutputValue, resolvedOutput))
                    .setConcrete(concrete)
                    .setGroupNum(problemCase == null ? result == null ? null : result.getGroupNum() : problemCase.getGroupNum())
                    .setScore(problemCase == null ? result == null ? null : result.getScore() : problemCase.getScore())
                    .setExecuted(result != null && !isPending(result.getStatus()))
                    .setJudgeStatus(result == null ? null : result.getStatus())
                    .setJudgeStatusText(result == null ? "NOT_RUN" : statusName(result.getStatus()))
                    .setTime(result == null ? null : result.getTime())
                    .setMemory(result == null ? null : result.getMemory())
                    .setStderr(stderr));
        }
        return new Snapshot(judge.getSubmitId(), judge.getStatus(), statusName(judge.getStatus()),
                judge.getErrorMessage(), points);
    }

    /**
     * Build read-only evidence for a report recheck.  The normal AI validation
     * path already resolves file-backed cases, but report-only rechecks used to
     * receive only values such as {@code 1.in}/{@code 1.out}.  Include the
     * resolved content and hashes so the model can distinguish an unchanged
     * case set from genuinely missing evidence without rerunning the judge.
     */
    public String currentTestcaseEvidence(Long pid, List<ProblemCase> cases, Long submitId) {
        StringBuilder text = new StringBuilder("当前实体测试点数量：").append(cases == null ? 0 : cases.size());
        if (cases == null) return text.toString();
        Map<Integer, JudgeCase> judgeCases = new HashMap<>();
        if (submitId != null) {
            List<JudgeCase> results = judgeCaseService.list(new QueryWrapper<JudgeCase>()
                    .eq("submit_id", submitId).orderByAsc("seq"));
            for (JudgeCase item : results) judgeCases.put(item.getSeq(), item);
        }
        boolean allMatch = !cases.isEmpty() && judgeCases.size() >= cases.size();
        for (int i = 0; i < cases.size(); i++) {
            ProblemCase item = cases.get(i);
            String input = resolveTestcaseFile(item.getInput(), pid);
            String output = resolveTestcaseFile(item.getOutput(), pid);
            JudgeCase formal = judgeCases.get(i + 1);
            String formalInput = formal == null ? "" : resolveTestcaseFile(formal.getInputData(), pid);
            String formalOutput = formal == null ? "" : resolveTestcaseFile(formal.getOutputData(), pid);
            boolean inputMatchesFormal = formal != null && Objects.equals(normalizeEvidence(input), normalizeEvidence(formalInput));
            boolean outputMatchesFormal = formal != null && Objects.equals(normalizeEvidence(output), normalizeEvidence(formalOutput));
            allMatch = allMatch && inputMatchesFormal && outputMatchesFormal;
            text.append("\n测试点 ").append(i + 1)
                    .append("，caseId=").append(item.getId())
                    .append("，inputName=").append(item.getInput())
                    .append("，outputName=").append(item.getOutput())
                    .append("，inputAvailable=").append(isContentAvailable(item.getInput(), input))
                    .append("，outputAvailable=").append(isContentAvailable(item.getOutput(), output))
                    .append("，formalJudgeStatus=").append(formal == null ? "NOT_RUN" : formal.getStatus())
                    .append("，matchesFormalSubmission=").append(inputMatchesFormal && outputMatchesFormal)
                    .append("，inputSha256=").append(sha256(input))
                    .append("，outputSha256=").append(sha256(output))
                    .append("\n输入内容：").append(limitEvidence(input))
                    .append("\n标准输出内容：").append(limitEvidence(output));
        }
        text.append("\n当前文件与 submitId=").append(submitId)
                .append(" 的正式判题输入/标准输出逐点完全一致：").append(allMatch);
        return text.toString();
    }

    /**
     * ProblemCase.input/output are either the actual content (database
     * fallback) or a testcase filename/path (normal file-backed judging).  The
     * AI prompt must receive the concrete file content whenever it is safely
     * available; showing "1.in" made the model report a spurious WARN even
     * though the formal judge had already Accepted the point.
     */
    private String resolveTestcaseFile(String value, Long pid) {
        if (isEmpty(value) || pid == null) return value;
        File root = new File(Constants.File.TESTCASE_BASE_FOLDER.getPath(), "problem_" + pid);
        List<File> candidates = new ArrayList<>();
        File supplied = new File(value.trim());
        if (supplied.isAbsolute()) candidates.add(supplied);
        candidates.add(new File(root, supplied.getName()));
        for (File candidate : candidates) {
            try {
                File canonicalRoot = root.getCanonicalFile();
                File canonical = candidate.getCanonicalFile();
                String rootPath = canonicalRoot.getPath() + File.separator;
                if (!canonical.getPath().startsWith(rootPath) || !canonical.isFile()) continue;
                byte[] content = Files.readAllBytes(canonical.toPath());
                return new String(content, StandardCharsets.UTF_8);
            } catch (IOException ignored) {
                // Keep the original DB value as a last-resort evidence marker.
            }
        }
        return value;
    }

    private boolean isEmpty(String value) {
        return value == null || value.trim().isEmpty();
    }

    private boolean isContentAvailable(String original, String resolved) {
        if (isEmpty(resolved)) return false;
        if (!Objects.equals(original, resolved)) return true;
        String text = original.trim();
        return !(text.matches("[^\\s]{1,160}\\.(?:in|out)")
                || text.startsWith(Constants.File.TESTCASE_BASE_FOLDER.getPath()));
    }

    private String limitEvidence(String value) {
        if (isEmpty(value)) return "[为空]";
        String text = value.trim();
        return text.length() > 1200 ? text.substring(0, 1200) + "…[截断]" : text;
    }

    private String sha256(String value) {
        if (value == null) return "";
        try {
            byte[] digest = MessageDigest.getInstance("SHA-256")
                    .digest(value.getBytes(StandardCharsets.UTF_8));
            StringBuilder hex = new StringBuilder();
            for (byte item : digest) hex.append(String.format("%02x", item));
            return hex.toString();
        } catch (Exception ignored) {
            return "不可用";
        }
    }

    private String normalizeEvidence(String value) {
        return value == null ? "" : value.replace("\r\n", "\n").trim();
    }

    private boolean isPending(Integer status) {
        return Objects.equals(status, Constants.Judge.STATUS_PENDING.getStatus())
                || Objects.equals(status, Constants.Judge.STATUS_COMPILING.getStatus())
                || Objects.equals(status, Constants.Judge.STATUS_JUDGING.getStatus());
    }

    private String statusName(Integer status) {
        for (Constants.Judge value : Constants.Judge.values()) {
            if (Objects.equals(value.getStatus(), status)) return value.getName();
        }
        return status == null ? "NOT_RUN" : "UNKNOWN(" + status + ")";
    }

    private String firstNonBlank(String... values) {
        for (String value : values) if (!StringUtils.isEmpty(value)) return value;
        return "";
    }

    @Getter
    @AllArgsConstructor
    public static class Snapshot {
        private final Long submitId;
        private final Integer status;
        private final String statusText;
        private final String judgeError;
        private final List<ProblemAiTestPoint> points;
    }
}
