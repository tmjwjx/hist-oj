package top.hcode.hoj.manager.oj;

import cn.hutool.crypto.digest.DigestUtil;
import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import com.baomidou.mybatisplus.core.metadata.IPage;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import org.apache.shiro.SecurityUtils;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.transaction.PlatformTransactionManager;
import org.springframework.transaction.TransactionDefinition;
import org.springframework.transaction.support.TransactionSynchronization;
import org.springframework.transaction.support.TransactionSynchronizationManager;
import org.springframework.transaction.support.TransactionTemplate;
import org.springframework.util.StringUtils;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.dao.judge.JudgeCaseEntityService;
import top.hcode.hoj.dao.judge.JudgeEntityService;
import top.hcode.hoj.dao.problem.*;
import top.hcode.hoj.service.problem.ProblemPendingSaveGuard;
import top.hcode.hoj.judge.self.JudgeDispatcher;
import top.hcode.hoj.mapper.ExternalVerificationSubmissionMapper;
import top.hcode.hoj.mapper.ProblemExternalVerificationMapper;
import top.hcode.hoj.pojo.dto.ExternalVerificationSettingDTO;
import top.hcode.hoj.pojo.dto.ExternalVerificationSubmitDTO;
import top.hcode.hoj.pojo.entity.judge.ExternalVerificationSubmission;
import top.hcode.hoj.pojo.entity.judge.Judge;
import top.hcode.hoj.pojo.entity.judge.JudgeCase;
import top.hcode.hoj.pojo.entity.problem.*;
import top.hcode.hoj.pojo.vo.ExternalVerificationLinkVO;
import top.hcode.hoj.pojo.vo.ExternalVerificationProblemVO;
import top.hcode.hoj.pojo.vo.ExternalVerificationSubmissionVO;
import top.hcode.hoj.service.problem.JudgeProgressService;
import top.hcode.hoj.shiro.AccountProfile;
import top.hcode.hoj.utils.Constants;
import top.hcode.hoj.utils.ProblemVerificationConstants;
import top.hcode.hoj.utils.ShiroUtils;

import javax.annotation.Resource;
import java.security.SecureRandom;
import java.util.*;
import java.util.stream.Collectors;

@Component
public class ExternalVerificationManager {
    @Resource private ProblemExternalVerificationMapper linkMapper;
    @Resource private ExternalVerificationSubmissionMapper submissionMapper;
    @Resource private ProblemEntityService problemService;
    @Resource private ProblemPendingSaveGuard pendingSaveGuard;
    @Resource private ProblemVerificationEntityService verificationService;
    @Resource private ProblemCaseEntityService problemCaseService;
    @Resource private ProblemLanguageEntityService problemLanguageService;
    @Resource private LanguageEntityService languageService;
    @Resource private JudgeEntityService judgeService;
    @Resource private JudgeCaseEntityService judgeCaseService;
    @Resource private JudgeDispatcher judgeDispatcher;
    @Resource private JudgeProgressService progressService;
    @Resource private ExternalVerificationRateLimiter rateLimiter;
    @Resource private PlatformTransactionManager transactionManager;
    private static final SecureRandom RANDOM = new SecureRandom();

    public ExternalVerificationLinkVO setting(Long pid) throws StatusFailException, StatusForbiddenException {
        checkOperator(requireProblem(pid));
        return linkVO(pid, linkMapper.selectById(pid));
    }

    public ExternalVerificationLinkVO saveSetting(ExternalVerificationSettingDTO dto)
            throws StatusFailException, StatusForbiddenException {
        if (dto == null || dto.getEnabled() == null) throw new StatusFailException("题目和开关状态不能为空");
        Problem problem = requireProblem(dto.getPid());
        AccountProfile profile = checkOperator(problem);
        requireLocal(problem);
        ProblemExternalVerification row = new ProblemExternalVerification().setPid(problem.getId())
                .setToken(newToken()).setEnabled(dto.getEnabled())
                .setOwnerUid(profile.getUid()).setOwnerUsername(profile.getUsername());
        linkMapper.saveSetting(row);
        return linkVO(problem.getId(), linkMapper.selectById(problem.getId()));
    }

