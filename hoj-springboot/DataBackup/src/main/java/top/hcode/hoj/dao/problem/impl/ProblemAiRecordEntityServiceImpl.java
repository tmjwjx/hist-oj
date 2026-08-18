package top.hcode.hoj.dao.problem.impl;

import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import org.springframework.stereotype.Service;
import top.hcode.hoj.dao.problem.ProblemAiRecordEntityService;
import top.hcode.hoj.mapper.ProblemAiRecordMapper;
import top.hcode.hoj.pojo.entity.problem.ProblemAiRecord;

@Service
public class ProblemAiRecordEntityServiceImpl extends ServiceImpl<ProblemAiRecordMapper, ProblemAiRecord>
        implements ProblemAiRecordEntityService {}
