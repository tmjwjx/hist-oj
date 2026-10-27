package top.hcode.hoj.manager.admin.problem;

import cn.hutool.core.io.FileUtil;
import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import com.baomidou.mybatisplus.core.conditions.update.UpdateWrapper;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.extern.slf4j.Slf4j;
import org.apache.shiro.SecurityUtils;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.transaction.support.TransactionSynchronizationAdapter;
import org.springframework.transaction.support.TransactionSynchronizationManager;
import org.springframework.util.StringUtils;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.dao.contest.ContestEntityService;
import top.hcode.hoj.dao.contest.ContestProblemEntityService;
import top.hcode.hoj.dao.judge.JudgeEntityService;
import top.hcode.hoj.dao.judge.JudgeCaseEntityService;
import top.hcode.hoj.dao.problem.*;
import top.hcode.hoj.manager.admin.contest.AdminContestProblemManager;
import top.hcode.hoj.mapper.ProblemPendingSaveMapper;
import top.hcode.hoj.pojo.dto.ProblemDTO;
import top.hcode.hoj.pojo.dto.ProblemPendingSaveDTO;
import top.hcode.hoj.pojo.entity.contest.Contest;
import top.hcode.hoj.pojo.entity.contest.ContestProblem;
import top.hcode.hoj.pojo.entity.judge.Judge;
import top.hcode.hoj.pojo.entity.judge.JudgeCase;
import top.hcode.hoj.pojo.entity.problem.*;
import top.hcode.hoj.pojo.vo.ProblemPendingSaveVO;
import top.hcode.hoj.service.problem.ProblemPendingSaveGuard;
import top.hcode.hoj.service.problem.ProblemTestCaseSyncManager;
import top.hcode.hoj.shiro.AccountProfile;
import top.hcode.hoj.utils.Constants;
import top.hcode.hoj.utils.ProblemVerificationConstants;
import top.hcode.hoj.utils.ShiroUtils;
import top.hcode.hoj.validator.GroupValidator;
import top.hcode.hoj.validator.ProblemValidator;

import javax.annotation.Resource;
import java.io.InputStream;
import java.nio.charset.StandardCharsets;
import java.nio.file.*;
import java.security.MessageDigest;
import java.util.*;
import java.util.stream.Collectors;
import java.util.stream.Stream;

/** Only candidate problems are written before the explicit verified commit. */
@Component
@Slf4j(topic = "hoj")
public class ProblemPendingSaveManager {
    @Resource private ProblemPendingSaveMapper pendingMapper;
    @Resource private ProblemEntityService problems;
    @Resource private ProblemCaseEntityService cases;
    @Resource private ProblemLanguageEntityService languages;
    @Resource private ProblemTagEntityService tags;
    @Resource private TagEntityService tagDefinitions;
    @Resource private CodeTemplateEntityService templates;
    @Resource private ProblemVerificationEntityService verifications;
    @Resource private JudgeEntityService judges;
    @Resource private JudgeCaseEntityService judgeCases;
    @Resource private ProblemVerificationManager verificationManager;
    @Resource private ProblemPendingSaveGuard guard;
    @Resource private AdminProblemManager adminProblemManager;
    @Resource private ProblemValidator validator;
    @Resource private ProblemTestCaseSyncManager syncManager;
    @Resource private ContestEntityService contests;
    @Resource private ContestProblemEntityService contestProblems;
    @Resource private AdminContestProblemManager contestProblemManager;
    @Resource private GroupValidator groupValidator;
    @Resource private ObjectMapper objectMapper;

