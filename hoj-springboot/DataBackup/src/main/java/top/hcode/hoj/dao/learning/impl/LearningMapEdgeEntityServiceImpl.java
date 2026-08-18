package top.hcode.hoj.dao.learning.impl;

import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import org.springframework.stereotype.Service;
import top.hcode.hoj.dao.learning.LearningMapEdgeEntityService;
import top.hcode.hoj.mapper.LearningMapEdgeMapper;
import top.hcode.hoj.pojo.entity.learning.LearningMapEdge;

@Service
public class LearningMapEdgeEntityServiceImpl extends ServiceImpl<LearningMapEdgeMapper, LearningMapEdge>
        implements LearningMapEdgeEntityService {
}
