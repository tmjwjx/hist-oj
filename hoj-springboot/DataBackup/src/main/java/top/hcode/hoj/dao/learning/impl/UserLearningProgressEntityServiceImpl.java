package top.hcode.hoj.dao.learning.impl;

import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import org.springframework.stereotype.Service;
import top.hcode.hoj.dao.learning.UserLearningProgressEntityService;
import top.hcode.hoj.mapper.UserLearningProgressMapper;
import top.hcode.hoj.pojo.entity.learning.UserLearningProgress;

@Service
public class UserLearningProgressEntityServiceImpl
        extends ServiceImpl<UserLearningProgressMapper, UserLearningProgress>
        implements UserLearningProgressEntityService {
}