    @Transactional(rollbackFor = Exception.class)
    public ProblemPendingSaveVO prepare(ProblemPendingSaveDTO request)
            throws StatusFailException, StatusForbiddenException {
        if (request == null || request.getProblemDTO() == null || request.getProblemDTO().getProblem() == null) {
            throw new StatusFailException("题目内容不能为空");
        }
        request = copy(request, ProblemPendingSaveDTO.class);
        ProblemDTO dto = request.getProblemDTO();
        Problem desired = dto.getProblem();
        // Always lock pending first, then the formal problem, matching commit's lock order.
        ProblemPendingSave previous = null;
        if (request.getPendingId() != null) {
            previous = requirePending(request.getPendingId(), true);
            if (!Objects.equals(previous.getTargetPid(), desired.getId())) {
                throw new StatusFailException("待保存记录与当前题目不一致，请重新打开题目");
            }
            if (previous.getPublishedPid() != null) {
                throw new StatusFailException("该版本已保存，请刷新页面后继续编辑");
            }
        }
        Problem stored = desired.getId() == null ? null : pendingMapper.lockProblem(desired.getId());
        if (desired.getId() != null && stored == null) throw new StatusFailException("原题目不存在");
        if (stored != null) { guard.requireFormal(stored.getId()); checkOperator(stored); }
        normalize(dto, stored);
        validateContest(request);
        checkDisplayId(desired.getProblemId(), stored == null ? null : stored.getId());
        if (stored != null && !guard.requiresVerification(dto, stored)) {
            adminProblemManager.updateProblem(dto);
            saveContest(request, stored.getId());
            if (previous != null) { previous.setSuperseded(true); pendingMapper.updateById(previous); }
            return new ProblemPendingSaveVO().setPid(stored.getId()).setTargetPid(stored.getId())
                    .setRequiresVerification(false);
        }
        if (Boolean.TRUE.equals(desired.getIsRemote())) throw new StatusFailException("远程题目不使用本地标准程序验题");

        request.setPendingId(null);
        String payload = json(request);
        String requestHash = hash(payload);
        if (previous != null && Objects.equals(previous.getRequestFingerprint(), requestHash)
                && Objects.equals(previous.getCandidateFingerprint(), fingerprint(previous.getCandidatePid()))) {
            return view(previous);
        }
        ProblemPendingSave pending = new ProblemPendingSave().setUid(profile().getUid())
                .setTargetPid(stored == null ? null : stored.getId()).setPayload(payload)
                .setRequestFingerprint(requestHash).setSuperseded(false)
                .setBaseFingerprint(stored == null ? null : fingerprint(stored.getId()));
        ProblemDTO candidateDto = copy(dto, ProblemDTO.class);
        Problem candidate = candidateDto.getProblem();
        candidate.setId(null).setProblemId("PENDING-" + UUID.randomUUID().toString().replace("-", ""))
                .setAuth(2).setIsRemote(false).setIsGroup(false).setGid(null)
                .setAuthor(profile().getUsername()).setGmtCreate(null).setGmtModified(null);
        clearChildIds(candidateDto);
        Path stage = null;
        try {
            if (Boolean.TRUE.equals(candidateDto.getIsUploadTestCase())) {
                Path source = uploadSource(dto, stored, previous);
                stage = stageCases(source, candidateDto.getSamples(), candidate.getJudgeMode());
                candidateDto.setUploadTestcaseDir(stage.toString());
            } else candidateDto.setUploadTestcaseDir(null);
            if (!problems.adminAddProblem(candidateDto)) throw new StatusFailException("准备验题失败，正式题目未修改");
            pending.setCandidatePid(candidate.getId());
            pending.setCandidateFingerprint(fingerprint(candidate.getId()));
            if (pendingMapper.insert(pending) != 1) throw new StatusFailException("保存待验题记录失败");
            if (previous != null) { previous.setSuperseded(true); pendingMapper.updateById(previous); }
            return view(pending);
        } finally {
            if (candidate.getId() != null) cleanupCandidateOnRollback(candidate.getId());
            if (stage != null) FileUtil.del(stage.toFile());
        }
    }

    public ProblemPendingSaveVO get(Long pendingId) throws StatusFailException, StatusForbiddenException {
        return view(requirePending(pendingId, false));
    }

