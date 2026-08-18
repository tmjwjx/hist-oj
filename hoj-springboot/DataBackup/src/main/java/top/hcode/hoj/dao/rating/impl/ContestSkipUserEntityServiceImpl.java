package top.hcode.hoj.dao.rating.impl;

import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import org.springframework.stereotype.Service;
import top.hcode.hoj.dao.rating.ContestSkipUserEntityService;
import top.hcode.hoj.mapper.ContestSkipUserMapper;
import top.hcode.hoj.pojo.entity.rating.ContestSkipUser;

@Service
public class ContestSkipUserEntityServiceImpl extends ServiceImpl<ContestSkipUserMapper, ContestSkipUser> implements ContestSkipUserEntityService { }
