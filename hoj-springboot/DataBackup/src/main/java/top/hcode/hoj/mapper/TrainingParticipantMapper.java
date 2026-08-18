package top.hcode.hoj.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.springframework.stereotype.Repository;
import top.hcode.hoj.pojo.entity.training.TrainingParticipant;
import top.hcode.hoj.pojo.vo.TrainingParticipantVO;

import java.util.List;

@Mapper
@Repository
public interface TrainingParticipantMapper extends BaseMapper<TrainingParticipant> {
    List<TrainingParticipantVO> getParticipants(@Param("tid") Long tid,
                                                @Param("uid") String uid);

    List<TrainingParticipantVO> getUserProgress(@Param("uid") String uid);
}