    @Transactional(rollbackFor = Exception.class)
    public ProblemPendingSaveVO commit(Long pendingId) throws StatusFailException, StatusForbiddenException {
        ProblemPendingSave pending = requirePending(pendingId, true);
        if (pending.getPublishedPid() != null) return saved(pending.getPublishedPid());
        ProblemPendingSaveDTO request = read(pending.getPayload(), ProblemPendingSaveDTO.class);
        ProblemDTO desired = request.getProblemDTO();
        Problem target = pending.getTargetPid() == null ? null : pendingMapper.lockProblem(pending.getTargetPid());
        if (pending.getTargetPid() != null) {
            if (target == null) throw new StatusFailException("原题目已删除，无法保存");
            checkOperator(target);
            if (!Objects.equals(pending.getBaseFingerprint(), fingerprint(target.getId()))) {
                throw new StatusFailException("原题目已被修改，请重新打开并验证当前版本；本次内容仍保留在待保存记录中");
            }
        }
        Problem candidate = pendingMapper.lockProblem(pending.getCandidatePid());
        if (candidate == null || !Objects.equals(pending.getCandidateFingerprint(), fingerprint(candidate.getId()))) {
            throw new StatusFailException("待保存题目内容已变化，请重新提交标准程序验证");
        }
        validateContest(request);
        checkDisplayId(desired.getProblem().getProblemId(), target == null ? null : target.getId());
        if (!Boolean.TRUE.equals(verificationManager.getStatus(candidate.getId()).getVerified())) {
            throw new StatusFailException("标准程序尚未通过当前版本验证，题目未保存");
        }
        ProblemVerification verification = verifications.getOne(new QueryWrapper<ProblemVerification>()
                .eq("pid", candidate.getId()), false);
        Judge accepted = verification == null || verification.getSubmitId() == null ? null
                : judges.getById(verification.getSubmitId());
        if (accepted == null || !Objects.equals(accepted.getPid(), candidate.getId())
                || !Boolean.TRUE.equals(accepted.getIsProblemVerification())
                || !Objects.equals(accepted.getStatus(), Constants.Judge.STATUS_ACCEPTED.getStatus())
                || !Objects.equals(verification.getCaseVersion(), candidate.getCaseVersion())
                || !(ProblemVerificationConstants.CREATOR_VALIDATION.equals(accepted.getSubmissionType())
                    || ProblemVerificationConstants.AI_VALIDATION.equals(accepted.getSubmissionType()))) {
            throw new StatusFailException("标准程序验证记录无效，题目未保存");
        }
        Long publishedPid;
        if (target == null) {
            Problem published = copy(desired.getProblem(), Problem.class);
            published.setId(candidate.getId()).setCaseVersion(candidate.getCaseVersion())
                    .setIoScore(candidate.getIoScore()).setGmtCreate(candidate.getGmtCreate())
                    .setModifiedUser(profile().getUsername());
            if (!problems.updateById(published)) throw new StatusFailException("保存题目失败");
            judges.update(new UpdateWrapper<Judge>().eq("pid", candidate.getId())
                    .set("display_pid", published.getProblemId()));
            publishedPid = candidate.getId();
        } else {
            preserveFilesOnRollback(target);
            ProblemDTO replacement = copy(desired, ProblemDTO.class);
            List<ProblemCase> candidateCases = caseList(candidate.getId());
            replacement.setSamples(copyCases(candidateCases));
            replacement.setIsUploadTestCase(candidate.getIsUploadCase());
            replacement.getProblem().setId(target.getId()).setIsUploadCase(candidate.getIsUploadCase())
                    .setGmtCreate(target.getGmtCreate()).setModifiedUser(profile().getUsername());
            clearChildIds(replacement);
            // Candidate creation already persisted new tag definitions; do not insert them again.
            for (Tag tag : replacement.getTags()) {
                if (tag.getId() == null) {
                    Tag existingTag = tagDefinitions.getOne(new QueryWrapper<Tag>().eq("name", tag.getName()).eq("oj", "ME"), false);
                    if (existingTag != null) tag.setId(existingTag.getId());
                }
            }
            Path stage = null;
            try {
                if (Boolean.TRUE.equals(candidate.getIsUploadCase())) {
                    stage = stageCases(problemDirectory(candidate.getId()), replacement.getSamples(), candidate.getJudgeMode());
                    replacement.setUploadTestcaseDir(stage.toString());
                } else replacement.setUploadTestcaseDir(null);
                if (!problems.adminUpdateProblem(replacement)) throw new StatusFailException("保存题目失败");
                Problem published = problems.getById(target.getId());
                ProblemVerification formalVerification = verifications.getOne(new QueryWrapper<ProblemVerification>()
                        .eq("pid", target.getId()), false);
                if (formalVerification == null || !Objects.equals(formalVerification.getSyncStatus(),
                        ProblemVerificationConstants.SYNC_SUCCESS)) {
                    throw new StatusFailException("正式题目测试数据同步失败，已取消本次保存");
                }
                Judge publishedJudge = copy(accepted, Judge.class).setSubmitId(null).setPid(target.getId())
                        .setDisplayPid(published.getProblemId()).setGid(target.getGid()).setGmtCreate(null).setGmtModified(null);
                if (!judges.save(publishedJudge)) throw new StatusFailException("保存标准程序验证记录失败");
                copyJudgeCases(accepted.getSubmitId(), publishedJudge.getSubmitId(), target.getId(),
                        candidateCases, caseList(target.getId()));
                formalVerification.setCaseVersion(published.getCaseVersion()).setJudgeMode(published.getJudgeMode())
                        .setVerificationStatus(ProblemVerificationConstants.PASSED)
                        .setSubmitId(publishedJudge.getSubmitId()).setVerifiedUid(verification.getVerifiedUid())
                        .setSampleVerified(verification.getSampleVerified()).setSyncMessage("当前版本已通过标准程序验证并保存");
                if (!verifications.updateById(formalVerification)) throw new StatusFailException("保存验证状态失败");
                judges.update(new UpdateWrapper<Judge>().eq("pid", target.getId())
                        .set("display_pid", published.getProblemId()));
                publishedPid = target.getId();
            } finally { if (stage != null) FileUtil.del(stage.toFile()); }
        }
        pending.setPublishedPid(publishedPid);
        if (pendingMapper.updateById(pending) != 1) throw new StatusFailException("保存完成状态失败");
        saveContest(request, publishedPid);
        return saved(publishedPid);
    }

