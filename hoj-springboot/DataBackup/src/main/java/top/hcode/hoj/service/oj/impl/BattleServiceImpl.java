package top.hcode.hoj.service.oj.impl;

import org.springframework.stereotype.Service;
import lombok.extern.slf4j.Slf4j;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.common.result.ResultStatus;
import top.hcode.hoj.manager.oj.battle.BattleMatchManager;
import top.hcode.hoj.manager.oj.battle.BattleQueryManager;
import top.hcode.hoj.manager.oj.battle.BattleRoomManager;
import top.hcode.hoj.pojo.dto.BattleExcludeDTO;
import top.hcode.hoj.pojo.dto.BattleReadyDTO;
import top.hcode.hoj.pojo.dto.BattleRoomDTO;
import top.hcode.hoj.pojo.dto.BattleSubmitDTO;
import top.hcode.hoj.pojo.entity.battle.BattleRoom;
import top.hcode.hoj.pojo.vo.BattleRankPageVO;
import top.hcode.hoj.pojo.vo.BattleRecordPageVO;
import top.hcode.hoj.pojo.vo.BattleResultVO;
import top.hcode.hoj.pojo.vo.BattleRoomInfoVO;
import top.hcode.hoj.service.oj.BattleService;

import javax.annotation.Resource;

@Service
@Slf4j
public class BattleServiceImpl implements BattleService {

    @Resource
    private BattleRoomManager roomManager;

    @Resource
    private BattleMatchManager matchManager;

    @Resource
    private BattleQueryManager queryManager;

    @Override
    public CommonResult<BattleRoom> createRoom() {
        return run(() -> roomManager.create());
    }

    @Override
    public CommonResult<BattleRoom> joinRoom(BattleRoomDTO dto) {
        return run(() -> roomManager.join(dto.getRoomId()));
    }

    @Override
    public CommonResult<BattleRoom> ready(BattleReadyDTO dto) {
        return run(() -> roomManager.ready(dto.getRoomId(), dto.getReady()));
    }

    @Override
    public CommonResult<BattleRoomInfoVO> roomInfo(String roomId) {
        return run(() -> roomManager.info(roomId));
    }

    @Override
    public CommonResult<BattleRoomInfoVO> start(BattleRoomDTO dto) {
        return run(() -> matchManager.start(dto.getRoomId()));
    }

    @Override
    public CommonResult<BattleResultVO> giveup(BattleRoomDTO dto) {
        return run(() -> matchManager.giveup(dto.getRoomId()));
    }

    @Override
    public CommonResult<BattleResultVO> submitAC(BattleSubmitDTO dto) {
        return run(() -> matchManager.submitAC(dto.getRoomId(), dto.getProblemId()));
    }

    @Override
    public CommonResult<Void> dissolve(BattleRoomDTO dto) {
        return run(() -> {
            matchManager.dissolve(dto.getRoomId());
            return null;
        });
    }

    @Override
    public CommonResult<Void> leave(BattleRoomDTO dto) {
        return run(() -> {
            roomManager.leave(dto.getRoomId());
            return null;
        });
    }

    @Override
    public CommonResult<Void> reset(BattleRoomDTO dto) {
        return run(() -> {
            roomManager.reset(dto.getRoomId());
            return null;
        });
    }

    @Override
    public CommonResult<BattleRecordPageVO> myRecords(Integer limit, Integer currentPage) {
        return CommonResult.successResponse(queryManager.myRecords(limit, currentPage));
    }

    @Override
    public CommonResult<BattleRecordPageVO> allRecords(Integer limit, Integer currentPage,
                                                        String username, String roomId, String isWinner) {
        return CommonResult.successResponse(queryManager.allRecords(limit, currentPage, username, roomId, isWinner));
    }

    @Override
    public CommonResult<BattleRankPageVO> rank(Integer limit, Integer currentPage, String username) {
        return CommonResult.successResponse(queryManager.rank(limit, currentPage, username));
    }

    @Override
    public CommonResult<Void> exclude(BattleExcludeDTO dto) {
        return run(() -> {
            queryManager.exclude(dto.getRecordId(), dto.getIsExcluded());
            return null;
        });
    }

    private <T> CommonResult<T> run(Operation<T> operation) {
        try {
            return CommonResult.successResponse(operation.execute());
        } catch (StatusNotFoundException e) {
            return CommonResult.errorResponse(e.getMessage(), ResultStatus.NOT_FOUND);
        } catch (StatusForbiddenException e) {
            return CommonResult.errorResponse(e.getMessage(), ResultStatus.FORBIDDEN);
        } catch (StatusFailException e) {
            return CommonResult.errorResponse(e.getMessage());
        } catch (Exception e) {
            log.error("Battle operation failed", e);
            return CommonResult.errorResponse("对战服务暂时不可用");
        }
    }

    @FunctionalInterface
    private interface Operation<T> {
        T execute() throws Exception;
    }
}
