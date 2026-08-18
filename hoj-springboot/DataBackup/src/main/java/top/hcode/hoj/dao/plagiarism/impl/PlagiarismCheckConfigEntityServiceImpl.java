package top.hcode.hoj.dao.plagiarism.impl;

import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import org.springframework.stereotype.Service;
import top.hcode.hoj.dao.plagiarism.PlagiarismCheckConfigEntityService;
import top.hcode.hoj.mapper.PlagiarismCheckConfigMapper;
import top.hcode.hoj.pojo.entity.plagiarism.PlagiarismCheckConfig;

@Service
public class PlagiarismCheckConfigEntityServiceImpl
        extends ServiceImpl<PlagiarismCheckConfigMapper, PlagiarismCheckConfig>
        implements PlagiarismCheckConfigEntityService {
}
