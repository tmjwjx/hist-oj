package top.hcode.hoj.service.oj;

import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.pojo.dto.ContestQuestionCreateDTO;
import top.hcode.hoj.pojo.vo.ContestQuestionPageVO;
import top.hcode.hoj.pojo.vo.ContestQuestionReplyVO;
import top.hcode.hoj.pojo.vo.ContestQuestionVO;

public interface ContestQuestionService {
    CommonResult<ContestQuestionVO> create(ContestQuestionCreateDTO dto);

    CommonResult<ContestQuestionPageVO> getPage(Long contestId, Integer page, Integer limit,
                                                String status, Boolean all);

    CommonResult<ContestQuestionVO> getDetail(Long questionId);

    CommonResult<ContestQuestionReplyVO> reply(Long questionId, String content);

    CommonResult<Void> updateStatus(Long questionId, String status);

    CommonResult<Void> delete(Long questionId);
}
