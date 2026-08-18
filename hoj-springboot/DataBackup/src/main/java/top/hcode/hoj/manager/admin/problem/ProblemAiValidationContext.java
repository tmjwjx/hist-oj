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
            points.add(new ProblemAiTestPoint().setIndex(i + 1)
                    .setCaseId(problemCase == null ? result == null ? null : result.getCaseId() : problemCase.getId())
                    .setInput(problemCase == null ? "验证样例或动态测试点，输入内容见题面样例" : problemCase.getInput())
                    .setExpectedOutput(problemCase == null ? "验证样例或动态测试点，输出内容见题面样例" : problemCase.getOutput())
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
