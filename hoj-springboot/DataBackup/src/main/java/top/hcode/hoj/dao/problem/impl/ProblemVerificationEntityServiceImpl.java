package top.hcode.hoj.dao.problem.impl;

import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import org.springframework.stereotype.Service;
import top.hcode.hoj.dao.problem.ProblemVerificationEntityService;
import top.hcode.hoj.mapper.ProblemVerificationMapper;
import top.hcode.hoj.pojo.entity.problem.ProblemVerification;

@Service
public class ProblemVerificationEntityServiceImpl
        extends ServiceImpl<ProblemVerificationMapper, ProblemVerification>
        implements ProblemVerificationEntityService {
}
