package top.hcode.hoj.dao.contest.impl;

import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import org.springframework.stereotype.Service;
import top.hcode.hoj.dao.contest.ContestQuestionReplyEntityService;
import top.hcode.hoj.mapper.ContestQuestionReplyMapper;
import top.hcode.hoj.pojo.entity.contest.ContestQuestionReply;

@Service
public class ContestQuestionReplyEntityServiceImpl
        extends ServiceImpl<ContestQuestionReplyMapper, ContestQuestionReply>
        implements ContestQuestionReplyEntityService {
}
