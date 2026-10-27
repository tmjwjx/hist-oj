package top.hcode.hoj.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.baomidou.mybatisplus.core.metadata.IPage;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import org.apache.ibatis.annotations.*;
import top.hcode.hoj.pojo.entity.judge.ExternalVerificationSubmission;
import top.hcode.hoj.pojo.vo.ExternalVerificationSubmissionVO;

@Mapper
public interface ExternalVerificationSubmissionMapper extends BaseMapper<ExternalVerificationSubmission> {
    @Select("<script>SELECT j.submit_id, e.nickname, j.language, j.status, j.time, j.memory, j.score, j.submit_time "
            + "FROM external_verification_submission e JOIN judge j ON j.submit_id = e.submit_id AND j.pid = e.pid "
            + "WHERE e.pid = #{pid} AND j.is_problem_verification = 1 AND j.submission_type = 'external_verification' "
            + "<if test='visitorHash != null'>AND e.visitor_token_hash = #{visitorHash} </if>"
            + "ORDER BY e.submit_id DESC</script>")
    IPage<ExternalVerificationSubmissionVO> submissions(Page<ExternalVerificationSubmissionVO> page,
                                                       @Param("pid") Long pid,
                                                       @Param("visitorHash") String visitorHash);
}
