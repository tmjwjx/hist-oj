package top.hcode.hoj.controller.classroom;

import org.apache.shiro.authz.annotation.Logical;
import org.apache.shiro.authz.annotation.RequiresAuthentication;
import org.apache.shiro.authz.annotation.RequiresRoles;
import org.springframework.web.bind.annotation.*;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.manager.classroom.ClassroomManager;
import top.hcode.hoj.manager.classroom.ClassroomTeacherManager;
import top.hcode.hoj.pojo.vo.classroom.ClassroomTeacherVO;
import top.hcode.hoj.pojo.vo.classroom.ClassroomUserVO;
import top.hcode.hoj.pojo.vo.classroom.ClassroomVO;

import javax.annotation.Resource;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/classroom")
public class ClassroomTeacherController extends ClassroomControllerSupport {
    @Resource private ClassroomTeacherManager teacherManager;
    @Resource private ClassroomManager classroomManager;

    @GetMapping("/{classroomId}/teachers")
    @RequiresAuthentication
    public CommonResult<List<ClassroomTeacherVO>> teachers(@PathVariable Long classroomId) {
        return run(() -> teacherManager.list(classroomId));
    }

    @GetMapping("/admin/classrooms")
    @RequiresAuthentication
    @RequiresRoles(value = {"root", "admin"}, logical = Logical.OR)
    public CommonResult<List<ClassroomVO>> classrooms() {
        return run(classroomManager::all);
    }

    @PostMapping("/admin/teacher/add")
    @RequiresAuthentication
    @RequiresRoles(value = {"root", "admin"}, logical = Logical.OR)
    public CommonResult<Void> add(@RequestBody Map<String, Object> request) {
        return runVoid(() -> teacherManager.add(request));
    }

    @DeleteMapping("/admin/teacher/remove")
    @RequiresAuthentication
    @RequiresRoles(value = {"root", "admin"}, logical = Logical.OR)
    public CommonResult<Void> remove(@RequestBody Map<String, Object> request) {
        return runVoid(() -> teacherManager.remove(request));
    }

    @GetMapping("/admin/teachers/search")
    @RequiresAuthentication
    @RequiresRoles(value = {"root", "admin"}, logical = Logical.OR)
    public CommonResult<List<ClassroomUserVO>> search(@RequestParam String keyword) {
        return run(() -> teacherManager.search(keyword));
    }
}
