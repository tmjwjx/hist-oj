package top.hcode.hoj.service.admin.problem;

import com.baomidou.mybatisplus.core.metadata.IPage;

import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.pojo.dto.ProblemDTO;
import top.hcode.hoj.pojo.dto.CompileDTO;
import top.hcode.hoj.pojo.entity.problem.Problem;
import top.hcode.hoj.pojo.entity.problem.ProblemCase;
import top.hcode.hoj.pojo.dto.ProblemVerificationSubmitDTO;
import top.hcode.hoj.pojo.dto.LastAcceptedCodeVO;
import top.hcode.hoj.pojo.vo.ProblemVerificationVO;
import top.hcode.hoj.pojo.entity.problem.ProblemVerificationDraft;
import java.util.List;
import java.util.Map;


public interface AdminProblemService {

    public CommonResult<IPage<Problem>> getProblemList(Integer limit, Integer currentPage, String keyword, Integer auth, String oj);

    public CommonResult<Problem> getProblem(Long pid);

    public CommonResult<Void> deleteProblem(Long pid);

    public CommonResult<Map<String, Object>> addProblem(ProblemDTO problemDto);

    public CommonResult<Void> updateProblem(ProblemDTO problemDto);

    public CommonResult<List<ProblemCase>> getProblemCases(Long pid, Boolean isUpload);

    public CommonResult compileSpj(CompileDTO compileDTO);

    public CommonResult compileInteractive(CompileDTO compileDTO);

    public CommonResult<Void> importRemoteOJProblem(String name,String problemId);

    public CommonResult<Void> changeProblemAuth(Problem problem);

    public CommonResult<ProblemVerificationVO> getVerification(Long pid);

    public CommonResult<ProblemVerificationVO> submitVerification(ProblemVerificationSubmitDTO dto);

    public CommonResult<ProblemVerificationVO> retryVerificationSync(Long pid);

    public CommonResult<LastAcceptedCodeVO> getLastPassedVerificationCode(Long pid);

    public CommonResult<ProblemVerificationDraft> getVerificationDraft(Long pid);
    public CommonResult<ProblemVerificationDraft> saveVerificationDraft(ProblemVerificationDraft draft);
    public CommonResult<Void> deleteVerificationDraft(Long pid);
    public CommonResult<Void> clearVerificationDraft(Long pid);
}
