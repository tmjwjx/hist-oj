package top.hcode.hoj.controller.oj;

import org.apache.shiro.authz.annotation.Logical;
import org.apache.shiro.authz.annotation.RequiresAuthentication;
import org.apache.shiro.authz.annotation.RequiresRoles;
import org.springframework.web.bind.annotation.*;
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
import top.hcode.hoj.service.oj.BattleService;

import javax.annotation.Resource;
import javax.validation.Valid;

@RestController
@RequestMapping("/api")
@RequiresAuthentication
public class BattleController {

    @Resource
    private BattleService battleService;

    @PostMapping("/battle/create-room")
    public CommonResult<BattleRoom> createRoom() {
        return battleService.createRoom();
    }

    @PostMapping("/battle/join-room")
    public CommonResult<BattleRoom> joinRoom(@Valid @RequestBody BattleRoomDTO dto) {
        return battleService.joinRoom(dto);
    }

    @PostMapping("/battle/ready")
    public CommonResult<BattleRoom> ready(@Valid @RequestBody BattleReadyDTO dto) {
        return battleService.ready(dto);
    }

    @GetMapping("/battle/room-info")
    public CommonResult<BattleRoomInfoVO> roomInfo(@RequestParam String roomId) {
        return battleService.roomInfo(roomId);
    }

    @PostMapping("/battle/start-battle")
    public CommonResult<BattleRoomInfoVO> start(@Valid @RequestBody BattleRoomDTO dto) {
        return battleService.start(dto);
    }

    @PostMapping("/battle/giveup")
    public CommonResult<BattleResultVO> giveup(@Valid @RequestBody BattleRoomDTO dto) {
        return battleService.giveup(dto);
    }

    @PostMapping("/battle/submit-ac")
    public CommonResult<BattleResultVO> submitAC(@Valid @RequestBody BattleSubmitDTO dto) {
        return battleService.submitAC(dto);
    }

    @PostMapping("/battle/dissolve-room")
    public CommonResult<Void> dissolve(@Valid @RequestBody BattleRoomDTO dto) {
        return battleService.dissolve(dto);
    }

    @PostMapping("/battle/leave-room")
    public CommonResult<Void> leave(@Valid @RequestBody BattleRoomDTO dto) {
        return battleService.leave(dto);
    }

    @PostMapping("/battle/reset-room")
    public CommonResult<Void> reset(@Valid @RequestBody BattleRoomDTO dto) {
        return battleService.reset(dto);
    }

    @GetMapping("/battle/my-records")
    public CommonResult<BattleRecordPageVO> myRecords(
            @RequestParam(required = false) Integer limit,
            @RequestParam(required = false) Integer currentPage) {
        return battleService.myRecords(limit, currentPage);
    }

    @GetMapping("/battle/all-records")
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<BattleRecordPageVO> allRecords(
            @RequestParam(required = false) Integer limit,
            @RequestParam(required = false) Integer currentPage,
            @RequestParam(required = false) String username,
            @RequestParam(required = false) String roomId,
            @RequestParam(required = false) String isWinner) {
        return battleService.allRecords(limit, currentPage, username, roomId, isWinner);
    }

    @GetMapping("/battle/rank")
    public CommonResult<BattleRankPageVO> rank(
            @RequestParam(required = false) Integer limit,
            @RequestParam(required = false) Integer currentPage,
            @RequestParam(required = false) String username) {
        return battleService.rank(limit, currentPage, username);
    }

    @PutMapping("/admin/battle/record/exclude")
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<Void> exclude(@Valid @RequestBody BattleExcludeDTO dto) {
        return battleService.exclude(dto);
    }
}
