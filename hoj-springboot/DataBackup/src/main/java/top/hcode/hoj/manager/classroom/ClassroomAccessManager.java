package top.hcode.hoj.manager.classroom;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import org.apache.shiro.SecurityUtils;
import org.springframework.stereotype.Component;
import top.hcode.hoj.common.exception.StatusForbiddenException;
import top.hcode.hoj.common.exception.StatusNotFoundException;
import top.hcode.hoj.mapper.classroom.ClassroomMapper;
import top.hcode.hoj.mapper.classroom.ClassroomStudentMapper;
import top.hcode.hoj.mapper.classroom.ClassroomTeacherMapper;
import top.hcode.hoj.pojo.entity.classroom.Classroom;
import top.hcode.hoj.pojo.entity.classroom.ClassroomStudent;
import top.hcode.hoj.pojo.entity.classroom.ClassroomTeacher;
import top.hcode.hoj.utils.ShiroUtils;

import javax.annotation.Resource;
import java.util.Objects;

@Component
public class ClassroomAccessManager {
    @Resource private ClassroomMapper classroomMapper;
    @Resource private ClassroomTeacherMapper teacherMapper;
    @Resource private ClassroomStudentMapper studentMapper;

    public String uid() {
        return ShiroUtils.getProfile().getUid();
    }

    public boolean isAdmin() {
        return SecurityUtils.getSubject().hasRole("root") || SecurityUtils.getSubject().hasRole("admin");
    }

    public Classroom requireClassroom(Long id) throws StatusNotFoundException {
        Classroom classroom = classroomMapper.selectOne(new QueryWrapper<Classroom>()
                .eq("id", id).eq("status", 1));
        if (classroom == null) throw new StatusNotFoundException("班级不存在");
        return classroom;
    }

    public boolean canManage(Classroom classroom, String uid) {
        if (isAdmin() || Objects.equals(classroom.getTeacherId(), uid)) return true;
        return teacherMapper.selectCount(new QueryWrapper<ClassroomTeacher>()
                .eq("classroom_id", classroom.getId()).eq("teacher_id", uid).eq("status", 1)) > 0;
    }

    public Classroom requireManager(Long id) throws Exception {
        Classroom classroom = requireClassroom(id);
        if (!canManage(classroom, uid())) throw new StatusForbiddenException("只有班级教师可以执行此操作");
        return classroom;
    }

    public boolean isStudent(Long classroomId, String uid) {
        return studentMapper.selectCount(new QueryWrapper<ClassroomStudent>()
                .eq("classroom_id", classroomId).eq("uid", uid).eq("status", 1)) > 0;
    }
}
