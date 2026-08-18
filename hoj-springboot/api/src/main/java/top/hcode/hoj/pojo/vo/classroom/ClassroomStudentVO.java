package top.hcode.hoj.pojo.vo.classroom;

import lombok.Data;
import lombok.experimental.Accessors;

import java.util.Date;

@Data
@Accessors(chain = true)
public class ClassroomStudentVO {
    private Long id;
    private Long classroomId;
    private String uid;
    private String realName;
    private String gender;
    private String studentClass;
    private String studentNo;
    private Integer status;
    private Date createdAt;
    private Date updatedAt;
    private ClassroomUserVO user;
}
