package top.hcode.hoj.manager.classroom;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import com.baomidou.mybatisplus.core.conditions.update.UpdateWrapper;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.dao.user.UserInfoEntityService;
import top.hcode.hoj.mapper.classroom.ClassroomMapper;
import top.hcode.hoj.mapper.classroom.ClassroomStudentMapper;
import top.hcode.hoj.pojo.entity.classroom.Classroom;
import top.hcode.hoj.pojo.entity.classroom.ClassroomStudent;
import top.hcode.hoj.pojo.entity.user.UserInfo;
import top.hcode.hoj.pojo.vo.classroom.ClassroomStudentVO;
import top.hcode.hoj.pojo.vo.classroom.ClassroomUserVO;
import top.hcode.hoj.pojo.vo.classroom.ClassroomVO;

import javax.annotation.Resource;
import java.util.*;
import java.util.stream.Collectors;

@Component
public class ClassroomStudentManager {
    @Resource private ClassroomMapper classroomMapper;
    @Resource private ClassroomStudentMapper studentMapper;
    @Resource private UserInfoEntityService userInfoService;
    @Resource private ClassroomAccessManager accessManager;
    @Resource private ClassroomViewBuilder viewBuilder;

    @Transactional(rollbackFor = Exception.class)
    public ClassroomVO join(Map<String, Object> request) throws Exception {
        String code = string(request, "classCode");
        String realName = string(request, "realName");
        if (blank(code) || blank(realName)) throw new StatusFailException("班级码和姓名不能为空");
        Classroom classroom = classroomMapper.selectOne(new QueryWrapper<Classroom>()
                .eq("class_code", code).eq("status", 1));
        if (classroom == null) throw new StatusNotFoundException("班级码不存在或班级已删除");
        saveStudent(classroom.getId(), accessManager.uid(), request);
        return viewBuilder.build(classroom);
    }

    public List<ClassroomStudentVO> list(Long classroomId) throws Exception {
        accessManager.requireManager(classroomId);
        List<ClassroomStudent> rows = studentMapper.selectList(new QueryWrapper<ClassroomStudent>()
                .eq("classroom_id", classroomId).eq("status", 1).orderByDesc("create_time"));
        return build(rows);
    }

    @Transactional(rollbackFor = Exception.class)
    public void add(Long classroomId, Map<String, Object> request) throws Exception {
        accessManager.requireManager(classroomId);
        String uid = string(request, "uid");
        if (blank(uid) || blank(string(request, "realName"))) throw new StatusFailException("用户和姓名不能为空");
        if (userInfoService.getById(uid) == null) throw new StatusNotFoundException("用户不存在");
        saveStudent(classroomId, uid, request);
    }

    public void remove(Long classroomId, String uid) throws Exception {
        accessManager.requireManager(classroomId);
        int changed = studentMapper.update(null, new UpdateWrapper<ClassroomStudent>()
                .eq("classroom_id", classroomId).eq("uid", uid).eq("status", 1)
                .set("status", 0).set("update_time", new Date()));
        if (changed == 0) throw new StatusNotFoundException("学生不存在");
    }

    public void update(Map<String, Object> request) throws Exception {
        Long classroomId = longValue(request.get("classroomId"));
        String uid = string(request, "uid");
        Classroom classroom = accessManager.requireClassroom(classroomId);
        if (!accessManager.uid().equals(uid) && !accessManager.canManage(classroom, accessManager.uid()))
            throw new StatusForbiddenException("无权修改该学生信息");
        UpdateWrapper<ClassroomStudent> update = new UpdateWrapper<ClassroomStudent>()
                .eq("classroom_id", classroomId).eq("uid", uid).eq("status", 1).set("update_time", new Date());
        setIfPresent(update, request, "realName", "real_name");
        setIfPresent(update, request, "gender", "gender");
        setIfPresent(update, request, "studentClass", "student_class");
        setIfPresent(update, request, "studentNo", "student_no");
        if (studentMapper.update(null, update) == 0) throw new StatusNotFoundException("学生不存在");
    }