    public ExternalVerificationProblemVO problem(String token) throws StatusFailException, StatusForbiddenException {
        Problem problem = requireProblem(requireLink(token, false).getPid());
        requireLocal(problem);
        ExternalVerificationProblemVO.Statement statement = new ExternalVerificationProblemVO.Statement()
                .setId(problem.getId()).setProblemId(problem.getProblemId()).setTitle(problem.getTitle())
                .setDescription(problem.getDescription()).setInput(problem.getInput()).setOutput(problem.getOutput())
                .setExamples(problem.getExamples()).setHint(problem.getHint()).setTimeLimit(problem.getTimeLimit())
                .setMemoryLimit(problem.getMemoryLimit()).setStackLimit(problem.getStackLimit())
                .setType(problem.getType()).setJudgeMode(problem.getJudgeMode())
                .setIsFileIO(problem.getIsFileIO()).setIoReadFileName(problem.getIoReadFileName())
                .setIoWriteFileName(problem.getIoWriteFileName());
        return new ExternalVerificationProblemVO().setProblem(statement).setLanguages(languages(problem.getId()).stream()
                .map(language -> new ExternalVerificationProblemVO.AllowedLanguage()
                        .setName(language.getName()).setContentType(language.getContentType())).collect(Collectors.toList()));
    }

    @Transactional(rollbackFor = Exception.class)
    public ExternalVerificationSubmissionVO submit(String token, ExternalVerificationSubmitDTO dto, String ip)
            throws StatusFailException, StatusForbiddenException {
        // Serialize submissions with disabling this link; a completed disable cannot admit another submission.
        ProblemExternalVerification link = requireLink(token, true);
        if (dto == null || !StringUtils.hasText(dto.getLanguage()) || dto.getLanguage().length() > 255
                || !StringUtils.hasText(dto.getCode()) || dto.getCode().length() > 65535
                || (dto.getNickname() != null && dto.getNickname().length() > 40)) {
            throw new StatusFailException("语言和代码不能为空，代码最多 65535 个字符，昵称最多 40 个字符");
        }
        String visitorHash = visitorHash(dto.getVisitorToken());
        Problem problem = requireProblem(link.getPid());
        requireLocal(problem);
        if (languages(problem.getId()).stream().noneMatch(language -> language.getName().equals(dto.getLanguage()))) {
            throw new StatusFailException("本题不支持该提交语言");
        }
        ProblemVerification ready = verificationService.getOne(new QueryWrapper<ProblemVerification>()
                .eq("pid", problem.getId()), false);
        if (ready == null || !StringUtils.hasText(problem.getCaseVersion())
                || !Objects.equals(ready.getCaseVersion(), problem.getCaseVersion())
                || !Objects.equals(ready.getSyncStatus(), ProblemVerificationConstants.SYNC_SUCCESS)
                || problemCaseService.count(new QueryWrapper<ProblemCase>().eq("pid", problem.getId()).eq("status", 0)) == 0) {
            throw new StatusFailException("题目测试数据尚未就绪，请联系出题人");
        }
        rateLimiter.checkSubmission(ip, problem.getId());
        Judge judge = new Judge().setPid(problem.getId()).setDisplayPid(problem.getProblemId())
                .setUid(link.getOwnerUid()).setUsername(link.getOwnerUsername()).setCode(dto.getCode())
                .setLanguage(dto.getLanguage()).setLength(dto.getCode().length()).setCid(0L).setCpid(0L)
                .setGid(problem.getGid()).setShare(false).setIsManual(false).setIsProblemVerification(true)
                .setSubmissionType(ProblemVerificationConstants.EXTERNAL_VERIFICATION)
                .setStatus(Constants.Judge.STATUS_PENDING.getStatus()).setSubmitTime(new Date()).setVersion(0);
        if (!judgeService.save(judge)) throw new StatusFailException("提交保存失败，请重试");
        ExternalVerificationSubmission external = new ExternalVerificationSubmission().setSubmitId(judge.getSubmitId())
                .setPid(problem.getId()).setVisitorTokenHash(visitorHash)
                .setNickname(StringUtils.hasText(dto.getNickname()) ? dto.getNickname().trim() : "匿名访客");
        if (submissionMapper.insert(external) != 1) throw new StatusFailException("提交保存失败，请重试");
        // The judge worker must not see the task until both ownership and judge rows have committed.
        if (TransactionSynchronizationManager.isSynchronizationActive()) {
            TransactionSynchronizationManager.registerSynchronization(new TransactionSynchronization() {
                @Override public void afterCommit() {
                    TransactionTemplate dispatch = new TransactionTemplate(transactionManager);
                    dispatch.setPropagationBehavior(TransactionDefinition.PROPAGATION_REQUIRES_NEW);
                    dispatch.execute(status -> {
                        judgeDispatcher.sendTask(judge.getSubmitId(), problem.getId(), false);
                        return null;
                    });
                }
            });
        } else {
            judgeDispatcher.sendTask(judge.getSubmitId(), problem.getId(), false);
        }
        return summary(judge, external.getNickname());
    }

