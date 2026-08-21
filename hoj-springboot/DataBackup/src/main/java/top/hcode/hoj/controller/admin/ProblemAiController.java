package top.hcode.hoj.controller.admin;

import org.apache.shiro.authz.annotation.Logical;
import org.apache.shiro.authz.annotation.RequiresAuthentication;
import org.apache.shiro.authz.annotation.RequiresRoles;
import org.springframework.web.bind.annotation.*;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.manager.admin.problem.ProblemAiManager;
import top.hcode.hoj.pojo.dto.ProblemAiChatDTO;
import top.hcode.hoj.pojo.dto.ProblemAiGenerateProgramDTO;
import top.hcode.hoj.pojo.dto.ProblemAiValidateDTO;
import top.hcode.hoj.pojo.dto.ProblemAiRecheckDTO;
import top.hcode.hoj.pojo.entity.problem.ProblemAiConfig;
import top.hcode.hoj.pojo.entity.problem.ProblemAiRecord;
import top.hcode.hoj.pojo.vo.ProblemAiConfigVO;
import top.hcode.hoj.pojo.vo.ProblemAiGeneratedProgramVO;

import javax.annotation.Resource;
import java.util.List;

@RestController
@RequestMapping("/api/admin/problem-ai")
@RequiresAuthentication
@RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
public class ProblemAiController {
    @Resource private ProblemAiManager manager;

    @GetMapping("/config")
    public CommonResult<ProblemAiConfigVO> config() {
        try { return CommonResult.successResponse(manager.getConfig()); }
        catch (Exception e) { return CommonResult.errorResponse(e.getMessage()); }
    }

    @PutMapping("/config")
    public CommonResult<Void> saveConfig(@RequestBody ProblemAiConfig config) {
        try { manager.saveConfig(config); return CommonResult.successResponse(); }
        catch (Exception e) { return CommonResult.errorResponse(e.getMessage()); }
    }

    @GetMapping("/records")
    public CommonResult<List<ProblemAiRecord>> records(@RequestParam Long pid) {
        try { return CommonResult.successResponse(manager.records(pid)); }
        catch (Exception e) { return CommonResult.errorResponse(e.getMessage()); }
    }

    @GetMapping("/records/{id}")
    public CommonResult<ProblemAiRecord> record(@PathVariable Long id) {
        try { return CommonResult.successResponse(manager.record(id)); }
        catch (Exception e) { return CommonResult.errorResponse(e.getMessage()); }
    }

    @PostMapping("/chat")
    public CommonResult<ProblemAiRecord> chat(@RequestBody ProblemAiChatDTO dto) {
        try { return CommonResult.successResponse(manager.chat(dto)); }
        catch (Exception e) { return CommonResult.errorResponse(e.getMessage()); }
    }

    @PostMapping("/validate")
    public CommonResult<ProblemAiRecord> validate(@RequestBody ProblemAiValidateDTO dto) {
        try { return CommonResult.successResponse(manager.validate(dto)); }
        catch (Exception e) { return CommonResult.errorResponse(e.getMessage()); }
    }

    @PostMapping("/recheck")
    public CommonResult<ProblemAiRecord> recheck(@RequestBody ProblemAiRecheckDTO dto) {
        try { return CommonResult.successResponse(manager.recheck(dto)); }
        catch (Exception e) { return CommonResult.errorResponse(e.getMessage()); }
    }

    @PostMapping("/generate-standard-program")
    public CommonResult<ProblemAiGeneratedProgramVO> generateProgram(@RequestBody ProblemAiGenerateProgramDTO dto) {
        try { return CommonResult.successResponse(manager.generateProgram(dto)); }
        catch (Exception e) { return CommonResult.errorResponse(e.getMessage()); }
    }
}
