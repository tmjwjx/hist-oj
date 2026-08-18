package top.hcode.hoj.manager.oj;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import org.apache.shiro.SecurityUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.dao.training.TrainingEntityService;
import top.hcode.hoj.dao.training.TrainingParticipantEntityService;
import top.hcode.hoj.dao.training.TrainingRegisterEntityService;
import top.hcode.hoj.pojo.entity.training.Training;
import top.hcode.hoj.pojo.entity.training.TrainingParticipant;
import top.hcode.hoj.pojo.entity.training.TrainingRegister;
import top.hcode.hoj.pojo.vo.TrainingParticipantVO;
import top.hcode.hoj.pojo.vo.TrainingProgressVO;
import top.hcode.hoj.utils.ShiroUtils;

import java.util.Date;
import java.util.List;
import java.util.stream.Collectors;

@Component
public class TrainingParticipantManager {

    @Autowired
    private TrainingEntityService trainingEntityService;

    @Autowired
    private TrainingRegisterEntityService registerEntityService;

    @Autowired
    private TrainingParticipantEntityService participantEntityService;

    public TrainingParticipantVO join(Long tid)
            throws StatusNotFoundException, StatusForbiddenException, StatusFailException {
        Training training = requireTraining(tid);
        requireTrainingAccess(training);
        String uid = currentUid();
        TrainingParticipant participant = participantEntityService.getOne(
                new QueryWrapper<TrainingParticipant>()
                        .eq("training_id", tid)
                        .eq("uid", uid), false);
        if (participant == null) {
            Date now = new Date();
            participant = new TrainingParticipant()
                    .setTrainingId(tid)
                    .setUid(uid)
                    .setStatus("not_started")
                    .setJoinTime(now)
                    .setGmtCreate(now)
                    .setGmtModified(now);
            if (!participantEntityService.save(participant)) {
                throw new StatusFailException("参加训练失败");
            }
        }
        return getMyRecord(tid);
    }

    public List<TrainingParticipantVO> getParticipants(Long tid)
            throws StatusNotFoundException, StatusForbiddenException {
        Training training = requireTraining(tid);
        if (!isTrainingAdmin(training)) {
            throw new StatusForbiddenException("无权限查看训练参与者");
        }
        return participantEntityService.getParticipants(tid, null);
    }

    public TrainingParticipantVO getMyRecord(Long tid) throws StatusNotFoundException {
        requireTraining(tid);
        List<TrainingParticipantVO> records = participantEntityService.getParticipants(tid, currentUid());
        if (records.isEmpty()) {
            throw new StatusNotFoundException("未找到训练记录");
        }
        return records.get(0);
    }

    public void refreshStatus(Long tid) throws StatusNotFoundException, StatusFailException {
        TrainingParticipantVO record = getMyRecord(tid);
        TrainingParticipant update = new TrainingParticipant()
                .setId(record.getId())
                .setStatus(calculateStatus(record.getProgress()))
                .setGmtModified(new Date());
        if (!participantEntityService.updateById(update)) {
            throw new StatusFailException("更新训练进度失败");
        }
    }

    public List<TrainingProgressVO> getMyProgress() {
        return participantEntityService.getUserProgress(currentUid()).stream()
                .map(TrainingParticipantVO::getProgress)
                .collect(Collectors.toList());
    }

    private void requireTrainingAccess(Training training) throws StatusForbiddenException {
        if (isTrainingAdmin(training) || !"Private".equals(training.getAuth())) {
            return;
        }
        TrainingRegister register = registerEntityService.getOne(
                new QueryWrapper<TrainingRegister>()
                        .eq("tid", training.getId())
                        .eq("uid", currentUid()), false);
        if (register == null || Boolean.FALSE.equals(register.getStatus())) {
            throw new StatusForbiddenException("私有训练需要先通过密码验证");
        }
    }

    private Training requireTraining(Long tid) throws StatusNotFoundException {
        Training training = trainingEntityService.getById(tid);
        if (training == null || !Boolean.TRUE.equals(training.getStatus())) {
            throw new StatusNotFoundException("训练不存在或已停用");
        }
        return training;
    }

    private boolean isTrainingAdmin(Training training) {
        return currentUsername().equals(training.getAuthor())
                || SecurityUtils.getSubject().hasRole("root")
                || SecurityUtils.getSubject().hasRole("admin");
    }

    private String calculateStatus(TrainingProgressVO progress) {
        if (progress == null || progress.getSolvedCount() == 0) {
            return "not_started";
        }
        if (progress.getTotalCount() > 0
                && progress.getSolvedCount() >= progress.getTotalCount()) {
            return "completed";
        }
        return "in_progress";
    }

    private String currentUid() {
        return ShiroUtils.getProfile().getUid();
    }

    private String currentUsername() {
        return ShiroUtils.getProfile().getUsername();
    }
}
