package top.hcode.hoj.manager.oj;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import top.hcode.hoj.dao.user.UserInfoEntityService;
import top.hcode.hoj.dao.user.UserRecordEntityService;
import top.hcode.hoj.pojo.entity.contest.ContestQuestion;
import top.hcode.hoj.pojo.entity.contest.ContestQuestionReply;
import top.hcode.hoj.pojo.entity.user.UserInfo;
import top.hcode.hoj.pojo.entity.user.UserRecord;
import top.hcode.hoj.pojo.vo.ContestQuestionReplyVO;
import top.hcode.hoj.pojo.vo.ContestQuestionUserVO;
import top.hcode.hoj.pojo.vo.ContestQuestionVO;

import java.util.*;
import java.util.function.Function;
import java.util.stream.Collectors;

@Component
public class ContestQuestionViewBuilder {

    @Autowired
    private UserInfoEntityService userInfoEntityService;

    @Autowired
    private UserRecordEntityService userRecordEntityService;

    public List<ContestQuestionVO> buildList(List<ContestQuestion> questions) {
        Set<String> uids = questions.stream()
                .map(ContestQuestion::getQuestionerId)
                .collect(Collectors.toSet());
        Map<String, ContestQuestionUserVO> users = loadUsers(uids);
        return questions.stream()
                .map(question -> buildQuestion(question, users, Collections.emptyList()))
                .collect(Collectors.toList());
    }

    public ContestQuestionVO buildDetail(ContestQuestion question, List<ContestQuestionReply> replies) {
        Set<String> uids = new HashSet<>();
        uids.add(question.getQuestionerId());
        replies.forEach(reply -> uids.add(reply.getSenderId()));
        Map<String, ContestQuestionUserVO> users = loadUsers(uids);
        List<ContestQuestionReplyVO> replyViews = replies.stream()
                .map(reply -> buildReply(reply, users))
                .collect(Collectors.toList());
        return buildQuestion(question, users, replyViews);
    }

    public ContestQuestionReplyVO buildReply(ContestQuestionReply reply) {
        Map<String, ContestQuestionUserVO> users = loadUsers(
                Collections.singleton(reply.getSenderId()));
        return buildReply(reply, users);
    }

    private ContestQuestionVO buildQuestion(ContestQuestion question,
                                            Map<String, ContestQuestionUserVO> users,
                                            List<ContestQuestionReplyVO> replies) {
        return new ContestQuestionVO()
                .setId(question.getId())
                .setContestId(question.getContestId())
                .setQuestionerId(question.getQuestionerId())
                .setTitle(question.getTitle())
                .setContent(question.getContent())
                .setStatus(question.getStatus())
                .setPriority(question.getPriority())
                .setCreatedAt(question.getCreatedAt())
                .setUpdatedAt(question.getUpdatedAt())
                .setQuestioner(users.get(question.getQuestionerId()))
                .setReplies(replies);
    }

    private ContestQuestionReplyVO buildReply(ContestQuestionReply reply,
                                               Map<String, ContestQuestionUserVO> users) {
        return new ContestQuestionReplyVO()
                .setId(reply.getId())
                .setQuestionId(reply.getQuestionId())
                .setSenderId(reply.getSenderId())
                .setContent(reply.getContent())
                .setCreatedAt(reply.getCreatedAt())
                .setSender(users.get(reply.getSenderId()));
    }

    private Map<String, ContestQuestionUserVO> loadUsers(Collection<String> uids) {
        if (uids.isEmpty()) {
            return Collections.emptyMap();
        }
        List<UserInfo> users = userInfoEntityService.list(
                new QueryWrapper<UserInfo>().in("uuid", uids));
        Map<String, Integer> ratings = userRecordEntityService.list(
                        new QueryWrapper<UserRecord>().in("uid", uids))
                .stream()
                .collect(Collectors.toMap(
                        UserRecord::getUid,
                        record -> Optional.ofNullable(record.getHistRating()).orElse(0),
                        (left, right) -> left));
        return users.stream().map(user -> new ContestQuestionUserVO()
                        .setUid(user.getUuid())
                        .setUsername(user.getUsername())
                        .setNickname(user.getNickname())
                        .setAvatar(user.getAvatar())
                        .setTitleName(user.getTitleName())
                        .setTitleColor(user.getTitleColor())
                        .setHistRating(ratings.getOrDefault(user.getUuid(), 0)))
                .collect(Collectors.toMap(ContestQuestionUserVO::getUid, Function.identity()));
    }
}
