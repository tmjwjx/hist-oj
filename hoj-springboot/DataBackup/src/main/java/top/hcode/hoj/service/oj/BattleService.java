package top.hcode.hoj.service.oj;

import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.pojo.dto.BattleExcludeDTO;
import top.hcode.hoj.pojo.dto.BattleReadyDTO;
import top.hcode.hoj.pojo.dto.BattleRoomDTO;
import top.hcode.hoj.pojo.dto.BattleSubmitDTO;
import top.hcode.hoj.pojo.entity.battle.BattleRoom;
import top.hcode.hoj.pojo.vo.BattleRankPageVO;
import top.hcode.hoj.pojo.vo.BattleRecordPageVO;
import top.hcode.hoj.pojo.vo.BattleResultVO;
import top.hcode.hoj.pojo.vo.BattleRoomInfoVO;

public interface BattleService {

    CommonResult<BattleRoom> createRoom();

    CommonResult<BattleRoom> joinRoom(BattleRoomDTO dto);

    CommonResult<BattleRoom> ready(BattleReadyDTO dto);

    CommonResult<BattleRoomInfoVO> roomInfo(String roomId);

    CommonResult<BattleRoomInfoVO> start(BattleRoomDTO dto);

    CommonResult<BattleResultVO> giveup(BattleRoomDTO dto);

    CommonResult<BattleResultVO> submitAC(BattleSubmitDTO dto);

    CommonResult<Void> dissolve(BattleRoomDTO dto);

    CommonResult<Void> leave(BattleRoomDTO dto);

    CommonResult<Void> reset(BattleRoomDTO dto);

    CommonResult<BattleRecordPageVO> myRecords(Integer limit, Integer currentPage);

    CommonResult<BattleRecordPageVO> allRecords(Integer limit, Integer currentPage,
                                                 String username, String roomId, String isWinner);

    CommonResult<BattleRankPageVO> rank(Integer limit, Integer currentPage, String username);

    CommonResult<Void> exclude(BattleExcludeDTO dto);
}
