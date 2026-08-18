package top.hcode.hoj.controller.classroom;

import org.apache.shiro.authz.annotation.Logical;
import org.apache.shiro.authz.annotation.RequiresAuthentication;
import org.apache.shiro.authz.annotation.RequiresRoles;
import org.springframework.web.bind.annotation.*;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.manager.classroom.QuestionBankManager;
import top.hcode.hoj.pojo.vo.classroom.QuestionBankVO;

import javax.annotation.Resource;
import java.util.Map;

@RestController
@RequestMapping("/api/classroom")
public class QuestionBankController extends ClassroomControllerSupport {
    @Resource private QuestionBankManager manager;

    @PostMapping("/question")
    @RequiresAuthentication
    public CommonResult<QuestionBankVO> create(@RequestBody Map<String, Object> request) {
        return run(() -> manager.create(request));
    }

    @GetMapping("/questions")
    @RequiresAuthentication
    public CommonResult<Map<String, Object>> list(@RequestParam Map<String, String> params) {
        return run(() -> manager.list(params, false));
    }

    @GetMapping("/question/{questionId}")
    public CommonResult<QuestionBankVO> detail(@PathVariable Long questionId) {
        return run(() -> manager.detail(questionId));
    }

    @PutMapping("/question/{questionId}")
    @RequiresAuthentication
    public CommonResult<Void> update(@PathVariable Long questionId, @RequestBody Map<String, Object> request) {
        return runVoid(() -> manager.update(questionId, request, false));
    }

    @DeleteMapping("/question/{questionId}")
    @RequiresAuthentication
    public CommonResult<Void> delete(@PathVariable Long questionId) {
        return runVoid(() -> manager.delete(questionId, false));
    }

    @GetMapping({"/admin/question-bank", "/admin/questions"})
    @RequiresAuthentication
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<Map<String, Object>> adminList(@RequestParam Map<String, String> params) {
        return run(() -> manager.list(params, true));
    }

    @PostMapping("/admin/question")
    @RequiresAuthentication
    @RequiresRoles(value = {"root", "admin"}, logical = Logical.OR)
    public CommonResult<QuestionBankVO> adminCreate(@RequestBody Map<String, Object> request) {
        return run(() -> manager.create(request));
    }

    @PutMapping("/admin/question/{questionId}")
    @RequiresAuthentication
    @RequiresRoles(value = {"root", "admin"}, logical = Logical.OR)
    public CommonResult<Void> adminUpdate(@PathVariable Long questionId, @RequestBody Map<String, Object> request) {
        return runVoid(() -> manager.update(questionId, request, true));
    }

    @DeleteMapping("/admin/question/{questionId}")
    @RequiresAuthentication
    @RequiresRoles(value = {"root", "admin"}, logical = Logical.OR)
    public CommonResult<Void> adminDelete(@PathVariable Long questionId) {
        return runVoid(() -> manager.delete(questionId, true));
    }
}
