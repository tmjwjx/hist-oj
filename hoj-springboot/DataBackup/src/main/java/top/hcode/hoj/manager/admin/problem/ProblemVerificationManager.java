package top.hcode.hoj.manager.admin.problem;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import cn.hutool.core.io.FileUtil;
import org.apache.shiro.SecurityUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.dao.contest.ContestProblemEntityService;
import top.hcode.hoj.dao.judge.JudgeEntityService;
import top.hcode.hoj.dao.problem.ProblemEntityService;
import top.hcode.hoj.dao.problem.ProblemCaseEntityService;
import top.hcode.hoj.dao.problem.ProblemVerificationDraftEntityService;
import top.hcode.hoj.dao.problem.ProblemVerificationEntityService;
import top.hcode.hoj.judge.self.JudgeDispatcher;
import top.hcode.hoj.pojo.dto.ProblemVerificationSubmitDTO;
import top.hcode.hoj.pojo.dto.LastAcceptedCodeVO;
import top.hcode.hoj.pojo.entity.contest.ContestProblem;
import top.hcode.hoj.pojo.entity.judge.Judge;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.pojo.entity.problem.ProblemVerification;
import top.hcode.hoj.pojo.entity.problem.ProblemVerificationDraft;
import top.hcode.hoj.pojo.vo.ContestVerificationVO;
import top.hcode.hoj.pojo.vo.ProblemVerificationVO;
import top.hcode.hoj.shiro.AccountProfile;
import top.hcode.hoj.utils.Constants;
import top.hcode.hoj.utils.ProblemVerificationConstants;
import top.hcode.hoj.utils.ShiroUtils;
import top.hcode.hoj.service.problem.ProblemVerificationLifecycle;
import top.hcode.hoj.service.problem.ProblemTestCaseSyncManager;
import top.hcode.hoj.service.problem.JudgeProgressService;

import java.util.ArrayList;
import java.util.Date;
import java.util.List;
import java.util.Objects;

@Component
public class ProblemVerificationManager {

    @Autowired
    private ProblemVerificationEntityService verificationService;

    @Autowired
    private ProblemEntityService problemEntityService;

    @Autowired
    private JudgeEntityService judgeEntityService;

    @Autowired
    private JudgeDispatcher judgeDispatcher;

    @Autowired
    private ContestProblemEntityService contestProblemEntityService;

    @Autowired
    private ProblemVerificationLifecycle verificationLifecycle;

    @Autowired
    private ProblemVerificationDraftEntityService draftService;

    @Autowired
    private ProblemCaseEntityService problemCaseEntityService;

    @Autowired
    private ProblemTestCaseSyncManager testCaseSyncManager;

    @Autowired
    private JudgeProgressService judgeProgressService;

    public ProblemVerificationVO getStatus(Long pid)
            throws StatusFailException, StatusForbiddenException {
        Problem problem = requireProblem(pid);
        checkOperator(problem);
        ProblemVerification record = find(pid);
        if (record == null) {
            verificationLifecycle.markTestCaseChanged(
                    problem.getId(), problem.getCaseVersion(), problem.getJudgeMode());
            record = find(pid);
        }
        refreshJudgeStatus(record);
        return toVO(problem, record);
    }

    public ProblemVerificationVO submit(ProblemVerificationSubmitDTO dto)
            throws StatusFailException, StatusForbiddenException {
        if (dto == null || dto.getPid() == null || StringUtils.isEmpty(dto.getLanguage())
                || StringUtils.isEmpty(dto.getCode()) || dto.getCode().length() > 65535) {
            throw new StatusFailException("题目、语言和标准程序不能为空，代码不能超过 65535 个字符");
        }
        Problem problem = requireProblem(dto.getPid());
        checkOperator(problem);
        ProblemVerification record = find(dto.getPid());
        if (record == null || !Objects.equals(record.getCaseVersion(), problem.getCaseVersion())) {
            throw new StatusFailException("题目测试数据尚未同步，请先重新保存题目");
        }
        if (!Objects.equals(record.getSyncStatus(), ProblemVerificationConstants.SYNC_SUCCESS)) {
            throw new StatusFailException("测试数据同步失败：" + record.getSyncMessage());
        }
        String submissionType = StringUtils.isEmpty(dto.getVerificationType())
                ? ProblemVerificationConstants.CREATOR_VALIDATION : dto.getVerificationType();
        if (!ProblemVerificationConstants.AI_VALIDATION.equals(submissionType)
                && !ProblemVerificationConstants.CREATOR_VALIDATION.equals(submissionType)) {
            throw new StatusFailException("无效的验题来源");
        }

        AccountProfile profile = ShiroUtils.getProfile();
        submitInternal(problem, dto, submissionType, profile.getUid(), profile.getUsername(), record);
        return toVO(problem, record);
    }

