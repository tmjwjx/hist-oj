package top.hcode.hoj.dao.training;

import com.baomidou.mybatisplus.extension.service.IService;
import top.hcode.hoj.pojo.entity.training.TrainingParticipant;
import top.hcode.hoj.pojo.vo.TrainingParticipantVO;

import java.util.List;

public interface TrainingParticipantEntityService extends IService<TrainingParticipant> {
    List<TrainingParticipantVO> getParticipants(Long tid, String uid);

    List<TrainingParticipantVO> getUserProgress(String uid);
}
