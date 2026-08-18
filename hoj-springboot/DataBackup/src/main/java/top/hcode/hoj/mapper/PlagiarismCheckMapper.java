package top.hcode.hoj.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import org.apache.ibatis.annotations.Mapper;
import org.springframework.stereotype.Repository;
import top.hcode.hoj.pojo.entity.plagiarism.PlagiarismCheck;

@Mapper
@Repository
public interface PlagiarismCheckMapper extends BaseMapper<PlagiarismCheck> {
}
