package top.hcode.hoj.dao.contest.impl;

import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import org.springframework.stereotype.Service;
import top.hcode.hoj.dao.contest.ContestQuestionEntityService;
import top.hcode.hoj.mapper.ContestQuestionMapper;
import top.hcode.hoj.pojo.entity.contest.ContestQuestion;

@Service
public class ContestQuestionEntityServiceImpl
        extends ServiceImpl<ContestQuestionMapper, ContestQuestion>
        implements ContestQuestionEntityService {
}
