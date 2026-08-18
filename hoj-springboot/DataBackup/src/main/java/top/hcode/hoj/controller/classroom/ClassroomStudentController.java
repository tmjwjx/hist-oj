package top.hcode.hoj.controller.classroom;

import org.apache.shiro.authz.annotation.RequiresAuthentication;
import org.springframework.web.bind.annotation.*;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.manager.classroom.ClassroomStudentManager;
import top.hcode.hoj.pojo.vo.classroom.ClassroomStudentVO;
import top.hcode.hoj.pojo.vo.classroom.ClassroomUserVO;
import top.hcode.hoj.pojo.vo.classroom.ClassroomVO;

import javax.annotation.Resource;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/classroom")
@RequiresAuthentication
public class ClassroomStudentController extends ClassroomControllerSupport {
    @Resource private ClassroomStudentManager manager;

    @PostMapping("/join")
    public CommonResult<ClassroomVO> join(@RequestBody Map<String, Object> request) {
        return run(() -> manager.join(request));
    }

    @GetMapping("/{classroomId}/students")
    public CommonResult<List<ClassroomStudentVO>> list(@PathVariable Long classroomId) {
        return run(() -> manager.list(classroomId));
    }

    @PostMapping("/{classroomId}/students")
    public CommonResult<Void> add(@PathVariable Long classroomId, @RequestBody Map<String, Object> request) {
        return runVoid(() -> manager.add(classroomId, request));
    }

    @GetMapping("/{classroomId}/students/search")
    public CommonResult<List<ClassroomUserVO>> search(@PathVariable Long classroomId, @RequestParam String keyword) {
        return run(() -> manager.search(classroomId, keyword));
    }

    @DeleteMapping("/student")
    public CommonResult<Void> remove(@RequestBody Map<String, Object> request) {
        return runVoid(() -> manager.remove(number(request.get("classroomId")), text(request.get("uid"))));
    }

    @PutMapping("/student")
    public CommonResult<Void> update(@RequestBody Map<String, Object> request) {
        return runVoid(() -> manager.update(request));
    }

    @GetMapping("/{classroomId}/my-info")
    public CommonResult<ClassroomStudentVO> myInfo(@PathVariable Long classroomId) {
        return run(() -> manager.myInfo(classroomId));
    }

    @PutMapping("/{classroomId}/my-info")
    public CommonResult<Void> updateMyInfo(@PathVariable Long classroomId, @RequestBody Map<String, Object> request) {
        return runVoid(() -> manager.updateMyInfo(classroomId, request));
    }

    private Long number(Object value) { if (value instanceof Number) return ((Number) value).longValue(); try { return value == null ? null : Long.valueOf(String.valueOf(value)); } catch (Exception e) { return null; } }
    private String text(Object value) { return value == null ? "" : String.valueOf(value); }
}
