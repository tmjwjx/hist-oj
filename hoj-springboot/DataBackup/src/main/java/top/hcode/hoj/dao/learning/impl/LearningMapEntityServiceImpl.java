package top.hcode.hoj.dao.learning.impl;

import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import org.springframework.stereotype.Service;
import top.hcode.hoj.dao.learning.LearningMapEntityService;
import top.hcode.hoj.mapper.LearningMapMapper;
import top.hcode.hoj.pojo.entity.learning.LearningMap;

@Service
public class LearningMapEntityServiceImpl extends ServiceImpl<LearningMapMapper, LearningMap>
        implements LearningMapEntityService {
}
