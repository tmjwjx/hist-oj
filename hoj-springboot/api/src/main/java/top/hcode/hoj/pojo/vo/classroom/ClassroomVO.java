package top.hcode.hoj.pojo.vo.classroom;

import lombok.Data;
import lombok.experimental.Accessors;

import java.util.Date;
import java.util.List;

@Data
@Accessors(chain = true)
public class ClassroomVO {
    private Long id;
    private String className;
    private String classBelong;
    private String classCode;
    private String teacherId;
    private Integer status;
    private Date createdAt;
    private Date updatedAt;
    private ClassroomUserVO teacher;
    private List<ClassroomTeacherVO> teachers;
}
