package top.hcode.hoj.controller.admin;

import com.baomidou.mybatisplus.core.metadata.IPage;
import org.apache.shiro.authz.annotation.RequiresAuthentication;
import org.apache.shiro.authz.annotation.Logical;
import org.apache.shiro.authz.annotation.RequiresRoles;
import org.springframework.web.bind.annotation.*;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.manager.oj.ExternalVerificationManager;
import top.hcode.hoj.pojo.dto.ExternalVerificationSettingDTO;
import top.hcode.hoj.pojo.vo.ExternalVerificationLinkVO;
import top.hcode.hoj.pojo.vo.ExternalVerificationSubmissionVO;
import javax.annotation.Resource;

@RestController
@RequestMapping("/api/admin/problem")
@RequiresAuthentication
@RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
public class AdminExternalVerificationController {
    @Resource private ExternalVerificationManager manager;

    @GetMapping("/external-verification")
    public CommonResult<ExternalVerificationLinkVO> setting(@RequestParam Long pid)
            throws StatusFailException, StatusForbiddenException {
        return CommonResult.successResponse(manager.setting(pid));
    }

    @PutMapping("/external-verification")
    public CommonResult<ExternalVerificationLinkVO> saveSetting(@RequestBody ExternalVerificationSettingDTO dto)
            throws StatusFailException, StatusForbiddenException {
        return CommonResult.successResponse(manager.saveSetting(dto));
    }

    @GetMapping("/external-submissions")
    public CommonResult<IPage<ExternalVerificationSubmissionVO>> submissions(@RequestParam Long pid,
            @RequestParam(required = false) Integer currentPage, @RequestParam(required = false) Integer limit)
            throws StatusFailException, StatusForbiddenException {
        return CommonResult.successResponse(manager.adminSubmissions(pid, currentPage, limit));
    }

    @GetMapping("/external-submission")
    public CommonResult<ExternalVerificationSubmissionVO> submission(@RequestParam Long pid, @RequestParam Long submitId)
            throws StatusFailException, StatusForbiddenException {
        return CommonResult.successResponse(manager.adminSubmission(pid, submitId));
    }
}
