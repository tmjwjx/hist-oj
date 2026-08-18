package top.hcode.hoj.controller.classroom;

import org.apache.shiro.authz.annotation.Logical;
import org.apache.shiro.authz.annotation.RequiresAuthentication;
import org.apache.shiro.authz.annotation.RequiresRoles;
import org.springframework.web.bind.annotation.*;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.manager.classroom.ExamPaperManager;
import top.hcode.hoj.pojo.entity.classroom.ExamPaper;

import javax.annotation.Resource;
import java.util.Map;

@RestController
@RequestMapping("/api/classroom")
public class ExamPaperController extends ClassroomControllerSupport {
    @Resource private ExamPaperManager manager;

    @PostMapping("/exam-paper")
    @RequiresAuthentication
    public CommonResult<ExamPaper> create(@RequestBody Map<String, Object> request) {
        return run(() -> manager.create(request));
    }

    @GetMapping("/exam-papers")
    @RequiresAuthentication
    public CommonResult<Map<String, Object>> list(@RequestParam Map<String, String> params) {
        return run(() -> manager.list(params, false));
    }

    @GetMapping("/exam-paper/{paperId}")
    @RequiresAuthentication
    public CommonResult<Map<String, Object>> detail(@PathVariable Long paperId) {
        return run(() -> manager.detail(paperId, false));
    }

    @GetMapping("/admin/exam-paper/{paperId}")
    @RequiresAuthentication
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<Map<String, Object>> adminDetail(@PathVariable Long paperId) {
        return run(() -> manager.adminDetail(paperId));
    }

    @PutMapping("/exam-paper/{paperId}")
    @RequiresAuthentication
    public CommonResult<Void> update(@PathVariable Long paperId, @RequestBody Map<String, Object> request) {
        return runVoid(() -> manager.update(paperId, request, false));
    }

    @DeleteMapping("/exam-paper/{paperId}")
    @RequiresAuthentication
    public CommonResult<Void> delete(@PathVariable Long paperId) {
        return runVoid(() -> manager.delete(paperId, false));
    }

    @PostMapping("/exam-paper/import")
    @RequiresAuthentication
    public CommonResult<Map<String, Object>> importPaper(@RequestBody Map<String, Object> request) {
        return run(() -> manager.importToHomework(number(request.get("paperId")), number(request.get("classroomId"))));
    }

    @GetMapping("/public/exam-papers")
    public CommonResult<Map<String, Object>> publicList(@RequestParam Map<String, String> params) {
        return run(() -> manager.publicList(params));
    }

    @GetMapping("/public/exam-paper/{paperId}")
    public CommonResult<Map<String, Object>> publicDetail(@PathVariable Long paperId) {
        return run(() -> manager.detail(paperId, true));
    }

    @GetMapping("/public/exam-paper/{paperId}/question/{questionId}/answer")
    public CommonResult<Map<String, Object>> publicAnswer(@PathVariable Long paperId, @PathVariable Long questionId) {
        return run(() -> manager.questionAnswer(paperId, questionId));
    }

    @GetMapping("/admin/exam-papers")
    @RequiresAuthentication
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<Map<String, Object>> adminList(@RequestParam Map<String, String> params) {
        return run(() -> manager.list(params, true));
    }

    @PutMapping("/admin/exam-paper/{paperId}")
    @RequiresAuthentication
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<Void> adminUpdate(@PathVariable Long paperId, @RequestBody Map<String, Object> request) {
        return runVoid(() -> manager.update(paperId, request, true));
    }

    @DeleteMapping("/admin/exam-paper/{paperId}")
    @RequiresAuthentication
    @RequiresRoles(value = {"root", "admin", "problem_admin"}, logical = Logical.OR)
    public CommonResult<Void> adminDelete(@PathVariable Long paperId) {
        return runVoid(() -> manager.delete(paperId, true));
    }

    private Long number(Object value) {
        if (value instanceof Number) return ((Number) value).longValue();
        try { return Long.valueOf(String.valueOf(value)); } catch (Exception e) { return null; }
    }
}
