package top.hcode.hoj.controller.classroom;

import io.swagger.annotations.Api;
import io.swagger.annotations.ApiOperation;
import org.apache.shiro.authz.annotation.RequiresAuthentication;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.manager.classroom.ClassroomAccessManager;
import top.hcode.hoj.manager.classroom.ClassroomCheckinManager;
import top.hcode.hoj.pojo.entity.classroom.Checkin;
import top.hcode.hoj.pojo.entity.classroom.CheckinRecord;
import top.hcode.hoj.pojo.vo.classroom.QrcodeCheckinVO;
import top.hcode.hoj.service.classroom.CheckinService;
import top.hcode.hoj.utils.ShiroUtils;

import java.util.List;
import java.util.Map;

/** 班级签到控制器。控制器放在后端模块，API 模块只保留可复用的领域代码。 */
@Api(tags = "班级签到管理")
@RestController
@RequestMapping("/api/classroom")
@RequiresAuthentication
public class CheckinController {

    @Autowired
    private CheckinService checkinService;

    @Autowired
    private ClassroomCheckinManager manager;

    @Autowired
    private ClassroomAccessManager accessManager;

    @PostMapping("/checkin")
    public CommonResult<Checkin> create(@RequestBody Map<String, Object> request) {
        try { return CommonResult.successResponse(manager.create(request)); }
        catch (Exception e) { return CommonResult.errorResponse(e.getMessage()); }
    }

    @PostMapping("/checkin/submit")
    public CommonResult<CheckinRecord> submit(@RequestBody Map<String, Object> request) {
        try { return CommonResult.successResponse(manager.checkin(request)); }
        catch (Exception e) { return CommonResult.errorResponse(e.getMessage()); }
    }

    @GetMapping("/{classroomId}/checkins")
    public CommonResult<List<Checkin>> list(@PathVariable Long classroomId) {
        try { return CommonResult.successResponse(manager.list(classroomId, false)); }
        catch (Exception e) { return CommonResult.errorResponse(e.getMessage()); }
    }

    @GetMapping("/{classroomId}/checkins/student")
    public CommonResult<List<Map<String, Object>>> studentList(@PathVariable Long classroomId) {
        try { return CommonResult.successResponse(manager.studentList(classroomId)); }
        catch (Exception e) { return CommonResult.errorResponse(e.getMessage()); }
    }

    @GetMapping("/checkin/{checkinId}/records")
    public CommonResult<List<Map<String, Object>>> records(@PathVariable Long checkinId) {
        try { return CommonResult.successResponse(manager.records(checkinId)); }
        catch (Exception e) { return CommonResult.errorResponse(e.getMessage()); }
    }

    @PostMapping("/checkin/{checkinId}/record")
    public CommonResult<Void> addRecord(@PathVariable Long checkinId, @RequestBody Map<String, Object> request) {
        try { manager.addRecord(checkinId, request); return CommonResult.successResponse(); }
        catch (Exception e) { return CommonResult.errorResponse(e.getMessage()); }
    }

    @PutMapping("/checkin/record")
    public CommonResult<Void> updateRecord(@RequestBody Map<String, Object> request) {
        try { manager.updateRecord(request); return CommonResult.successResponse(); }
        catch (Exception e) { return CommonResult.errorResponse(e.getMessage()); }
    }

    @PostMapping("/checkin/{checkinId}/end")
    public CommonResult<Void> end(@PathVariable Long checkinId) {
        try { manager.end(checkinId); return CommonResult.successResponse(); }
        catch (Exception e) { return CommonResult.errorResponse(e.getMessage()); }
    }

    @PutMapping("/checkin/{checkinId}")
    public CommonResult<Void> update(@PathVariable Long checkinId, @RequestBody Map<String, Object> request) {
        try { manager.update(checkinId, request); return CommonResult.successResponse(); }
        catch (Exception e) { return CommonResult.errorResponse(e.getMessage()); }
    }

    @DeleteMapping("/checkin/{checkinId}")
    public CommonResult<Void> delete(@PathVariable Long checkinId) {
        try { manager.delete(checkinId); return CommonResult.successResponse(); }
        catch (Exception e) { return CommonResult.errorResponse(e.getMessage()); }
    }

    @ApiOperation(value = "获取二维码信息")
    @GetMapping("/checkin/{checkinId}/qrcode")
    public CommonResult<QrcodeCheckinVO> getQrcode(@PathVariable Long checkinId) {
        try {
            accessManager.requireManager(manager.classroomId(checkinId));
            return CommonResult.successResponse(checkinService.getQrcodeInfo(checkinId));
        } catch (Exception e) {
            return CommonResult.errorResponse(e.getMessage());
        }
    }

    @ApiOperation(value = "刷新二维码")
    @PostMapping("/checkin/{checkinId}/qrcode/refresh")
    public CommonResult<QrcodeCheckinVO> refreshQrcode(@PathVariable Long checkinId) {
        try {
            accessManager.requireManager(manager.classroomId(checkinId));
            return CommonResult.successResponse(checkinService.refreshQrcode(checkinId));
        } catch (Exception e) {
            return CommonResult.errorResponse(e.getMessage());
        }
    }

    @ApiOperation(value = "二维码签到")
    @PostMapping("/checkin/qrcode/submit")
    public CommonResult<Void> submitQrcodeCheckin(@RequestBody QrcodeCheckinDTO dto) {
        try {
            String userId = ShiroUtils.getProfile().getUid();
            checkinService.submitQrcodeCheckin(dto.getToken(), dto.getCheckinId(), userId);
            return CommonResult.successResponse();
        } catch (Exception e) {
            return CommonResult.errorResponse(e.getMessage());
        }
    }

    static class QrcodeCheckinDTO {
        private String token;
        private Long checkinId;

        public String getToken() {
            return token;
        }

        public void setToken(String token) {
            this.token = token;
        }

        public Long getCheckinId() {
            return checkinId;
        }

        public void setCheckinId(Long checkinId) {
            this.checkinId = checkinId;
        }
    }
}
