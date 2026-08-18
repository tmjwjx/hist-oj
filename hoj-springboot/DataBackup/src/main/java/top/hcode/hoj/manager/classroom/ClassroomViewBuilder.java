package top.hcode.hoj.manager.classroom;

import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import org.springframework.stereotype.Component;
import top.hcode.hoj.dao.user.UserInfoEntityService;
import top.hcode.hoj.mapper.classroom.ClassroomTeacherMapper;
import top.hcode.hoj.pojo.entity.classroom.Classroom;
import top.hcode.hoj.pojo.entity.classroom.ClassroomTeacher;
import top.hcode.hoj.pojo.entity.user.UserInfo;
import top.hcode.hoj.pojo.vo.classroom.ClassroomTeacherVO;
import top.hcode.hoj.pojo.vo.classroom.ClassroomUserVO;
import top.hcode.hoj.pojo.vo.classroom.ClassroomVO;

import javax.annotation.Resource;
import java.util.*;
import java.util.stream.Collectors;

@Component
public class ClassroomViewBuilder {
    @Resource private ClassroomTeacherMapper teacherMapper;
    @Resource private UserInfoEntityService userInfoService;

    public ClassroomVO build(Classroom classroom) {
        return buildAll(Collections.singletonList(classroom)).get(0);
    }

    public List<ClassroomVO> buildAll(List<Classroom> classrooms) {
        if (classrooms.isEmpty()) return Collections.emptyList();
        List<Long> ids = classrooms.stream().map(Classroom::getId).collect(Collectors.toList());
        List<ClassroomTeacher> relations = teacherMapper.selectList(new QueryWrapper<ClassroomTeacher>()
                .in("classroom_id", ids).eq("status", 1).orderByAsc("create_time"));
        Set<String> userIds = classrooms.stream().map(Classroom::getTeacherId).filter(Objects::nonNull)
                .collect(Collectors.toSet());
        relations.stream().map(ClassroomTeacher::getTeacherId).forEach(userIds::add);
        Map<String, UserInfo> users = userIds.isEmpty() ? Collections.emptyMap()
                : userInfoService.listByIds(userIds).stream().collect(Collectors.toMap(UserInfo::getUuid, u -> u));
        Map<Long, List<ClassroomTeacher>> grouped = relations.stream()
                .collect(Collectors.groupingBy(ClassroomTeacher::getClassroomId));
        return classrooms.stream().map(c -> toVO(c, grouped.getOrDefault(c.getId(), Collections.emptyList()), users))
                .collect(Collectors.toList());
    }

    public ClassroomUserVO user(UserInfo user) {
        if (user == null) return null;
        return new ClassroomUserVO().setUid(user.getUuid()).setUuid(user.getUuid())
                .setUsername(user.getUsername()).setNickname(user.getNickname()).setRealname(user.getRealname())
                .setEmail(user.getEmail()).setAvatar(user.getAvatar()).setStatus(user.getStatus());
    }

    private ClassroomVO toVO(Classroom classroom, List<ClassroomTeacher> relations, Map<String, UserInfo> users) {
        List<ClassroomTeacherVO> teachers = relations.stream().map(row -> teacher(row, users.get(row.getTeacherId())))
                .collect(Collectors.toList());
        if (classroom.getTeacherId() != null && teachers.stream()
                .noneMatch(row -> classroom.getTeacherId().equals(row.getTeacherId()))) {
            ClassroomTeacher primary = new ClassroomTeacher().setId(classroom.getId() * 1000000L)
                    .setClassroomId(classroom.getId()).setTeacherId(classroom.getTeacherId()).setStatus(1);
            teachers.add(0, teacher(primary, users.get(classroom.getTeacherId())));
        }
        return new ClassroomVO().setId(classroom.getId()).setClassName(classroom.getClassName())
                .setClassBelong(classroom.getClassBelong()).setClassCode(classroom.getClassCode())
                .setTeacherId(classroom.getTeacherId()).setStatus(classroom.getStatus())
                .setCreatedAt(classroom.getCreateTime()).setUpdatedAt(classroom.getUpdateTime())
                .setTeacher(user(users.get(classroom.getTeacherId()))).setTeachers(teachers);
    }

    private ClassroomTeacherVO teacher(ClassroomTeacher row, UserInfo user) {
        return new ClassroomTeacherVO().setId(row.getId()).setClassroomId(row.getClassroomId())
                .setTeacherId(row.getTeacherId()).setStatus(row.getStatus()).setCreatedAt(row.getCreateTime())
                .setUpdatedAt(row.getUpdateTime()).setTeacher(user(user));
    }
}
