package top.hcode.hoj.manager.classroom;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import com.baomidou.mybatisplus.core.conditions.update.UpdateWrapper;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.dao.user.UserInfoEntityService;
import top.hcode.hoj.mapper.classroom.ClassroomMapper;
import top.hcode.hoj.mapper.classroom.ClassroomStudentMapper;
import top.hcode.hoj.mapper.classroom.ClassroomTeacherMapper;
import top.hcode.hoj.pojo.entity.classroom.Classroom;
import top.hcode.hoj.pojo.entity.classroom.ClassroomStudent;
import top.hcode.hoj.pojo.entity.classroom.ClassroomTeacher;
import top.hcode.hoj.pojo.entity.user.UserInfo;
import top.hcode.hoj.pojo.vo.classroom.ClassroomTeacherVO;
import top.hcode.hoj.pojo.vo.classroom.ClassroomUserVO;

import javax.annotation.Resource;
import java.util.Date;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

@Component
public class ClassroomTeacherManager {
    @Resource private ClassroomMapper classroomMapper;
    @Resource private ClassroomTeacherMapper teacherMapper;
    @Resource private ClassroomStudentMapper studentMapper;
    @Resource private UserInfoEntityService userInfoService;
    @Resource private ClassroomAccessManager accessManager;
    @Resource private ClassroomViewBuilder viewBuilder;

    @Transactional(rollbackFor = Exception.class)
    public void add(Map<String, Object> request) throws Exception {
        Long classroomId = number(request.get("classroomId"));
        Classroom classroom = accessManager.requireClassroom(classroomId);
        String teacherId = text(request.get("teacherId"));
        if (teacherId.isEmpty() && request.get("username") != null) {
            UserInfo user = userInfoService.getOne(new QueryWrapper<UserInfo>()
                    .eq("username", text(request.get("username"))));
            if (user != null) teacherId = user.getUuid();
        }
        if (teacherId.isEmpty() || userInfoService.getById(teacherId) == null)
            throw new StatusNotFoundException("用户不存在");
        if (classroom.getTeacherId().equals(teacherId)) throw new StatusFailException("该用户已经是班级的主教师");
        if (studentMapper.selectCount(new QueryWrapper<ClassroomStudent>()
                .eq("classroom_id", classroomId).eq("uid", teacherId).eq("status", 1)) > 0)
            throw new StatusFailException("该用户是班级学生，不能同时添加为教师");
        ClassroomTeacher relation = teacherMapper.selectOne(new QueryWrapper<ClassroomTeacher>()
                .eq("classroom_id", classroomId).eq("teacher_id", teacherId));
        if (relation != null && relation.getStatus() == 1) throw new StatusFailException("该教师已经是班级教师");
        if (relation == null) relation = new ClassroomTeacher().setClassroomId(classroomId)
                .setTeacherId(teacherId).setCreateTime(new Date());
        relation.setStatus(1).setUpdateTime(new Date());
        if (relation.getId() == null) teacherMapper.insert(relation); else teacherMapper.updateById(relation);
    }

    @Transactional(rollbackFor = Exception.class)
    public void remove(Map<String, Object> request) throws Exception {
        Long classroomId = number(request.get("classroomId"));
        String teacherId = text(request.get("teacherId"));
        String newTeacherId = text(request.get("newTeacherId"));
        Classroom classroom = accessManager.requireClassroom(classroomId);
        boolean primaryTeacher = classroom.getTeacherId().equals(teacherId);
        if (primaryTeacher) {
            if (newTeacherId.isEmpty()) throw new StatusFailException("删除主教师必须指定新的主教师");
            if (teacherMapper.selectCount(new QueryWrapper<ClassroomTeacher>().eq("classroom_id", classroomId)
                    .eq("teacher_id", newTeacherId).eq("status", 1)) == 0)
                throw new StatusFailException("新主教师不是该班级的教师");
            classroomMapper.updateById(classroom.setTeacherId(newTeacherId).setUpdateTime(new Date()));
        }
        int changed = teacherMapper.update(null, new UpdateWrapper<ClassroomTeacher>()
                .eq("classroom_id", classroomId).eq("teacher_id", teacherId).eq("status", 1)
                .set("status", 0).set("update_time", new Date()));
        if (changed == 0 && !primaryTeacher)
            throw new StatusNotFoundException("教师关联不存在");
    }

    public List<ClassroomTeacherVO> list(Long classroomId) throws Exception {
        return viewBuilder.build(accessManager.requireClassroom(classroomId)).getTeachers();
    }

    public List<ClassroomUserVO> search(String keyword) throws StatusFailException {
        if (keyword == null || keyword.trim().isEmpty()) throw new StatusFailException("关键词不能为空");
        return userInfoService.list(new QueryWrapper<UserInfo>().eq("status", 0)
                .and(q -> q.like("username", keyword).or().like("realname", keyword).or().like("nickname", keyword))
                .orderByAsc("username").last("LIMIT 20")).stream().map(viewBuilder::user).collect(Collectors.toList());
    }

    private Long number(Object value) throws StatusFailException {
        if (!(value instanceof Number)) throw new StatusFailException("classroomId不能为空");
        return ((Number) value).longValue();
    }
    private String text(Object value) { return value == null ? "" : String.valueOf(value).trim(); }
}