    public IPage<ExternalVerificationSubmissionVO> submissions(String token, String visitorToken, Integer currentPage, Integer limit)
            throws StatusFailException, StatusForbiddenException {
        return list(requireLink(token, false).getPid(), visitorHash(visitorToken), currentPage, limit);
    }

    public ExternalVerificationSubmissionVO submission(String token, Long submitId, String visitorToken)
            throws StatusFailException, StatusForbiddenException {
        return detail(requireLink(token, false).getPid(), submitId, visitorHash(visitorToken), false);
    }

    public IPage<ExternalVerificationSubmissionVO> adminSubmissions(Long pid, Integer currentPage, Integer limit)
            throws StatusFailException, StatusForbiddenException {
        checkOperator(requireProblem(pid));
        return list(pid, null, currentPage, limit);
    }

    public ExternalVerificationSubmissionVO adminSubmission(Long pid, Long submitId)
            throws StatusFailException, StatusForbiddenException {
        checkOperator(requireProblem(pid));
        return detail(pid, submitId, null, true);
    }

    private IPage<ExternalVerificationSubmissionVO> list(Long pid, String hash, Integer currentPage, Integer limit)
            throws StatusFailException {
        int page = currentPage == null ? 1 : currentPage;
        int size = limit == null ? 20 : limit;
        if (page < 1 || page > 100000 || size < 1 || size > 100) throw new StatusFailException("分页参数无效");
        IPage<ExternalVerificationSubmissionVO> result = submissionMapper.submissions(new Page<>(page, size), pid, hash);
        for (ExternalVerificationSubmissionVO row : result.getRecords()) setProgress(row);
        return result;
    }

    private ExternalVerificationSubmissionVO detail(Long pid, Long submitId, String hash, boolean admin)
            throws StatusForbiddenException {
        if (submitId == null || submitId <= 0) throw new StatusForbiddenException("提交不存在或无权查看");
        ExternalVerificationSubmission external = submissionMapper.selectById(submitId);
        if (external == null || !Objects.equals(external.getPid(), pid)
                || (!admin && !Objects.equals(external.getVisitorTokenHash(), hash))) {
            throw new StatusForbiddenException("提交不存在或无权查看");
        }
        Judge judge = judgeService.getById(submitId);
        if (judge == null || !Objects.equals(judge.getPid(), pid)
                || !Boolean.TRUE.equals(judge.getIsProblemVerification())
                || !ProblemVerificationConstants.EXTERNAL_VERIFICATION.equals(judge.getSubmissionType())) {
            throw new StatusForbiddenException("提交不存在或无权查看");
        }
        ExternalVerificationSubmissionVO result = summary(judge, external.getNickname()).setCode(judge.getCode());
        if (admin || Objects.equals(judge.getStatus(), Constants.Judge.STATUS_COMPILE_ERROR.getStatus())) {
            result.setErrorMessage(judge.getErrorMessage());
        }
        List<JudgeCase> cases = judgeCaseService.list(new QueryWrapper<JudgeCase>()
                .eq("submit_id", submitId).eq("pid", pid).orderByAsc("seq"));
        result.setCaseResults(cases.stream().map(row -> {
            ExternalVerificationSubmissionVO.CaseResult item = new ExternalVerificationSubmissionVO.CaseResult()
                    .setSeq(row.getSeq()).setStatus(row.getStatus()).setTime(row.getTime()).setMemory(row.getMemory())
                    .setScore(row.getScore()).setGroupNum(row.getGroupNum()).setMode(row.getMode());
            if (admin) item.setInputData(row.getInputData()).setOutputData(row.getOutputData())
                    .setUserOutput(row.getUserOutput()).setStderr(row.getStderr());
            return item;
        }).collect(Collectors.toList()));
        return result;
    }

