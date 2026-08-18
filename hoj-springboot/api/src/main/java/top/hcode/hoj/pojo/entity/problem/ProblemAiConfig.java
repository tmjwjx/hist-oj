package top.hcode.hoj.pojo.entity.problem;

import com.baomidou.mybatisplus.annotation.*;
import lombok.Data;
import lombok.experimental.Accessors;

import java.io.Serializable;
import java.util.Date;

@Data
@Accessors(chain = true)
@TableName("problem_ai_config")
public class ProblemAiConfig implements Serializable {
    @TableId(value = "id", type = IdType.INPUT)
    private Long id;
    private Boolean enabled;
    private String apiUrl;
    private String apiKey;
    private String model;
    private Integer timeoutSeconds;
    private String systemPrompt;
    private String validationPrompt;
    @TableField(fill = FieldFill.INSERT)
    private Date gmtCreate;
    @TableField(fill = FieldFill.INSERT_UPDATE)
    private Date gmtModified;
}
