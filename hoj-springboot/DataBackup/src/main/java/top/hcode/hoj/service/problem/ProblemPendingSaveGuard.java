package top.hcode.hoj.service.problem;

import cn.hutool.core.bean.BeanUtil;
import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.dao.problem.ProblemCaseEntityService;
import top.hcode.hoj.pojo.dto.ProblemDTO;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.pojo.entity.problem.ProblemCase;

import javax.annotation.Resource;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.Set;

/** Existing formal problems may save unchanged judging inputs without repeating historical verification. */
@Component
public class ProblemPendingSaveGuard {
    @Resource private JdbcTemplate jdbcTemplate;
    @Resource private ProblemCaseEntityService cases;

    public boolean isCandidate(Long pid) {
        if (pid == null) return false;
        Integer count = jdbcTemplate.queryForObject("SELECT COUNT(*) FROM problem_pending_save "
                + "WHERE candidate_pid=? AND (published_pid IS NULL OR published_pid<>candidate_pid)", Integer.class, pid);
        return count != null && count > 0;
    }

    public void requireFormal(Long pid) throws StatusFailException {
        if (isCandidate(pid)) throw new StatusFailException("此题目尚未完成验题保存，请返回录题页面完成验证后保存");
    }

    public boolean requiresVerification(ProblemDTO dto, Problem stored) throws StatusFailException {
        if (dto == null || dto.getProblem() == null) throw new StatusFailException("题目内容不能为空");
        Problem input = dto.getProblem();
        if (stored == null) return !Boolean.TRUE.equals(input.getIsRemote());
        requireFormal(stored.getId());
        if (input.getIsRemote() != null && !Objects.equals(input.getIsRemote(), stored.getIsRemote())) {
            throw new StatusFailException("不能通过保存切换本地与远程题目类型");
        }
        if (Boolean.TRUE.equals(stored.getIsRemote())) return false;
        Problem after = BeanUtil.copyProperties(input, Problem.class);
        after.setCaseVersion(stored.getCaseVersion());
        if (after.getIsRemote() == null) after.setIsRemote(stored.getIsRemote());
        if (after.getIsGroup() == null) after.setIsGroup(stored.getIsGroup());
        Boolean upload = dto.getIsUploadTestCase() != null ? dto.getIsUploadTestCase()
                : after.getIsUploadCase() != null ? after.getIsUploadCase() : stored.getIsUploadCase();
        if (dto.getIsUploadTestCase() != null && input.getIsUploadCase() != null
                && !Objects.equals(dto.getIsUploadTestCase(), input.getIsUploadCase())) {
            throw new StatusFailException("测试数据上传方式不一致");
        }
        after.setIsUploadCase(upload);
        if (ProblemVerificationChangeDetector.changed(stored, after)
                || !Objects.equals(stored.getIsUploadCase(), upload)
                || StringUtils.hasText(dto.getUploadTestcaseDir())) return true;
        if (dto.getSamples() == null || dto.getSamples().isEmpty()) {
            throw new StatusFailException("题目的测试数据不能为空，请加载完成后重试");
        }
        List<ProblemCase> previous = cases.list(new QueryWrapper<ProblemCase>().eq("pid", stored.getId()).eq("status", 0));
        if (previous.size() != dto.getSamples().size()) return true;
        Map<Long, ProblemCase> byId = new HashMap<>();
        for (ProblemCase row : previous) byId.put(row.getId(), row);
        Set<Long> seen = new HashSet<>();
        for (ProblemCase row : dto.getSamples()) {
            if (row == null) throw new StatusFailException("测试数据内容不完整");
            ProblemCase old = byId.get(row.getId());
            if (old == null || !seen.add(row.getId())) return true;
            if (!Objects.equals(old.getInput(), row.getInput()) || !Objects.equals(old.getOutput(), row.getOutput())
                    || (Objects.equals(after.getType(), 1)
                    && (!Objects.equals(old.getScore(), row.getScore()) || !Objects.equals(old.getGroupNum(), row.getGroupNum())))) return true;
        }
        return false;
    }
}