    /**
     * 后台 AI 验题专用提交入口。调用方已经在请求线程完成管理员鉴权，
     * 因此后台线程不依赖 Shiro 的请求线程 Subject。
     */
    Long submitForAi(ProblemVerificationSubmitDTO dto, String uid, String username)
            throws StatusFailException {
        if (dto == null || dto.getPid() == null || StringUtils.isEmpty(dto.getLanguage())
                || StringUtils.isEmpty(dto.getCode()) || dto.getCode().length() > 65535) {
            throw new StatusFailException("题目、语言和标准程序不能为空，代码不能超过 65535 个字符");
        }
        Problem problem = requireProblem(dto.getPid());
        ProblemVerification record = find(dto.getPid());
        if (record == null || !Objects.equals(record.getCaseVersion(), problem.getCaseVersion())) {
            throw new StatusFailException("题目测试数据尚未同步，请先重新保存题目");
        }
        if (!Objects.equals(record.getSyncStatus(), ProblemVerificationConstants.SYNC_SUCCESS)) {
            throw new StatusFailException("测试数据同步失败：" + record.getSyncMessage());
        }
        String submissionType = StringUtils.isEmpty(dto.getVerificationType())
                ? ProblemVerificationConstants.AI_VALIDATION : dto.getVerificationType();
        if (!ProblemVerificationConstants.AI_VALIDATION.equals(submissionType)) {
            throw new StatusFailException("后台 AI 验题提交来源无效");
        }
        return submitInternal(problem, dto, submissionType, uid, username, record);
    }

    private Long submitInternal(Problem problem, ProblemVerificationSubmitDTO dto,
                                String submissionType, String uid, String username,
                                ProblemVerification record) throws StatusFailException {
        Judge judge = new Judge().setPid(problem.getId()).setDisplayPid(problem.getProblemId())
                .setUid(uid).setUsername(username).setCode(dto.getCode()).setLanguage(dto.getLanguage())
                .setLength(dto.getCode().length()).setCid(0L).setCpid(0L).setGid(problem.getGid())
                .setShare(false).setIsProblemVerification(true).setSubmissionType(submissionType)
                .setStatus(Constants.Judge.STATUS_PENDING.getStatus()).setSubmitTime(new Date()).setVersion(0);
        judgeEntityService.save(judge);

        record.setSubmitId(judge.getSubmitId()).setVerifiedUid(uid)
                .setVerificationStatus(ProblemVerificationConstants.JUDGING);
        verificationService.updateById(record);
        judgeDispatcher.sendTask(judge.getSubmitId(), problem.getId(), false);
        return judge.getSubmitId();
    }

    /** 等待已有标准程序判题结束，供后台 AI 任务串行化使用。 */
    void waitForAiJudge(Long pid, long timeoutMs) throws StatusFailException {
        ProblemVerification record = find(pid);
        if (record == null || record.getSubmitId() == null
                || !Objects.equals(record.getVerificationStatus(), ProblemVerificationConstants.JUDGING)) {
            return;
        }
        waitForAiJudge(pid, record.getSubmitId(), timeoutMs);
    }

    /** 等待指定标准程序提交完成，整个过程不依赖前端轮询。 */
    void waitForAiJudge(Long pid, Long submitId, long timeoutMs) throws StatusFailException {
        long deadline = System.currentTimeMillis() + timeoutMs;
        while (true) {
            Judge judge = judgeEntityService.getById(submitId);
            if (judge != null && !ProblemVerificationConstants.isJudgeRunning(judge.getStatus())) {
                ProblemVerification record = find(pid);
                if (record != null) refreshJudgeStatus(record);
                return;
            }
            if (System.currentTimeMillis() >= deadline) {
                throw new StatusFailException("正式判题等待超时，请检查判题机状态后重试");
            }
            try {
                Thread.sleep(1000L);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                throw new StatusFailException("后台 AI 验题任务被中断");
            }
        }
    }

    public ProblemVerificationVO retrySync(Long pid)
            throws StatusFailException, StatusForbiddenException {
        Problem problem = requireProblem(pid);
        checkOperator(problem);
        verificationLifecycle.markTestCaseChanged(pid, problem.getCaseVersion(), problem.getJudgeMode());
        return getStatus(pid);
    }

