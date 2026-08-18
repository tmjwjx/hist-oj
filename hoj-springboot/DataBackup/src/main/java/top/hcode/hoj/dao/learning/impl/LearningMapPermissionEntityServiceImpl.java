package top.hcode.hoj.dao.learning.impl;

import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import org.springframework.stereotype.Service;
import top.hcode.hoj.dao.learning.LearningMapPermissionEntityService;
import top.hcode.hoj.mapper.LearningMapPermissionMapper;
import top.hcode.hoj.pojo.entity.learning.LearningMapPermission;

@Service
public class LearningMapPermissionEntityServiceImpl
        extends ServiceImpl<LearningMapPermissionMapper, LearningMapPermission>
        implements LearningMapPermissionEntityService {
}
