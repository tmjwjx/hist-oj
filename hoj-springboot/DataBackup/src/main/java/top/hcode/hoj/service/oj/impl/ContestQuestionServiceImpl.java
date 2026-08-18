package top.hcode.hoj.service.oj.impl;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.common.result.ResultStatus;
import top.hcode.hoj.manager.oj.ContestQuestionManager;
import top.hcode.hoj.pojo.dto.ContestQuestionCreateDTO;
import top.hcode.hoj.pojo.vo.ContestQuestionPageVO;
import top.hcode.hoj.pojo.vo.ContestQuestionReplyVO;
import top.hcode.hoj.pojo.vo.ContestQuestionVO;
import top.hcode.hoj.service.oj.ContestQuestionService;

@Service
public class ContestQuestionServiceImpl implements ContestQuestionService {

    @Autowired
    private ContestQuestionManager manager;

    @Override
    public CommonResult<ContestQuestionVO> create(ContestQuestionCreateDTO dto) {
        try {
            return CommonResult.successResponse(manager.create(dto));
        } catch (StatusNotFoundException e) {
            return CommonResult.errorResponse(e.getMessage(), ResultStatus.NOT_FOUND);
        } catch (StatusFailException e) {
            return CommonResult.errorResponse(e.getMessage());
        }
    }

    @Override
    public CommonResult<ContestQuestionPageVO> getPage(Long contestId, Integer page, Integer limit,
                                                       String status, Boolean all) {
        try {
            return CommonResult.successResponse(manager.getPage(contestId, page, limit, status, all));
        } catch (StatusNotFoundException e) {
            return CommonResult.errorResponse(e.getMessage(), ResultStatus.NOT_FOUND);
        } catch (StatusForbiddenException e) {
            return CommonResult.errorResponse(e.getMessage(), ResultStatus.FORBIDDEN);
        }
    }

    @Override
    public CommonResult<ContestQuestionVO> getDetail(Long questionId) {
        try {
            return CommonResult.successResponse(manager.getDetail(questionId));
        } catch (StatusNotFoundException e) {
            return CommonResult.errorResponse(e.getMessage(), ResultStatus.NOT_FOUND);
        } catch (StatusForbiddenException e) {
            return CommonResult.errorResponse(e.getMessage(), ResultStatus.FORBIDDEN);
        }
    }

    @Override
    public CommonResult<ContestQuestionReplyVO> reply(Long questionId, String content) {
        try {
            return CommonResult.successResponse(manager.reply(questionId, content));
        } catch (StatusNotFoundException e) {
            return CommonResult.errorResponse(e.getMessage(), ResultStatus.NOT_FOUND);
        } catch (StatusForbiddenException e) {
            return CommonResult.errorResponse(e.getMessage(), ResultStatus.FORBIDDEN);
        } catch (StatusFailException e) {
            return CommonResult.errorResponse(e.getMessage());
        }
    }

    @Override
    public CommonResult<Void> updateStatus(Long questionId, String status) {
        try {
            manager.updateStatus(questionId, status);
            return CommonResult.successResponse();
        } catch (StatusNotFoundException e) {
            return CommonResult.errorResponse(e.getMessage(), ResultStatus.NOT_FOUND);
        } catch (StatusForbiddenException e) {
            return CommonResult.errorResponse(e.getMessage(), ResultStatus.FORBIDDEN);
        } catch (StatusFailException e) {
            return CommonResult.errorResponse(e.getMessage());
        }
    }

    @Override
    public CommonResult<Void> delete(Long questionId) {
        try {
            manager.delete(questionId);
            return CommonResult.successResponse();
        } catch (StatusNotFoundException e) {
            return CommonResult.errorResponse(e.getMessage(), ResultStatus.NOT_FOUND);
        } catch (StatusForbiddenException e) {
            return CommonResult.errorResponse(e.getMessage(), ResultStatus.FORBIDDEN);
        } catch (StatusFailException e) {
            return CommonResult.errorResponse(e.getMessage());
        }
    }
}
