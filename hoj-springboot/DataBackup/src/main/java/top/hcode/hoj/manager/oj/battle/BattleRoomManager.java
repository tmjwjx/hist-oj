package top.hcode.hoj.manager.oj.battle;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.dao.battle.BattleRoomEntityService;
import top.hcode.hoj.dao.problem.ProblemEntityService;
import top.hcode.hoj.dao.user.UserRecordEntityService;
import top.hcode.hoj.mapper.BattleRoomMapper;
import top.hcode.hoj.pojo.entity.battle.BattleRoom;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.pojo.entity.user.UserRecord;
import top.hcode.hoj.pojo.vo.BattleProblemVO;
import top.hcode.hoj.pojo.vo.BattleRoomInfoVO;
import top.hcode.hoj.utils.ShiroUtils;

import javax.annotation.Resource;
import java.security.SecureRandom;
import java.util.Date;

@Component
public class BattleRoomManager {

    private static final int WAITING = 0;
    private static final int IN_PROGRESS = 1;
    private static final int ENDED = 2;

    private static final String ROOM_CHARS = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
    private static final SecureRandom RANDOM = new SecureRandom();

    @Resource
    private BattleRoomEntityService roomService;

    @Resource
    private BattleRoomMapper roomMapper;

    @Resource
    private UserRecordEntityService userRecordService;

    @Resource
    private ProblemEntityService problemService;

    @Transactional(rollbackFor = Exception.class)
    public BattleRoom create() throws StatusFailException {
        String uid = uid();
        BattleRoom active = roomMapper.selectActiveRoomForUpdate(uid);
        if (active != null) {
            throw new StatusFailException("你已在房间 " + active.getRoomId() + " 中，请先退出当前房间");
        }
        for (int i = 0; i < 5; i++) {
            BattleRoom room = new BattleRoom()
                    .setRoomId(generateRoomId())
                    .setHostId(uid)
                    .setHostUsername(username())
                    .setStatus(WAITING)
                    .setChallengerReady(false)
                    .setGmtCreate(new Date())
                    .setGmtModified(new Date());
            if (roomService.save(room)) {
                return withRatings(room);
            }
        }
        throw new StatusFailException("创建房间失败，请重试");
    }

    @Transactional(rollbackFor = Exception.class)
    public BattleRoom join(String roomId) throws StatusFailException, StatusNotFoundException {
        String uid = uid();
        BattleRoom active = roomMapper.selectActiveRoomForUpdate(uid);
        if (active != null) {
            throw new StatusFailException("你已在房间 " + active.getRoomId() + " 中，请先退出当前房间");
        }
        BattleRoom room = lockRoom(roomId);
        if (room == null) {
            throw new StatusNotFoundException("房间不存在");
        }
        if (!Integer.valueOf(WAITING).equals(room.getStatus())) {
            throw new StatusFailException("房间已开始或已结束");
        }
        if (uid.equals(room.getHostId())) {
            throw new StatusFailException("不能加入自己创建的房间");
        }
        if (room.getChallengerId() != null) {
            throw new StatusFailException("房间已满员");
        }
        room.setChallengerId(uid)
                .setChallengerUsername(username())
                .setChallengerReady(false)
                .setGmtModified(new Date());
        roomServiceUpdate(room);
        return withRatings(room);
    }

    @Transactional(rollbackFor = Exception.class)
    public BattleRoom ready(String roomId, boolean ready)
            throws StatusFailException, StatusNotFoundException {
        BattleRoom room = lockRoom(roomId);
        if (room == null) {
            throw new StatusNotFoundException("房间不存在");
        }
        if (!Integer.valueOf(WAITING).equals(room.getStatus())) {
            throw new StatusFailException("房间已开始或已结束");
        }
        if (!uid().equals(room.getChallengerId())) {
            throw new StatusFailException("只有挑战者可以准备");
        }
        room.setChallengerReady(ready).setGmtModified(new Date());
        roomServiceUpdate(room);
        return withRatings(room);
    }

    public BattleRoomInfoVO info(String roomId) throws StatusNotFoundException {
        BattleRoom room = roomService.getOne(new QueryWrapper<BattleRoom>().eq("room_id", normalize(roomId)));
        if (room == null) {
            throw new StatusNotFoundException("房间不存在");
        }
        BattleRoomInfoVO info = new BattleRoomInfoVO().setRoom(withRatings(room));
        if (room.getProblemId() != null) {
            Problem problem = problemService.getOne(new QueryWrapper<Problem>()
                    .eq("problem_id", room.getProblemId()));
            if (problem != null) {
                info.setProblem(toProblem(problem));
            }
        }
        return info;
    }

