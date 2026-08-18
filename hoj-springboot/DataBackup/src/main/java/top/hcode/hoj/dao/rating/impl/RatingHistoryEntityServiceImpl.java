package top.hcode.hoj.dao.rating.impl;

import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import org.springframework.stereotype.Service;
import top.hcode.hoj.dao.rating.RatingHistoryEntityService;
import top.hcode.hoj.mapper.RatingHistoryMapper;
import top.hcode.hoj.pojo.entity.rating.RatingHistory;

@Service
public class RatingHistoryEntityServiceImpl extends ServiceImpl<RatingHistoryMapper, RatingHistory> implements RatingHistoryEntityService { }
