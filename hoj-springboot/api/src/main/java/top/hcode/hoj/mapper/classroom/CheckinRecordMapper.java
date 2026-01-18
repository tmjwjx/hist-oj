package top.hcode.hoj.mapper.classroom;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import org.apache.ibatis.annotations.Param;
import top.hcode.hoj.pojo.entity.classroom.CheckinRecord;

/**
 * 签到记录Mapper接口
 */
public interface CheckinRecordMapper extends BaseMapper<CheckinRecord> {

    /**
     * 根据签到ID和用户ID查询记录
     */
    CheckinRecord selectByCheckinIdAndUserId(@Param("checkinId") Long checkinId,
                                             @Param("uid") String uid);
}
