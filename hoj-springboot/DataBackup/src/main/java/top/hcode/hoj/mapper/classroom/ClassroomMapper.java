package top.hcode.hoj.mapper.classroom;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import org.apache.ibatis.annotations.Mapper;
import top.hcode.hoj.pojo.entity.classroom.Classroom;

@Mapper
public interface ClassroomMapper extends BaseMapper<Classroom> {
}