    private void normalize(ProblemDTO dto, Problem stored) throws StatusFailException {
        Problem p = dto.getProblem();
        if (p.getTimeLimit() == null || p.getMemoryLimit() == null || p.getStackLimit() == null) {
            throw new StatusFailException("时间、内存和栈限制不能为空");
        }
        if (dto.getTags() == null || dto.getCodeTemplates() == null || dto.getLanguages() == null
                || dto.getLanguages().isEmpty() || dto.getSamples() == null) {
            throw new StatusFailException("题目关联配置尚未加载完整，请稍后重试");
        }
        if (stored != null && p.getIsRemote() != null && !Objects.equals(p.getIsRemote(), stored.getIsRemote())) {
            throw new StatusFailException("不能通过保存切换本地与远程题目类型");
        }
        if (dto.getIsUploadTestCase() != null && p.getIsUploadCase() != null
                && !Objects.equals(dto.getIsUploadTestCase(), p.getIsUploadCase())) {
            throw new StatusFailException("测试数据上传方式不一致");
        }
        p.setIsRemote(stored == null ? false : stored.getIsRemote());
        p.setIsGroup(stored == null ? false : stored.getIsGroup()).setGid(stored == null ? null : stored.getGid());
        if (p.getIsFileIO() == null) p.setIsFileIO(stored != null && Boolean.TRUE.equals(stored.getIsFileIO()));
        if (p.getIsRemoveEndBlank() == null) p.setIsRemoveEndBlank(stored != null && Boolean.TRUE.equals(stored.getIsRemoveEndBlank()));
        if (Boolean.TRUE.equals(p.getIsFileIO())) {
            if (p.getIoReadFileName() == null && stored != null) p.setIoReadFileName(stored.getIoReadFileName());
            if (p.getIoWriteFileName() == null && stored != null) p.setIoWriteFileName(stored.getIoWriteFileName());
            if (!StringUtils.hasText(p.getIoReadFileName()) || !StringUtils.hasText(p.getIoWriteFileName())) {
                throw new StatusFailException("文件读写模式必须提供输入和输出文件名");
            }
        }
        if (Boolean.TRUE.equals(p.getIsGroup())) throw new StatusFailException("请从团队题目管理页面保存团队题目");
        p.setAuthor(stored == null ? profile().getUsername() : stored.getAuthor());
        p.setModifiedUser(stored == null ? null : stored.getModifiedUser());
        p.setGmtCreate(stored == null ? null : stored.getGmtCreate()).setGmtModified(stored == null ? null : stored.getGmtModified());
        p.setCaseVersion(stored == null ? null : stored.getCaseVersion());
        if (dto.getIsUploadTestCase() == null) dto.setIsUploadTestCase(p.getIsUploadCase());
        if (!Boolean.TRUE.equals(p.getIsRemote()) && (dto.getIsUploadTestCase() == null || dto.getSamples().isEmpty())) {
            throw new StatusFailException("测试数据不能为空，请先加载或上传测试数据");
        }
        p.setIsUploadCase(dto.getIsUploadTestCase());
        dto.setJudgeMode(p.getJudgeMode());
        validator.validateProblem(p);
        p.setProblemId(p.getProblemId().trim().toUpperCase(Locale.ROOT));
        if ("default".equals(p.getJudgeMode())) p.setSpjCode(null).setSpjLanguage(null);
        Set<Long> originalIds = stored == null ? Collections.emptySet() : caseList(stored.getId()).stream()
                .map(ProblemCase::getId).collect(Collectors.toSet());
        Set<Long> seen = new HashSet<>();
        for (ProblemCase c : dto.getSamples()) {
            if (c != null && c.getInput() != null && Boolean.TRUE.equals(dto.getIsUploadTestCase())
                    && !"default".equals(p.getJudgeMode())
                    && (!StringUtils.hasText(c.getOutput()) || "-".equals(c.getOutput()))) {
                String input = c.getInput();
                int extension = input.lastIndexOf('.');
                c.setOutput((extension > 0 ? input.substring(0, extension) : input) + ".out");
            }
            if (c == null || c.getInput() == null || c.getOutput() == null || (c.getId() != null
                    && (!originalIds.contains(c.getId()) || !seen.add(c.getId())))) {
                throw new StatusFailException("测试数据不完整或不属于当前题目");
            }
            if (Boolean.TRUE.equals(dto.getIsUploadTestCase())) { safeFilename(c.getInput()); safeFilename(c.getOutput()); }
        }
        Set<Integer> templateIds = stored == null ? Collections.emptySet() : templates.list(
                new QueryWrapper<CodeTemplate>().eq("pid", stored.getId())).stream()
                .map(CodeTemplate::getId).collect(Collectors.toSet());
        Set<Integer> seenTemplates = new HashSet<>();
        for (CodeTemplate template : dto.getCodeTemplates()) {
            if (template == null || template.getLid() == null || (template.getId() != null
                    && (!templateIds.contains(template.getId()) || !seenTemplates.add(template.getId())))) {
                throw new StatusFailException("代码模板配置不完整或不属于当前题目");
            }
        }
        for (Language language : dto.getLanguages()) {
            if (language == null || language.getId() == null) throw new StatusFailException("题目语言配置不完整");
        }
    }

