package top.hcode.hoj.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import org.apache.ibatis.annotations.Mapper;
import top.hcode.hoj.pojo.entity.rating.ContestRatingStatus;

@Mapper
public interface ContestRatingStatusMapper extends BaseMapper<ContestRatingStatus> { }
