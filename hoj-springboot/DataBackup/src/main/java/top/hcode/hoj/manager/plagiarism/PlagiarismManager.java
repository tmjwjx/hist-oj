package top.hcode.hoj.manager.plagiarism;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import org.apache.shiro.SecurityUtils;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.dao.contest.ContestEntityService;
import top.hcode.hoj.dao.contest.ContestProblemEntityService;
import top.hcode.hoj.dao.contest.ContestRecordEntityService;
import top.hcode.hoj.dao.judge.JudgeEntityService;
import top.hcode.hoj.dao.plagiarism.PlagiarismCheckConfigEntityService;
import top.hcode.hoj.dao.plagiarism.PlagiarismCheckEntityService;
import top.hcode.hoj.dao.plagiarism.PlagiarismResultEntityService;
import top.hcode.hoj.pojo.dto.PlagiarismConfigDTO;
import top.hcode.hoj.pojo.dto.PlagiarismConfigItemDTO;
import top.hcode.hoj.pojo.entity.contest.Contest;
import top.hcode.hoj.pojo.entity.contest.ContestProblem;
import top.hcode.hoj.pojo.entity.judge.Judge;
import top.hcode.hoj.pojo.entity.plagiarism.PlagiarismCheck;
import top.hcode.hoj.pojo.entity.plagiarism.PlagiarismCheckConfig;
import top.hcode.hoj.pojo.entity.plagiarism.PlagiarismResult;
import top.hcode.hoj.utils.ShiroUtils;

import javax.annotation.Resource;
import java.util.List;

@Component
public class PlagiarismManager {
    @Resource private ContestEntityService contestService;
    @Resource private ContestProblemEntityService contestProblemService;
    @Resource private ContestRecordEntityService contestRecordService;
    @Resource private JudgeEntityService judgeService;
    @Resource private PlagiarismCheckConfigEntityService configService;
    @Resource private PlagiarismCheckEntityService checkService;
    @Resource private PlagiarismResultEntityService resultService;

    public Contest requireAccess(Long cid) throws StatusNotFoundException, StatusForbiddenException {
        Contest contest = contestService.getById(cid);
        if (contest == null) throw new StatusNotFoundException("比赛不存在");
        String uid = ShiroUtils.getProfile().getUid();
        if (!uid.equals(contest.getUid()) && !isAdmin()) {
            throw new StatusForbiddenException("无权限访问查重功能");
        }
        return contest;
    }

    public List<PlagiarismCheckConfig> configs(Long cid)
            throws StatusNotFoundException, StatusForbiddenException {
        requireAccess(cid);
        return configService.list(new QueryWrapper<PlagiarismCheckConfig>().eq("cid", cid)
                .orderByAsc("cpid"));
    }

    public List<ContestProblem> contestProblems(Long cid)
            throws StatusNotFoundException, StatusForbiddenException {
        requireAccess(cid);
        return contestProblemService.list(new QueryWrapper<ContestProblem>().eq("cid", cid)
                .orderByAsc("id"));
    }

    @Transactional(rollbackFor = Exception.class)
    public void saveConfig(Long cid, PlagiarismConfigDTO dto)
            throws StatusNotFoundException, StatusForbiddenException, StatusFailException {
        requireAccess(cid);
        configService.remove(new QueryWrapper<PlagiarismCheckConfig>().eq("cid", cid));
        if (dto == null || dto.getConfigs() == null) return;
        for (PlagiarismConfigItemDTO item : dto.getConfigs()) {
            ContestProblem problem = contestProblemService.getById(item.getCpid());
            if (problem == null || !cid.equals(problem.getCid())) {
                throw new StatusFailException("比赛题目不存在: " + item.getCpid());
            }
            int threshold = item.getThreshold() == null ? 50 : item.getThreshold();
            if (threshold < 0 || threshold > 100) {
                throw new StatusFailException("查重阈值必须在 0 到 100 之间");
            }
            configService.save(new PlagiarismCheckConfig().setCid(cid).setPid(problem.getPid())
                    .setCpid(problem.getId()).setThreshold(threshold)
                    .setCreatedBy(ShiroUtils.getProfile().getUid()));
        }
    }

    @Transactional(rollbackFor = Exception.class)
    public PlagiarismCheck createCheck(Long cid)
            throws StatusNotFoundException, StatusForbiddenException, StatusFailException {
        Contest contest = requireAccess(cid);
        if (!Integer.valueOf(1).equals(contest.getStatus())) {
            throw new StatusFailException("比赛未结束，无法进行查重");
        }
        if (configService.count(new QueryWrapper<PlagiarismCheckConfig>().eq("cid", cid)) == 0) {
            throw new StatusFailException("请先设置查重配置");
        }
        resultService.remove(new QueryWrapper<PlagiarismResult>().eq("cid", cid));
        checkService.remove(new QueryWrapper<PlagiarismCheck>().eq("cid", cid)
                .in("status", "pending", "running", "failed"));
        PlagiarismCheck check = new PlagiarismCheck().setCid(cid).setStatus("pending")
                .setTotalPairs(0).setCheckedPairs(0).setTotalSubmissions(0).setProgress(0D)
                .setStartedBy(ShiroUtils.getProfile().getUid());
        if (!checkService.save(check)) throw new StatusFailException("创建查重任务失败");
        return check;
    }

    public PlagiarismCheck latest(Long cid)
            throws StatusNotFoundException, StatusForbiddenException {
        requireAccess(cid);
        return checkService.getOne(new QueryWrapper<PlagiarismCheck>().eq("cid", cid)
                .orderByDesc("id").last("LIMIT 1"));
    }

    public PlagiarismCheck check(Long checkId)
            throws StatusNotFoundException, StatusForbiddenException {
        PlagiarismCheck check = checkService.getById(checkId);
        if (check == null) throw new StatusNotFoundException("查重任务不存在");
        requireAccess(check.getCid());
        return check;
    }

    public Judge submission(Long submitId)
            throws StatusNotFoundException, StatusForbiddenException {
        Judge judge = judgeService.getById(submitId);
        if (judge == null) throw new StatusNotFoundException("提交不存在");
        requireAccess(judge.getCid());
        return judge;
    }

    public List<PlagiarismCheckConfig> allConfigs(Long cid) {
        return configService.list(new QueryWrapper<PlagiarismCheckConfig>().eq("cid", cid));
    }

    public List<ContestProblem> allContestProblems(Long cid) {
        return contestProblemService.list(new QueryWrapper<ContestProblem>().eq("cid", cid));
    }

    public List<Judge> acSubmissions(Long cid, java.util.Date start, java.util.Date end) {
        return judgeService.list(new QueryWrapper<Judge>().eq("cid", cid).eq("status", 0)
                .ge("submit_time", start).le("submit_time", end).orderByAsc("submit_time"));
    }

    public List<top.hcode.hoj.pojo.entity.contest.ContestRecord> records(Long cid) {
        return contestRecordService.list(new QueryWrapper<top.hcode.hoj.pojo.entity.contest.ContestRecord>()
                .eq("cid", cid));
    }

    public PlagiarismCheckEntityService checks() { return checkService; }
    public PlagiarismResultEntityService results() { return resultService; }
    public ContestEntityService contests() { return contestService; }

    private boolean isAdmin() {
        return SecurityUtils.getSubject().hasRole("root") || SecurityUtils.getSubject().hasRole("admin");
    }
}
