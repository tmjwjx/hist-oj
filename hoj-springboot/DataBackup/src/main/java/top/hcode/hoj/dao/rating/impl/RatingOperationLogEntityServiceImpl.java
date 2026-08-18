package top.hcode.hoj.dao.rating.impl;

import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import org.springframework.stereotype.Service;
import top.hcode.hoj.dao.rating.RatingOperationLogEntityService;
import top.hcode.hoj.mapper.RatingOperationLogMapper;
import top.hcode.hoj.pojo.entity.rating.RatingOperationLog;

@Service
public class RatingOperationLogEntityServiceImpl extends ServiceImpl<RatingOperationLogMapper, RatingOperationLog> implements RatingOperationLogEntityService { }
