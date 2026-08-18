package top.hcode.hoj.dao.problem.impl;

import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import org.springframework.stereotype.Service;
import top.hcode.hoj.dao.problem.ProblemAiConfigEntityService;
import top.hcode.hoj.mapper.ProblemAiConfigMapper;
import top.hcode.hoj.pojo.entity.problem.ProblemAiConfig;

@Service
public class ProblemAiConfigEntityServiceImpl extends ServiceImpl<ProblemAiConfigMapper, ProblemAiConfig>
        implements ProblemAiConfigEntityService {}
