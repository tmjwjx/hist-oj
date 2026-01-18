package top.hcode.hoj.controller.classroom;

import io.swagger.annotations.Api;
import io.swagger.annotations.ApiOperation;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.pojo.vo.classroom.QrcodeCheckinVO;
import top.hcode.hoj.service.classroom.CheckinService;
import top.hcode.hoj.util.SessionUtils;

/**
 * 班级签到控制器
 */
@Api(tags = "班级签到管理")
@RestController
@RequestMapping("/api/classroom")
public class CheckinController {

    @Autowired
    private CheckinService checkinService;

    /**
     * 获取当前二维码信息
     */
    @ApiOperation(value = "获取二维码信息")
    @GetMapping("/checkin/{checkinId}/qrcode")
    public CommonResult<QrcodeCheckinVO> getQrcode(@PathVariable Long checkinId) {
        try {
            QrcodeCheckinVO qrcodeInfo = checkinService.getQrcodeInfo(checkinId);
            return CommonResult.success(qrcodeInfo);
        } catch (Exception e) {
            return CommonResult.error(e.getMessage());
        }
    }

    /**
     * 刷新二维码
     */
    @ApiOperation(value = "刷新二维码")
    @PostMapping("/checkin/{checkinId}/qrcode/refresh")
    public CommonResult<QrcodeCheckinVO> refreshQrcode(@PathVariable Long checkinId) {
        try {
            QrcodeCheckinVO qrcodeInfo = checkinService.refreshQrcode(checkinId);
            return CommonResult.success(qrcodeInfo);
        } catch (Exception e) {
            return CommonResult.error(e.getMessage());
        }
    }

    /**
     * 学生通过二维码签到
     */
    @ApiOperation(value = "二维码签到")
    @PostMapping("/checkin/qrcode/submit")
    public CommonResult<Void> submitQrcodeCheckin(@RequestBody QrcodeCheckinDTO dto) {
        try {
            // 从Session获取当前用户ID
            String userId = SessionUtils.getUserInfo().getUuid();

            checkinService.submitQrcodeCheckin(dto.getToken(), dto.getCheckinId(), userId);

            return CommonResult.success();
        } catch (Exception e) {
            return CommonResult.error(e.getMessage());
        }
    }

    /**
     * 二维码签到请求DTO
     */
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
