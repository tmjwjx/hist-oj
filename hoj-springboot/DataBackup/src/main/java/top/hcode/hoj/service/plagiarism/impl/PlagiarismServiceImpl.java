package top.hcode.hoj.service.plagiarism.impl;

import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.common.result.ResultStatus;
import top.hcode.hoj.manager.plagiarism.PlagiarismExecutionManager;
import top.hcode.hoj.manager.plagiarism.PlagiarismManager;
import top.hcode.hoj.manager.plagiarism.PlagiarismQueryManager;
import top.hcode.hoj.pojo.dto.PlagiarismConfigDTO;
import top.hcode.hoj.pojo.entity.contest.ContestProblem;
import top.hcode.hoj.pojo.entity.judge.Judge;
import top.hcode.hoj.pojo.entity.plagiarism.PlagiarismCheck;
import top.hcode.hoj.pojo.entity.plagiarism.PlagiarismCheckConfig;
import top.hcode.hoj.pojo.vo.PlagiarismResultPageVO;
import top.hcode.hoj.service.plagiarism.PlagiarismService;

import javax.annotation.Resource;
import java.util.List;

@Slf4j
@Service
public class PlagiarismServiceImpl implements PlagiarismService {
    @Resource private PlagiarismManager manager;
    @Resource private PlagiarismExecutionManager execution;
    @Resource private PlagiarismQueryManager query;

    @Override public CommonResult<List<PlagiarismCheckConfig>> configs(Long cid) {
        return run(() -> manager.configs(cid));
    }
    @Override public CommonResult<List<ContestProblem>> contestProblems(Long cid) {
        return run(() -> manager.contestProblems(cid));
    }
    @Override public CommonResult<Void> saveConfig(Long cid, PlagiarismConfigDTO dto) {
        return runVoid(() -> manager.saveConfig(cid, dto));
    }
    @Override public CommonResult<PlagiarismCheck> start(Long cid) {
        return run(() -> {
            PlagiarismCheck check = manager.createCheck(cid);
            execution.start(check.getId());
            return check;
        });
    }
    @Override public CommonResult<PlagiarismCheck> progress(Long cid) {
        return run(() -> manager.latest(cid));
    }
    @Override public CommonResult<PlagiarismCheck> latest(Long cid) {
        return run(() -> manager.latest(cid));
    }
    @Override public CommonResult<PlagiarismResultPageVO> results(Long checkId, String displayId) {
        return run(() -> query.results(checkId, displayId));
    }
    @Override public CommonResult<Judge> submission(Long submitId) {
        return run(() -> query.submission(submitId));
    }

    @Override
    public byte[] export(Long checkId) throws StatusNotFoundException, StatusForbiddenException {
        return query.export(checkId);
    }

    private <T> CommonResult<T> run(CheckedSupplier<T> action) {
        try {
            return CommonResult.successResponse(action.get());
        } catch (StatusNotFoundException e) {
            return CommonResult.errorResponse(e.getMessage(), ResultStatus.NOT_FOUND);
        } catch (StatusForbiddenException e) {
            return CommonResult.errorResponse(e.getMessage(), ResultStatus.FORBIDDEN);
        } catch (StatusFailException e) {
            return CommonResult.errorResponse(e.getMessage(), ResultStatus.FAIL);
        } catch (Exception e) {
            log.error("plagiarism request failed", e);
            return CommonResult.errorResponse("查重服务暂时不可用", ResultStatus.SYSTEM_ERROR);
        }
    }

    private CommonResult<Void> runVoid(CheckedRunnable action) {
        return run(() -> { action.run(); return null; });
    }

    @FunctionalInterface
    private interface CheckedSupplier<T> { T get() throws Exception; }

    @FunctionalInterface
    private interface CheckedRunnable { void run() throws Exception; }
}
