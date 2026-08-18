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
@TableName("student_question_order")
public class StudentQuestionOrder implements Serializable {
    @TableId(type = IdType.AUTO) private Long id;
    private Long homeworkId;
    private String uid;
    @TableField("order_mapping") private String questionOrder;
    private Date createTime;
    private Date updateTime;
}
