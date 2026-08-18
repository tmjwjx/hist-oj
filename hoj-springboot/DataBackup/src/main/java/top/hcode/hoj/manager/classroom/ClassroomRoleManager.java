package top.hcode.hoj.manager.classroom;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import com.baomidou.mybatisplus.core.conditions.update.UpdateWrapper;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.dao.user.UserInfoEntityService;
import top.hcode.hoj.mapper.classroom.ClassroomRoleRequestMapper;
import top.hcode.hoj.mapper.classroom.ClassroomUserRoleMapper;
import top.hcode.hoj.pojo.entity.classroom.ClassroomRoleRequest;
import top.hcode.hoj.pojo.entity.classroom.ClassroomUserRole;
import top.hcode.hoj.pojo.entity.user.UserInfo;
import top.hcode.hoj.pojo.vo.classroom.ClassroomUserVO;

import javax.annotation.Resource;
import java.text.SimpleDateFormat;
import java.util.*;
import java.util.stream.Collectors;

@Component
public class ClassroomRoleManager {
    private static final Set<String> ROLES = new HashSet<>(Arrays.asList("teacher", "student"));
    @Resource(name = "classroomUserRoleMapper")
    private ClassroomUserRoleMapper roleMapper;
    @Resource private ClassroomRoleRequestMapper requestMapper;
    @Resource private UserInfoEntityService userInfoService;
    @Resource private ClassroomAccessManager accessManager;
    @Resource private ClassroomViewBuilder viewBuilder;

    public List<ClassroomUserRole> currentRoles() {
        return roles(accessManager.uid());
    }

    public List<ClassroomUserRole> roles(String uid) {
        return roleMapper.selectList(new QueryWrapper<ClassroomUserRole>().eq("uid", uid));
    }

    public ClassroomUserRole grant(String uid, String role) throws StatusFailException {
        validateRole(role);
        if (roleMapper.selectCount(new QueryWrapper<ClassroomUserRole>().eq("uid", uid).eq("role", role)) > 0)
            throw new StatusFailException("该用户已拥有此角色");
        ClassroomUserRole row = new ClassroomUserRole().setUid(uid).setRole(role)
                .setCreateTime(new Date()).setUpdateTime(new Date());
        roleMapper.insert(row);
        return row;
    }

    public void revoke(String uid, String role) throws StatusNotFoundException {
        int changed = roleMapper.delete(new QueryWrapper<ClassroomUserRole>().eq("uid", uid).eq("role", role));
        if (changed == 0) throw new StatusNotFoundException("角色不存在");
    }

    @Transactional(rollbackFor = Exception.class)
    public Map<String, Object> replace(String uid, List<String> roles) throws StatusFailException {
        List<String> normalized = roles == null ? Collections.emptyList() : roles.stream().distinct().collect(Collectors.toList());
        for (String role : normalized) validateRole(role);
        roleMapper.delete(new QueryWrapper<ClassroomUserRole>().eq("uid", uid));
        for (String role : normalized) grant(uid, role);
        if (!normalized.isEmpty()) requestMapper.update(null, new UpdateWrapper<ClassroomRoleRequest>()
                .eq("uid", uid).in("role", normalized).eq("status", 0).set("status", 1)
                .set("reviewer_uid", "system").set("review_time", new Date())
                .set("review_note", "管理员手动添加角色").set("update_time", new Date()));
        Map<String, Object> result = new LinkedHashMap<>();
        result.put("uid", uid); result.put("roles", normalized);
        return result;
    }

    public Map<String, Object> apply(String role, String reason) throws StatusFailException {
        validateRole(role);
        String uid = accessManager.uid();
        if (roleMapper.selectCount(new QueryWrapper<ClassroomUserRole>().eq("uid", uid).eq("role", role)) > 0)
            throw new StatusFailException("您已经拥有该角色，无需申请");
        if (requestMapper.selectCount(new QueryWrapper<ClassroomRoleRequest>()
                .eq("uid", uid).eq("role", role).eq("status", 0)) > 0)
            throw new StatusFailException("您已有待审批的" + roleName(role) + "申请，请耐心等待");
        ClassroomRoleRequest row = new ClassroomRoleRequest().setUid(uid).setRole(role).setReason(reason)
                .setStatus(0).setCreateTime(new Date()).setUpdateTime(new Date());
        requestMapper.insert(row);
        Map<String, Object> result = new LinkedHashMap<>();
        result.put("applicationId", row.getId()); result.put("message", "申请已提交，请等待管理员审批");
        return result;
    }

    public Map<String, Object> myApplications(String status) {
        QueryWrapper<ClassroomRoleRequest> query = new QueryWrapper<ClassroomRoleRequest>()
                .eq("uid", accessManager.uid()).orderByDesc("create_time");
        if (status == null || status.isEmpty()) query.eq("status", 0);
        else if (!"all".equals(status)) query.eq("status", Integer.parseInt(status));
        SimpleDateFormat format = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss");
        List<Map<String, Object>> rows = requestMapper.selectList(query).stream().map(row -> {
            Map<String, Object> item = new LinkedHashMap<>();
            item.put("id", row.getId()); item.put("role", row.getRole()); item.put("reason", row.getReason());
            item.put("status", row.getStatus()); item.put("createdAt", format.format(row.getCreateTime()));
            return item;
        }).collect(Collectors.toList());
        Map<String, Object> result = new LinkedHashMap<>();
        result.put("applications", rows); result.put("count", rows.size());
        return result;
    }