    private void checkOperator(Problem p) throws StatusForbiddenException {
        if (!SecurityUtils.getSubject().hasRole("root") && !SecurityUtils.getSubject().hasRole("problem_admin")
                && !Objects.equals(profile().getUsername(), p.getAuthor())) {
            throw new StatusForbiddenException("只有题目创建者或题目管理员可以修改本题");
        }
    }

    private ProblemPendingSave requirePending(Long id, boolean lock) throws StatusFailException, StatusForbiddenException {
        if (id == null || id <= 0) throw new StatusFailException("待保存记录不能为空");
        ProblemPendingSave p = lock ? pendingMapper.lock(id) : pendingMapper.selectById(id);
        if (p == null) throw new StatusFailException("待保存记录不存在");
        if (!Objects.equals(p.getUid(), profile().getUid())) throw new StatusForbiddenException("无权限操作他人的待保存记录");
        if (Boolean.TRUE.equals(p.getSuperseded())) throw new StatusFailException("该待保存版本已更新，请使用最新版本重新验题");
        return p;
    }

    private void checkDisplayId(String displayId, Long targetPid) throws StatusFailException {
        Problem p = problems.getOne(new QueryWrapper<Problem>().eq("problem_id", displayId), false);
        if (p != null && !Objects.equals(p.getId(), targetPid)) throw new StatusFailException("该题目的展示ID已存在，请更换");
    }

    private ProblemPendingSaveVO view(ProblemPendingSave p) throws StatusFailException {
        if (p.getPublishedPid() != null) return saved(p.getPublishedPid());
        ProblemDTO payload = read(p.getPayload(), ProblemPendingSaveDTO.class).getProblemDTO();
        Problem candidate = problems.getById(p.getCandidatePid());
        if (candidate == null) throw new StatusFailException("待验题目不存在");
        candidate.setProblemId(payload.getProblem().getProblemId());
        return new ProblemPendingSaveVO().setPendingId(p.getId()).setPid(p.getCandidatePid())
                .setTargetPid(p.getTargetPid()).setRequiresVerification(true).setProblem(candidate).setPayload(payload);
    }

    private ProblemPendingSaveVO saved(Long pid) {
        return new ProblemPendingSaveVO().setPid(pid).setTargetPid(pid).setRequiresVerification(false);
    }

