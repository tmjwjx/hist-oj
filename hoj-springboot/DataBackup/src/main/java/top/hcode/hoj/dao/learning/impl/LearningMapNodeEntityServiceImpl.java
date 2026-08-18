package top.hcode.hoj.dao.learning.impl;

import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import org.springframework.stereotype.Service;
import top.hcode.hoj.dao.learning.LearningMapNodeEntityService;
import top.hcode.hoj.mapper.LearningMapNodeMapper;
import top.hcode.hoj.pojo.entity.learning.LearningMapNode;

@Service
public class LearningMapNodeEntityServiceImpl extends ServiceImpl<LearningMapNodeMapper, LearningMapNode>
        implements LearningMapNodeEntityService {
}
