package top.hcode.hoj.dao.plagiarism.impl;

import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import org.springframework.stereotype.Service;
import top.hcode.hoj.dao.plagiarism.PlagiarismResultEntityService;
import top.hcode.hoj.mapper.PlagiarismResultMapper;
import top.hcode.hoj.pojo.entity.plagiarism.PlagiarismResult;

@Service
public class PlagiarismResultEntityServiceImpl
        extends ServiceImpl<PlagiarismResultMapper, PlagiarismResult>
        implements PlagiarismResultEntityService {
}
