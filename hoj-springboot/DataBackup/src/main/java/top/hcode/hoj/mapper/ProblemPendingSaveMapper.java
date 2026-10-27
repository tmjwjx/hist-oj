package top.hcode.hoj.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import org.apache.ibatis.annotations.*;
import org.springframework.stereotype.Repository;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.pojo.entity.problem.ProblemPendingSave;

@Mapper
@Repository
public interface ProblemPendingSaveMapper extends BaseMapper<ProblemPendingSave> {
    @Select("SELECT * FROM problem_pending_save WHERE id = #{id} FOR UPDATE")
    ProblemPendingSave lock(@Param("id") Long id);

    @Select("SELECT * FROM problem WHERE id = #{id} FOR UPDATE")
    Problem lockProblem(@Param("id") Long id);
}
