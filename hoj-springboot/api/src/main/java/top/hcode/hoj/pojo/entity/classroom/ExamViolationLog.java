package top.hcode.hoj.pojo.entity.classroom;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import com.baomidou.mybatisplus.annotation.TableField;
import lombok.Data;
import lombok.experimental.Accessors;

import java.io.Serializable;
import java.util.Date;

@Data
@Accessors(chain = true)
@TableName("exam_violation_log")
public class ExamViolationLog implements Serializable {
    @TableId(type = IdType.AUTO) private Long id;
    private Long homeworkId;
    private String uid;
    private String violationType;
    @TableField("description") private String details;
    private Date createTime;
}