    private void clearChildIds(ProblemDTO dto) {
        for (ProblemCase c : dto.getSamples()) c.setId(null).setPid(null).setGmtCreate(null).setGmtModified(null);
        for (CodeTemplate c : dto.getCodeTemplates()) c.setId(null).setPid(null).setGmtCreate(null).setGmtModified(null);
    }

    private List<ProblemCase> caseList(Long pid) {
        return cases.list(new QueryWrapper<ProblemCase>().eq("pid", pid).orderByAsc("id"));
    }

    private List<ProblemCase> copyCases(List<ProblemCase> source) throws StatusFailException {
        List<ProblemCase> result = new ArrayList<>();
        for (ProblemCase c : source) result.add(copy(c, ProblemCase.class));
        return result;
    }

    private void copyJudgeCases(Long from, Long to, Long pid, List<ProblemCase> oldCases, List<ProblemCase> newCases)
            throws StatusFailException {
        if (oldCases.size() != newCases.size()) throw new StatusFailException("验证测试点数量不一致，已取消保存");
        Map<Long, Long> ids = new HashMap<>();
        for (int i = 0; i < oldCases.size(); i++) ids.put(oldCases.get(i).getId(), newCases.get(i).getId());
        List<JudgeCase> results = judgeCases.list(new QueryWrapper<JudgeCase>().eq("submit_id", from).orderByAsc("seq"));
        for (JudgeCase result : results) {
            Long oldId = result.getCaseId();
            if (oldId != null && oldId > 0 && !ids.containsKey(oldId)) throw new StatusFailException("验证测试点关联不一致");
            result.setId(null).setPid(pid).setSubmitId(to).setCaseId(ids.getOrDefault(oldId, oldId))
                    .setGmtCreate(null).setGmtModified(null);
        }
        if (!results.isEmpty() && !judgeCases.saveBatch(results)) throw new StatusFailException("保存标准程序测试点结果失败");
    }

    private Path uploadSource(ProblemDTO dto, Problem stored, ProblemPendingSave previous) throws StatusFailException {
        if (StringUtils.hasText(dto.getUploadTestcaseDir())) {
            try {
                Path root = Paths.get(Constants.File.TESTCASE_TMP_FOLDER.getPath()).toRealPath();
                Path source = Paths.get(dto.getUploadTestcaseDir()).toAbsolutePath().normalize();
                if (!Objects.equals(source.getParent(), root) || !source.getFileName().toString().matches("[a-fA-F0-9]{32}")) {
                    throw new StatusFailException("上传测试数据目录无效，请重新上传");
                }
                if (!Files.isSymbolicLink(source) && Files.isDirectory(source) && source.toRealPath().startsWith(root)) return source;
            } catch (java.io.IOException e) { throw new StatusFailException("上传测试数据已失效，请重新上传", e); }
            throw new StatusFailException("上传测试数据已失效，请重新上传");
        }
        if (stored != null && Boolean.TRUE.equals(stored.getIsUploadCase())) return problemDirectory(stored.getId());
        if (previous != null) return problemDirectory(previous.getCandidatePid());
        throw new StatusFailException("请先上传测试数据包");
    }

    private Path stageCases(Path source, List<ProblemCase> samples, String judgeMode) throws StatusFailException {
        Path stage = null;
        try {
            Path root = Paths.get(Constants.File.TESTCASE_TMP_FOLDER.getPath());
            Files.createDirectories(root);
            stage = Files.createTempDirectory(root, "pending-");
            Path realSource = source.toRealPath();
            if (Files.isSymbolicLink(source)) throw new StatusFailException("测试数据目录不能为链接");
            Set<String> copied = new HashSet<>();
            for (ProblemCase c : samples) for (String name : Arrays.asList(c.getInput(), c.getOutput())) {
                safeFilename(name);
                if (!copied.add(name)) continue;
                Path file = source.resolve(name);
                if (!name.equals(c.getInput()) && !"default".equals(judgeMode)
                        && !Files.exists(file, LinkOption.NOFOLLOW_LINKS)) {
                    Files.createFile(stage.resolve(name));
                    continue;
                }
                if (!Files.isRegularFile(file, LinkOption.NOFOLLOW_LINKS) || !file.toRealPath().getParent().equals(realSource)) {
                    throw new StatusFailException("测试数据文件缺失或路径无效：" + name);
                }
                Files.copy(file, stage.resolve(name));
            }
            return stage;
        } catch (Exception e) {
            if (stage != null) FileUtil.del(stage.toFile());
            if (e instanceof StatusFailException) throw (StatusFailException) e;
            throw new StatusFailException("复制测试数据失败，正式题目未修改", e);
        }
    }

