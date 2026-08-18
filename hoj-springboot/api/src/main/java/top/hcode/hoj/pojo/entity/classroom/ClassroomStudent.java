package top.hcode.hoj.pojo.entity.classroom;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import lombok.experimental.Accessors;

import java.io.Serializable;
import java.util.Date;

@Data
@Accessors(chain = true)
@TableName("classroom_student")
public class ClassroomStudent implements Serializable {
    @TableId(type = IdType.AUTO)
    private Long id;
    private Long classroomId;
    private String uid;
    private String realName;
    private String gender;
    private String studentClass;
    private String studentNo;
    private Integer status;
    private Date createTime;
    private Date updateTime;
}
