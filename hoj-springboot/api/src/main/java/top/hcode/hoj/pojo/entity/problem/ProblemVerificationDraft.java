package top.hcode.hoj.pojo.entity.problem;

import com.baomidou.mybatisplus.annotation.*;
import lombok.Data;
import lombok.experimental.Accessors;

import java.io.Serializable;
import java.util.Date;

@Data
@Accessors(chain = true)
@TableName("problem_verification_draft")
public class ProblemVerificationDraft implements Serializable {
    @TableId(value = "id", type = IdType.AUTO)
    private Long id;
    private Long pid;
    private String uid;
    private String language;
    private String code;
    private String caseVersion;
    @TableField(fill = FieldFill.INSERT)
    private Date gmtCreate;
    @TableField(fill = FieldFill.INSERT_UPDATE)
    private Date gmtModified;
}