    private ExternalVerificationSubmissionVO summary(Judge judge, String nickname) {
        ExternalVerificationSubmissionVO row = new ExternalVerificationSubmissionVO().setSubmitId(judge.getSubmitId())
                .setNickname(nickname).setLanguage(judge.getLanguage()).setStatus(judge.getStatus())
                .setTime(judge.getTime()).setMemory(judge.getMemory()).setScore(judge.getScore()).setSubmitTime(judge.getSubmitTime());
        setProgress(row);
        return row;
    }

    private void setProgress(ExternalVerificationSubmissionVO row) {
        JudgeProgressService.Progress progress = progressService.resolve(row.getSubmitId(), row.getStatus(), null);
        row.setStatusText(progress.getText()).setCurrentTest(progress.getCurrentTest());
    }

    private List<Language> languages(Long pid) {
        List<Long> ids = problemLanguageService.list(new QueryWrapper<ProblemLanguage>().eq("pid", pid))
                .stream().map(ProblemLanguage::getLid).collect(Collectors.toList());
        if (ids.isEmpty()) return Collections.emptyList();
        return languageService.list(new QueryWrapper<Language>().in("id", ids).eq("oj", "ME")
                .orderByDesc("seq").orderByAsc("id"));
    }

    private Problem requireProblem(Long pid) throws StatusFailException {
        pendingSaveGuard.requireFormal(pid);
        Problem problem = pid == null || pid <= 0 ? null : problemService.getById(pid);
        if (problem == null) throw new StatusFailException("题目不存在");
        return problem;
    }

    private void requireLocal(Problem problem) throws StatusFailException {
        if (Boolean.TRUE.equals(problem.getIsRemote())) throw new StatusFailException("外部验题仅支持本地题目");
    }

    private ProblemExternalVerification requireLink(String token, boolean lock) throws StatusForbiddenException {
        if (token == null || !token.matches("[0-9a-f]{64}")) throw new StatusForbiddenException("验题链接不存在或已关闭");
        ProblemExternalVerification link = lock ? linkMapper.lockByToken(token)
                : linkMapper.selectOne(new QueryWrapper<ProblemExternalVerification>().eq("token", token));
        if (link == null || !Boolean.TRUE.equals(link.getEnabled()) || pendingSaveGuard.isCandidate(link.getPid())) {
            throw new StatusForbiddenException("验题链接不存在或已关闭");
        }
        return link;
    }

    private AccountProfile checkOperator(Problem problem) throws StatusForbiddenException {
        AccountProfile profile = ShiroUtils.getProfile();
        boolean admin = SecurityUtils.getSubject().hasRole("root")
                || SecurityUtils.getSubject().hasRole("problem_admin");
        if (profile == null || (!admin && !Objects.equals(profile.getUsername(), problem.getAuthor()))) {
            throw new StatusForbiddenException("对不起，你无权限查看该题目的验题链接或外部提交！");
        }
        return profile;
    }

    private String visitorHash(String token) throws StatusFailException {
        if (token == null || !token.matches("[0-9a-f]{64}")) throw new StatusFailException("访客凭证无效，请刷新页面重试");
        return DigestUtil.sha256Hex(token);
    }

    private ExternalVerificationLinkVO linkVO(Long pid, ProblemExternalVerification row) {
        return new ExternalVerificationLinkVO().setPid(pid).setToken(row == null ? null : row.getToken())
                .setEnabled(row != null && Boolean.TRUE.equals(row.getEnabled()));
    }

    private String newToken() {
        byte[] bytes = new byte[32];
        RANDOM.nextBytes(bytes);
        StringBuilder value = new StringBuilder(64);
        for (byte item : bytes) value.append(String.format("%02x", item & 255));
        return value.toString();
    }
}