    @Transactional(rollbackFor = Exception.class)
    public void leave(String roomId) throws StatusFailException, StatusNotFoundException {
        BattleRoom room = lockRoom(roomId);
        if (room == null) {
            throw new StatusNotFoundException("房间不存在");
        }
        if (Integer.valueOf(IN_PROGRESS).equals(room.getStatus())) {
            throw new StatusFailException("对战中不能退出房间");
        }
        String uid = uid();
        if (uid.equals(room.getHostId())) {
            if (room.getChallengerId() == null) {
                removeRoom(room);
            } else {
                room.setHostId(room.getChallengerId())
                        .setHostUsername(room.getChallengerUsername())
                        .setChallengerId(null)
                        .setChallengerUsername(null)
                        .setChallengerReady(false)
                        .setGmtModified(new Date());
                roomServiceUpdate(room);
            }
            return;
        }
        if (!uid.equals(room.getChallengerId())) {
            throw new StatusFailException("你不在该房间中");
        }
        room.setChallengerId(null)
                .setChallengerUsername(null)
                .setChallengerReady(false)
                .setGmtModified(new Date());
        roomServiceUpdate(room);
    }

    @Transactional(rollbackFor = Exception.class)
    public void reset(String roomId) throws StatusFailException, StatusNotFoundException {
        BattleRoom room = lockRoom(roomId);
        if (room == null) {
            throw new StatusNotFoundException("房间不存在");
        }
        if (!isParticipant(room)) {
            throw new StatusFailException("你不在该房间中");
        }
        if (Integer.valueOf(WAITING).equals(room.getStatus())) {
            return;
        }
        if (!Integer.valueOf(ENDED).equals(room.getStatus())) {
            throw new StatusFailException("只能重置已结束的对局");
        }
        room.setStatus(WAITING)
                .setProblemId(null)
                .setWinnerId(null)
                .setEndReason(null)
                .setStartTime(null)
                .setEndTime(null)
                .setChallengerReady(false)
                .setGmtModified(new Date());
        roomServiceUpdate(room);
    }

    BattleRoom lockRoom(String roomId) {
        return roomMapper.selectByRoomIdForUpdate(normalize(roomId));
    }

    BattleRoom withRatings(BattleRoom room) {
        UserRecord host = userRecordService.getOne(new QueryWrapper<UserRecord>().eq("uid", room.getHostId()));
        room.setHostRating(host == null ? null : host.getHistRating());
        if (room.getChallengerId() != null) {
            UserRecord challenger = userRecordService.getOne(
                    new QueryWrapper<UserRecord>().eq("uid", room.getChallengerId()));
            room.setChallengerRating(challenger == null ? null : challenger.getHistRating());
        }
        return room;
    }

    void roomServiceUpdate(BattleRoom room) throws StatusFailException {
        if (!roomService.updateById(room)) {
            throw new StatusFailException("保存房间状态失败");
        }
    }

    void removeRoom(BattleRoom room) throws StatusFailException {
        if (!roomService.removeById(room.getId())) {
            throw new StatusFailException("删除房间失败");
        }
    }

    BattleProblemVO toProblem(Problem problem) {
        return new BattleProblemVO()
                .setPid(problem.getId())
                .setDisplayId(problem.getProblemId())
                .setProblemId(problem.getProblemId())
                .setTitle(problem.getTitle())
                .setDifficulty(problem.getDifficulty())
                .setAuth(problem.getAuth());
    }

    boolean isParticipant(BattleRoom room) {
        return uid().equals(room.getHostId()) || uid().equals(room.getChallengerId());
    }

    String uid() {
        return ShiroUtils.getProfile().getUid();
    }

    String username() {
        return ShiroUtils.getProfile().getUsername();
    }

    static String normalize(String roomId) {
        return roomId == null ? "" : roomId.trim().toUpperCase();
    }

    private String generateRoomId() {
        StringBuilder roomId = new StringBuilder(6);
        for (int i = 0; i < 6; i++) {
            roomId.append(ROOM_CHARS.charAt(RANDOM.nextInt(ROOM_CHARS.length())));
        }
        return roomId.toString();
    }
}