    public ContestVerificationVO getContestStatus(Long cid) {
        List<ProblemVerificationVO> pending = new ArrayList<>();
        List<ContestProblem> contestProblems = contestProblemEntityService.list(
                new QueryWrapper<ContestProblem>().eq("cid", cid));
        for (ContestProblem contestProblem : contestProblems) {
            Problem problem = problemEntityService.getById(contestProblem.getPid());
            if (problem == null || Boolean.TRUE.equals(problem.getIsRemote())) {
                continue;
            }
            ProblemVerification record = find(problem.getId());
            if (record == null) {
                continue;
            }
            refreshJudgeStatus(record);
            ProblemVerificationVO status = toVO(problem, record);
            if (!Boolean.TRUE.equals(status.getVerified())) {
                pending.add(status);
            }
        }
        return new ContestVerificationVO().setReady(pending.isEmpty())
                .setTotalProblems(contestProblems.size()).setProblems(pending);
    }

    private ProblemVerification find(Long pid) {
        return verificationService.getOne(new QueryWrapper<ProblemVerification>().eq("pid", pid), false);
    }

    private Problem requireProblem(Long pid) throws StatusFailException {
        Problem problem = problemEntityService.getById(pid);
        if (problem == null) {
            throw new StatusFailException("题目不存在");
        }
        return problem;
    }

    private void checkOperator(Problem problem) throws StatusForbiddenException {
        AccountProfile profile = ShiroUtils.getProfile();
        boolean admin = SecurityUtils.getSubject().hasRole("root")
                || SecurityUtils.getSubject().hasRole("admin")
                || SecurityUtils.getSubject().hasRole("problem_admin");
        if (!admin && (profile == null || !Objects.equals(profile.getUsername(), problem.getAuthor()))) {
            throw new StatusForbiddenException("只有题目创建者或管理员可以提交标准程序");
        }
    }

    private void refreshJudgeStatus(ProblemVerification record) {
        if (!Objects.equals(record.getVerificationStatus(), ProblemVerificationConstants.JUDGING)
                || record.getSubmitId() == null) {
            return;
        }
        Judge judge = judgeEntityService.getById(record.getSubmitId());
        if (judge == null || ProblemVerificationConstants.isJudgeRunning(judge.getStatus())) {
            return;
        }
        if (Objects.equals(judge.getStatus(), Constants.Judge.STATUS_ACCEPTED.getStatus())) {
            record.setVerificationStatus(ProblemVerificationConstants.PASSED)
                    .setSubmitId(judge.getSubmitId())
                    .setSampleVerified(true)
                    .setSyncMessage("标准程序已通过正式判题，样例也随同测试点完成校验");
        } else {
            JudgeProgressService.Progress progress = judgeProgressService.resolve(
                    judge.getSubmitId(), judge.getStatus(), judge.getErrorMessage());
            record.setVerificationStatus(ProblemVerificationConstants.FAILED)
                    .setSampleVerified(false)
                    .setSyncMessage("标准程序验题失败：" + trimMessage(progress.getMessage()));
        }
        verificationService.updateById(record);
    }

    private ProblemVerificationVO toVO(Problem problem, ProblemVerification record) {
        Integer judgeStatus = null;
        String judgeStatusText = null;
        Integer currentTest = null;
        String judgeMessage = null;
        if (record.getSubmitId() != null) {
            Judge judge = judgeEntityService.getById(record.getSubmitId());
            if (judge != null) {
                judgeStatus = judge.getStatus();
                JudgeProgressService.Progress progress = judgeProgressService.resolve(
                        judge.getSubmitId(), judge.getStatus(), judge.getErrorMessage());
                judgeStatusText = progress.getText();
                currentTest = progress.getCurrentTest();
                judgeMessage = progress.getMessage();
            }
        }
        boolean verified = Objects.equals(record.getSyncStatus(), ProblemVerificationConstants.SYNC_SUCCESS)
                && Objects.equals(record.getVerificationStatus(), ProblemVerificationConstants.PASSED)
                && Objects.equals(record.getCaseVersion(), problem.getCaseVersion());
        boolean canSubmit = Objects.equals(record.getSyncStatus(), ProblemVerificationConstants.SYNC_SUCCESS)
                && !Objects.equals(record.getVerificationStatus(), ProblemVerificationConstants.JUDGING)
                && !verified;
        return new ProblemVerificationVO().setPid(problem.getId()).setProblemId(problem.getProblemId())
                .setTitle(problem.getTitle()).setCaseVersion(record.getCaseVersion())
                .setJudgeMode(record.getJudgeMode()).setSyncStatus(record.getSyncStatus())
                .setSyncMessage(record.getSyncMessage()).setVerificationStatus(record.getVerificationStatus())
                .setSubmitId(record.getSubmitId()).setJudgeStatus(judgeStatus)
                .setJudgeStatusText(judgeStatusText).setCurrentTest(currentTest).setJudgeMessage(judgeMessage)
                .setCanSubmit(canSubmit).setVerified(verified)
                .setSampleVerified(Boolean.TRUE.equals(record.getSampleVerified()));
    }

