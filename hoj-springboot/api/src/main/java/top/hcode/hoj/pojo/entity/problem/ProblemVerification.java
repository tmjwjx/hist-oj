package top.hcode.hoj.pojo.entity.problem;

import com.baomidou.mybatisplus.annotation.*;
import lombok.Data;
import lombok.experimental.Accessors;

import java.io.Serializable;
import java.util.Date;

@Data
@Accessors(chain = true)
@TableName("problem_verification")
public class ProblemVerification implements Serializable {

    @TableId(value = "id", type = IdType.AUTO)
    private Long id;

    private Long pid;

    private String caseVersion;

    private String judgeMode;

    private Integer syncStatus;

    private String syncMessage;

    private Integer verificationStatus;

    private Long submitId;

    private String verifiedUid;

    private Boolean sampleVerified;

    @TableField(fill = FieldFill.INSERT)
    private Date gmtCreate;

    @TableField(fill = FieldFill.INSERT_UPDATE)
    private Date gmtModified;
}
