package top.hcode.hoj.dao.training.impl;

import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import top.hcode.hoj.dao.training.TrainingParticipantEntityService;
import top.hcode.hoj.mapper.TrainingParticipantMapper;
import top.hcode.hoj.pojo.entity.training.TrainingParticipant;
import top.hcode.hoj.pojo.vo.TrainingParticipantVO;

import java.util.List;

@Service
public class TrainingParticipantEntityServiceImpl
        extends ServiceImpl<TrainingParticipantMapper, TrainingParticipant>
        implements TrainingParticipantEntityService {

    @Autowired
    private TrainingParticipantMapper mapper;

    @Override
    public List<TrainingParticipantVO> getParticipants(Long tid, String uid) {
        return mapper.getParticipants(tid, uid);
    }

    @Override
    public List<TrainingParticipantVO> getUserProgress(String uid) {
        return mapper.getUserProgress(uid);
    }
}
