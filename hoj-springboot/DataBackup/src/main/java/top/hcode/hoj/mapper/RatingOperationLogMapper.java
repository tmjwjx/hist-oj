package top.hcode.hoj.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import org.apache.ibatis.annotations.Mapper;
import top.hcode.hoj.pojo.entity.rating.RatingOperationLog;

@Mapper
public interface RatingOperationLogMapper extends BaseMapper<RatingOperationLog> { }