    public ClassroomStudentVO myInfo(Long classroomId) throws Exception {
        ClassroomStudent row = studentMapper.selectOne(new QueryWrapper<ClassroomStudent>()
                .eq("classroom_id", classroomId).eq("uid", accessManager.uid()).eq("status", 1));
        if (row == null) throw new StatusNotFoundException("未找到班级信息或你不在该班级中");
        return build(Collections.singletonList(row)).get(0);
    }

    public void updateMyInfo(Long classroomId, Map<String, Object> request) throws Exception {
        request.put("classroomId", classroomId);
        request.put("uid", accessManager.uid());
        update(request);
    }

    public List<ClassroomUserVO> search(Long classroomId, String keyword) throws Exception {
        accessManager.requireManager(classroomId);
        if (blank(keyword)) throw new StatusFailException("keyword参数不能为空");
        Set<String> existing = studentMapper.selectList(new QueryWrapper<ClassroomStudent>()
                .eq("classroom_id", classroomId).eq("status", 1)).stream()
                .map(ClassroomStudent::getUid).collect(Collectors.toSet());
        QueryWrapper<UserInfo> query = new QueryWrapper<UserInfo>().eq("status", 0)
                .and(q -> q.like("username", keyword).or().like("realname", keyword).or().like("nickname", keyword))
                .orderByAsc("username").last("LIMIT 20");
        return userInfoService.list(query).stream().filter(user -> !existing.contains(user.getUuid()))
                .map(viewBuilder::user).collect(Collectors.toList());
    }

    private void saveStudent(Long classroomId, String uid, Map<String, Object> request) throws StatusFailException {
        ClassroomStudent row = studentMapper.selectOne(new QueryWrapper<ClassroomStudent>()
                .eq("classroom_id", classroomId).eq("uid", uid));
        if (row != null && Objects.equals(row.getStatus(), 1)) throw new StatusFailException("学生已在班级中");
        if (row == null) row = new ClassroomStudent().setClassroomId(classroomId).setUid(uid).setCreateTime(new Date());
        row.setRealName(string(request, "realName")).setGender(string(request, "gender"))
                .setStudentClass(string(request, "studentClass")).setStudentNo(string(request, "studentNo"))
                .setStatus(1).setUpdateTime(new Date());
        if (row.getId() == null) studentMapper.insert(row); else studentMapper.updateById(row);
    }

    private List<ClassroomStudentVO> build(List<ClassroomStudent> rows) {
        Set<String> ids = rows.stream().map(ClassroomStudent::getUid).collect(Collectors.toSet());
        Map<String, UserInfo> users = ids.isEmpty() ? Collections.emptyMap()
                : userInfoService.listByIds(ids).stream().collect(Collectors.toMap(UserInfo::getUuid, u -> u));
        return rows.stream().map(row -> new ClassroomStudentVO().setId(row.getId())
                .setClassroomId(row.getClassroomId()).setUid(row.getUid()).setRealName(row.getRealName())
                .setGender(row.getGender()).setStudentClass(row.getStudentClass()).setStudentNo(row.getStudentNo())
                .setStatus(row.getStatus()).setCreatedAt(row.getCreateTime()).setUpdatedAt(row.getUpdateTime())
                .setUser(viewBuilder.user(users.get(row.getUid())))).collect(Collectors.toList());
    }

    private void setIfPresent(UpdateWrapper<ClassroomStudent> update, Map<String, Object> request, String key, String column) {
        if (request.containsKey(key)) update.set(column, request.get(key));
    }
    private String string(Map<String, Object> map, String key) { return map.get(key) == null ? "" : String.valueOf(map.get(key)); }
    private Long longValue(Object value) throws StatusFailException { if (!(value instanceof Number)) throw new StatusFailException("classroomId不能为空"); return ((Number) value).longValue(); }
    private boolean blank(String value) { return value == null || value.trim().isEmpty(); }
}