    public void cancel(Long applicationId) throws Exception {
        ClassroomRoleRequest row = requestMapper.selectOne(new QueryWrapper<ClassroomRoleRequest>()
                .eq("id", applicationId).eq("uid", accessManager.uid()));
        if (row == null) throw new StatusNotFoundException("申请不存在");
        if (!Objects.equals(row.getStatus(), 0))
            throw new StatusFailException("只能取消待审批的申请，当前状态：" + statusName(row.getStatus()));
        requestMapper.deleteById(applicationId);
    }

    public Map<String, Object> applications(String role, String status, int page, int limit) {
        QueryWrapper<ClassroomRoleRequest> query = new QueryWrapper<>();
        if (role != null && !role.isEmpty() && !"all".equals(role)) query.eq("role", role);
        if (status != null && !status.isEmpty() && !"all".equals(status)) query.eq("status", Integer.parseInt(status));
        Page<ClassroomRoleRequest> result = new Page<>(Math.max(1, page), Math.max(1, limit));
        List<ClassroomRoleRequest> rows = requestMapper.selectPage(result, query.orderByDesc("create_time")).getRecords();
        Set<String> ids = new HashSet<>();
        rows.forEach(row -> { ids.add(row.getUid()); if (row.getReviewerUid() != null) ids.add(row.getReviewerUid()); });
        Map<String, UserInfo> users = ids.isEmpty() ? Collections.emptyMap()
                : userInfoService.listByIds(ids).stream().collect(Collectors.toMap(UserInfo::getUuid, u -> u));
        List<Map<String, Object>> records = rows.stream().map(row -> application(row, users)).collect(Collectors.toList());
        Map<String, Object> data = new LinkedHashMap<>(); data.put("records", records); data.put("total", result.getTotal()); return data;
    }

    @Transactional(rollbackFor = Exception.class)
    public Map<String, Object> review(List<Long> ids, String action, String note) throws StatusFailException {
        if (ids == null || ids.isEmpty() || !("approve".equals(action) || "reject".equals(action)))
            throw new StatusFailException("审批参数错误");
        int success = 0, failed = 0;
        for (ClassroomRoleRequest row : requestMapper.selectBatchIds(ids)) {
            if (!Objects.equals(row.getStatus(), 0)) { failed++; continue; }
            if ("approve".equals(action) && roleMapper.selectCount(new QueryWrapper<ClassroomUserRole>()
                    .eq("uid", row.getUid()).eq("role", row.getRole())) == 0) grant(row.getUid(), row.getRole());
            row.setStatus("approve".equals(action) ? 1 : 2).setReviewerUid(accessManager.uid())
                    .setReviewTime(new Date()).setReviewNote(note).setUpdateTime(new Date());
            requestMapper.updateById(row); success++;
        }
        if (success == 0) throw new StatusFailException("审批失败，没有申请被处理");
        Map<String, Object> result = new LinkedHashMap<>(); result.put("successCount", success);
        result.put("failedCount", failed); result.put("total", ids.size()); result.put("message", "成功审批 " + success + " 个申请");
        return result;
    }

    public Map<String, Object> searchUsers(String keyword, int page, int limit) throws StatusFailException {
        if (keyword == null || keyword.trim().isEmpty()) throw new StatusFailException("关键词不能为空");
        Page<UserInfo> result = new Page<>(Math.max(1, page), Math.min(100, Math.max(1, limit)));
        userInfoService.page(result, new QueryWrapper<UserInfo>()
                .like("username", keyword).or().like("realname", keyword).or().like("nickname", keyword)
                .orderByAsc("uuid"));
        List<ClassroomUserVO> records = result.getRecords().stream().map(viewBuilder::user).collect(Collectors.toList());
        Map<String, Object> data = new LinkedHashMap<>(); data.put("records", records); data.put("total", result.getTotal()); return data;
    }

    private Map<String, Object> application(ClassroomRoleRequest row, Map<String, UserInfo> users) {
        UserInfo applicant = users.get(row.getUid()), reviewer = users.get(row.getReviewerUid());
        Map<String, Object> item = new LinkedHashMap<>(); item.put("id", row.getId()); item.put("uid", row.getUid());
        item.put("username", applicant == null ? "" : applicant.getUsername()); item.put("realname", applicant == null ? "" : applicant.getRealname());
        item.put("email", applicant == null ? "" : applicant.getEmail()); item.put("role", row.getRole()); item.put("roleName", roleName(row.getRole()));
        item.put("reason", row.getReason()); item.put("status", row.getStatus()); item.put("statusName", statusName(row.getStatus()));
        item.put("reviewerUid", row.getReviewerUid()); item.put("reviewerName", reviewer == null ? "" : reviewer.getUsername());
        item.put("reviewTime", row.getReviewTime()); item.put("reviewNote", row.getReviewNote()); item.put("createdAt", row.getCreateTime());
        return item;
    }

    private void validateRole(String role) throws StatusFailException { if (!ROLES.contains(role)) throw new StatusFailException("角色类型必须是 teacher 或 student"); }
    private String roleName(String role) { return "teacher".equals(role) ? "教师" : "student".equals(role) ? "学生" : role; }
    private String statusName(Integer status) { return Objects.equals(status, 0) ? "待审批" : Objects.equals(status, 1) ? "已批准" : Objects.equals(status, 2) ? "已拒绝" : "未知"; }
}
