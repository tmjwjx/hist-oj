package top.hcode.hoj.service.plagiarism;

import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.pojo.dto.PlagiarismConfigDTO;
import top.hcode.hoj.pojo.entity.contest.ContestProblem;
import top.hcode.hoj.pojo.entity.judge.Judge;
import top.hcode.hoj.pojo.entity.plagiarism.PlagiarismCheck;
import top.hcode.hoj.pojo.entity.plagiarism.PlagiarismCheckConfig;
import top.hcode.hoj.pojo.vo.PlagiarismResultPageVO;

import java.util.List;

public interface PlagiarismService {
    CommonResult<List<PlagiarismCheckConfig>> configs(Long cid);
    CommonResult<List<ContestProblem>> contestProblems(Long cid);
    CommonResult<Void> saveConfig(Long cid, PlagiarismConfigDTO dto);
    CommonResult<PlagiarismCheck> start(Long cid);
    CommonResult<PlagiarismCheck> progress(Long cid);
    CommonResult<PlagiarismCheck> latest(Long cid);
    CommonResult<PlagiarismResultPageVO> results(Long checkId, String displayId);
    CommonResult<Judge> submission(Long submitId);
    byte[] export(Long checkId) throws StatusNotFoundException, StatusForbiddenException;
}
