package top.hcode.hoj.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import org.apache.ibatis.annotations.*;
import top.hcode.hoj.pojo.entity.problem.ProblemExternalVerification;

@Mapper
public interface ProblemExternalVerificationMapper extends BaseMapper<ProblemExternalVerification> {
    @Insert("INSERT INTO problem_external_verification (pid, token, enabled, owner_uid, owner_username) "
            + "VALUES (#{pid}, #{token}, #{enabled}, #{ownerUid}, #{ownerUsername}) "
            + "ON DUPLICATE KEY UPDATE enabled = VALUES(enabled), owner_uid = VALUES(owner_uid), "
            + "owner_username = VALUES(owner_username)")
    int saveSetting(ProblemExternalVerification setting);

    @Select("SELECT * FROM problem_external_verification WHERE token = #{token} FOR UPDATE")
    ProblemExternalVerification lockByToken(@Param("token") String token);
}
