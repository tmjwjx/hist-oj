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
@TableName("classroom")
public class Classroom implements Serializable {
    @TableId(type = IdType.AUTO)
    private Long id;
    private String className;
    private String classBelong;
    private String classCode;
    private String teacherId;
    private Integer status;
    private Date createTime;
    private Date updateTime;
}
