package top.hcode.hoj.pojo.entity.problem;

import com.baomidou.mybatisplus.annotation.*;
import lombok.Data;
import lombok.experimental.Accessors;

import java.io.Serializable;
import java.util.Date;

@Data
@Accessors(chain = true)
@TableName("problem_ai_record")
public class ProblemAiRecord implements Serializable {
    @TableId(value = "id", type = IdType.AUTO)
    private Long id;
    private Long pid;
    private String sessionId;
    private String uid;
    private String question;
    private String response;
    private String status;
    private Integer durationMs;
    private String errorMessage;
    @TableField(fill = FieldFill.INSERT)
    private Date gmtCreate;
    @TableField(fill = FieldFill.INSERT_UPDATE)
    private Date gmtModified;
}
