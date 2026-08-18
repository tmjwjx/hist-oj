package top.hcode.hoj.dao.plagiarism.impl;

import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import org.springframework.stereotype.Service;
import top.hcode.hoj.dao.plagiarism.PlagiarismCheckEntityService;
import top.hcode.hoj.mapper.PlagiarismCheckMapper;
import top.hcode.hoj.pojo.entity.plagiarism.PlagiarismCheck;

@Service
public class PlagiarismCheckEntityServiceImpl
        extends ServiceImpl<PlagiarismCheckMapper, PlagiarismCheck>
        implements PlagiarismCheckEntityService {
}
