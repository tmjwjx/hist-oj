package top.hcode.hoj.manager.oj.battle;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.dao.battle.BattleRecordEntityService;
import top.hcode.hoj.dao.battle.BattleRoomEntityService;
import top.hcode.hoj.dao.judge.JudgeEntityService;
import top.hcode.hoj.dao.problem.ProblemEntityService;
import top.hcode.hoj.dao.user.UserRecordEntityService;
import top.hcode.hoj.pojo.entity.battle.BattleRecord;
import top.hcode.hoj.pojo.entity.battle.BattleRoom;
import top.hcode.hoj.pojo.entity.judge.Judge;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.pojo.entity.user.UserRecord;

import javax.annotation.Resource;
import java.util.Date;
import java.util.UUID;

@Component
public class BattleResultManager {

    @Resource
    private BattleRoomEntityService roomService;

    @Resource
    private BattleRecordEntityService recordService;

    @Resource
    private ProblemEntityService problemService;

    @Resource
    private UserRecordEntityService userRecordService;

    @Resource
    private JudgeEntityService judgeService;

    @Transactional(rollbackFor = Exception.class)
    public void finish(BattleRoom room, String winnerId, String endReason)
            throws StatusFailException {
        if (room.getChallengerId() == null || room.getProblemId() == null || room.getStartTime() == null) {
            throw new StatusFailException("对战信息不完整，无法结算");
        }
        Date endTime = new Date();
        int battleTime = Math.max(0, (int) ((endTime.getTime() - room.getStartTime().getTime()) / 1000));
        room.setStatus(2)
                .setWinnerId(winnerId)
                .setEndReason(endReason)
                .setEndTime(endTime)
                .setGmtModified(endTime);
        if (!roomService.updateById(room)) {
            throw new StatusFailException("保存对战结果失败");
        }

        Problem problem = problemService.getOne(new QueryWrapper<Problem>()
                .eq("problem_id", room.getProblemId()));
        String title = problem == null ? room.getProblemId() : problem.getTitle();
        UserRecord hostRecord = getRecord(room.getHostId());
        UserRecord challengerRecord = getRecord(room.getChallengerId());
        String pairId = UUID.randomUUID().toString().replace("-", "");

        BattleRecord host = buildRecord(room, pairId, room.getHostId(), room.getHostUsername(),
                room.getChallengerId(), room.getChallengerUsername(),
                challengerRecord == null ? null : challengerRecord.getHistRating(),
                winnerId, endReason, title, battleTime, countSubmits(room.getHostId(), room));
        BattleRecord challenger = buildRecord(room, pairId, room.getChallengerId(), room.getChallengerUsername(),
                room.getHostId(), room.getHostUsername(),
                hostRecord == null ? null : hostRecord.getHistRating(),
                winnerId, endReason, title, battleTime, countSubmits(room.getChallengerId(), room));
        if (!recordService.save(host) || !recordService.save(challenger)) {
            throw new StatusFailException("创建对战记录失败");
        }
    }

    private BattleRecord buildRecord(BattleRoom room, String pairId, String uid, String username,
                                     String opponentId, String opponentUsername, Integer opponentRating,
                                     String winnerId, String endReason, String title, int battleTime,
                                     int submitCount) {
        return new BattleRecord()
                .setBattlePairId(pairId)
                .setRoomId(room.getRoomId())
                .setUserId(uid)
                .setUsername(username)
                .setOpponentId(opponentId)
                .setOpponentUsername(opponentUsername)
                .setOpponentRating(opponentRating)
                .setProblemId(room.getProblemId())
                .setProblemTitle(title)
                .setIsWinner(uid.equals(winnerId))
                .setEndReason(endReason)
                .setSubmitCount(submitCount)
                .setBattleTime(battleTime)
                .setIsExcluded(false)
                .setGmtCreate(new Date());
    }

    private int countSubmits(String uid, BattleRoom room) {
        return (int) judgeService.count(new QueryWrapper<Judge>()
                .eq("uid", uid)
                .eq("display_pid", room.getProblemId())
                .ge("submit_time", room.getStartTime()));
    }

    private UserRecord getRecord(String uid) {
        return userRecordService.getOne(new QueryWrapper<UserRecord>().eq("uid", uid));
    }
}
