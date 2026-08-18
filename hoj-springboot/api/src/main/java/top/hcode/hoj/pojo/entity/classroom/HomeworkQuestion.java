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
@TableName("homework_question")
public class HomeworkQuestion implements Serializable {
    @TableId(type = IdType.AUTO) private Long id;
    private Long homeworkId;
    private Long questionId;
    private String problemId;
    private Integer questionOrder;
    private Integer score;
    private Date createTime;
}
