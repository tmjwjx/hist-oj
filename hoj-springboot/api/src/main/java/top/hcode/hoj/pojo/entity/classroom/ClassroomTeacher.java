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
@TableName("classroom_teacher")
public class ClassroomTeacher implements Serializable {
    @TableId(type = IdType.AUTO)
    private Long id;
    private Long classroomId;
    private String teacherId;
    private Integer status;
    private Date createTime;
    private Date updateTime;
}
