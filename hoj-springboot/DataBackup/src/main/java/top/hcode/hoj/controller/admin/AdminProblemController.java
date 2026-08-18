package top.hcode.hoj.controller.admin;


import com.baomidou.mybatisplus.core.metadata.IPage;

import org.apache.shiro.authz.annotation.Logical;
import org.apache.shiro.authz.annotation.RequiresAuthentication;
import org.apache.shiro.authz.annotation.RequiresRoles;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.pojo.dto.ProblemDTO;
import top.hcode.hoj.pojo.dto.CompileDTO;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.pojo.entity.problem.ProblemCase;
import top.hcode.hoj.pojo.entity.problem.ProblemVerificationDraft;
import top.hcode.hoj.pojo.dto.ProblemVerificationSubmitDTO;
import top.hcode.hoj.pojo.dto.LastAcceptedCodeVO;
import top.hcode.hoj.pojo.vo.ProblemVerificationVO;
import top.hcode.hoj.service.admin.problem.AdminProblemService;

import java.util.*;

/**
 * @Author: Himit_ZH
 * @Date: 2020/12/11 21:45
 * @Description:
 */
@RestController
@RequestMapping("/api/admin/problem")
public class AdminProblemController {

    @Autowired
    private AdminProblemService adminProblemService;

    @GetMapping("/get-problem-list")
    @RequiresAuthentication
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<IPage<Problem>> getProblemList(@RequestParam(value = "limit", required = false) Integer limit,
                                                       @RequestParam(value = "currentPage", required = false) Integer currentPage,
                                                       @RequestParam(value = "keyword", required = false) String keyword,
                                                       @RequestParam(value = "auth", required = false) Integer auth,
                                                       @RequestParam(value = "oj", required = false) String oj) {
        return adminProblemService.getProblemList(limit, currentPage, keyword, auth, oj);
    }

    @GetMapping("")
    @RequiresAuthentication
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<Problem> getProblem(@RequestParam("pid") Long pid) {
        return adminProblemService.getProblem(pid);
    }

    @DeleteMapping("")
    @RequiresAuthentication
    @RequiresRoles(value = {"root", "problem_admin"}, logical = Logical.OR)
    public CommonResult<Void> deleteProblem(@RequestParam("pid") Long pid) {
        return adminProblemService.deleteProblem(pid);
    }

    @PostMapping("")
    @RequiresAuthentication
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<Map<String, Object>> addProblem(@RequestBody ProblemDTO problemDto) {
        return adminProblemService.addProblem(problemDto);
    }

    @PutMapping("")
    @RequiresAuthentication
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<Void> updateProblem(@RequestBody ProblemDTO problemDto) {
        return adminProblemService.updateProblem(problemDto);
    }

    @GetMapping("/get-problem-cases")
    @RequiresAuthentication
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<List<ProblemCase>> getProblemCases(@RequestParam("pid") Long pid,
                                                           @RequestParam(value = "isUpload", defaultValue = "true") Boolean isUpload) {
        return adminProblemService.getProblemCases(pid, isUpload);
    }

    @PostMapping("/compile-spj")
    @RequiresAuthentication
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult compileSpj(@RequestBody CompileDTO compileDTO) {
        return adminProblemService.compileSpj(compileDTO);
    }

    @PostMapping("/compile-interactive")
    @RequiresAuthentication
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult compileInteractive(@RequestBody CompileDTO compileDTO) {
        return adminProblemService.compileInteractive(compileDTO);
    }

    @GetMapping("/import-remote-oj-problem")
    @RequiresAuthentication
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<Void> importRemoteOJProblem(@RequestParam("name") String name,
                                                    @RequestParam("problemId") String problemId) {
        return adminProblemService.importRemoteOJProblem(name, problemId);
    }

    @PutMapping("/change-problem-auth")
    @RequiresAuthentication
    @RequiresRoles(value = {"root", "problem_admin", "admin"}, logical = Logical.OR)
    public CommonResult<Void> changeProblemAuth(@RequestBody Problem problem) {
        return adminProblemService.changeProblemAuth(problem);
    }

    @GetMapping("/verification")
    @RequiresAuthentication
    public CommonResult<ProblemVerificationVO> getVerification(@RequestParam("pid") Long pid) {
        return adminProblemService.getVerification(pid);
    }

    @GetMapping("/verification/last-passed-code")
    @RequiresAuthentication
    public CommonResult<LastAcceptedCodeVO> getLastPassedVerificationCode(@RequestParam("pid") Long pid) {
        return adminProblemService.getLastPassedVerificationCode(pid);
    }

    @PostMapping("/verification/submit")
    @RequiresAuthentication
    public CommonResult<ProblemVerificationVO> submitVerification(
            @RequestBody ProblemVerificationSubmitDTO dto) {
        return adminProblemService.submitVerification(dto);
    }

    @PostMapping("/verification/sync")
    @RequiresAuthentication
    public CommonResult<ProblemVerificationVO> retryVerificationSync(@RequestParam("pid") Long pid) {
        return adminProblemService.retryVerificationSync(pid);
    }

    @GetMapping("/verification/draft")
    @RequiresAuthentication
    public CommonResult<ProblemVerificationDraft> getVerificationDraft(@RequestParam("pid") Long pid) {
        return adminProblemService.getVerificationDraft(pid);
    }

    @PostMapping("/verification/draft")
    @RequiresAuthentication
    public CommonResult<ProblemVerificationDraft> saveVerificationDraft(@RequestBody ProblemVerificationDraft draft) {
        return adminProblemService.saveVerificationDraft(draft);
    }

    @DeleteMapping("/verification/draft")
    @RequiresAuthentication
    public CommonResult<Void> deleteVerificationDraft(@RequestParam("pid") Long pid) {
        return adminProblemService.deleteVerificationDraft(pid);
    }

    @PostMapping("/verification/draft/clear")
    @RequiresAuthentication
    public CommonResult<Void> clearVerificationDraft(@RequestParam("pid") Long pid) {
        return adminProblemService.clearVerificationDraft(pid);
    }

}
