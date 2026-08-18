package top.hcode.hoj.controller.classroom;

import org.apache.shiro.authz.annotation.Logical;
import org.apache.shiro.authz.annotation.RequiresAuthentication;
import org.apache.shiro.authz.annotation.RequiresRoles;
import org.springframework.web.bind.annotation.*;
import top.hcode.hoj.common.result.CommonResult;
import top.hcode.hoj.manager.classroom.ClassroomRoleManager;
import top.hcode.hoj.pojo.entity.classroom.ClassroomUserRole;

import javax.annotation.Resource;
import java.util.Collections;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/classroom")
public class ClassroomRoleController extends ClassroomControllerSupport {
    @Resource private ClassroomRoleManager manager;

    @GetMapping("/user/roles")
    @RequiresAuthentication
    public CommonResult<List<ClassroomUserRole>> currentRoles() { return run(manager::currentRoles); }

    @PostMapping("/role/apply")
    @RequiresAuthentication
    public CommonResult<Map<String, Object>> apply(@RequestBody Map<String, Object> request) {
        return run(() -> manager.apply(text(request.get("role")), text(request.get("reason"))));
    }

    @GetMapping("/role/my_applications")
    @RequiresAuthentication
    public CommonResult<Map<String, Object>> myApplications(@RequestParam(defaultValue = "0") String status) {
        return run(() -> manager.myApplications(status));
    }

    @DeleteMapping("/role/application/{applicationId}")
    @RequiresAuthentication
    public CommonResult<Map<String, Object>> cancel(@PathVariable Long applicationId) {
        return run(() -> { manager.cancel(applicationId); return Collections.singletonMap("message", "申请已取消"); });
    }

    @PostMapping("/admin/role/grant")
    @RequiresAuthentication
    @RequiresRoles(value = {"root", "admin"}, logical = Logical.OR)
    public CommonResult<ClassroomUserRole> grant(@RequestBody Map<String, Object> request) {
        return run(() -> manager.grant(text(request.get("uid")), text(request.get("role"))));
    }

    @DeleteMapping("/admin/role/revoke")
    @RequiresAuthentication
    @RequiresRoles(value = {"root", "admin"}, logical = Logical.OR)
    public CommonResult<Void> revoke(@RequestBody Map<String, Object> request) {
        return runVoid(() -> manager.revoke(text(request.get("uid")), text(request.get("role"))));
    }

    @GetMapping("/admin/roles/{userId}")
    @RequiresAuthentication
    @RequiresRoles(value = {"root", "admin"}, logical = Logical.OR)
    public CommonResult<List<ClassroomUserRole>> roles(@PathVariable String userId) { return run(() -> manager.roles(userId)); }

    @PutMapping("/admin/user/{uid}/roles")
    @RequiresAuthentication
    @RequiresRoles(value = {"root", "admin"}, logical = Logical.OR)
    public CommonResult<Map<String, Object>> replace(@PathVariable String uid, @RequestBody Map<String, Object> request) {
        return run(() -> manager.replace(uid, strings(request.get("roles"))));
    }

    @GetMapping("/admin/users/search")
    @RequiresAuthentication
    @RequiresRoles(value = {"root", "admin"}, logical = Logical.OR)
    public CommonResult<Map<String, Object>> users(@RequestParam String keyword,
                                                   @RequestParam(defaultValue = "1") int currentPage,
                                                   @RequestParam(defaultValue = "20") int limit) {
        return run(() -> manager.searchUsers(keyword, currentPage, limit));
    }

    @GetMapping("/admin/role/applications")
    @RequiresAuthentication
    @RequiresRoles(value = {"root", "admin"}, logical = Logical.OR)
    public CommonResult<Map<String, Object>> applications(@RequestParam(defaultValue = "all") String role,
                                                          @RequestParam(defaultValue = "all") String status,
                                                          @RequestParam(defaultValue = "1") int currentPage,
                                                          @RequestParam(defaultValue = "20") int limit) {
        return run(() -> manager.applications(role, status, currentPage, limit));
    }

    @PostMapping("/admin/role/review")
    @RequiresAuthentication
    @RequiresRoles(value = {"root", "admin"}, logical = Logical.OR)
    public CommonResult<Map<String, Object>> review(@RequestBody Map<String, Object> request) {
        return run(() -> manager.review(longs(request.get("applicationIds")), text(request.get("action")),
                text(request.get("reviewNote"))));
    }

    private String text(Object value) { return value == null ? "" : String.valueOf(value); }
    private List<String> strings(Object value) { return value instanceof List ? (List<String>) value : Collections.emptyList(); }
    private List<Long> longs(Object value) {
        if (!(value instanceof List)) return Collections.emptyList();
        java.util.ArrayList<Long> ids = new java.util.ArrayList<>();
        for (Object item : (List<?>) value) if (item instanceof Number) ids.add(((Number) item).longValue());
        return ids;
    }
}
