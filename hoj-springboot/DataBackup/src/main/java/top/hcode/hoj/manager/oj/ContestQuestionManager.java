package top.hcode.hoj.manager.oj;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import com.baomidou.mybatisplus.core.metadata.IPage;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import org.apache.shiro.SecurityUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.dao.contest.ContestEntityService;
import top.hcode.hoj.dao.contest.ContestQuestionEntityService;
import top.hcode.hoj.dao.contest.ContestQuestionReplyEntityService;
import top.hcode.hoj.pojo.dto.ContestQuestionCreateDTO;
import top.hcode.hoj.pojo.entity.contest.Contest;
import top.hcode.hoj.pojo.entity.contest.ContestQuestion;
import top.hcode.hoj.pojo.entity.contest.ContestQuestionReply;
import top.hcode.hoj.pojo.vo.*;
import top.hcode.hoj.utils.ShiroUtils;

import java.util.Arrays;
import java.util.Date;
import java.util.List;

@Component
public class ContestQuestionManager {

    @Autowired
    private ContestEntityService contestEntityService;

    @Autowired
    private ContestQuestionEntityService questionService;

    @Autowired
    private ContestQuestionReplyEntityService replyService;

    @Autowired
    private ContestQuestionViewBuilder viewBuilder;

    public ContestQuestionVO create(ContestQuestionCreateDTO dto) throws StatusNotFoundException, StatusFailException {
        requireContest(dto.getContestId());
        Date now = new Date();
        ContestQuestion question = new ContestQuestion()
                .setContestId(dto.getContestId())
                .setQuestionerId(currentUid())
                .setTitle(dto.getTitle().trim())
                .setContent(dto.getContent().trim())
                .setStatus("pending")
                .setPriority(0)
                .setCreatedAt(now)
                .setUpdatedAt(now);
        if (!questionService.save(question)) {
            throw new StatusFailException("创建问题失败");
        }
        return viewBuilder.buildDetail(question, java.util.Collections.emptyList());
    }

    public ContestQuestionPageVO getPage(Long contestId, Integer page, Integer limit,
                                         String status, Boolean all)
            throws StatusNotFoundException, StatusForbiddenException {
        Contest contest = requireContest(contestId);
        boolean showAll = Boolean.TRUE.equals(all);
        if (showAll && !isContestAdmin(contest)) {
            throw new StatusForbiddenException("无权限访问全部问题");
        }
        int current = page == null || page < 1 ? 1 : page;
        int size = limit == null || limit < 1 ? 20 : Math.min(limit, 100);
        QueryWrapper<ContestQuestion> query = new QueryWrapper<ContestQuestion>()
                .eq("contest_id", contestId)
                .orderByDesc("created_at");
        if (!showAll) {
            query.eq("questioner_id", currentUid());
        }
        if (status != null && !status.trim().isEmpty() && !"all".equals(status)) {
            query.eq("status", status);
        }
        IPage<ContestQuestion> result = questionService.page(new Page<>(current, size), query);
        return new ContestQuestionPageVO()
                .setList(viewBuilder.buildList(result.getRecords()))
                .setTotal(result.getTotal());
    }

    public ContestQuestionVO getDetail(Long questionId)
            throws StatusNotFoundException, StatusForbiddenException {
        ContestQuestion question = requireQuestion(questionId);
        Contest contest = requireContest(question.getContestId());
        requireParticipantOrAdmin(question, contest);
        List<ContestQuestionReply> replies = replyService.list(
                new QueryWrapper<ContestQuestionReply>()
                        .eq("question_id", questionId)
                        .orderByAsc("created_at"));
        return viewBuilder.buildDetail(question, replies);
    }

    @Transactional(rollbackFor = Exception.class)
    public ContestQuestionReplyVO reply(Long questionId, String content)
            throws StatusNotFoundException, StatusForbiddenException, StatusFailException {
        ContestQuestion question = requireQuestion(questionId);
        Contest contest = requireContest(question.getContestId());
        requireParticipantOrAdmin(question, contest);
        ContestQuestionReply reply = new ContestQuestionReply()
                .setQuestionId(questionId)
                .setSenderId(currentUid())
                .setContent(content.trim())
                .setCreatedAt(new Date());
        if (!replyService.save(reply)) {
            throw new StatusFailException("发送回复失败");
        }
        if ("pending".equals(question.getStatus())) {
            question.setStatus("answered").setUpdatedAt(new Date());
            questionService.updateById(question);
        }
        return viewBuilder.buildReply(reply);
    }

    public void updateStatus(Long questionId, String status)
            throws StatusNotFoundException, StatusForbiddenException, StatusFailException {
        if (!Arrays.asList("pending", "answered", "closed").contains(status)) {
            throw new StatusFailException("无效的问题状态");
        }
        ContestQuestion question = requireQuestion(questionId);
        if (!isContestAdmin(requireContest(question.getContestId()))) {
            throw new StatusForbiddenException("无权限操作");
        }
        question.setStatus(status).setUpdatedAt(new Date());
        if (!questionService.updateById(question)) {
            throw new StatusFailException("更新问题状态失败");
        }
    }

    @Transactional(rollbackFor = Exception.class)
    public void delete(Long questionId)
            throws StatusNotFoundException, StatusForbiddenException, StatusFailException {
        ContestQuestion question = requireQuestion(questionId);
        Contest contest = requireContest(question.getContestId());
        requireParticipantOrAdmin(question, contest);
        replyService.remove(new QueryWrapper<ContestQuestionReply>().eq("question_id", questionId));
        if (!questionService.removeById(questionId)) {
            throw new StatusFailException("删除问题失败");
        }
    }

    private Contest requireContest(Long contestId) throws StatusNotFoundException {
        Contest contest = contestEntityService.getById(contestId);
        if (contest == null) {
            throw new StatusNotFoundException("比赛不存在");
        }
        return contest;
    }

    private ContestQuestion requireQuestion(Long questionId) throws StatusNotFoundException {
        ContestQuestion question = questionService.getById(questionId);
        if (question == null) {
            throw new StatusNotFoundException("问题不存在");
        }
        return question;
    }

    private void requireParticipantOrAdmin(ContestQuestion question, Contest contest)
            throws StatusForbiddenException {
        if (!currentUid().equals(question.getQuestionerId()) && !isContestAdmin(contest)) {
            throw new StatusForbiddenException("无权限访问");
        }
    }

    private boolean isContestAdmin(Contest contest) {
        return currentUid().equals(contest.getUid())
                || SecurityUtils.getSubject().hasRole("root")
                || SecurityUtils.getSubject().hasRole("admin");
    }

    private String currentUid() {
        return ShiroUtils.getProfile().getUid();
    }
}
