package top.hcode.hoj.dao.rating.impl;

import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import org.springframework.stereotype.Service;
import top.hcode.hoj.dao.rating.ContestRatingStatusEntityService;
import top.hcode.hoj.mapper.ContestRatingStatusMapper;
import top.hcode.hoj.pojo.entity.rating.ContestRatingStatus;

@Service
public class ContestRatingStatusEntityServiceImpl extends ServiceImpl<ContestRatingStatusMapper, ContestRatingStatus> implements ContestRatingStatusEntityService { }
