package top.hcode.hoj.dao.rating.impl;

import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import org.springframework.stereotype.Service;
import top.hcode.hoj.dao.rating.RatingRecalculateQueueEntityService;
import top.hcode.hoj.mapper.RatingRecalculateQueueMapper;
import top.hcode.hoj.pojo.entity.rating.RatingRecalculateQueue;

@Service
public class RatingRecalculateQueueEntityServiceImpl extends ServiceImpl<RatingRecalculateQueueMapper, RatingRecalculateQueue> implements RatingRecalculateQueueEntityService { }