    private void safeFilename(String name) throws StatusFailException {
        if (name == null || name.isEmpty() || name.equals(".") || name.equals("..") || name.equals("info")
                || name.contains("/") || name.contains("\\") || name.indexOf('\0') >= 0) {
            throw new StatusFailException("测试数据文件名无效");
        }
    }

    private Path problemDirectory(Long pid) throws StatusFailException {
        if (pid == null || pid <= 0) throw new StatusFailException("测试数据题目ID无效");
        return Paths.get(Constants.File.TESTCASE_BASE_FOLDER.getPath(), "problem_" + pid);
    }

    private String fingerprint(Long pid) throws StatusFailException {
        Problem problem = problems.getById(pid);
        if (problem == null) throw new StatusFailException("题目不存在");
        Map<String, Object> snapshot = new TreeMap<>();
        snapshot.put("problem", problem);
        snapshot.put("cases", caseList(pid));
        snapshot.put("languages", languages.list(new QueryWrapper<ProblemLanguage>().eq("pid", pid).orderByAsc("lid")));
        snapshot.put("tags", tags.list(new QueryWrapper<ProblemTag>().eq("pid", pid).orderByAsc("tid")));
        snapshot.put("templates", templates.list(new QueryWrapper<CodeTemplate>().eq("pid", pid).orderByAsc("id")));
        try {
            MessageDigest digest = MessageDigest.getInstance("SHA-256");
            digest.update(json(snapshot).getBytes(StandardCharsets.UTF_8));
            Path directory = problemDirectory(pid);
            if (Files.isDirectory(directory)) {
                Path info = directory.resolve("info");
                if (!Files.isRegularFile(info, LinkOption.NOFOLLOW_LINKS)) throw new StatusFailException("测试数据 info 文件缺失");
                JsonNode metadata = objectMapper.readTree(info.toFile());
                Set<String> names = new TreeSet<>();
                names.add("info");
                for (JsonNode entry : metadata.path("testCases")) {
                    String input = entry.path("inputName").asText();
                    String output = entry.path("outputName").asText();
                    safeFilename(input); safeFilename(output);
                    names.add(input); names.add(output);
                }
                byte[] buffer = new byte[65536];
                for (String name : names) {
                    Path path = directory.resolve(name);
                    if (!Files.isRegularFile(path, LinkOption.NOFOLLOW_LINKS)) throw new StatusFailException("测试数据包含无效文件");
                    digest.update(path.getFileName().toString().getBytes(StandardCharsets.UTF_8));
                    try (InputStream input = Files.newInputStream(path)) {
                        int count;
                        while ((count = input.read(buffer)) >= 0) digest.update(buffer, 0, count);
                    }
                }
            }
            return hex(digest.digest());
        } catch (StatusFailException e) { throw e; }
        catch (Exception e) { throw new StatusFailException("读取题目版本失败，请稍后重试", e); }
    }

    private void cleanupCandidateOnRollback(Long pid) {
        TransactionSynchronizationManager.registerSynchronization(new TransactionSynchronizationAdapter() {
            @Override public void afterCompletion(int status) {
                if (status == STATUS_COMMITTED) return;
                try { FileUtil.del(problemDirectory(pid).toFile()); syncManager.delete(pid); }
                catch (Exception e) { log.error("清理失败的候选题目 {} 数据失败", pid, e); }
            }
        });
    }

    private void preserveFilesOnRollback(Problem original) throws StatusFailException {
        try {
            Path source = problemDirectory(original.getId());
            if (!Files.isDirectory(source)) throw new StatusFailException("正式题目测试数据缺失，已取消保存");
            Path backup = Files.createTempDirectory(Paths.get(Constants.File.TESTCASE_BASE_FOLDER.getPath()), ".pending-backup-");
            copyDirectory(source, backup);
            TransactionSynchronizationManager.registerSynchronization(new TransactionSynchronizationAdapter() {
                @Override public void afterCompletion(int status) {
                    if (status == STATUS_COMMITTED) { FileUtil.del(backup.toFile()); return; }
                    try {
                        FileUtil.del(source.toFile()); Files.createDirectories(source); copyDirectory(backup, source);
                        syncManager.sync(original.getId(), original.getCaseVersion());
                        FileUtil.del(backup.toFile());
                    } catch (Exception e) {
                        log.error("恢复题目 {} 测试数据失败，保留备份 {}", original.getId(), backup, e);
                    }
                }
            });
        } catch (StatusFailException e) { throw e; }
        catch (Exception e) { throw new StatusFailException("备份原测试数据失败，正式题目未修改", e); }
    }

