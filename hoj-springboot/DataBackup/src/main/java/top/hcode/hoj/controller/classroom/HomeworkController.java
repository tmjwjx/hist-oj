package top.hcode.hoj.controller.classroom;

import org.apache.shiro.authz.annotation.RequiresAuthentication;
import org.springframework.web.bind.annotation.*;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.manager.classroom.ExamManager;
import top.hcode.hoj.manager.classroom.HomeworkAnalysisManager;
import top.hcode.hoj.manager.classroom.HomeworkManager;
import top.hcode.hoj.manager.classroom.HomeworkRankingManager;

import javax.annotation.Resource;
import java.util.Collections;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/classroom")
public class HomeworkController extends ClassroomControllerSupport {
    @Resource private HomeworkManager manager;
    @Resource private HomeworkAnalysisManager analysisManager;
    @Resource private HomeworkRankingManager rankingManager;
    @Resource private ExamManager examManager;

    @PostMapping("/homework") @RequiresAuthentication
    public CommonResult<Map<String, Object>> create(@RequestBody Map<String, Object> request) { return run(() -> manager.create(request)); }
    @GetMapping("/{classroomId}/homeworks") @RequiresAuthentication
    public CommonResult<List<Map<String, Object>>> list(@PathVariable Long classroomId) { return run(() -> manager.list(classroomId)); }
    @GetMapping("/homework/{homeworkId}") @RequiresAuthentication
    public CommonResult<Map<String, Object>> detail(@PathVariable Long homeworkId) { return run(() -> manager.detail(homeworkId)); }
    @PutMapping("/homework/{homeworkId}") @RequiresAuthentication
    public CommonResult<Map<String, Object>> update(@PathVariable Long homeworkId, @RequestBody Map<String, Object> request) { return run(() -> manager.update(homeworkId, request)); }
    @DeleteMapping("/homework/{homeworkId}") @RequiresAuthentication
    public CommonResult<Void> delete(@PathVariable Long homeworkId) { return runVoid(() -> manager.delete(homeworkId)); }
    @PostMapping("/homework/draft") @RequiresAuthentication
    public CommonResult<Map<String, Object>> draft(@RequestBody Map<String, Object> request) { return run(() -> manager.saveDraft(request)); }
    @PostMapping("/homework/submit") @RequiresAuthentication
    public CommonResult<Map<String, Object>> submit(@RequestBody Map<String, Object> request) { return run(() -> manager.submit(request)); }
    @GetMapping("/homework/{homeworkId}/submissions") @RequiresAuthentication
    public CommonResult<List<Map<String, Object>>> submissions(@PathVariable Long homeworkId) { return run(() -> manager.submissions(homeworkId)); }
    @GetMapping("/homework/{homeworkId}/status") @RequiresAuthentication
    public CommonResult<Map<String, Object>> status(@PathVariable Long homeworkId) { return run(() -> manager.status(homeworkId)); }
    @GetMapping("/homework/{homeworkId}/my-detail") @RequiresAuthentication
    public CommonResult<Map<String, Object>> myDetail(@PathVariable Long homeworkId) { return run(() -> manager.myDetail(homeworkId)); }
    @PostMapping("/homework/grade") @RequiresAuthentication
    public CommonResult<Void> grade(@RequestBody Map<String, Object> request) { return runVoid(() -> manager.grade(request)); }
    @PostMapping("/homework/programming/grade") @RequiresAuthentication
    public CommonResult<Void> gradeProgramming(@RequestBody Map<String, Object> request) { return runVoid(() -> manager.grade(request)); }
    @PostMapping("/homework/recalculate") @RequiresAuthentication
    public CommonResult<Void> recalculate(@RequestBody Map<String, Object> request) { return runVoid(() -> manager.recalculate(request)); }
    @GetMapping("/homework/{homeworkId}/ranking") @RequiresAuthentication
    public CommonResult<Map<String, Object>> ranking(@PathVariable Long homeworkId) { return run(() -> rankingManager.ranking(homeworkId)); }
    @GetMapping("/homework/{homeworkId}/analysis") @RequiresAuthentication
    public CommonResult<Map<String, Object>> analysis(@PathVariable Long homeworkId) { return run(() -> analysisManager.analysis(homeworkId)); }
    @PostMapping("/programming/submission") @RequiresAuthentication
    public CommonResult<Map<String, Object>> programming(@RequestBody Map<String, Object> request) { return run(() -> manager.saveProgramming(request)); }
    @GetMapping("/programming/submissions") @RequiresAuthentication
    public CommonResult<List<Map<String, Object>>> programmingList(@RequestParam Map<String, String> params) { return run(() -> manager.programmingSubmissions(params)); }

    @PostMapping("/homework/{homeworkId}/start-exam") @RequiresAuthentication
    public CommonResult<Map<String, Object>> startExam(@PathVariable Long homeworkId, @RequestBody(required = false) Map<String, Object> request) { return run(() -> examManager.start(homeworkId, request == null ? Collections.emptyMap() : request)); }
    @GetMapping("/homework/{homeworkId}/exam-status") @RequiresAuthentication
    public CommonResult<Map<String, Object>> examStatus(@PathVariable Long homeworkId) { return run(() -> examManager.status(homeworkId)); }
    @PostMapping("/homework/violation") @RequiresAuthentication
    public CommonResult<Void> violation(@RequestBody Map<String, Object> request) { return runVoid(() -> examManager.violation(request)); }
    @GetMapping("/homework/{homeworkId}/exam-monitoring") @RequiresAuthentication
    public CommonResult<Map<String, Object>> monitoring(@PathVariable Long homeworkId) { return run(() -> examManager.monitoring(homeworkId)); }
    @PostMapping("/homework/force-submit") @RequiresAuthentication
    public CommonResult<Void> force(@RequestBody Map<String, Object> request) { return runVoid(() -> examManager.force(number(request.get("homeworkId")), text(request.get("uid")), text(request.get("reason")))); }
    @PostMapping("/homework/{homeworkId}/force-submit-all") @RequiresAuthentication
    public CommonResult<Map<String, Object>> forceAll(@PathVariable Long homeworkId) { return run(() -> examManager.forceAll(homeworkId)); }

    private Long number(Object value) { try { return value instanceof Number ? ((Number) value).longValue() : Long.parseLong(String.valueOf(value)); } catch (Exception e) { return null; } }
    private String text(Object value) { return value == null ? "" : String.valueOf(value); }
}
