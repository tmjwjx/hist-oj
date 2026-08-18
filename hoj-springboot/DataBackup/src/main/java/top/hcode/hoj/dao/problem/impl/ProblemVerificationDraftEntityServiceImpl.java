package top.hcode.hoj.dao.problem.impl;

import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import org.springframework.stereotype.Service;
import top.hcode.hoj.dao.problem.ProblemVerificationDraftEntityService;
import top.hcode.hoj.mapper.ProblemVerificationDraftMapper;
import top.hcode.hoj.pojo.entity.problem.ProblemVerificationDraft;

@Service
public class ProblemVerificationDraftEntityServiceImpl
        extends ServiceImpl<ProblemVerificationDraftMapper, ProblemVerificationDraft>
        implements ProblemVerificationDraftEntityService {}
