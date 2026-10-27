package top.hcode.hoj.controller.admin;

import org.apache.shiro.authz.annotation.*;
import org.springframework.web.bind.annotation.*;
import javax.annotation.Resource;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.manager.admin.problem.ProblemPendingSaveManager;
import top.hcode.hoj.pojo.dto.ProblemPendingSaveDTO;
import top.hcode.hoj.pojo.vo.ProblemPendingSaveVO;

@RestController
@RequestMapping("/api/admin/problem/pending-save")
@RequiresAuthentication
@RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
public class AdminProblemPendingSaveController {
    @Resource private ProblemPendingSaveManager manager;

    @PostMapping
    public CommonResult<ProblemPendingSaveVO> prepare(@RequestBody ProblemPendingSaveDTO dto)
            throws StatusFailException, StatusForbiddenException {
        return CommonResult.successResponse(manager.prepare(dto));
    }

    @GetMapping
    public CommonResult<ProblemPendingSaveVO> get(@RequestParam Long pendingId)
            throws StatusFailException, StatusForbiddenException {
        return CommonResult.successResponse(manager.get(pendingId));
    }

    @PostMapping("/commit")
    public CommonResult<ProblemPendingSaveVO> commit(@RequestBody ProblemPendingSaveDTO dto)
            throws StatusFailException, StatusForbiddenException {
        return CommonResult.successResponse(manager.commit(dto == null ? null : dto.getPendingId()));
    }
}