    public LastAcceptedCodeVO getLastPassedCode(Long pid)
            throws StatusFailException, StatusForbiddenException {
        Problem problem = requireProblem(pid);
        checkOperator(problem);
        Judge judge = judgeEntityService.getOne(new QueryWrapper<Judge>()
                .select("submit_id", "language", "code")
                .eq("pid", pid).eq("is_problem_verification", true)
                .eq("status", Constants.Judge.STATUS_ACCEPTED.getStatus())
                .orderByDesc("submit_id").last("LIMIT 1"), false);
        LastAcceptedCodeVO result = new LastAcceptedCodeVO();
        if (judge != null) {
            result.setSubmitId(judge.getSubmitId());
            result.setLanguage(judge.getLanguage());
            result.setCode(judge.getCode());
        } else {
            result.setCode("");
        }
        return result;
    }

    public ProblemVerificationDraft getDraft(Long pid) throws StatusFailException, StatusForbiddenException {
        Problem problem = requireProblem(pid);
        checkOperator(problem);
        return draftService.getOne(new QueryWrapper<ProblemVerificationDraft>().eq("pid", pid)
                .eq("uid", ShiroUtils.getProfile().getUid()), false);
    }

    /**
     * Read the current verification binding for report-only AI rechecks.
     * The caller still performs the administrator authorization check through
     * this manager, while the returned record lets the recheck prompt explain
     * whether the historical formal submission is bound to the current case
     * version instead of guessing from the test-point count.
     */
    public ProblemVerification getVerification(Long pid)
            throws StatusFailException, StatusForbiddenException {
        Problem problem = requireProblem(pid);
        checkOperator(problem);
        return find(problem.getId());
    }

    public ProblemVerificationDraft saveDraft(ProblemVerificationDraft draft)
            throws StatusFailException, StatusForbiddenException {
        if (draft == null || draft.getPid() == null) {
            throw new StatusFailException("题目不能为空");
        }
        Problem problem = requireProblem(draft.getPid());
        checkOperator(problem);
        if (StringUtils.isEmpty(draft.getLanguage()) || StringUtils.isEmpty(draft.getCode())
                || draft.getCode().length() > 65535) {
            throw new StatusFailException("语言和标准程序不能为空，代码不能超过 65535 个字符");
        }
        ProblemVerificationDraft existing = draftService.getOne(new QueryWrapper<ProblemVerificationDraft>()
                .eq("pid", problem.getId()).eq("uid", ShiroUtils.getProfile().getUid()), false);
        draft.setUid(ShiroUtils.getProfile().getUid()).setCaseVersion(problem.getCaseVersion());
        if (existing != null) draft.setId(existing.getId());
        draftService.saveOrUpdate(draft);
        return draft;
    }

    public void deleteDraft(Long pid) throws StatusFailException, StatusForbiddenException {
        Problem problem = requireProblem(pid);
        checkOperator(problem);
        draftService.remove(new QueryWrapper<ProblemVerificationDraft>().eq("pid", pid)
                .eq("uid", ShiroUtils.getProfile().getUid()));
        problemCaseEntityService.remove(new QueryWrapper<top.hcode.hoj.pojo.entity.problem.ProblemCase>()
                .eq("pid", pid));
        FileUtil.del(Constants.File.TESTCASE_BASE_FOLDER.getPath() + java.io.File.separator + "problem_" + pid);
        String cleanupError = null;
        try {
            testCaseSyncManager.delete(pid);
        } catch (Exception e) {
            cleanupError = e.getMessage();
        }
        ProblemVerification record = find(pid);
        if (record != null) {
            record.setSyncStatus(ProblemVerificationConstants.SYNC_FAILED)
                    .setVerificationStatus(ProblemVerificationConstants.REQUIRED)
                    .setSyncMessage("测试点已删除，请重新上传测试数据")
                    .setSubmitId(null).setSampleVerified(false);
            verificationService.updateById(record);
        }
        if (cleanupError != null) {
            throw new StatusFailException("草稿已删除，但清理判题服务器测试点失败：" + cleanupError);
        }
    }

    public void clearDraft(Long pid) throws StatusFailException, StatusForbiddenException {
        Problem problem = requireProblem(pid);
        checkOperator(problem);
        draftService.remove(new QueryWrapper<ProblemVerificationDraft>().eq("pid", pid)
                .eq("uid", ShiroUtils.getProfile().getUid()));
    }

    private String trimMessage(String message) {
        if (message == null || message.trim().isEmpty()) {
            return "未知错误";
        }
        String value = message.trim();
        return value.length() > 900 ? value.substring(0, 900) : value;
    }
}
