package top.hcode.hoj.service.admin.problem.impl;

import com.baomidou.mybatisplus.core.metadata.IPage;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.common.result.ResultStatus;
import top.hcode.hoj.manager.admin.problem.AdminProblemManager;
import top.hcode.hoj.manager.admin.problem.ProblemVerificationManager;
import top.hcode.hoj.pojo.dto.ProblemDTO;
import top.hcode.hoj.pojo.dto.CompileDTO;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.pojo.entity.problem.ProblemCase;
import top.hcode.hoj.pojo.dto.ProblemVerificationSubmitDTO;
import top.hcode.hoj.pojo.dto.LastAcceptedCodeVO;
import top.hcode.hoj.pojo.vo.ProblemVerificationVO;
import top.hcode.hoj.pojo.entity.problem.ProblemVerificationDraft;
import top.hcode.hoj.service.admin.problem.AdminProblemService;

import java.util.List;
import java.util.HashMap;
import java.util.Map;

/**
 * @Author: Himit_ZH
 * @Date: 2022/3/9 16:33
 * @Description:
 */

@Service
public class AdminProblemServiceImpl implements AdminProblemService {

    @Autowired
    private AdminProblemManager adminProblemManager;

    @Autowired
    private ProblemVerificationManager problemVerificationManager;

    @Override
    public CommonResult<IPage<Problem>> getProblemList(Integer limit, Integer currentPage, String keyword, Integer auth, String oj) {
        IPage<Problem> problemList = adminProblemManager.getProblemList(limit, currentPage, keyword, auth, oj);
        return CommonResult.successResponse(problemList);
    }

    @Override
    public CommonResult<Problem> getProblem(Long pid) {
        try {
            Problem problem = adminProblemManager.getProblem(pid);
            return CommonResult.successResponse(problem);
        } catch (StatusForbiddenException e) {
            return CommonResult.errorResponse(e.getMessage(), ResultStatus.FORBIDDEN);
        } catch (StatusFailException e) {
            return CommonResult.errorResponse(e.getMessage());
        }
    }

    @Override
    public CommonResult<Void> deleteProblem(Long pid) {
        try {
            adminProblemManager.deleteProblem(pid);
            return CommonResult.successResponse();
        } catch (StatusFailException e) {
            return CommonResult.errorResponse(e.getMessage());
        }
    }

    @Override
    public CommonResult<Map<String, Object>> addProblem(ProblemDTO problemDto) {
        try {
            Long pid = adminProblemManager.addProblem(problemDto);
            Map<String, Object> result = new HashMap<>();
            result.put("pid", pid);
            return CommonResult.successResponse(result);
        } catch (StatusFailException e) {
            return CommonResult.errorResponse(e.getMessage());
        }
    }

    @Override
    public CommonResult<Void> updateProblem(ProblemDTO problemDto) {
        try {
            adminProblemManager.updateProblem(problemDto);
            return CommonResult.successResponse();
        } catch (StatusFailException e) {
            return CommonResult.errorResponse(e.getMessage());
        } catch (StatusForbiddenException e) {
            return CommonResult.errorResponse(e.getMessage(), ResultStatus.FORBIDDEN);
        }
    }

    @Override
    public CommonResult<List<ProblemCase>> getProblemCases(Long pid, Boolean isUpload) {

        List<ProblemCase> problemCaseList = adminProblemManager.getProblemCases(pid, isUpload);
        return CommonResult.successResponse(problemCaseList);
    }

    @Override
    public CommonResult compileSpj(CompileDTO compileDTO) {
        return adminProblemManager.compileSpj(compileDTO);
    }

    @Override
    public CommonResult compileInteractive(CompileDTO compileDTO) {
        return adminProblemManager.compileInteractive(compileDTO);
    }

    @Override
    public CommonResult<Void> importRemoteOJProblem(String name, String problemId) {
        try {
            adminProblemManager.importRemoteOJProblem(name, problemId);
            return CommonResult.successResponse("导入新题目成功");
        } catch (StatusFailException e) {
            return CommonResult.errorResponse(e.getMessage());
        }
    }

    @Override
    public CommonResult<Void> changeProblemAuth(Problem problem) {
        try {
            adminProblemManager.changeProblemAuth(problem);
            return CommonResult.successResponse();
        } catch (StatusFailException e) {
            return CommonResult.errorResponse(e.getMessage());
        } catch (StatusForbiddenException e) {
            return CommonResult.errorResponse(e.getMessage(), ResultStatus.FORBIDDEN);
        }
    }

    @Override
    public CommonResult<ProblemVerificationVO> getVerification(Long pid) {
        try {
            return CommonResult.successResponse(problemVerificationManager.getStatus(pid));
        } catch (StatusForbiddenException e) {
            return CommonResult.errorResponse(e.getMessage(), ResultStatus.FORBIDDEN);
        } catch (StatusFailException e) {
            return CommonResult.errorResponse(e.getMessage());
        }
    }

    @Override
    public CommonResult<ProblemVerificationVO> submitVerification(ProblemVerificationSubmitDTO dto) {
        try {
            return CommonResult.successResponse(problemVerificationManager.submit(dto));
        } catch (StatusForbiddenException e) {
            return CommonResult.errorResponse(e.getMessage(), ResultStatus.FORBIDDEN);
        } catch (StatusFailException e) {
            return CommonResult.errorResponse(e.getMessage());
        }
    }

    @Override
    public CommonResult<ProblemVerificationVO> retryVerificationSync(Long pid) {
        try {
            return CommonResult.successResponse(problemVerificationManager.retrySync(pid));
        } catch (StatusForbiddenException e) {
            return CommonResult.errorResponse(e.getMessage(), ResultStatus.FORBIDDEN);
        } catch (StatusFailException e) {
            return CommonResult.errorResponse(e.getMessage());
        }
    }

    @Override
    public CommonResult<LastAcceptedCodeVO> getLastPassedVerificationCode(Long pid) {
        try {
            return CommonResult.successResponse(problemVerificationManager.getLastPassedCode(pid));
        } catch (StatusForbiddenException e) {
            return CommonResult.errorResponse(e.getMessage(), ResultStatus.FORBIDDEN);
        } catch (StatusFailException e) {
            return CommonResult.errorResponse(e.getMessage());
        }
    }

    @Override
    public CommonResult<ProblemVerificationDraft> getVerificationDraft(Long pid) {
        try { return CommonResult.successResponse(problemVerificationManager.getDraft(pid)); }
        catch (Exception e) { return CommonResult.errorResponse(e.getMessage()); }
    }

    @Override
    public CommonResult<ProblemVerificationDraft> saveVerificationDraft(ProblemVerificationDraft draft) {
        try { return CommonResult.successResponse(problemVerificationManager.saveDraft(draft)); }
        catch (Exception e) { return CommonResult.errorResponse(e.getMessage()); }
    }

    @Override
    public CommonResult<Void> deleteVerificationDraft(Long pid) {
        try { problemVerificationManager.deleteDraft(pid); return CommonResult.successResponse(); }
        catch (Exception e) { return CommonResult.errorResponse(e.getMessage()); }
    }

    @Override
    public CommonResult<Void> clearVerificationDraft(Long pid) {
        try { problemVerificationManager.clearDraft(pid); return CommonResult.successResponse(); }
        catch (Exception e) { return CommonResult.errorResponse(e.getMessage()); }
    }
}
