package top.hcode.hoj.pojo.vo.classroom;

import lombok.Data;
import lombok.experimental.Accessors;

import java.util.Date;

@Data
@Accessors(chain = true)
public class ClassroomTeacherVO {
    private Long id;
    private Long classroomId;
    private String teacherId;
    private Integer status;
    private Date createdAt;
    private Date updatedAt;
    private ClassroomUserVO teacher;
}
