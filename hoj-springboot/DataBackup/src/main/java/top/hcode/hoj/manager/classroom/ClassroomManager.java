package top.hcode.hoj.manager.classroom;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import com.baomidou.mybatisplus.core.conditions.update.UpdateWrapper;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import top.hcode.hoj.common.exception.StatusFailException;
import top.hcode.hoj.mapper.classroom.ClassroomMapper;
import top.hcode.hoj.mapper.classroom.ClassroomStudentMapper;
import top.hcode.hoj.mapper.classroom.ClassroomTeacherMapper;
import top.hcode.hoj.pojo.entity.classroom.Classroom;
import top.hcode.hoj.pojo.entity.classroom.ClassroomStudent;
import top.hcode.hoj.pojo.entity.classroom.ClassroomTeacher;
import top.hcode.hoj.pojo.vo.classroom.ClassroomVO;

import javax.annotation.Resource;
import java.util.*;
import java.util.stream.Collectors;

@Component
public class ClassroomManager {
    @Resource private ClassroomMapper classroomMapper;
    @Resource private ClassroomTeacherMapper teacherMapper;
    @Resource private ClassroomStudentMapper studentMapper;
    @Resource private ClassroomAccessManager accessManager;
    @Resource private ClassroomViewBuilder viewBuilder;

    @Transactional(rollbackFor = Exception.class)
    public ClassroomVO create(String className, String classBelong) throws StatusFailException {
        if (blank(className) || blank(classBelong)) throw new StatusFailException("班级名称和所属信息不能为空");
        Classroom classroom = new Classroom().setClassName(className.trim()).setClassBelong(classBelong.trim())
                .setClassCode(uniqueCode()).setTeacherId(accessManager.uid()).setStatus(1)
                .setCreateTime(new Date()).setUpdateTime(new Date());
        classroomMapper.insert(classroom);
        teacherMapper.insert(new ClassroomTeacher().setClassroomId(classroom.getId())
                .setTeacherId(accessManager.uid()).setStatus(1).setCreateTime(new Date()).setUpdateTime(new Date()));
        return viewBuilder.build(classroom);
    }

    @Transactional(rollbackFor = Exception.class)
    public void delete(Long classroomId) throws Exception {
        accessManager.requireManager(classroomId);
        classroomMapper.update(null, new UpdateWrapper<Classroom>().eq("id", classroomId).set("status", 0));
        teacherMapper.update(null, new UpdateWrapper<ClassroomTeacher>()
                .eq("classroom_id", classroomId).set("status", 0).set("update_time", new Date()));
        studentMapper.update(null, new UpdateWrapper<ClassroomStudent>()
                .eq("classroom_id", classroomId).set("status", 0).set("update_time", new Date()));
    }

    public List<ClassroomVO> teacherClassrooms() {
        String uid = accessManager.uid();
        Set<Long> ids = teacherMapper.selectList(new QueryWrapper<ClassroomTeacher>()
                .eq("teacher_id", uid).eq("status", 1)).stream()
                .map(ClassroomTeacher::getClassroomId).collect(Collectors.toSet());
        classroomMapper.selectList(new QueryWrapper<Classroom>().eq("teacher_id", uid).eq("status", 1))
                .forEach(row -> ids.add(row.getId()));
        return findByIds(ids);
    }

    public List<ClassroomVO> studentClassrooms() {
        Set<Long> ids = studentMapper.selectList(new QueryWrapper<ClassroomStudent>()
                .eq("uid", accessManager.uid()).eq("status", 1)).stream()
                .map(ClassroomStudent::getClassroomId).collect(Collectors.toSet());
        return findByIds(ids);
    }

    public List<ClassroomVO> all() {
        return viewBuilder.buildAll(classroomMapper.selectList(new QueryWrapper<Classroom>()
                .eq("status", 1).orderByDesc("create_time")));
    }

    public ClassroomVO detail(Long classroomId) throws Exception {
        return viewBuilder.build(accessManager.requireClassroom(classroomId));
    }

    private List<ClassroomVO> findByIds(Set<Long> ids) {
        if (ids.isEmpty()) return Collections.emptyList();
        return viewBuilder.buildAll(classroomMapper.selectList(new QueryWrapper<Classroom>()
                .in("id", ids).eq("status", 1).orderByDesc("create_time")));
    }

    private String uniqueCode() {
        for (int i = 0; i < 10; i++) {
            String code = UUID.randomUUID().toString().replace("-", "").substring(0, 8);
            if (classroomMapper.selectCount(new QueryWrapper<Classroom>().eq("class_code", code)) == 0) return code;
        }
        return Long.toHexString(System.nanoTime()).substring(0, 8);
    }

    private boolean blank(String value) {
        return value == null || value.trim().isEmpty();
    }
}
