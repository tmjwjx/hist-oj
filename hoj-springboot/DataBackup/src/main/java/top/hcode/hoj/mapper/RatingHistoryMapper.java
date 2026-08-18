package top.hcode.hoj.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import org.apache.ibatis.annotations.Mapper;
import top.hcode.hoj.pojo.entity.rating.RatingHistory;

@Mapper
public interface RatingHistoryMapper extends BaseMapper<RatingHistory> { }
