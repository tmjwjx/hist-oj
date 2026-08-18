package top.hcode.hoj.manager.oj.battle;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.dao.judge.JudgeEntityService;
import top.hcode.hoj.dao.problem.ProblemEntityService;
import top.hcode.hoj.dao.user.UserAcproblemEntityService;
import top.hcode.hoj.pojo.entity.battle.BattleRoom;
import top.hcode.hoj.pojo.entity.judge.Judge;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.pojo.entity.user.UserAcproblem;
import top.hcode.hoj.pojo.vo.BattleProblemVO;
import top.hcode.hoj.pojo.vo.BattleResultVO;
import top.hcode.hoj.pojo.vo.BattleRoomInfoVO;

import javax.annotation.Resource;
import java.util.Date;
import java.util.HashSet;
import java.util.List;
import java.util.Set;
import java.util.concurrent.ThreadLocalRandom;

@Component
public class BattleMatchManager {

    @Resource
    private BattleRoomManager roomManager;

    @Resource
    private BattleResultManager resultManager;

    @Resource
    private ProblemEntityService problemService;

    @Resource
    private UserAcproblemEntityService acproblemService;

    @Resource
    private JudgeEntityService judgeService;

    @Transactional(rollbackFor = Exception.class)
    public BattleRoomInfoVO start(String roomId)
            throws StatusFailException, StatusNotFoundException {
        BattleRoom room = requireLockedRoom(roomId);
        if (!roomManager.uid().equals(room.getHostId())) {
            throw new StatusFailException("只有房主可以开始对战");
        }
        if (!Integer.valueOf(0).equals(room.getStatus())) {
            throw new StatusFailException("房间状态不正确");
        }
        if (room.getChallengerId() == null || !Boolean.TRUE.equals(room.getChallengerReady())) {
            throw new StatusFailException("挑战者尚未准备");
        }
        Problem problem = selectProblem(room.getHostId(), room.getChallengerId());
        Date now = new Date();
        room.setProblemId(problem.getProblemId())
                .setStatus(1)
                .setStartTime(now)
                .setGmtModified(now);
        roomManager.roomServiceUpdate(room);
        return new BattleRoomInfoVO()
                .setRoom(roomManager.withRatings(room))
                .setProblem(roomManager.toProblem(problem));
    }

    @Transactional(rollbackFor = Exception.class)
    public BattleResultVO giveup(String roomId)
            throws StatusFailException, StatusNotFoundException {
        BattleRoom room = requireLockedRoom(roomId);
        if (!Integer.valueOf(1).equals(room.getStatus())) {
            throw new StatusFailException("房间未开始或已结束");
        }
        String uid = roomManager.uid();
        String winner;
        if (uid.equals(room.getHostId())) {
            winner = room.getChallengerId();
        } else if (uid.equals(room.getChallengerId())) {
            winner = room.getHostId();
        } else {
            throw new StatusFailException("你不在该房间中");
        }
        resultManager.finish(room, winner, "giveup");
        return result(room, winner, "giveup");
    }

    @Transactional(rollbackFor = Exception.class)
    public BattleResultVO submitAC(String roomId, String problemId)
            throws StatusFailException, StatusNotFoundException {
        BattleRoom room = requireLockedRoom(roomId);
        if (!Integer.valueOf(1).equals(room.getStatus())) {
            throw new StatusFailException("房间未开始或已结束");
        }
        if (room.getProblemId() == null
                || !BattleRoomManager.normalize(room.getProblemId())
                .equals(BattleRoomManager.normalize(problemId))) {
            throw new StatusFailException("不是本场对战题目");
        }
        String uid = roomManager.uid();
        if (!uid.equals(room.getHostId()) && !uid.equals(room.getChallengerId())) {
            throw new StatusFailException("你不在该房间中");
        }
        long accepted = judgeService.count(new QueryWrapper<Judge>()
                .eq("uid", uid)
                .eq("display_pid", room.getProblemId())
                .eq("status", 0)
                .ge("submit_time", room.getStartTime()));
        if (accepted == 0) {
            throw new StatusFailException("尚未检测到本局有效的 Accepted 提交");
        }
        resultManager.finish(room, uid, "ac");
        return result(room, uid, "ac");
    }

    @Transactional(rollbackFor = Exception.class)
    public void dissolve(String roomId) throws StatusFailException, StatusNotFoundException {
        BattleRoom room = requireLockedRoom(roomId);
        if (!roomManager.uid().equals(room.getHostId())) {
            throw new StatusFailException("只有房主可以解散房间");
        }
        if (Integer.valueOf(1).equals(room.getStatus())) {
            if (room.getChallengerId() == null) {
                throw new StatusFailException("房间缺少挑战者信息");
            }
            resultManager.finish(room, room.getChallengerId(), "dissolve");
        } else if (Integer.valueOf(0).equals(room.getStatus()) || Integer.valueOf(2).equals(room.getStatus())) {
            roomManager.removeRoom(room);
        } else {
            throw new StatusFailException("房间状态不允许解散");
        }
    }

    private Problem selectProblem(String firstUid, String secondUid) throws StatusFailException {
        List<Problem> problems = problemService.list(new QueryWrapper<Problem>()
                .select("id", "problem_id", "title", "difficulty", "auth")
                .eq("auth", 1));
        Set<Long> solved = new HashSet<>();
        List<UserAcproblem> solvedRows = acproblemService.list(new QueryWrapper<UserAcproblem>()
                .select("pid")
                .in("uid", firstUid, secondUid));
        for (UserAcproblem row : solvedRows) {
            solved.add(row.getPid());
        }
        problems.removeIf(problem -> solved.contains(problem.getId())
                || problem.getProblemId() == null || problem.getProblemId().trim().isEmpty());
        if (problems.isEmpty()) {
            throw new StatusFailException("没有找到双方都未通过的公开题目");
        }
        return problems.get(ThreadLocalRandom.current().nextInt(problems.size()));
    }

    private BattleRoom requireLockedRoom(String roomId) throws StatusNotFoundException {
        BattleRoom room = roomManager.lockRoom(roomId);
        if (room == null) {
            throw new StatusNotFoundException("房间不存在");
        }
        return room;
    }

    private BattleResultVO result(BattleRoom room, String winner, String reason) {
        String winnerName = winner.equals(room.getHostId())
                ? room.getHostUsername() : room.getChallengerUsername();
        return new BattleResultVO()
                .setWinner(winnerName)
                .setWinnerId(winner)
                .setEndReason(reason)
                .setIsWinner(winner.equals(roomManager.uid()));
    }
}