    private void copyDirectory(Path source, Path target) throws java.io.IOException {
        try (Stream<Path> stream = Files.list(source)) {
            for (Path file : stream.collect(Collectors.toList())) {
                if (!Files.isRegularFile(file, LinkOption.NOFOLLOW_LINKS)) throw new java.io.IOException("测试数据包含无效文件");
                Files.copy(file, target.resolve(file.getFileName()), StandardCopyOption.REPLACE_EXISTING);
            }
        }
    }

    private void validateContest(ProblemPendingSaveDTO request) throws StatusFailException, StatusForbiddenException {
        if (request.getContestId() == null) {
            if (request.getContestProblem() != null) throw new StatusFailException("比赛信息不完整");
            return;
        }
        Contest contest = contests.getById(request.getContestId());
        if (contest == null) throw new StatusFailException("比赛不存在");
        if (!SecurityUtils.getSubject().hasRole("root") && !Objects.equals(contest.getUid(), profile().getUid())
                && !(Boolean.TRUE.equals(contest.getIsGroup()) && groupValidator.isGroupRoot(profile().getUid(), contest.getGid()))) {
            throw new StatusForbiddenException("只有比赛创建者或管理员可以保存比赛题目");
        }
        ContestProblem cp = request.getContestProblem();
        if (cp == null || !StringUtils.hasText(cp.getDisplayId())) throw new StatusFailException("比赛题目展示ID不能为空");
        Long targetPid = request.getProblemDTO().getProblem().getId();
        ContestProblem duplicate = contestProblems.getOne(new QueryWrapper<ContestProblem>()
                .eq("cid", contest.getId()).eq("display_id", cp.getDisplayId()), false);
        if (duplicate != null && !Objects.equals(duplicate.getPid(), targetPid)) throw new StatusFailException("比赛题目展示ID已存在");
        cp.setCid(contest.getId());
    }

    private void saveContest(ProblemPendingSaveDTO request, Long pid) throws StatusFailException {
        if (request.getContestId() == null) return;
        ContestProblem desired = request.getContestProblem();
        ContestProblem existing = contestProblems.getOne(new QueryWrapper<ContestProblem>()
                .eq("cid", request.getContestId()).eq("pid", pid), false);
        desired.setId(existing == null ? null : existing.getId()).setPid(pid).setCid(request.getContestId());
        contestProblemManager.setContestProblem(desired);
    }

    private AccountProfile profile() { return ShiroUtils.getProfile(); }
    private <T> T copy(Object value, Class<T> type) throws StatusFailException { return read(json(value), type); }
    private <T> T read(String value, Class<T> type) throws StatusFailException {
        try { return objectMapper.readValue(value, type); }
        catch (Exception e) { throw new StatusFailException("题目保存内容无效，请重新加载", e); }
    }
    private String json(Object value) throws StatusFailException {
        try { return objectMapper.writeValueAsString(sorted(objectMapper.valueToTree(value))); }
        catch (Exception e) { throw new StatusFailException("题目保存内容无法序列化", e); }
    }
    private Object sorted(JsonNode node) {
        if (node.isObject()) {
            Map<String, Object> result = new TreeMap<>();
            node.fields().forEachRemaining(entry -> result.put(entry.getKey(), sorted(entry.getValue())));
            return result;
        }
        if (node.isArray()) { List<Object> result = new ArrayList<>(); node.forEach(item -> result.add(sorted(item))); return result; }
        return node;
    }
    private String hash(String value) throws StatusFailException {
        try { return hex(MessageDigest.getInstance("SHA-256").digest(value.getBytes(StandardCharsets.UTF_8))); }
        catch (Exception e) { throw new StatusFailException("题目版本计算失败", e); }
    }
    private String hex(byte[] bytes) {
        StringBuilder value = new StringBuilder();
        for (byte b : bytes) value.append(String.format("%02x", b & 255));
        return value.toString();
    }
}
