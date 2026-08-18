package top.hcode.hoj.controller.classroom;

import org.apache.shiro.authz.annotation.RequiresAuthentication;
import org.springframework.web.bind.annotation.*;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.manager.classroom.ClassroomManager;
import top.hcode.hoj.pojo.vo.classroom.ClassroomVO;

import javax.annotation.Resource;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/classroom")
public class ClassroomBaseController extends ClassroomControllerSupport {
    @Resource private ClassroomManager manager;

    @PostMapping("/create")
    @RequiresAuthentication
    public CommonResult<ClassroomVO> create(@RequestBody Map<String, Object> request) {
        return run(() -> manager.create(text(request, "className"), text(request, "classBelong")));
    }

    @DeleteMapping("/{classroomId}")
    @RequiresAuthentication
    public CommonResult<Void> delete(@PathVariable Long classroomId) {
        return runVoid(() -> manager.delete(classroomId));
    }

    @GetMapping("/list")
    @RequiresAuthentication
    public CommonResult<List<ClassroomVO>> list() {
        return run(manager::teacherClassrooms);
    }

    @GetMapping("/{classroomId}")
    public CommonResult<ClassroomVO> detail(@PathVariable Long classroomId) {
        return run(() -> manager.detail(classroomId));
    }

    @GetMapping("/my-classrooms")
    @RequiresAuthentication
    public CommonResult<List<ClassroomVO>> studentClassrooms() {
        return run(manager::studentClassrooms);
    }

    @GetMapping("/teacher-classrooms")
    @RequiresAuthentication
    public CommonResult<List<ClassroomVO>> teacherClassrooms() {
        return run(manager::teacherClassrooms);
    }

    private String text(Map<String, Object> request, String key) {
        return request.get(key) == null ? "" : String.valueOf(request.get(key));
    }
}
