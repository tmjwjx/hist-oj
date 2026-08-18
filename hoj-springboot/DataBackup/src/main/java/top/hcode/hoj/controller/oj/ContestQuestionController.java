package top.hcode.hoj.controller.oj;

import org.apache.shiro.authz.annotation.RequiresAuthentication;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.pojo.dto.ContestQuestionCreateDTO;
import top.hcode.hoj.pojo.dto.ContestQuestionReplyDTO;
import top.hcode.hoj.pojo.dto.ContestQuestionStatusDTO;
import top.hcode.hoj.pojo.vo.ContestQuestionPageVO;
import top.hcode.hoj.pojo.vo.ContestQuestionReplyVO;
import top.hcode.hoj.pojo.vo.ContestQuestionVO;
import top.hcode.hoj.service.oj.ContestQuestionService;

import javax.validation.Valid;

@RestController
@RequestMapping("/api/contest-question")
@RequiresAuthentication
public class ContestQuestionController {

    @Autowired
    private ContestQuestionService service;

    @PostMapping
    public CommonResult<ContestQuestionVO> create(@Valid @RequestBody ContestQuestionCreateDTO dto) {
        return service.create(dto);
    }

    @GetMapping("/{contestId}/questions")
    public CommonResult<ContestQuestionPageVO> getPage(
            @PathVariable Long contestId,
            @RequestParam(defaultValue = "1") Integer page,
            @RequestParam(defaultValue = "20") Integer limit,
            @RequestParam(required = false) String status,
            @RequestParam(name = "all", defaultValue = "false") Boolean all) {
        return service.getPage(contestId, page, limit, status, all);
    }

    @GetMapping("/question/{questionId}")
    public CommonResult<ContestQuestionVO> getDetail(@PathVariable Long questionId) {
        return service.getDetail(questionId);
    }

    @PostMapping("/question/{questionId}/reply")
    public CommonResult<ContestQuestionReplyVO> reply(
            @PathVariable Long questionId,
            @Valid @RequestBody ContestQuestionReplyDTO dto) {
        return service.reply(questionId, dto.getContent());
    }

    @PutMapping("/question/{questionId}/status")
    public CommonResult<Void> updateStatus(
            @PathVariable Long questionId,
            @Valid @RequestBody ContestQuestionStatusDTO dto) {
        return service.updateStatus(questionId, dto.getStatus());
    }

    @DeleteMapping("/question/{questionId}")
    public CommonResult<Void> delete(@PathVariable Long questionId) {
        return service.delete(questionId);
    }
}
